//go:build ignore

// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/kui_api_media_lottie.go; edit that file, not this copy.
// API tour: game audio, media elements and Lottie (examples/api_tours/media_lottie). At start-up
// the host drives the document's audio manager (play by path into groups, per-handle volume /
// pan / pitch / pause / seek, group volume and mute, pause-all / stop-all, preload / unload,
// counts), an <audio> and a <video> element (src, load, loop / autoplay / muted / volume flags,
// seek, play / pause, network and ready state) and a <lottie-player> (state, natural size, speed,
// loop, pause, seek, stop), checking every call against the state read back. Time-dependent
// results are read once frames have run; #tour-result ends as "PASS n/n". Works without an audio
// device (KUI_MEDIA_NO_DEVICE=1): the checks read state, they do not listen.
package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

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

// ---- Small helpers ----------------------------------------------------------------------

func near(a, b, tolerance float64) bool { return math.Abs(a-b) <= tolerance }

// status is the KuiStatus behind an error (0 for nil).
func status(err error) kui.Status {
	var kerr *kui.Error
	if errors.As(err, &kerr) {
		return kerr.Status
	}
	if err != nil {
		return -1
	}
	return 0
}

func b2i(value bool) int {
	if value {
		return 1
	}
	return 0
}

const (
	audioSeek = 300.0 // far from where playback is after a second (the file is 675 s)
	videoSeek = 5.0   // the clip is 13.3 s
)

type host struct {
	doc   *kui.Document
	t     *tour
	audio *kui.Audio

	mp3, mp4, lottie, missing string

	aud, vid, lot *kui.Element // kept for the whole run (the buttons use them)

	available   bool
	music       uint32 // the looping music handle the buttons drive
	lottieMark  float64
	auPaused    bool
	auQuiet     bool
	auRight     bool
	auFast      bool
	auAllPaused bool
	auMuted     bool
	lotFast     bool
	vidControls bool
}

func (h *host) setText(id, text string) {
	if element, err := h.doc.GetElementByID(id); err == nil && element != nil {
		_ = element.SetTextContent(text)
		element.Close()
	}
}

func (h *host) active(group kui.AudioGroup) uint32 {
	active, _, _ := h.audio.Counts(group)
	return active
}

func (h *host) pooled() uint32 {
	_, pooled, _ := h.audio.Counts(kui.AudioAllGroups)
	return pooled
}

// ---- Game audio: availability, one looping music sound, sfx, groups, pool ---------------

func (h *host) checkAudioGroups() {
	// Group volume / mute live in the manager and work with or without a device.
	_ = h.audio.SetGroup(kui.AudioMusic, 0.4, true)
	volume, muted, _ := h.audio.Group(kui.AudioMusic)
	badVolume := h.audio.SetGroup(kui.AudioMusic, 2.0, false)
	_, _, badGroup := h.audio.Group(kui.AudioGroup(7))
	_ = h.audio.SetGroup(kui.AudioMusic, 1.0, false)
	restored, restoredMuted, _ := h.audio.Group(kui.AudioMusic)
	h.t.check("audio.set_group", near(float64(volume), 0.4, 0.001) && muted && status(badVolume) == kui.ErrInvalidArg &&
		status(badGroup) == kui.ErrInvalidArg && near(float64(restored), 1, 0.001) && !restoredMuted,
		fmt.Sprintf("music 0.4/muted -> %.2f/%v, volume 2 %v, group 7 %v, restored %.2f/%v", volume, muted, badVolume,
			badGroup, restored, restoredMuted))
}

func (h *host) checkAudioUnavailable() {
	handle, err := h.audio.Play(h.mp3, &kui.AudioPlayOptions{Group: kui.AudioSFX, Volume: 1, Pitch: 1})
	loaded, _ := h.audio.Preload(h.mp3, 2)
	h.t.check("audio.play-unavailable", status(err) == kui.ErrNotSupported && handle == 0 && loaded == 0,
		fmt.Sprintf("play %v handle %d, preload loaded %d", err, handle, loaded))
}

