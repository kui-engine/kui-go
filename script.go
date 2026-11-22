//go:build (windows && amd64) || (linux && amd64) || (darwin && arm64)

package kui

/*
#cgo CFLAGS: -std=c11 -I${SRCDIR}
#include <stdlib.h>
#include "loader.h"
#include "callbacks.h"
*/
import "C"

import (
	"runtime/cgo"
	"unsafe"
)

// 3.61.8 起的脚本引擎控制（C ABI 第 16 节）：带返回值的执行、调用 Lua 函数、把 Go 函数交给 Lua、
// 全局变量、print 重定向、内存与指令上限。全部只能在 UI 线程调用，回调也在 UI 线程执行。

// ScriptValueType 是脚本值的种类（与原生 ScriptValue 一致；Lua 表没有对应形式，读出来是 nil）。
type ScriptValueType int32

const (
	ScriptNil ScriptValueType = iota
	ScriptBool
	ScriptInt
	ScriptFloat
	ScriptString
	ScriptFunc
	ScriptUserData
)

// ScriptValue 是 Go 与 Lua 之间传递的值，按 Type 读对应字段。
// UserData 是 Lua 的 light userdata（一个不透明整数，例如 cgo.Handle 的值）。
type ScriptValue struct {
	Type     ScriptValueType
	Bool     bool
	Int      int64
	Float    float64
	String   string
	Function *ScriptFunction
	UserData uintptr
}

func ScriptNilValue() ScriptValue            { return ScriptValue{} }
func ScriptBoolValue(v bool) ScriptValue     { return ScriptValue{Type: ScriptBool, Bool: v} }
func ScriptIntValue(v int64) ScriptValue     { return ScriptValue{Type: ScriptInt, Int: v} }
func ScriptFloatValue(v float64) ScriptValue { return ScriptValue{Type: ScriptFloat, Float: v} }
func ScriptStringValue(v string) ScriptValue { return ScriptValue{Type: ScriptString, String: v} }
func ScriptUserDataValue(v uintptr) ScriptValue {
	return ScriptValue{Type: ScriptUserData, UserData: v}
}
func ScriptFunctionValue(f *ScriptFunction) ScriptValue {
	return ScriptValue{Type: ScriptFunc, Function: f}
}

// ScriptResult 是一次执行的结果：Success 为 false 时 Error 是 Lua 错误文本，Value 是第一个返回值。
type ScriptResult struct {
	Success bool
	Error   string
	Value   ScriptValue
}

// ScriptFunction 是一个可调用的脚本函数（Lua 函数，或交给 Lua 的 Go 函数）。
// 从结果、Global 拿到的要自己 Close；作为 Go 函数的参数收到的只在那次调用期间有效，
// 要留用就 Clone。引擎关闭后调用返回 ErrInvalidArg。
type ScriptFunction struct {
	handle unsafe.Pointer
	engine *ScriptEngine
}

func (f *ScriptFunction) raw() *C.KuiScriptFunction {
	if f == nil {
		return nil
	}
	return (*C.KuiScriptFunction)(f.handle)
}

// Call 调用函数。Go 函数返回错误时这里返回错误；Lua 函数内部出错得到 nil（与原生一致）。
func (f *ScriptFunction) Call(args ...ScriptValue) (ScriptValue, error) {
	if f == nil || f.handle == nil {
		return ScriptValue{}, &Error{Status: ErrInvalidArg, Context: "ScriptFunction.Call", Message: "closed function"}
	}
	values, release, err := scriptArgs(args, "ScriptFunction.Call")
	if err != nil {
		return ScriptValue{}, err
	}
	defer release()
	var success C.int
	var out C.KuiCScriptValue
	if err := statusError(C.kui_script_function_call(f.raw(), firstValue(values), C.size_t(len(values)), &success, &out), "ScriptFunction.Call"); err != nil {
		return ScriptValue{}, err
	}
	value, err := scriptValueFromC(f.engine, &out)
	if err != nil {
		return ScriptValue{}, err
	}
	if success == 0 {
		return value, &Error{Status: ErrInternal, Context: "ScriptFunction.Call", Message: f.engine.lastError()}
	}
	return value, nil
}

// Clone 得到一个独立的句柄（要自己 Close）。
func (f *ScriptFunction) Clone() (*ScriptFunction, error) {
	var owned *C.KuiScriptFunction
	if err := statusError(C.kui_script_function_retain(f.raw(), &owned), "ScriptFunction.Clone"); err != nil {
		return nil, err
	}
	return &ScriptFunction{handle: unsafe.Pointer(owned), engine: f.engine}, nil
}

