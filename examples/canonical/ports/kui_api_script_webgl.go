//go:build ignore

// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/kui_api_script_webgl.go; edit that file, not this copy.
// API tour: script engine control and a host-driven WebGL page (examples/api_tours/script_webgl).
// On start the host exchanges every kind of script value with the page's Lua, calls Lua and lets
// Lua call the host, exercises limits, print, EvalFile and document re-binding, then watches the
// WebGL scene advance, pause and resume. One row per check goes into #tour-log; #tour-result ends
// as "PASS n/n". The panel buttons repeat the calls by hand.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

// ---- Self-check bookkeeping: count, log one <tr> per check, publish the verdict ----------

type tour struct {
	doc         *kui.Document
	total       int
	failed      int
	failedNames []string
}

func (t *tour) row(name, result, cell string) {
	log, err := t.doc.GetElementByID("tour-log")
	if err != nil || log == nil {
		return
	}
	defer log.Close()
	row, err := t.doc.CreateElement("tr")
	if err != nil || row == nil {
		return
	}
	defer row.Close()
	_ = row.SetAttribute("data-check", name)
	_ = row.SetAttribute("data-result", result)
	if result != "ok" {
		_ = row.AddClass("bad")
	}
	for _, text := range []string{name, cell} {
		if td, err := t.doc.CreateElement("td"); err == nil && td != nil {
			_ = td.SetTextContent(text)
			_ = row.AppendChild(td)
			td.Close()
		}
	}
	_ = log.AppendChild(row)
}

func (t *tour) check(name string, ok bool, detail string) {
	t.total++
	if ok {
		t.row(name, "ok", "ok")
		return
	}
	t.failed++
	t.failedNames = append(t.failedNames, name)
	fmt.Printf("tour check failed: %s: %s\n", name, detail)
	t.row(name, "fail", "FAIL: "+detail)
}

func (t *tour) finish() {
	result, err := t.doc.GetElementByID("tour-result")
	if err != nil || result == nil {
		return
	}
	defer result.Close()
	text := fmt.Sprintf("PASS %d/%d", t.total, t.total)
	if t.failed > 0 {
		text = fmt.Sprintf("FAIL %d/%d: %s", t.failed, t.total, strings.Join(t.failedNames, ", "))
	}
	_ = result.SetTextContent(text)
	_ = result.SetAttribute("data-total", fmt.Sprint(t.total))
	_ = result.SetAttribute("data-failed", fmt.Sprint(t.failed))
	_ = result.SetAttribute("data-known", "0")
	if t.failed == 0 {
		_ = result.AddClass("pass")
	} else {
		_ = result.AddClass("fail")
	}
}

// ---- Small helpers ------------------------------------------------------------------------

var kindNames = map[kui.ScriptValueType]string{
	kui.ScriptNil: "nil", kui.ScriptBool: "bool", kui.ScriptInt: "int", kui.ScriptFloat: "float",
	kui.ScriptString: "string", kui.ScriptFunc: "function", kui.ScriptUserData: "userdata",
}

// number reads a value of either Lua number subtype.
func number(v kui.ScriptValue) float64 {
	if v.Type == kui.ScriptInt {
		return float64(v.Int)
	}
	return v.Float
}

// closeFunc releases a function handle a result or Global handed out.
func closeFunc(v kui.ScriptValue) {
	if v.Function != nil {
		_ = v.Function.Close()
	}
}

func setText(doc *kui.Document, id, text string) {
	if element, err := doc.GetElementByID(id); err == nil && element != nil {
		_ = element.SetTextContent(text)
		element.Close()
	}
}

func getText(doc *kui.Document, id string) string {
	element, err := doc.GetElementByID(id)
	if err != nil || element == nil {
		return ""
	}
	defer element.Close()
	text, _ := element.TextContent()
	return text
}

type host struct {
	doc     *kui.Document
	script  *kui.ScriptEngine
	t       *tour
	reports int
	prints  int
	double  *kui.ScriptFunction // host_double, also stored in the Lua global host_double
	counter *kui.ScriptFunction // a Lua closure the "Lua counter++" button keeps calling
	mark    int64
}