func (h *host) checkAudioHandles() {
	loaded, err := h.audio.Preload(h.mp3, 2)
	pooled := h.pooled()
	h.t.check("audio.preload", err == nil && loaded == 2 && pooled >= 2, fmt.Sprintf("%v loaded %d pooled %d", err, loaded, pooled))

	h.music, err = h.audio.Play(h.mp3, &kui.AudioPlayOptions{Group: kui.AudioMusic, Volume: 0.5, Pan: -0.5, Pitch: 1.25, Loop: true})
	info, _, _ := h.audio.Info(h.music)
	h.t.check("audio.play", err == nil && h.music != 0 && info.Group == kui.AudioMusic && near(float64(info.Volume), 0.5, 0.001) &&
		near(float64(info.Pan), -0.5, 0.001) && near(float64(info.Pitch), 1.25, 0.001) && info.Looping && info.Playing,
		fmt.Sprintf("%v handle %d info %+v", err, h.music, info))

	_ = h.audio.SetVolume(h.music, 0.3)
	_ = h.audio.SetPan(h.music, 0.7)
	_ = h.audio.SetPitch(h.music, 0.8)
	info, _, _ = h.audio.Info(h.music)
	h.t.check("audio.set_volume_pan_pitch", near(float64(info.Volume), 0.3, 0.001) && near(float64(info.Pan), 0.7, 0.001) &&
		near(float64(info.Pitch), 0.8, 0.001), fmt.Sprintf("vol %.2f pan %.2f pitch %.2f", info.Volume, info.Pan, info.Pitch))

	_ = h.audio.SetPaused(h.music, true)
	pausedInfo, _, _ := h.audio.Info(h.music)
	_ = h.audio.SetPaused(h.music, false)
	info, _, _ = h.audio.Info(h.music)
	h.t.check("audio.set_paused", !pausedInfo.Playing && info.Playing,
		fmt.Sprintf("playing while paused %v, after resume %v", pausedInfo.Playing, info.Playing))

	seekErr := h.audio.Seek(h.music, 1.5)
	bad := h.audio.Seek(h.music, -1)
	h.t.check("audio.seek-args", seekErr == nil && status(bad) == kui.ErrInvalidArg, fmt.Sprintf("seek %v, seek(-1) %v", seekErr, bad))

	missing, err := h.audio.Play(h.missing, nil)
	h.t.check("audio.play-missing-file", status(err) == kui.ErrIO && missing == 0, fmt.Sprintf("%v handle %d", err, missing))
}

func (h *host) checkAudioBulk() {
	sfx := &kui.AudioPlayOptions{Group: kui.AudioSFX, Volume: 0.2, Pitch: 1}
	first, _ := h.audio.Play(h.mp3, sfx)
	_, _ = h.audio.Play(h.mp3, sfx)
	music, sfxActive, all := h.active(kui.AudioMusic), h.active(kui.AudioSFX), h.active(kui.AudioAllGroups)
	h.t.check("audio.get_counts", music == 1 && sfxActive == 2 && all == 3, fmt.Sprintf("music %d sfx %d all %d", music, sfxActive, all))

	_ = h.audio.SetAllPaused(kui.AudioMusic, true)
	musicInfo, _, _ := h.audio.Info(h.music)
	sfxInfo, _, _ := h.audio.Info(first)
	_ = h.audio.SetAllPaused(kui.AudioAllGroups, false)
	resumed, _, _ := h.audio.Info(h.music)
	h.t.check("audio.set_all_paused", !musicInfo.Playing && sfxInfo.Playing && resumed.Playing,
		fmt.Sprintf("music while group paused %v, sfx %v, music after resume all %v", musicInfo.Playing, sfxInfo.Playing, resumed.Playing))

	_ = h.audio.StopAll(kui.AudioSFX)
	sfxActive, music = h.active(kui.AudioSFX), h.active(kui.AudioMusic)
	h.t.check("audio.stop_all", sfxActive == 0 && music == 1, fmt.Sprintf("sfx %d music %d", sfxActive, music))
}

// checkAudioLate runs once frames have run: the seek landed and the position moves; then stop
// and unload.
func (h *host) checkAudioLate() {
	info, _, _ := h.audio.Info(h.music)
	h.t.check("audio.get_info-time", info.Duration > 1 && info.CurrentTime > 1.45,
		fmt.Sprintf("duration %.2f current %.2f (seeked to 1.50)", info.Duration, info.CurrentTime))

	stopped := h.music
	_ = h.audio.Stop(h.music)
	info, found, _ := h.audio.Info(stopped)
	h.t.check("audio.stop", (!found || !info.Playing) && h.active(kui.AudioMusic) == 0,
		fmt.Sprintf("found %v playing %v, music active %d", found, info.Playing, h.active(kui.AudioMusic)))
	err := h.audio.Stop(987654)
	h.t.check("audio.stop-unknown", err == nil, fmt.Sprint(err))

	before := h.pooled()
	_ = h.audio.Unload(h.mp3)
	after := h.pooled()
	_ = h.audio.UnloadAll()
	h.t.check("audio.unload", before > 0 && after == 0 && h.pooled() == 0,
		fmt.Sprintf("pooled %d -> %d after Unload(path), %d after UnloadAll", before, after, h.pooled()))
	h.music = 0
}

