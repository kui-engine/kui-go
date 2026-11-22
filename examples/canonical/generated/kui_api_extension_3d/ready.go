// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/kui_api_extension_3d.go; edit that file, not this copy.
// API tour: dynamic extensions and a host-driven 3D scene (examples/api_tours/extension_3d).
// The host builds a scene in <kui-3d-view> through Document.InvokeExtension (meshes, lights,
// materials, a glTF model, an orbit camera driven by mouse events), inspects the kui_ext_3d
// extension (state, perf stats, watch), unloads and reloads it and reads every invoke result.
// One #tour-log row per check; #tour-result ends as "PASS n/n". The panel buttons repeat the
// calls by hand.
package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

const tag3D = "kui-3d-view"

// ---- Self-check bookkeeping: count, log one <tr> per check, publish the verdict ----------
// check:    a documented behaviour; a failure makes the tour FAIL.
// knownBug: a documented behaviour the engine currently gets wrong (reported bug). A failure is
//           logged as "KNOWN BUG" and kept out of the verdict; once it passes it counts as ok.

type tour struct {
	doc         *kui.Document
	total       int
	failed      int
	known       int
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

func (t *tour) knownBug(name string, ok bool, bug, detail string) {
	if ok {
		t.total++
		t.row(name, "ok", "ok")
		return
	}
	t.known++
	fmt.Printf("tour known bug: %s: %s\n", name, detail)
	t.row(name, "known", fmt.Sprintf("KNOWN BUG: %s (%s)", bug, detail))
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
	if t.known > 0 {
		text += fmt.Sprintf(" (known bugs: %d)", t.known)
	}
	_ = result.SetTextContent(text)
	_ = result.SetAttribute("data-total", fmt.Sprint(t.total))
	_ = result.SetAttribute("data-failed", fmt.Sprint(t.failed))
	_ = result.SetAttribute("data-known", fmt.Sprint(t.known))
	if t.failed == 0 {
		_ = result.AddClass("pass")
	} else {
		_ = result.AddClass("fail")
	}
}

// ---- Host state: the document, paths next to the executable, camera and scene ------------

type scene struct {
	ground, box, sphere, cylinder, sun, lamp, tank int64
	transformsOK, materialsOK, tankTransformOK     bool
}

type host struct {
	doc        *kui.Document
	exeDir     string
	extPath    string
	tankPath   string
	az, el     float64
	dist       float64
	scene      scene
	color      int
	spinning   bool
	spinAngle  float64
	spinTimer  uint64
	tankOn     bool
	dragButton kui.MouseButton
	lastX      float32
	lastY      float32
}

var boxColors = [4][3]float64{{0.20, 0.55, 0.95}, {0.95, 0.65, 0.15}, {0.65, 0.30, 0.90}, {0.95, 0.95, 0.95}}

func extFileName() string {
	switch runtime.GOOS {
	case "windows":
		return "kui_ext_3d.dll"
	case "darwin":
		return "libkui_ext_3d.dylib"
	}
	return "libkui_ext_3d.so"
}

func (h *host) setText(id, text string) {
	if element, err := h.doc.GetElementByID(id); err == nil && element != nil {
		_ = element.SetTextContent(text)
		element.Close()
	}
}

// ---- Extension calls: InvokeExtension returns the values, or an error carrying the result's
// error text when the call itself failed (unknown method, wrong argument count, ...). ---------

func num(value float64) kui.ExtensionValue {
	return kui.ExtensionValue{Type: kui.ExtensionFloat, Float: value}
}
func integer(value int64) kui.ExtensionValue {
	return kui.ExtensionValue{Type: kui.ExtensionInt, Int: value}
}
func nums(values ...float64) []kui.ExtensionValue {
	out := make([]kui.ExtensionValue, len(values))
	for i, value := range values {
		out[i] = num(value)
	}
	return out
}

func (h *host) call(method string, args []kui.ExtensionValue, maxResults int) ([]kui.ExtensionValue, error) {
	view, err := h.doc.GetElementByID("view")
	if err != nil || view == nil {
		return nil, errors.New("no #view element")
	}
	defer view.Close()
	return h.doc.InvokeExtension(view, method, args, maxResults)
}

// First result as an entity / asset handle; 0 when the call failed.
func (h *host) callInt(method string, args []kui.ExtensionValue) int64 {
	values, err := h.call(method, args, 1)
	if err != nil || len(values) < 1 || values[0].Type != kui.ExtensionInt {
		return 0
	}
	return values[0].Int
}

// True when the call succeeded and its first result (if any) is true.
func (h *host) callOK(method string, args []kui.ExtensionValue) bool {
	values, err := h.call(method, args, 1)
	if err != nil {
		return false
	}
	if len(values) >= 1 && values[0].Type == kui.ExtensionBool {
		return values[0].Bool
	}
	return true
}

func describe(values []kui.ExtensionValue, err error) string {
	if err != nil {
		return "error=" + err.Error()
	}
	parts := make([]string, len(values))
	for i, value := range values {
		parts[i] = fmt.Sprint(value.Int)
	}
	return fmt.Sprintf("count=%d values=%s", len(values), strings.Join(parts, ","))
}

// ---- Orbit camera: azimuth / elevation / distance around the origin -> setCameraLookAt ------

func (h *host) applyCamera() error {
	h.el = math.Max(-10, math.Min(80, h.el))
	h.dist = math.Max(3, math.Min(30, h.dist))
	h.az = math.Mod(h.az+360, 360)
	az, el := h.az*math.Pi/180, h.el*math.Pi/180
	_, err := h.call("setCameraLookAt", nums(h.dist*math.Cos(el)*math.Sin(az), 0.3+h.dist*math.Sin(el),
		h.dist*math.Cos(el)*math.Cos(az), 0, 0.3, 0), 0)
	h.setText("cam-out", fmt.Sprintf("az=%.0f el=%.0f dist=%.1f", h.az, h.el, h.dist))
	return err
}

func (h *host) cameraStep(dAz, dEl, zoom float64) {
	h.az += dAz
	h.el += dEl
	h.dist *= zoom
	_ = h.applyCamera()
}

func (h *host) cameraReset() {
	h.az, h.el, h.dist = 35, 22, 8
	_ = h.applyCamera()
}

// ---- Scene: ground, three meshes, two lights, PBR materials and the glTF tank -------------

// setEntityTransform(entity, position, rotation quaternion about Y, uniform scale) -> bool.
func (h *host) place(entity int64, x, y, z, angle, scale float64) bool {
	args := append([]kui.ExtensionValue{integer(entity)},
		nums(x, y, z, 0, math.Sin(angle/2), 0, math.Cos(angle/2), scale, scale, scale)...)
	return h.callOK("setEntityTransform", args)
}

// createBox / createSphere / ...(r, g, b) -> entity, then addToScene and place it.
func (h *host) addMesh(method string, r, g, b, x, y, z, scale float64) int64 {
	entity := h.callInt(method, nums(r, g, b))
	if entity == 0 || !h.callOK("addToScene", []kui.ExtensionValue{integer(entity)}) {
		return 0
	}
	if !h.place(entity, x, y, z, 0, scale) {
		h.scene.transformsOK = false
	}
	return entity
}

// setMaterialProperties(entity, r, g, b, metallic, roughness) -> bool.
func (h *host) setMaterial(entity int64, r, g, b, metallic, roughness float64) bool {
	return h.callOK("setMaterialProperties",
		append([]kui.ExtensionValue{integer(entity)}, nums(r, g, b, metallic, roughness)...))
}

// loadGltf(path) -> asset handle (also adds it to the scene); setAssetTransform places it.
func (h *host) loadTank() {
	h.scene.tank = h.callInt("loadGltf", []kui.ExtensionValue{{Type: kui.ExtensionString, String: h.tankPath}})
	h.scene.tankTransformOK = h.scene.tank != 0 && h.callOK("setAssetTransform",
		append([]kui.ExtensionValue{integer(h.scene.tank)}, nums(0, -0.75, -2.6, 0, 0, 0, 1, 0.9, 0.9, 0.9)...))
}

func (h *host) recolorBox() {
	c := boxColors[h.color]
	h.setMaterial(h.scene.box, c[0], c[1], c[2], 0.1, 0.45)
}

func (h *host) buildScene() {
	h.scene = scene{transformsOK: true}
	h.spinAngle = 0
	h.callOK("setCameraProjection", nums(55, 0.1, 300))
	h.callOK("setSkyboxColor", nums(0.04, 0.07, 0.14))
	// createDirectionalLight(r, g, b, intensity, dirX, dirY, dirZ); createPointLight(..., x, y, z, range)
	h.scene.sun = h.callInt("createDirectionalLight", nums(1, 0.95, 0.88, 110000, -0.8, -1, -0.5))
	h.scene.lamp = h.callInt("createPointLight", nums(0.35, 0.6, 1, 60000, -2.5, 3.5, 2, 12))
	h.scene.ground = h.addMesh("createPlane", 0.20, 0.24, 0.30, 0, -0.75, 0, 9)
	h.scene.box = h.addMesh("createBox", 0.2, 0.55, 0.95, 0, 0, 0, 1.4)
	h.scene.sphere = h.addMesh("createSphere", 0.9, 0.25, 0.2, -2.4, 0, 0, 1.3)
	h.scene.cylinder = h.addMesh("createCylinder", 0.25, 0.85, 0.45, 2.4, 0, 0, 1.1)
	h.scene.materialsOK = h.setMaterial(h.scene.sphere, 0.9, 0.25, 0.2, 0.9, 0.2) &&
		h.setMaterial(h.scene.cylinder, 0.25, 0.85, 0.45, 0, 0.8)
	h.recolorBox()
	if h.tankOn {
		h.loadTank()
	}
	_ = h.applyCamera()
}

// A fresh <kui-3d-view id="view"> in #view-host.
func (h *host) createView() {
	parent, err := h.doc.GetElementByID("view-host")
	if err != nil || parent == nil {
		return
	}
	defer parent.Close()
	if view, err := h.doc.CreateElement(tag3D); err == nil && view != nil {
		_ = view.SetAttribute("id", "view")
		_ = parent.AppendChild(view)
		view.Close()
	}
}

// Removes the 3D element before its extension goes away.
func (h *host) removeView() {
	if view, err := h.doc.GetElementByID("view"); err == nil && view != nil {
		_ = view.Remove()
		view.Close()
	}
}

func (h *host) showScene() {
	onOff := func(on bool) string {
		if on {
			return "on"
		}
		return "off"
	}
	h.setText("scene-out", fmt.Sprintf("color=%d spin=%s tank=%s entities=%d,%d,%d,%d", h.color,
		onOff(h.spinning), onOff(h.scene.tank != 0), h.scene.ground, h.scene.box, h.scene.sphere, h.scene.cylinder))
}

// ---- Mouse orbit: listeners on #view-host (events bubble up from the 3D element) ----------
// Left drag orbits, right drag (vertical) zooms, every wheel event zooms in one step: the C ABI
// has no accessor for a DOM wheel event's delta, so the direction is not available.

func (h *host) listenMouse(ctx *runner.Context) error {
	viewHost, err := h.doc.GetElementByID("view-host")
	if err != nil || viewHost == nil {
		return fmt.Errorf("no #view-host: %v", err)
	}
	ctx.Keep(viewHost)
	listeners := map[string]func(*kui.Event){
		"mousedown": func(e *kui.Event) {
			point, err := e.MouseClient()
			if err != nil {
				return
			}
			h.lastX, h.lastY = point.X, point.Y
			h.dragButton, _ = e.MouseButton()
		},
		"mousemove": func(e *kui.Event) {
			point, err := e.MouseClient()
			if h.dragButton == kui.MouseButtonNone || err != nil {
				return
			}
			switch h.dragButton {
			case kui.MouseButtonLeft:
				h.cameraStep(-float64(point.X-h.lastX)*0.5, float64(point.Y-h.lastY)*0.3, 1)
			case kui.MouseButtonRight:
				h.cameraStep(0, 0, 1+float64(point.Y-h.lastY)*0.01)
			}
			h.lastX, h.lastY = point.X, point.Y
		},
		"mouseup":    func(*kui.Event) { h.dragButton = kui.MouseButtonNone },
		"mouseleave": func(*kui.Event) { h.dragButton = kui.MouseButtonNone },
		"wheel":      func(*kui.Event) { h.cameraStep(0, 0, 0.9) },
	}
	for name, fn := range listeners {
		if _, err := viewHost.AddEventListener(name, fn); err != nil {
			return err
		}
	}
	return nil
}

// ---- Panel buttons: camera, scene, extension manager, invoke results ------------------------

func (h *host) showExtState(prefix string) {
	count, watching, _ := h.doc.ExtensionState()
	loaded, _ := h.doc.ExtensionLoaded(tag3D)
	h.setText("ext-out", fmt.Sprintf("%sloaded=%d extensions=%d watching=%d", prefix, b2i(loaded), count, b2i(watching)))
}

func (h *host) showPerf(prefix string) {
	stats, _ := h.doc.ExtensionPerfStats(tag3D)
	h.setText("ext-out", fmt.Sprintf("%sticks=%d paints=%d max_tick=%.2fms max_paint=%.2fms", prefix,
		stats.TickCount, stats.PaintCount, stats.MaxTickMS, stats.MaxPaintMS))
}

func b2i(value bool) int {
	if value {
		return 1
	}
	return 0
}

func (h *host) buttons(ctx *runner.Context) error {
	actions := map[string]func(){
		"btn-cam-left":  func() { h.cameraStep(-15, 0, 1) },
		"btn-cam-right": func() { h.cameraStep(15, 0, 1) },
		"btn-cam-up":    func() { h.cameraStep(0, 10, 1) },
		"btn-cam-down":  func() { h.cameraStep(0, -10, 1) },
		"btn-cam-in":    func() { h.cameraStep(0, 0, 0.8) },
		"btn-cam-out":   func() { h.cameraStep(0, 0, 1.25) },
		"btn-cam-reset": h.cameraReset,
		"btn-scene-color": func() {
			h.color = (h.color + 1) % 4
			h.recolorBox()
			h.showScene()
		},
		"btn-scene-spin": func() {
			if h.spinning {
				_ = kui.ClearTimer(h.spinTimer)
				h.spinning = false
			} else if id, err := kui.SetInterval(func() {
				h.spinAngle += 0.06
				h.place(h.scene.box, 0, 0, 0, h.spinAngle, 1.4)
				h.place(h.scene.cylinder, 2.4, 0, 0, -h.spinAngle, 1.1)
			}, 33); err == nil {
				h.spinTimer, h.spinning = id, true
			}
			h.showScene()
		},
		"btn-scene-tank": func() {
			if h.scene.tank != 0 {
				h.callOK("destroyAsset", []kui.ExtensionValue{integer(h.scene.tank)})
				h.scene.tank, h.tankOn = 0, false
			} else {
				h.tankOn = true
				h.loadTank()
			}
			h.showScene()
		},
		"btn-ext-state": func() { h.showExtState("") },
		"btn-ext-perf":  func() { h.showPerf("") },
		"btn-ext-reset": func() {
			_ = h.doc.ResetExtensionPerfStats()
			h.showPerf("reset: ")
		},
		// Unload -> reload from the DLL path -> fresh element -> rebuilt scene.
		"btn-ext-reload": func() {
			if h.spinning {
				_ = kui.ClearTimer(h.spinTimer)
				h.spinning = false
			}
			h.removeView()
			_ = h.doc.UnloadExtension(tag3D)
			reloaded, _ := h.doc.ReloadExtension(h.extPath)
			h.createView()
			h.buildScene()
			state := "missing"
			if h.scene.box != 0 {
				state = "rebuilt"
			}
			h.showExtState(fmt.Sprintf("reloaded=%d box=%s | ", b2i(reloaded), state))
		},
		"btn-ext-watch": func() {
			if _, watching, _ := h.doc.ExtensionState(); watching {
				_ = h.doc.StopWatchingExtensions()
			} else {
				_ = h.doc.WatchExtensions(h.exeDir)
			}
			h.showExtState("")
		},
		"btn-res-ok": func() { h.setText("res-out", "status=0 "+describe(h.call("getRenderStats", nil, 4))) },
		"btn-res-unknown": func() {
			values, err := h.call("noSuchMethod", nil, 1)
			h.setText("res-out", "status=3 count=0 "+describe(values, err))
		},
		"btn-res-args": func() {
			values, err := h.call("addToScene", nil, 1)
			h.setText("res-out", "status=4 count=0 "+describe(values, err))
		},
	}
	for id, action := range actions {
		button, err := h.doc.GetElementByID(id)
		if err != nil || button == nil {
			return fmt.Errorf("no #%s: %v", id, err)
		}
		ctx.Keep(button)
		action := action
		if _, err := button.AddEventListener("click", func(*kui.Event) { action() }); err != nil {
			return err
		}
	}
	return nil
}

// ---- Self-check: the scene built through extension calls -----------------------------------

func (h *host) checkScene(t *tour, prefix string) {
	s := h.scene
	t.check(prefix+".lights", s.sun != 0 && s.lamp != 0, fmt.Sprintf("sun %d, point light %d", s.sun, s.lamp))
	t.check(prefix+".meshes", s.ground != 0 && s.box != 0 && s.sphere != 0 && s.cylinder != 0 &&
		s.box != s.sphere && s.sphere != s.cylinder && s.ground != s.box,
		fmt.Sprintf("entities %d %d %d %d", s.ground, s.box, s.sphere, s.cylinder))
	t.check(prefix+".transforms", s.transformsOK, "setEntityTransform returned false")
	t.check(prefix+".materials", s.materialsOK, "setMaterialProperties returned false")
	t.check(prefix+".gltf", s.tank != 0 && s.tankTransformOK,
		fmt.Sprintf("loadGltf(%s) -> %d, setAssetTransform %v", h.tankPath, s.tank, s.tankTransformOK))
}

// ---- Self-check: invoke results (values / count, and the error of a failed call) ----------

func (h *host) checkResults(t *tour) {
	values, err := h.call("getRenderStats", nil, 4)
	allInt := true
	for _, value := range values {
		allInt = allInt && value.Type == kui.ExtensionInt
	}
	t.check("res.get_count", err == nil && len(values) == 4 && allInt,
		fmt.Sprintf("getRenderStats: %d values, err %v", len(values), err))
	t.check("res.get_error-empty", err == nil, fmt.Sprintf("error on success: %v", err))

	values, err = h.call("noSuchMethod", nil, 1)
	t.check("res.status-unknown-method", err != nil && len(values) == 0 && strings.Contains(err.Error(), "noSuchMethod"),
		fmt.Sprintf("values %d err %v", len(values), err))
	values, err = h.call("addToScene", nil, 1)
	t.check("res.status-arg-count", err != nil && len(values) == 0 && strings.Contains(err.Error(), "argument count"),
		fmt.Sprintf("values %d err %v", len(values), err))
}

// ---- Self-check: the orbit camera, through the buttons' function and real mouse events -----

func approx(a, b float64) bool { return math.Abs(a-b) < 0.01 }

func (h *host) checkCamera(t *tour) {
	t.check("cam.set_camera_look_at", h.applyCamera() == nil, "setCameraLookAt failed")
	h.cameraStep(15, 0, 1)
	t.check("cam.step", approx(h.az, 50), fmt.Sprintf("azimuth %.2f after +15", h.az))

	var rect kui.Rect
	if viewHost, err := h.doc.GetElementByID("view-host"); err == nil && viewHost != nil {
		rect, _ = viewHost.BoundingRect()
		viewHost.Close()
	}
	cx, cy := rect.X+rect.Width/2, rect.Y+rect.Height/2
	inject := func(kind kui.MouseEventType, x, y float32, button kui.MouseButton, clicks int, dy float32) {
		_ = h.doc.DispatchMouseEvent(kui.MouseInjection{Kind: kind, X: x, Y: y, Button: button, ClickCount: clicks, DeltaY: dy})
	}

	// Left drag 40 px to the right: the listeners turn it into -20 degrees of azimuth.
	before := h.az
	inject(kui.MousePressed, cx, cy, kui.MouseButtonLeft, 1, 0)
	inject(kui.MouseMoved, cx+40, cy, kui.MouseButtonLeft, 0, 0)
	inject(kui.MouseReleased, cx+40, cy, kui.MouseButtonLeft, 1, 0)
	t.check("cam.mouse-drag-orbit", approx(h.az, before-20) && h.dragButton == kui.MouseButtonNone,
		fmt.Sprintf("azimuth %.2f -> %.2f, drag button %d", before, h.az, h.dragButton))

	// Right drag 20 px down zooms out by 20 %.
	before = h.dist
	inject(kui.MousePressed, cx, cy, kui.MouseButtonRight, 1, 0)
	inject(kui.MouseMoved, cx, cy+20, kui.MouseButtonRight, 0, 0)
	inject(kui.MouseReleased, cx, cy+20, kui.MouseButtonRight, 1, 0)
	t.check("cam.mouse-drag-zoom", approx(h.dist, before*1.2), fmt.Sprintf("distance %.2f -> %.2f", before, h.dist))

	before = h.dist
	inject(kui.MouseWheel, cx, cy, kui.MouseButtonNone, 0, -120)
	t.check("cam.mouse-wheel", approx(h.dist, before*0.9), fmt.Sprintf("distance %.2f -> %.2f", before, h.dist))
	h.cameraReset()
}

// ---- Self-check: extension manager state, directory watching, unload / reload / load ------

func (h *host) extCount() int {
	count, _, _ := h.doc.ExtensionState()
	return count
}

func (h *host) extLoaded() bool {
	loaded, _ := h.doc.ExtensionLoaded(tag3D)
	return loaded
}

// Returns the initial number of loaded extensions.
func (h *host) checkManager(t *tour) int {
	t.check("ext.is_loaded", h.extLoaded(), "kui-3d-view not loaded")
	initial, watching, err := h.doc.ExtensionState()
	t.check("ext.get_extension_state", err == nil && initial >= 1 && !watching,
		fmt.Sprintf("count %d watching %v err %v", initial, watching, err))
	err = h.doc.WatchExtensions(h.exeDir)
	_, watching, _ = h.doc.ExtensionState()
	t.check("ext.watch_extensions-on", err == nil && watching, fmt.Sprintf("watching %v err %v", watching, err))
	err = h.doc.StopWatchingExtensions()
	_, watching, _ = h.doc.ExtensionState()
	t.check("ext.watch_extensions-off", err == nil && !watching, fmt.Sprintf("watching %v err %v", watching, err))
	_, err = h.doc.ExtensionPerfStats(tag3D)
	t.check("perf.get_perf_stats", err == nil, fmt.Sprint(err))
	return initial
}

// Unload -> confirm -> reload -> load, before the page has any 3D element: the element is
// created afterwards, on the extension that is loaded at the end. (The "Unload + reload 3D" button
// does the same with a live view; the example's pixel check covers that second view.)
func (h *host) checkLifecycle(t *tour, initial int) {
	err := h.doc.UnloadExtension(tag3D)
	t.check("ext.unload_extension", err == nil && !h.extLoaded() && h.extCount() == initial-1,
		fmt.Sprintf("err %v loaded %v count %d", err, h.extLoaded(), h.extCount()))
	err = h.doc.UnloadExtension(tag3D)
	var kerr *kui.Error
	t.check("ext.unload_extension-again", errors.As(err, &kerr) && kerr.Status == kui.ErrNotFound, fmt.Sprint(err))

	reloaded, err := h.doc.ReloadExtension(h.extPath)
	t.check("ext.reload_extension", err == nil && reloaded && h.extLoaded() && h.extCount() == initial,
		fmt.Sprintf("err %v reloaded %v count %d", err, reloaded, h.extCount()))

	loaded, err := h.doc.LoadExtension(filepath.Join(h.exeDir, "no_such_extension_"+extFileName()))
	t.check("ext.load_extension-missing", err == nil && !loaded && h.extCount() == initial,
		fmt.Sprintf("err %v loaded %v count %d", err, loaded, h.extCount()))

	// unload_all removes the custom element definitions of every extension it unloads (like
	// unload_extension does), so the DLLs can be loaded again below.
	err = h.doc.UnloadAllExtensions()
	t.check("ext.unload_all_extensions", err == nil && h.extCount() == 0 && !h.extLoaded(),
		fmt.Sprintf("err %v count %d", err, h.extCount()))
	defined, _ := h.doc.CustomElementDefined("kui-editor")
	t.check("ext.unload_all-undefines-elements", !defined, "<kui-editor> still defined")

	loaded, err = h.doc.LoadExtension(h.extPath)
	t.check("ext.load_extension", err == nil && loaded && h.extCount() == 1 && h.extLoaded(),
		fmt.Sprintf("err %v loaded %v count %d", err, loaded, h.extCount()))
}

// After a few frames: what the perf counters saw, then reset them; publish the verdict.
func (h *host) checkPerfAndFinish(t *tour) {
	stats, _ := h.doc.ExtensionPerfStats(tag3D)
	t.knownBug("perf.paints-recorded", stats.PaintCount > 0 || stats.TickCount > 0,
		"ExtensionManager never records tick / paint timings (recordTickPerf / recordPaintPerf have no caller)",
		fmt.Sprintf("ticks %d paints %d after rendering", stats.TickCount, stats.PaintCount))
	err := h.doc.ResetExtensionPerfStats()
	stats, _ = h.doc.ExtensionPerfStats(tag3D)
	t.check("perf.reset_perf_stats", err == nil && stats.TickCount == 0 && stats.PaintCount == 0 && stats.TotalPaintMS == 0,
		fmt.Sprintf("err %v ticks %d paints %d", err, stats.TickCount, stats.PaintCount))
	h.showExtState("self-check done | ")
	t.finish()
}

func ready(ctx *runner.Context) error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	h := &host{doc: ctx.Document, exeDir: filepath.Dir(executable), az: 35, el: 22, dist: 8, tankOn: true}
	h.extPath = filepath.Join(h.exeDir, extFileName())
	h.tankPath = filepath.Join(h.exeDir, "assets", "native_3d", "stylized_tank.glb")
	t := &tour{doc: ctx.Document}

	if err := h.listenMouse(ctx); err != nil {
		return err
	}
	initial := h.checkManager(t)
	h.checkLifecycle(t, initial)
	h.createView()
	h.buildScene()
	h.checkScene(t, "scene")
	h.checkResults(t)
	h.checkCamera(t)
	h.showScene()
	if _, err := kui.SetTimeout(func() { h.checkPerfAndFinish(t) }, 400); err != nil {
		t.check("perf.timer", false, err.Error())
		t.finish()
	}
	return h.buttons(ctx)
}