func (h *host) eval(code string) kui.ScriptResult {
	result, err := h.script.Eval(code, "tour")
	if err != nil {
		return kui.ScriptResult{Error: err.Error()}
	}
	return result
}

func (h *host) evalBool(code string) bool {
	result := h.eval(code)
	return result.Success && result.Value.Type == kui.ScriptBool && result.Value.Bool
}

func (h *host) call(name string, args ...kui.ScriptValue) kui.ScriptResult {
	result, err := h.script.CallFunction(name, args...)
	if err != nil {
		return kui.ScriptResult{Error: err.Error()}
	}
	return result
}

func (h *host) frames() int64 {
	result := h.call("getFrameCount")
	if !result.Success {
		return -1
	}
	return result.Value.Int
}

func (h *host) global(name string) kui.ScriptValue {
	value, _ := h.script.Global(name)
	return value
}

// ---- Host functions Lua can call (RegisterFunction / NewFunction) ------------------------

func (h *host) registerHostFunctions() {
	// host_report(frames, fps): Lua's render loop reports progress; the host shows it.
	_ = h.script.RegisterFunction("host_report", func(args []kui.ScriptValue) (kui.ScriptValue, error) {
		h.reports++
		var frames, fps float64
		if len(args) > 0 {
			frames = number(args[0])
		}
		if len(args) > 1 {
			fps = number(args[1])
		}
		setText(h.doc, "gl-report", fmt.Sprintf("frames=%.0f | fps=%.0f", frames, fps))
		return kui.ScriptNilValue(), nil
	})
	// host_kind(v): names the kind of value the host received.
	_ = h.script.RegisterFunction("host_kind", func(args []kui.ScriptValue) (kui.ScriptValue, error) {
		kind := kui.ScriptNil
		if len(args) > 0 {
			kind = args[0].Type
		}
		return kui.ScriptStringValue(kindNames[kind]), nil
	})
	// host_fail(): a host error becomes a Lua error with this message.
	_ = h.script.RegisterFunction("host_fail", func([]kui.ScriptValue) (kui.ScriptValue, error) {
		return kui.ScriptNilValue(), fmt.Errorf("host says no")
	})
}

// ---- Globals: every value kind through SetGlobal / Global ---------------------------------

func (h *host) checkGlobals() {
	t := h.t
	_ = h.script.SetGlobal("t_nil", kui.ScriptNilValue())
	value, missing := h.global("t_nil"), h.global("no_such_global")
	t.check("global.nil", value.Type == kui.ScriptNil && missing.Type == kui.ScriptNil,
		fmt.Sprintf("types %d / %d", value.Type, missing.Type))

	_ = h.script.SetGlobal("t_bool", kui.ScriptBoolValue(true))
	value = h.global("t_bool")
	t.check("global.bool", value.Type == kui.ScriptBool && value.Bool && h.evalBool("return t_bool == true"),
		fmt.Sprintf("value %+v", value))

	_ = h.script.SetGlobal("t_int", kui.ScriptIntValue(42))
	value = h.global("t_int")
	t.check("global.int", value.Type == kui.ScriptInt && value.Int == 42 &&
		h.evalBool("return math.type(t_int) == 'integer'"), fmt.Sprintf("value %+v", value))

	_ = h.script.SetGlobal("t_float", kui.ScriptFloatValue(2.5))
	value = h.global("t_float")
	t.check("global.float", value.Type == kui.ScriptFloat && value.Float == 2.5 &&
		h.evalBool("return math.type(t_float) == 'float'"), fmt.Sprintf("value %+v", value))

	_ = h.script.SetGlobal("t_str", kui.ScriptStringValue("héllo, KUI"))
	value = h.global("t_str")
	t.check("global.string", value.Type == kui.ScriptString && value.String == "héllo, KUI" &&
		h.evalBool("return #t_str == 11"), fmt.Sprintf("value %+v", value))

	// The page's own globals are what the scene reads every frame.
	value = h.global("gl_ready")
	t.check("webgl.page_ready", value.Type == kui.ScriptBool && value.Bool,
		fmt.Sprintf("gl_ready %+v, #gl-status '%s'", value, getText(h.doc, "gl-status")))
}

// ---- Calls: host -> Lua (CallFunction) and Lua -> host (RegisterFunction) ----------------