// Close 释放句柄；Lua 里仍在用的函数不受影响。
func (f *ScriptFunction) Close() error {
	if f == nil || f.handle == nil {
		return nil
	}
	err := statusError(C.kui_script_function_release(f.raw()), "ScriptFunction.Close")
	f.handle = nil
	return err
}

func firstValue(values []C.KuiCScriptValue) *C.KuiCScriptValue {
	if len(values) == 0 {
		return nil
	}
	return &values[0]
}

// scriptArgs 把 Go 值转成 C 值；字符串放在 C 内存里，release 释放。
func scriptArgs(args []ScriptValue, context string) ([]C.KuiCScriptValue, func(), error) {
	values := make([]C.KuiCScriptValue, len(args))
	var strings []*C.char
	release := func() {
		for _, s := range strings {
			C.free(unsafe.Pointer(s))
		}
	}
	for index, arg := range args {
		value := &values[index]
		value._type = C.KuiCScriptValueType(arg.Type)
		switch arg.Type {
		case ScriptNil:
		case ScriptBool:
			value.bool_value = boolInt(arg.Bool)
		case ScriptInt:
			value.int_value = C.int64_t(arg.Int)
		case ScriptFloat:
			value.float_value = C.double(arg.Float)
		case ScriptString:
			raw := C.CString(arg.String)
			strings = append(strings, raw)
			value.string_value = raw
			value.string_len = C.size_t(len(arg.String))
		case ScriptFunc:
			if arg.Function == nil || arg.Function.handle == nil {
				release()
				return nil, nil, &Error{Status: ErrInvalidArg, Context: context, Message: "function value without a function"}
			}
			value.function_value = arg.Function.raw()
		case ScriptUserData:
			value.userdata = C.kui_go_handle_pointer(C.uintptr_t(arg.UserData))
		default:
			release()
			return nil, nil, &Error{Status: ErrInvalidArg, Context: context, Message: "unknown script value type"}
		}
	}
	return values, release, nil
}

// scriptValueFromC 复制一个 C 值；函数会被 retain 成一个要 Close 的句柄。
func scriptValueFromC(engine *ScriptEngine, value *C.KuiCScriptValue) (ScriptValue, error) {
	switch ScriptValueType(value._type) {
	case ScriptBool:
		return ScriptBoolValue(value.bool_value != 0), nil
	case ScriptInt:
		return ScriptIntValue(int64(value.int_value)), nil
	case ScriptFloat:
		return ScriptFloatValue(float64(value.float_value)), nil
	case ScriptString:
		return ScriptStringValue(goStringN(value.string_value, value.string_len)), nil
	case ScriptFunc:
		var owned *C.KuiScriptFunction
		if err := statusError(C.kui_script_function_retain(value.function_value, &owned), "ScriptValue"); err != nil {
			return ScriptValue{}, err
		}
		return ScriptFunctionValue(&ScriptFunction{handle: unsafe.Pointer(owned), engine: engine}), nil
	case ScriptUserData:
		return ScriptUserDataValue(uintptr(value.userdata)), nil
	default:
		return ScriptValue{}, nil
	}
}

// ---------------------------------------------------------------------------
// Go 函数交给 Lua
// ---------------------------------------------------------------------------

type scriptFunctionState struct {
	fn     func(args []ScriptValue) (ScriptValue, error)
	engine *ScriptEngine
	// 上一次返回的字符串（C 内存）：KUI 在回调返回后立刻复制，下一次调用或释放时 free。
	result *C.char
}

func (s *scriptFunctionState) setResult(text string) *C.char {
	if s.result != nil {
		C.free(unsafe.Pointer(s.result))
	}
	s.result = C.CString(text)
	return s.result
}