func (h *host) checkAudio() {
	available, err := h.audio.Available()
	h.available = available
	// Recorded as is: with KUI_MEDIA_NO_DEVICE=1 the plugin runs without a device.
	h.t.check("audio.is_available", err == nil, fmt.Sprintf("%v available %v", err, available))
	h.setText("au-avail", fmt.Sprintf("available=%d", b2i(available)))
	h.checkAudioGroups()
	if !available {
		h.checkAudioUnavailable()
		return
	}
	h.checkAudioHandles()
	h.checkAudioBulk()
}

// ---- Media elements: flags, volume, mute, src / load, seek, play / pause ---------------

func (h *host) mediaTime(media *kui.Element) float64 {
	t, err := media.CurrentTime()
	if err != nil {
		return -1
	}
	return t
}

func (h *host) mediaPaused(media *kui.Element) int {
	paused, err := media.Paused()
	if err != nil {
		return -1
	}
	return b2i(paused)
}

func (h *host) checkMediaSetup(name string, media *kui.Element, src string) {
	state, _ := media.MediaNetworkState()
	h.t.check(name+".network_state-empty", state == 0, fmt.Sprintf("networkState %d before src (expected 0 NETWORK_EMPTY)", state))

	_ = media.SetMediaLoop(true)
	_ = media.SetMediaAutoplay(false)
	loop, autoplay, ended, _ := media.MediaPlaybackFlags()
	h.t.check(name+".playback_flags", loop && !autoplay && !ended, fmt.Sprintf("loop %v autoplay %v ended %v", loop, autoplay, ended))

	_ = media.SetMuted(true)
	on, _ := media.Muted()
	_ = media.SetMuted(false)
	off, _ := media.Muted()
	_ = media.SetVolume(0.4)
	volume, _ := media.Volume()
	h.t.check(name+".muted_volume", on && !off && near(float64(volume), 0.4, 0.001),
		fmt.Sprintf("muted on %v off %v, volume %.2f", on, off, volume))

	_, availErr := media.MediaAvailable()
	_ = media.SetMediaSource(src)
	err := media.MediaLoad()
	state, _ = media.MediaNetworkState()
	h.t.check(name+".load", err == nil && availErr == nil && state == 2 && h.mediaPaused(media) == 1 && h.mediaTime(media) == 0,
		fmt.Sprintf("%v networkState %d (expected 2 LOADING) paused %d available err %v", err, state, h.mediaPaused(media), availErr))
}

func (h *host) checkMediaNotMedia() {
	plain, _ := h.doc.GetElementByID("plain")
	if plain == nil {
		h.t.check("media.not_media_element", false, "no #plain")
		return
	}
	defer plain.Close()
	_, mediaErr := plain.Paused()
	_, lottieErr := plain.LottieState()
	h.t.check("media.not_media_element", status(mediaErr) == kui.ErrNotSupported && status(lottieErr) == kui.ErrNotSupported,
		fmt.Sprintf("Paused %v, LottieState %v", mediaErr, lottieErr))
}

// checkMediaMetadata: a second after load(), metadata should be there without playing.
func (h *host) checkMediaMetadata(name string, media *kui.Element) {
	duration, _ := media.Duration()
	ready, _ := media.MediaReadyState()
	h.t.knownBug(name+".metadata-before-play", duration > 0 && ready >= 1,
		"media load lazily on the first play(): load() leaves readyState 0 and duration 0",
		fmt.Sprintf("duration %.2f readyState %d a second after load()", duration, ready))
}

// checkMediaPlay: play loads the file (duration, network / ready state); then seek while playing.
func (h *host) checkMediaPlay(name string, media *kui.Element, seekTo float64) {
	_ = media.MediaPlay()
	duration, _ := media.Duration()
	network, _ := media.MediaNetworkState()
	ready, _ := media.MediaReadyState()
	h.t.check(name+".play", h.mediaPaused(media) == 0 && duration > seekTo+1 && network == 1 && ready >= 1,
		fmt.Sprintf("paused %d duration %.2f networkState %d readyState %d", h.mediaPaused(media), duration, network, ready))
	_ = media.SetCurrentTime(seekTo)
	after := h.mediaTime(media)
	h.t.check(name+".set_current_time", near(after, seekTo, 0.3), fmt.Sprintf("currentTime %.2f right after seek to %.2f", after, seekTo))
}