func (h *host) checkCalls() {
	t := h.t
	result := h.call("setSpeed", kui.ScriptFloatValue(1.5))
	speed := h.global("speed")
	t.check("call_function.setSpeed", result.Success && number(result.Value) == 1.5 && number(speed) == 1.5,
		fmt.Sprintf("result %+v, global speed %+v", result, speed))
	h.call("setSpeed", kui.ScriptFloatValue(1.0))

	first, second := h.call("togglePause"), h.call("togglePause")
	t.check("call_function.togglePause", first.Value.Type == kui.ScriptBool && first.Value.Bool &&
		second.Value.Type == kui.ScriptBool && !second.Value.Bool,
		fmt.Sprintf("first %+v second %+v", first.Value, second.Value))

	// ScriptResult.Error is the engine's last error (kui_script_engine_get_last_error).
	result = h.call("noSuchFunction")
	t.check("get_last_error.missing_function", !result.Success && strings.Contains(result.Error, "noSuchFunction"),
		fmt.Sprintf("result %+v", result))
	result = h.call("getFrameCount")
	t.check("get_last_error.cleared", result.Success && result.Value.Type == kui.ScriptInt && result.Error == "",
		fmt.Sprintf("result %+v", result))

	h.registerHostFunctions()
	result = h.eval("return hostKinds()")
	t.check("register_function.kinds", result.Success && result.Value.String == "nil,bool,int,float,string,function",
		fmt.Sprintf("result %+v", result))

	result = h.eval("local ok, e = pcall(host_fail) return (not ok) and tostring(e) or 'no error'")
	t.check("register_function.error", result.Success && strings.Contains(result.Value.String, "host says no"),
		fmt.Sprintf("result %+v", result))
}

// ---- Function values: create, call, retain (Clone), release (Close) ----------------------

func (h *host) checkFunctions() {
	t := h.t
	double, err := h.script.NewFunction(func(args []kui.ScriptValue) (kui.ScriptValue, error) {
		if len(args) < 1 {
			return kui.ScriptNilValue(), fmt.Errorf("host_double needs a number")
		}
		if args[0].Type == kui.ScriptInt {
			return kui.ScriptIntValue(args[0].Int * 2), nil
		}
		return kui.ScriptFloatValue(number(args[0]) * 2), nil
	})
	h.double = double
	var result kui.ScriptResult
	if err == nil {
		_ = h.script.SetGlobal("host_double", kui.ScriptFunctionValue(double))
		result = h.eval("return applyTwice(host_double, 5)")
	}
	t.check("function_create.as_lua_value", err == nil && result.Success && result.Value.Int == 20,
		fmt.Sprintf("err %v result %+v", err, result))

	var value kui.ScriptValue
	if double != nil {
		value, err = double.Call(kui.ScriptIntValue(21))
	}
	t.check("function_call.host", double != nil && err == nil && value.Type == kui.ScriptInt && value.Int == 42,
		fmt.Sprintf("err %v value %+v", err, value))

	// A second owned handle outlives the first; Lua's reference is independent of both.
	if double != nil {
		copied, cloneErr := double.Clone()
		var callErr, closeErr error = cloneErr, cloneErr
		value = kui.ScriptValue{}
		if cloneErr == nil {
			value, callErr = copied.Call(kui.ScriptIntValue(4))
			closeErr = copied.Close()
		}
		t.check("function_retain_release", cloneErr == nil && callErr == nil && value.Int == 8 &&
			closeErr == nil && h.evalBool("return host_double(3) == 6"),
			fmt.Sprintf("clone %v call %v value %+v close %v", cloneErr, callErr, value, closeErr))
	}

	// Global hands out an owned copy of a Lua function; calling it returns another one.
	maker := h.global("makeCounter")
	var got []int64
	if maker.Function != nil {
		counter, _ := maker.Function.Call()
		h.counter = counter.Function
		for i := 0; i < 3 && h.counter != nil; i++ {
			next, _ := h.counter.Call()
			got = append(got, next.Int)
		}
	}
	closeFunc(maker)
	t.check("get_global.function", h.counter != nil && fmt.Sprint(got) == "[1 2 3]",
		fmt.Sprintf("makeCounter %+v values %v", maker.Type, got))

	// A Lua error inside a Lua function yields nil (kui_c.h, as in native C++).
	fails := h.global("luaFails")
	value = kui.ScriptIntValue(-1)
	if fails.Function != nil {
		value, _ = fails.Function.Call()
	}
	closeFunc(fails)
	t.check("function_call.lua_error_is_nil", fails.Function != nil && value.Type == kui.ScriptNil,
		fmt.Sprintf("value %+v", value))
}