//export goKuiScriptFunctionCallback
func goKuiScriptFunctionCallback(args *C.KuiCScriptValue, count C.size_t, out *C.KuiCScriptValue, handle C.uintptr_t) (status C.KuiStatus) {
	state, ok := cgo.Handle(handle).Value().(*scriptFunctionState)
	if !ok {
		return C.KUI_ERR_INTERNAL
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			out._type = C.KUI_C_SCRIPT_STRING
			out.string_value = state.setResult("Go function panicked")
			out.string_len = C.size_t(len("Go function panicked"))
			status = C.KUI_ERR_INTERNAL
		}
	}()
	values := make([]ScriptValue, 0, int(count))
	var borrowed []*ScriptFunction
	defer func() {
		for _, function := range borrowed {
			_ = function.Close()
		}
	}()
	if count > 0 {
		for _, raw := range unsafe.Slice(args, int(count)) {
			value, err := scriptValueFromC(state.engine, &raw)
			if err != nil {
				return C.KUI_ERR_INTERNAL
			}
			if value.Function != nil {
				borrowed = append(borrowed, value.Function)
			}
			values = append(values, value)
		}
	}
	result, err := state.fn(values)
	if err != nil {
		message := err.Error()
		out._type = C.KUI_C_SCRIPT_STRING
		out.string_value = state.setResult(message)
		out.string_len = C.size_t(len(message))
		return C.KUI_ERR_INTERNAL
	}
	out._type = C.KuiCScriptValueType(result.Type)
	switch result.Type {
	case ScriptNil:
	case ScriptBool:
		out.bool_value = boolInt(result.Bool)
	case ScriptInt:
		out.int_value = C.int64_t(result.Int)
	case ScriptFloat:
		out.float_value = C.double(result.Float)
	case ScriptString:
		out.string_value = state.setResult(result.String)
		out.string_len = C.size_t(len(result.String))
	case ScriptFunc:
		if result.Function == nil || result.Function.handle == nil {
			return C.KUI_ERR_INVALID_ARG
		}
		out.function_value = result.Function.raw()
	case ScriptUserData:
		out.userdata = C.kui_go_handle_pointer(C.uintptr_t(result.UserData))
	default:
		return C.KUI_ERR_INVALID_ARG
	}
	return C.KUI_OK
}

//export goKuiScriptFunctionRelease
func goKuiScriptFunctionRelease(handle C.uintptr_t) {
	defer containCallbackPanic()
	h := cgo.Handle(handle)
	if state, ok := h.Value().(*scriptFunctionState); ok && state.result != nil {
		C.free(unsafe.Pointer(state.result))
		state.result = nil
	}
	h.Delete()
}

// RegisterFunction 把 Go 函数设为 Lua 全局函数。参数里的函数只在这次调用期间有效（要留用就 Clone）；
// 返回错误会在 Lua 里抛出同样文字的错误。同一个名字再登记会替换旧的。
func (s *ScriptEngine) RegisterFunction(name string, fn func(args []ScriptValue) (ScriptValue, error)) error {
	if fn == nil {
		return &Error{Status: ErrInvalidArg, Context: "ScriptEngine.RegisterFunction", Message: "nil function"}
	}
	rawName, size, release := inputString(name)
	defer release()
	handle, user := callbackHandle(&scriptFunctionState{fn: fn, engine: s})
	status := C.kui_script_engine_register_function(s.raw(), rawName, size, C.kui_go_script_function_trampoline(), user, C.kui_go_script_function_release_trampoline())
	if err := statusError(status, "ScriptEngine.RegisterFunction"); err != nil {
		handle.Delete()
		return err
	}
	return nil
}

// NewFunction 把 Go 函数包成一个可以作为值交给 Lua 的函数（要自己 Close；Lua 里仍在用时不受影响）。
func (s *ScriptEngine) NewFunction(fn func(args []ScriptValue) (ScriptValue, error)) (*ScriptFunction, error) {
	if fn == nil {
		return nil, &Error{Status: ErrInvalidArg, Context: "ScriptEngine.NewFunction", Message: "nil function"}
	}
	handle, user := callbackHandle(&scriptFunctionState{fn: fn, engine: s})
	var raw *C.KuiScriptFunction
	status := C.kui_script_function_create(s.raw(), C.kui_go_script_function_trampoline(), user, C.kui_go_script_function_release_trampoline(), &raw)
	if err := statusError(status, "ScriptEngine.NewFunction"); err != nil {
		handle.Delete()
		return nil, err
	}
	return &ScriptFunction{handle: unsafe.Pointer(raw), engine: s}, nil
}

// ---------------------------------------------------------------------------
// 执行、全局变量与上限
// ---------------------------------------------------------------------------

func (s *ScriptEngine) lastError() string {
	var size C.size_t
	if C.kui_script_engine_get_last_error(s.raw(), nil, 0, &size) != C.KUI_OK || size == 0 {
		return ""
	}
	buffer := make([]byte, int(size)+1)
	if C.kui_script_engine_get_last_error(s.raw(), (*C.char)(unsafe.Pointer(&buffer[0])), C.size_t(len(buffer)), &size) != C.KUI_OK {
		return ""
	}
	return string(buffer[:int(size)])
}

func (s *ScriptEngine) finishRun(success C.int, out *C.KuiCScriptValue) (ScriptResult, error) {
	value, err := scriptValueFromC(s, out)
	if err != nil {
		return ScriptResult{}, err
	}
	result := ScriptResult{Success: success != 0, Value: value}
	if !result.Success {
		result.Error = s.lastError()
	}
	return result, nil
}