// checkMediaPlaying: a while after the seek playback continues from there; pause holds it.
func (h *host) checkMediaPlaying(name string, media *kui.Element, seekTo float64, seekBug bool) {
	now := h.mediaTime(media)
	detail := fmt.Sprintf("currentTime %.2f, seeked to %.2f", now, seekTo)
	if seekBug {
		h.t.knownBug(name+".current_time-after-seek", now >= seekTo && now < seekTo+60,
			"seeking an <audio> element never reaches the player: playback goes on from the old position and currentTime falls back to it", detail)
	} else {
		h.t.check(name+".current_time-after-seek", now >= seekTo && now < seekTo+5, detail)
	}
	_ = media.MediaPause()
	h.t.check(name+".pause", h.mediaPaused(media) == 1 && h.mediaTime(media) > 0,
		fmt.Sprintf("paused %d currentTime %.2f", h.mediaPaused(media), h.mediaTime(media)))
}

// ---- Lottie: load, state, speed / loop, pause / seek / stop ---------------------------

func (h *host) lottieState() kui.LottieState {
	state, _ := h.lot.LottieState()
	return state
}

func (h *host) checkLottieLoaded() {
	state := h.lottieState()
	size, _ := h.lot.LottieNaturalSize()
	h.t.check("lottie.get_state-loaded", state.Load == kui.LottieComplete && state.Playing && near(float64(state.Speed), 1, 0.001),
		fmt.Sprintf("load %d playing %v speed %.2f", state.Load, state.Playing, state.Speed))
	h.t.check("lottie.get_natural_size", near(float64(size.Width), 200, 0.5) && near(float64(size.Height), 200, 0.5),
		fmt.Sprintf("%.0fx%.0f (pulse.json is 200x200)", size.Width, size.Height))
	h.lottieMark = state.Position
}

func (h *host) checkLottieControls() {
	state := h.lottieState()
	h.t.check("lottie.position-advances", state.Position != h.lottieMark, fmt.Sprintf("position %.3f then %.3f", h.lottieMark, state.Position))

	_ = h.lot.SetLottieSpeed(2)
	after := h.lottieState()
	bad := h.lot.SetLottieSpeed(-1)
	h.t.check("lottie.set_speed", near(float64(after.Speed), 2, 0.001) && status(bad) == kui.ErrInvalidArg,
		fmt.Sprintf("speed %.2f, speed -1 %v", after.Speed, bad))
	_ = h.lot.SetLottieSpeed(1)

	_ = h.lot.SetLottieLoop(false)
	off := h.lottieState().Looping
	_ = h.lot.SetLottieLoop(true)
	on := h.lottieState().Looping
	h.t.check("lottie.set_loop", !off && on, fmt.Sprintf("loop off %v, on %v", off, on))

	_ = h.lot.LottiePause()
	_ = h.lot.LottieSeek(0.5)
	after = h.lottieState()
	h.t.check("lottie.pause_seek", !after.Playing && near(after.Position, 0.5, 0.02),
		fmt.Sprintf("playing %v position %.3f", after.Playing, after.Position))
}

func (h *host) checkLottieHeld() {
	state := h.lottieState()
	h.t.check("lottie.paused-holds", !state.Playing && near(state.Position, 0.5, 0.02),
		fmt.Sprintf("playing %v position %.3f after a while", state.Playing, state.Position))
	_ = h.lot.LottieStop()
	stopped := h.lottieState()
	h.t.check("lottie.stop", !stopped.Playing && near(stopped.Position, 0, 0.001),
		fmt.Sprintf("playing %v position %.3f", stopped.Playing, stopped.Position))
	_ = h.lot.LottiePlay() // leave it playing for the panel
}

// ---- Phases: immediate calls, then timers for results that need frames -----------------

func (h *host) phaseThree() {
	h.checkMediaPlaying("audio", h.aud, audioSeek, true)
	h.checkMediaPlaying("video", h.vid, videoSeek, false)
	h.checkLottieHeld()
	if h.available {
		h.checkAudioLate()
	}
	h.t.finish()
}

func (h *host) phaseTwo() {
	h.checkMediaMetadata("audio", h.aud)
	h.checkMediaMetadata("video", h.vid)
	h.checkMediaPlay("audio", h.aud, audioSeek)
	h.checkMediaPlay("video", h.vid, videoSeek)
	h.checkLottieControls()
	_, _ = kui.SetTimeout(h.phaseThree, 700)
}