// ---- Timers: the engine runs due timers only when the host ticks it -----------------------------

func (h *host) checkTimers() {
	// A timer is due delay after the engine's clock when the script scheduled it; ticking the
	// current time first makes the test independent of any earlier tick.
	now, _ := kui.Now()
	_ = h.script.TickTimers(now)
	scheduled := h.eval("tour_timer_fired = false; setTimeout(function() tour_timer_fired = true end, 50)")
	_ = h.script.TickTimers(now + 10)
	before := h.evalBool("return tour_timer_fired == false")
	tickErr := h.script.TickTimers(now + 100)
	after := h.evalBool("return tour_timer_fired == true")
	h.t.check("timers.tick_timers", scheduled.Success && tickErr == nil && before && after,
		fmt.Sprintf("scheduled %v, not fired early %v, fired after the tick %v, err %v", scheduled.Success, before, after, tickErr))
}

// ---- Limits and resource usage -----------------------------------------------------------

func (h *host) checkLimits() {
	t := h.t
	usage, _ := h.script.ResourceUsage()
	t.check("get_resource_usage.defaults", usage.MemoryBytes > 0 && usage.MemoryLimit == 256*1024*1024 &&
		usage.InstructionLimit == 100000000, fmt.Sprintf("usage %+v", usage))

	_ = h.script.SetInstructionLimit(10000)
	now, _ := h.script.ResourceUsage()
	result := h.eval("while true do end")
	t.check("set_instruction_limit.stops_loop", now.InstructionLimit == 10000 && !result.Success &&
		strings.Contains(result.Error, "instruction limit"), fmt.Sprintf("limit %d result %+v", now.InstructionLimit, result))
	_ = h.script.SetInstructionLimit(usage.InstructionLimit)
	now, _ = h.script.ResourceUsage()
	result = h.eval("local n = 0 for i = 1, 100000 do n = n + i end return n")
	t.check("set_instruction_limit.restore", now.InstructionLimit == usage.InstructionLimit && result.Success &&
		result.Value.Int == 5000050000, fmt.Sprintf("limit %d result %+v", now.InstructionLimit, result))

	_ = h.script.SetMemoryLimit(usage.MemoryBytes + 1024*1024)
	now, _ = h.script.ResourceUsage()
	result = h.eval("local s = string.rep('x', 16 * 1024 * 1024) return #s")
	t.check("set_memory_limit.refuses", now.MemoryLimit == usage.MemoryBytes+1024*1024 && !result.Success &&
		strings.Contains(result.Error, "memory"), fmt.Sprintf("limit %d result %+v", now.MemoryLimit, result))
	_ = h.script.SetMemoryLimit(usage.MemoryLimit)
	now, _ = h.script.ResourceUsage()
	result = h.eval("local s = string.rep('x', 16 * 1024 * 1024) local n = #s s = nil collectgarbage() return n")
	t.check("set_memory_limit.restore", now.MemoryLimit == usage.MemoryLimit && result.Success &&
		result.Value.Int == 16*1024*1024, fmt.Sprintf("limit %d result %+v", now.MemoryLimit, result))
}

// ---- print, EvalFile, document binding ---------------------------------------------------

// writeLuaFile puts a small Lua chunk into the temp directory.
func writeLuaFile() (string, error) {
	path := filepath.Join(os.TempDir(), "kui_api_script_webgl_go.lua")
	return path, os.WriteFile(path, []byte("from_file = 'loaded'\nreturn 6 * 7\n"), 0o644)
}