// EvalFile 执行一个 Lua 文件。
func (s *ScriptEngine) EvalFile(path string) (ScriptResult, error) {
	rawPath, size, release := inputString(path)
	defer release()
	var success C.int
	var out C.KuiCScriptValue
	if err := statusError(C.kui_script_engine_eval_file(s.raw(), rawPath, size, &success, &out), "ScriptEngine.EvalFile"); err != nil {
		return ScriptResult{}, err
	}
	return s.finishRun(success, &out)
}

// CallFunction 调用 Lua 全局函数；函数不存在或出错时 Success 为 false、Error 是错误文本。
func (s *ScriptEngine) CallFunction(name string, args ...ScriptValue) (ScriptResult, error) {
	rawName, size, releaseName := inputString(name)
	defer releaseName()
	values, release, err := scriptArgs(args, "ScriptEngine.CallFunction")
	if err != nil {
		return ScriptResult{}, err
	}
	defer release()
	var success C.int
	var out C.KuiCScriptValue
	if err := statusError(C.kui_script_engine_call_function(s.raw(), rawName, size, firstValue(values), C.size_t(len(values)), &success, &out), "ScriptEngine.CallFunction"); err != nil {
		return ScriptResult{}, err
	}
	return s.finishRun(success, &out)
}

func (s *ScriptEngine) SetGlobal(name string, value ScriptValue) error {
	rawName, size, releaseName := inputString(name)
	defer releaseName()
	values, release, err := scriptArgs([]ScriptValue{value}, "ScriptEngine.SetGlobal")
	if err != nil {
		return err
	}
	defer release()
	return statusError(C.kui_script_engine_set_global(s.raw(), rawName, size, &values[0]), "ScriptEngine.SetGlobal")
}

// Global 读 Lua 全局变量；不存在是 nil。拿到的函数要自己 Close。
func (s *ScriptEngine) Global(name string) (ScriptValue, error) {
	rawName, size, release := inputString(name)
	defer release()
	var out C.KuiCScriptValue
	if err := statusError(C.kui_script_engine_get_global(s.raw(), rawName, size, &out), "ScriptEngine.Global"); err != nil {
		return ScriptValue{}, err
	}
	return scriptValueFromC(s, &out)
}

type scriptPrintState struct{ fn func(string) }

//export goKuiScriptPrintCallback
func goKuiScriptPrintCallback(message *C.char, size C.size_t, handle C.uintptr_t) {
	defer containCallbackPanic()
	if state, ok := cgo.Handle(handle).Value().(*scriptPrintState); ok {
		state.fn(goStringN(message, size))
	}
}

// SetPrintHandler 接收 Lua print() 的输出（参数以制表符分隔）；nil 恢复默认（只进 DevTools 控制台）。
func (s *ScriptEngine) SetPrintHandler(fn func(message string)) error {
	if fn == nil {
		return statusError(C.kui_script_engine_set_print_handler(s.raw(), nil, nil, nil), "ScriptEngine.SetPrintHandler")
	}
	handle, user := callbackHandle(&scriptPrintState{fn: fn})
	status := C.kui_script_engine_set_print_handler(s.raw(), C.kui_go_script_print_trampoline(), user, C.kui_go_release_trampoline())
	if err := statusError(status, "ScriptEngine.SetPrintHandler"); err != nil {
		handle.Delete()
		return err
	}
	return nil
}

// SetMemoryLimit 设内存上限（字节，0 表示不限；默认 256 MiB）。
func (s *ScriptEngine) SetMemoryLimit(bytes uint64) error {
	return statusError(C.kui_script_engine_set_memory_limit(s.raw(), C.size_t(bytes)), "ScriptEngine.SetMemoryLimit")
}

// SetInstructionLimit 设每次执行的指令上限（0 表示不限；默认 1 亿），从下一次执行起生效。
func (s *ScriptEngine) SetInstructionLimit(max uint64) error {
	return statusError(C.kui_script_engine_set_instruction_limit(s.raw(), C.uint64_t(max)), "ScriptEngine.SetInstructionLimit")
}

// ScriptResources 是脚本引擎的内存用量与上限（上限 0 表示不限）。
type ScriptResources struct {
	MemoryBytes      uint64
	MemoryLimit      uint64
	InstructionLimit uint64
}

func (s *ScriptEngine) ResourceUsage() (ScriptResources, error) {
	var memory, limit C.size_t
	var instructions C.uint64_t
	err := statusError(C.kui_script_engine_get_resource_usage(s.raw(), &memory, &limit, &instructions), "ScriptEngine.ResourceUsage")
	return ScriptResources{MemoryBytes: uint64(memory), MemoryLimit: uint64(limit), InstructionLimit: uint64(instructions)}, err
}