func (h *host) phaseOne() {
	h.checkLottieLoaded()
	_, _ = kui.SetTimeout(h.phaseTwo, 900)
}

// ---- Panel buttons ----------------------------------------------------------------------

func (h *host) reportAudio() {
	info, found, err := h.audio.Info(h.music)
	if found {
		h.setText("au-out", fmt.Sprintf("music #%d playing=%d vol=%.2f pan=%.2f pitch=%.2f loop=%d t=%.1f/%.1f", h.music,
			b2i(info.Playing), info.Volume, info.Pan, info.Pitch, b2i(info.Looping), info.CurrentTime, info.Duration))
	} else {
		h.setText("au-out", fmt.Sprintf("music: none (handle %d, %v)", h.music, err))
	}
	volume, muted, _ := h.audio.Group(kui.AudioMusic)
	suffix := ""
	if muted {
		suffix = " muted"
	}
	h.setText("au-counts", fmt.Sprintf("active=%d sfx=%d pooled=%d music-group=%.1f%s", h.active(kui.AudioAllGroups),
		h.active(kui.AudioSFX), h.pooled(), volume, suffix))
}

func (h *host) reportLottie() {
	s := h.lottieState()
	h.setText("lot-out", fmt.Sprintf("playing=%d loop=%d speed=%.1f position=%.2f load=%d", b2i(s.Playing), b2i(s.Looping),
		s.Speed, s.Position, s.Load))
}

func (h *host) reportMedia(id string, media *kui.Element) {
	loop, autoplay, ended, _ := media.MediaPlaybackFlags()
	muted, _ := media.Muted()
	volume, _ := media.Volume()
	duration, _ := media.Duration()
	network, _ := media.MediaNetworkState()
	ready, _ := media.MediaReadyState()
	h.setText(id, fmt.Sprintf("paused=%d t=%.1f/%.1f muted=%d vol=%.1f loop=%d autoplay=%d ended=%d net=%d ready=%d",
		h.mediaPaused(media), h.mediaTime(media), duration, b2i(muted), volume, b2i(loop), b2i(autoplay), b2i(ended), network, ready))
}