func (h *host) checkPrintFileBinding() {
	t := h.t
	captured := ""
	_ = h.script.SetPrintHandler(func(message string) { captured = message })
	h.eval("print('hello', 42)")
	t.check("set_print_handler.captures", captured == "hello\t42", fmt.Sprintf("captured %q", captured))
	// The handler kept after the tour shows Lua's print() output in #print-out.
	_ = h.script.SetPrintHandler(func(message string) {
		h.prints++
		setText(h.doc, "print-out", fmt.Sprintf("#%d %s", h.prints, message))
	})

	path, err := writeLuaFile()
	var result kui.ScriptResult
	if err == nil {
		result, err = h.script.EvalFile(path)
		_ = os.Remove(path)
	}
	loaded := h.global("from_file")
	t.check("eval_file.runs", err == nil && result.Success && result.Value.Type == kui.ScriptInt &&
		result.Value.Int == 42 && loaded.String == "loaded", fmt.Sprintf("err %v result %+v from_file %+v", err, result, loaded))
	result, err = h.script.EvalFile("no/such/file.lua")
	t.check("eval_file.missing", err == nil && !result.Success && result.Error != "", fmt.Sprintf("err %v result %+v", err, result))

	_ = h.script.UnbindDocument()
	t.check("unbind_document", h.evalBool("return document == nil"), "document still bound")
	_ = h.script.BindDocument(h.doc)
	t.check("bind_document.again", h.evalBool("return document:getElementById('gl-canvas') ~= nil"),
		"document not usable after re-binding")
}

// ---- Asynchronous part: the WebGL scene advances, pauses and resumes -------------------

func (h *host) phaseRunning() {
	t := h.t
	now := h.frames()
	t.check("webgl.frames_advance", now > h.mark, fmt.Sprintf("frames %d -> %d", h.mark, now))
	report := getText(h.doc, "gl-report")
	t.check("register_function.host_report", h.reports > 0 && strings.HasPrefix(report, "frames="),
		fmt.Sprintf("host_report calls %d, #gl-report '%s'", h.reports, report))
	h.call("togglePause")
	paused := h.global("paused")
	t.check("webgl.paused_global", paused.Type == kui.ScriptBool && paused.Bool, fmt.Sprintf("paused %+v", paused))
	// Let a frame already in flight land before taking the paused mark.
	_, _ = kui.SetTimeout(func() {
		h.mark = h.frames()
		_, _ = kui.SetTimeout(h.phasePaused, 300)
	}, 100)
}

func (h *host) phasePaused() {
	now := h.frames()
	h.t.check("webgl.pause_stops", now == h.mark, fmt.Sprintf("frames %d -> %d while paused", h.mark, now))
	h.call("togglePause")
	_, _ = kui.SetTimeout(func() {
		now := h.frames()
		h.t.check("webgl.resume", now > h.mark, fmt.Sprintf("frames %d after resume, %d while paused", now, h.mark))
		h.t.finish()
	}, 400)
}

// ---- Panel buttons -----------------------------------------------------------------------

