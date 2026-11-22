//go:build (windows && amd64) || (linux && amd64) || (darwin && arm64)

package kui

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// C ABI 第 16 节（3.61.8）：脚本引擎控制。
func TestScriptEngineControl(t *testing.T) {
	onMain(t, func(t *mainT) {
		loadTestRuntime(t)
		engine, err := NewScriptEngine()
		if err != nil {
			t.Fatal(err)
		}
		defer engine.Close()

		result, err := engine.Eval("return 40 + 2", "go-test")
		if err != nil || !result.Success || result.Value.Type != ScriptInt || result.Value.Int != 42 {
			t.Fatalf("Eval value: %+v %v", result, err)
		}
		result, _ = engine.Eval("error('"+strings.Repeat("x", 5000)+"')", "go-test")
		if result.Success || len(result.Error) < 5000 {
			t.Fatalf("long Lua errors must not be cut (len %d)", len(result.Error))
		}

		if err := engine.RegisterFunction("go_add", func(args []ScriptValue) (ScriptValue, error) {
			if len(args) != 2 {
				return ScriptValue{}, errors.New("go_add wants two arguments")
			}
			if args[0].Type == ScriptString {
				return ScriptStringValue(args[0].String + args[1].String), nil
			}
			return ScriptIntValue(args[0].Int + args[1].Int), nil
		}); err != nil {
			t.Fatal(err)
		}
		if result, _ = engine.Eval("return go_add(2, 3)", ""); result.Value.Int != 5 {
			t.Fatalf("go_add: %+v", result)
		}
		if result, _ = engine.Eval("return go_add('ab', 'cd')", ""); result.Value.String != "abcd" {
			t.Fatalf("go_add strings: %+v", result)
		}
		if result, _ = engine.Eval("go_add(1)", ""); result.Success || !strings.Contains(result.Error, "two arguments") {
			t.Fatalf("Go error should become a Lua error: %+v", result)
		}

		// Lua 函数交给 Go：参数里的函数在调用期间可用，Clone 之后可以留用。
		var kept *ScriptFunction
		if err := engine.RegisterFunction("go_apply", func(args []ScriptValue) (ScriptValue, error) {
			value, err := args[0].Function.Call(args[1])
			if err == nil && kept == nil {
				kept, err = args[0].Function.Clone()
			}
			return value, err
		}); err != nil {
			t.Fatal(err)
		}
		if result, _ = engine.Eval("return go_apply(function(x) return x * 10 end, 4)", ""); result.Value.Int != 40 {
			t.Fatalf("go_apply: %+v", result)
		}
		if kept == nil {
			t.Fatal("Clone failed")
		}
		if value, err := kept.Call(ScriptIntValue(9)); err != nil || value.Int != 90 {
			t.Fatalf("kept function: %+v %v", value, err)
		}

		// Go 函数作为值交给 Lua，以及 Lua 函数作为全局读出。
		made, err := engine.NewFunction(func(args []ScriptValue) (ScriptValue, error) {
			return ScriptIntValue(args[0].Int * 100), nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := engine.SetGlobal("go_made", ScriptFunctionValue(made)); err != nil {
			t.Fatal(err)
		}
		_ = made.Close()
		if result, _ = engine.Eval("return go_made(3)", ""); result.Value.Int != 300 {
			t.Fatalf("go_made: %+v", result)
		}
		if _, err := engine.Eval("function mul(a, b) return a * b end", ""); err != nil {
			t.Fatal(err)
		}
		if result, err = engine.CallFunction("mul", ScriptIntValue(6), ScriptIntValue(7)); err != nil || result.Value.Int != 42 {
			t.Fatalf("CallFunction: %+v %v", result, err)
		}
		if result, _ = engine.CallFunction("no_such_fn"); result.Success || !strings.Contains(result.Error, "no_such_fn") {
			t.Fatalf("missing function: %+v", result)
		}
		mul, err := engine.Global("mul")
		if err != nil || mul.Type != ScriptFunc {
			t.Fatalf("Global(mul): %+v %v", mul, err)
		}
		if value, err := mul.Function.Call(ScriptIntValue(3), ScriptIntValue(4)); err != nil || value.Int != 12 {
			t.Fatalf("mul handle: %+v %v", value, err)
		}
		_ = mul.Function.Close()

		if err := engine.SetGlobal("g_text", ScriptStringValue("text")); err != nil {
			t.Fatal(err)
		}
		if err := engine.SetGlobal("g_ud", ScriptUserDataValue(1234)); err != nil {
			t.Fatal(err)
		}
		if value, _ := engine.Global("g_text"); value.String != "text" {
			t.Fatalf("g_text: %+v", value)
		}
		if value, _ := engine.Global("g_ud"); value.Type != ScriptUserData || value.UserData != 1234 {
			t.Fatalf("g_ud: %+v", value)
		}
		if value, _ := engine.Global("missing"); value.Type != ScriptNil {
			t.Fatalf("missing global: %+v", value)
		}

		path := filepath.Join(t.TempDir(), "eval_file.lua")
		if err := os.WriteFile(path, []byte("return 'from file'"), 0o644); err != nil {
			t.Fatal(err)
		}
		if result, err = engine.EvalFile(path); err != nil || result.Value.String != "from file" {
			t.Fatalf("EvalFile: %+v %v", result, err)
		}

		printed := ""
		if err := engine.SetPrintHandler(func(message string) { printed = message }); err != nil {
			t.Fatal(err)
		}
		_, _ = engine.Eval("print('hi', 1, true)", "")
		if printed != "hi\t1\ttrue" {
			t.Fatalf("print: %q", printed)
		}
		_ = engine.SetPrintHandler(nil)
		_, _ = engine.Eval("print('quiet')", "")
		if printed != "hi\t1\ttrue" {
			t.Fatal("print handler should be gone")
		}

		usage, err := engine.ResourceUsage()
		if err != nil || usage.MemoryBytes == 0 || usage.MemoryLimit != 256*1024*1024 || usage.InstructionLimit != 100000000 {
			t.Fatalf("ResourceUsage: %+v %v", usage, err)
		}
		_ = engine.SetInstructionLimit(100000)
		if result, _ = engine.Eval("while true do end", ""); result.Success || !strings.Contains(result.Error, "instruction limit") {
			t.Fatalf("instruction limit: %+v", result)
		}
		_ = engine.SetMemoryLimit(usage.MemoryBytes + 1024*1024)
		if result, _ = engine.Eval("local s = string.rep('x', 8 * 1024 * 1024)", ""); result.Success || !strings.Contains(result.Error, "memory") {
			t.Fatalf("memory limit: %+v", result)
		}
		_ = engine.SetMemoryLimit(0)
		if usage, _ = engine.ResourceUsage(); usage.MemoryLimit != 0 || usage.InstructionLimit != 100000 {
			t.Fatalf("limits read back: %+v", usage)
		}

		if err := engine.Close(); err != nil {
			t.Fatal(err)
		}
		var kuiErr *Error
		if _, err := kept.Call(ScriptIntValue(1)); !errors.As(err, &kuiErr) || kuiErr.Status != ErrInvalidArg {
			t.Fatalf("a kept function reports its engine is gone: %v", err)
		}
		if err := kept.Close(); err != nil {
			t.Fatal(err)
		}
	})
}