func (h *host) buttons() map[string]func() {
	audio := func(fn func()) func() { return func() { fn(); h.reportAudio() } }
	lottie := func(fn func()) func() { return func() { fn(); h.reportLottie() } }
	aud := func(fn func()) func() { return func() { fn(); h.reportMedia("aud-out", h.aud) } }
	vid := func(fn func()) func() { return func() { fn(); h.reportMedia("vid-out", h.vid) } }
	sfx := &kui.AudioPlayOptions{Group: kui.AudioSFX, Volume: 0.3, Pitch: 1}
	return map[string]func(){
		"btn-au-play": audio(func() {
			_ = h.audio.Stop(h.music)
			h.music, _ = h.audio.Play(h.mp3, &kui.AudioPlayOptions{Group: kui.AudioMusic, Volume: 1, Pitch: 1, Loop: true})
			h.auPaused, h.auQuiet, h.auRight, h.auFast = false, false, false, false
		}),
		"btn-au-pause": audio(func() { h.auPaused = !h.auPaused; _ = h.audio.SetPaused(h.music, h.auPaused) }),
		"btn-au-volume": audio(func() {
			h.auQuiet = !h.auQuiet
			volume := float32(1)
			if h.auQuiet {
				volume = 0.3
			}
			_ = h.audio.SetVolume(h.music, volume)
		}),
		"btn-au-pan": audio(func() {
			h.auRight = !h.auRight
			pan := float32(-1)
			if h.auRight {
				pan = 1
			}
			_ = h.audio.SetPan(h.music, pan)
		}),
		"btn-au-pitch": audio(func() {
			h.auFast = !h.auFast
			pitch := float32(1)
			if h.auFast {
				pitch = 1.5
			}
			_ = h.audio.SetPitch(h.music, pitch)
		}),
		"btn-au-seek":    audio(func() { _ = h.audio.Seek(h.music, 2) }),
		"btn-au-stop":    audio(func() { _ = h.audio.Stop(h.music) }),
		"btn-au-sfx":     audio(func() { _, _ = h.audio.Play(h.mp3, sfx); _, _ = h.audio.Play(h.mp3, sfx) }),
		"btn-au-stopsfx": audio(func() { _ = h.audio.StopAll(kui.AudioSFX) }),
		"btn-au-pauseall": audio(func() {
			h.auAllPaused = !h.auAllPaused
			_ = h.audio.SetAllPaused(kui.AudioAllGroups, h.auAllPaused)
		}),
		"btn-au-mute": audio(func() {
			h.auMuted = !h.auMuted
			volume := float32(1)
			if h.auMuted {
				volume = 0.5
			}
			_ = h.audio.SetGroup(kui.AudioMusic, volume, h.auMuted)
		}),
		"btn-au-preload": audio(func() { _, _ = h.audio.Preload(h.mp3, 3) }),
		"btn-au-unload":  audio(func() { _ = h.audio.UnloadAll() }),
		"btn-lot-play":   lottie(func() { _ = h.lot.LottiePlay() }),
		"btn-lot-pause":  lottie(func() { _ = h.lot.LottiePause() }),
		"btn-lot-seek0":  lottie(func() { _ = h.lot.LottieSeek(0) }),
		"btn-lot-seek50": lottie(func() { _ = h.lot.LottieSeek(0.5) }),
		"btn-lot-stop":   lottie(func() { _ = h.lot.LottieStop() }),
		"btn-lot-speed": lottie(func() {
			h.lotFast = !h.lotFast
			speed := float32(1)
			if h.lotFast {
				speed = 2
			}
			_ = h.lot.SetLottieSpeed(speed)
		}),
		"btn-lot-loop":  lottie(func() { _ = h.lot.SetLottieLoop(!h.lottieState().Looping) }),
		"btn-aud-play":  aud(func() { _ = h.aud.MediaPlay() }),
		"btn-aud-pause": aud(func() { _ = h.aud.MediaPause() }),
		"btn-aud-seek":  aud(func() { _ = h.aud.SetCurrentTime(3) }),
		"btn-aud-mute":  aud(func() { muted, _ := h.aud.Muted(); _ = h.aud.SetMuted(!muted) }),
		"btn-aud-loop": aud(func() {
			loop, _, _, _ := h.aud.MediaPlaybackFlags()
			_ = h.aud.SetMediaLoop(!loop)
		}),
		"btn-aud-load":     aud(func() { _ = h.aud.MediaLoad() }),
		"btn-vid-play":     vid(func() { _ = h.vid.MediaPlay() }),
		"btn-vid-pause":    vid(func() { _ = h.vid.MediaPause() }),
		"btn-vid-seek":     vid(func() { _ = h.vid.SetCurrentTime(1) }),
		"btn-vid-controls": vid(func() { h.vidControls = !h.vidControls; _ = h.vid.SetMediaControls(h.vidControls) }),
		"btn-vid-autoplay": vid(func() {
			_, autoplay, _, _ := h.vid.MediaPlaybackFlags()
			_ = h.vid.SetMediaAutoplay(!autoplay)
		}),
	}
}

func ready(ctx *runner.Context) error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	assets := filepath.Join(filepath.Dir(executable), "assets")
	h := &host{doc: ctx.Document, t: &tour{doc: ctx.Document}, audio: ctx.Document.Audio(),
		mp3:     filepath.Join(assets, "native_media", "sample.mp3"),
		mp4:     filepath.Join(assets, "native_media", "sample.mp4"),
		lottie:  filepath.Join(assets, "native_animation", "pulse.json"),
		missing: filepath.Join(assets, "native_media", "no_such_sound.mp3")}
	for id, slot := range map[string]**kui.Element{"aud": &h.aud, "vid": &h.vid, "lot": &h.lot} {
		element, err := ctx.Document.GetElementByID(id)
		if err != nil || element == nil {
			return fmt.Errorf("page has no #%s: %v", id, err)
		}
		*slot = element
	}

	_ = h.lot.SetLottieSource(h.lottie)
	_ = h.lot.SetLottieLoop(true)
	_ = h.lot.LottiePlay()

	h.checkAudio()
	h.checkMediaSetup("audio", h.aud, h.mp3)
	h.checkMediaSetup("video", h.vid, h.mp4)
	h.checkMediaNotMedia()
	if _, err := kui.SetTimeout(h.phaseOne, 300); err != nil {
		return err
	}

	for id, fn := range h.buttons() {
		fn := fn
		element, err := ctx.Document.GetElementByID(id)
		if err != nil || element == nil {
			return fmt.Errorf("page has no #%s: %v", id, err)
		}
		ctx.Keep(element)
		if _, err := element.AddEventListener("click", func(*kui.Event) { fn() }); err != nil {
			return err
		}
	}
	return nil
}