func (h *host) bindButtons(ctx *runner.Context) error {
	doc := h.doc
	buttons := map[string]func(){
		"btn-gl-pause": func() {
			result := h.call("togglePause")
			state := "running"
			if result.Value.Bool {
				state = "paused"
			}
			setText(doc, "gl-state", state)
		},
		"btn-gl-speed": func() {
			result := h.call("setSpeed", kui.ScriptFloatValue(3.0))
			setText(doc, "gl-out", fmt.Sprintf("speed=%g", number(result.Value)))
		},
		"btn-gl-tint": func() {
			tints := []string{"sky", "rose", "lime"}
			current := h.global("tint").String
			index := 0
			for index < len(tints) && tints[index] != current {
				index++
			}
			next := tints[(index+1)%len(tints)]
			_ = h.script.SetGlobal("tint", kui.ScriptStringValue(next))
			setText(doc, "gl-out", "tint="+next)
		},
		"btn-gl-scale": func() {
			next := 0.5
			if number(h.global("scale")) < 0.65 {
				next = 0.8
			}
			_ = h.script.SetGlobal("scale", kui.ScriptFloatValue(next))
			setText(doc, "gl-out", fmt.Sprintf("scale=%g", next))
		},
		"btn-gl-frames": func() { setText(doc, "gl-out", fmt.Sprintf("frames=%d", h.frames())) },
		"btn-val-roundtrip": func() {
			var kinds []string
			for _, name := range []string{"t_nil", "t_bool", "t_int", "t_float", "t_str", "makeCounter"} {
				value := h.global(name)
				kinds = append(kinds, kindNames[value.Type])
				closeFunc(value)
			}
			setText(doc, "val-out", strings.Join(kinds, ","))
		},
		"btn-val-kinds": func() {
			setText(doc, "val-out", "host saw: "+h.eval("return hostKinds()").Value.String)
		},
		"btn-val-double": func() {
			var value kui.ScriptValue
			if h.double != nil {
				value, _ = h.double.Call(kui.ScriptIntValue(21))
			}
			setText(doc, "val-out", fmt.Sprintf("host_double(21) = %d", value.Int))
		},
		"btn-val-counter": func() {
			var value kui.ScriptValue
			if h.counter != nil {
				value, _ = h.counter.Call()
			}
			setText(doc, "val-out", fmt.Sprintf("counter=%d", value.Int))
		},
		"btn-lim-loop": func() {
			saved, _ := h.script.ResourceUsage()
			_ = h.script.SetInstructionLimit(10000)
			result := h.eval("while true do end")
			_ = h.script.SetInstructionLimit(saved.InstructionLimit)
			setText(doc, "lim-out", "stopped: "+result.Error)
		},
		"btn-lim-mem": func() {
			saved, _ := h.script.ResourceUsage()
			_ = h.script.SetMemoryLimit(saved.MemoryBytes + 1024*1024)
			result := h.eval("local s = string.rep('x', 16 * 1024 * 1024) return #s")
			_ = h.script.SetMemoryLimit(saved.MemoryLimit)
			setText(doc, "lim-out", "refused: "+result.Error)
		},
		"btn-lim-usage": func() {
			usage, _ := h.script.ResourceUsage()
			setText(doc, "lim-out", fmt.Sprintf("memory=%d KiB | limit=%d MiB | instructions=%d",
				usage.MemoryBytes/1024, usage.MemoryLimit/(1024*1024), usage.InstructionLimit))
		},
		"btn-lim-error": func() { setText(doc, "lim-out", "last error: "+h.call("noSuchFunction").Error) },
		"btn-io-print":  func() { h.eval("print('hello from Lua', frames > 0)") },
		"btn-io-file": func() {
			value := int64(-1)
			if path, err := writeLuaFile(); err == nil {
				if result, err := h.script.EvalFile(path); err == nil && result.Success {
					value = result.Value.Int
				}
				_ = os.Remove(path)
			}
			setText(doc, "io-out", fmt.Sprintf("eval_file -> %d", value))
		},
		"btn-io-rebind": func() {
			_ = h.script.UnbindDocument()
			unbound := h.evalBool("return document == nil")
			_ = h.script.BindDocument(doc)
			rebound := h.evalBool("return document ~= nil")
			state := func(set bool) string {
				if set {
					return "set"
				}
				return "nil"
			}
			setText(doc, "io-out", fmt.Sprintf("unbound: document=%s | rebound: document=%s",
				state(!unbound), state(rebound)))
		},
	}
	for id, fn := range buttons {
		element, err := doc.GetElementByID(id)
		if err != nil || element == nil {
			return fmt.Errorf("没有 #%s: %v", id, err)
		}
		ctx.Keep(element)
		handler := fn
		if _, err := element.AddEventListener("click", func(*kui.Event) { handler() }); err != nil {
			return err
		}
	}
	return nil
}

func ready(ctx *runner.Context) error {
	h := &host{doc: ctx.Document, script: ctx.Script, t: &tour{doc: ctx.Document}}
	h.checkGlobals()
	h.checkCalls()
	h.checkFunctions()
	h.checkLimits()
	h.checkTimers()
	h.checkPrintFileBinding()
	// The rest needs frames: the scene must advance, stop while paused and resume.
	h.mark = h.frames()
	if _, err := kui.SetTimeout(h.phaseRunning, 600); err != nil {
		return err
	}
	ctx.OnClose(func() {
		if h.double != nil {
			_ = h.double.Close()
		}
		if h.counter != nil {
			_ = h.counter.Close()
		}
	})
	return h.bindButtons(ctx)
}
