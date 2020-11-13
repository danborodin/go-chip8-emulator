package emulator

import (
	"encoding/binary"
	"log"
	"math"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	VBlankFrequency = 60
	DisplayScale    = 10
	EmulatorW       = Chip8DisplayW * DisplayScale
	EmulatorH       = Chip8DisplayH * DisplayScale
	WindowW         = EmulatorW
	WindowH         = EmulatorH
	FontPerW        = 32
	AudioSamples    = 64
)

type Emulator struct {
	rom      []byte
	chip8    *Chip8
	renderer *sdl.Renderer
	audio    sdl.AudioDeviceID
	running  bool
	focus    bool
}

// Map keyboard key to Chip8 keys
var scanCodeToKey = map[int]byte{
	sdl.SCANCODE_1: 0x1,
	sdl.SCANCODE_2: 0x2,
	sdl.SCANCODE_3: 0x3,
	sdl.SCANCODE_4: 0xc,
	sdl.SCANCODE_Q: 0x4,
	sdl.SCANCODE_W: 0x5,
	sdl.SCANCODE_E: 0x6,
	sdl.SCANCODE_R: 0xd,
	sdl.SCANCODE_A: 0x7,
	sdl.SCANCODE_S: 0x8,
	sdl.SCANCODE_D: 0x9,
	sdl.SCANCODE_F: 0xe,
	sdl.SCANCODE_Z: 0xa,
	sdl.SCANCODE_X: 0x0,
	sdl.SCANCODE_C: 0xb,
	sdl.SCANCODE_V: 0xf,
}

func initRenderer() *sdl.Renderer {
	window, err := sdl.CreateWindow("Chip-8 Emulator", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, WindowW, WindowH, sdl.WINDOW_SHOWN)
	if err != nil {
		log.Println(err)
	}

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_PRESENTVSYNC)
	if err != nil {
		log.Println(err)
	}

	window.Hide()
	sdl.PumpEvents()
	window.Show()

	return renderer
}

func initAudio() sdl.AudioDeviceID {
	want := &sdl.AudioSpec{
		Freq:     AudioSamples * VBlankFrequency,
		Format:   sdl.AUDIO_F32LSB,
		Channels: 1,
		Samples:  AudioSamples,
	}
	have := &sdl.AudioSpec{}
	audio, err := sdl.OpenAudioDevice("", false, want, have, sdl.AUDIO_ALLOW_ANY_CHANGE)
	if err != nil {
		log.Println("Open Audio Device ", err)
	}

	sdl.PauseAudioDevice(audio, false)
	return audio
}

// Return pointer to new Emulator struct
func NewEmulator(b []byte) *Emulator {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		log.Println("sdl.Init ", err)
	}

	renderer := initRenderer()
	audio := initAudio()

	return &Emulator{
		rom:      b,
		chip8:    newChip8(b),
		renderer: renderer,
		audio:    audio,
		running:  true,
		focus:    true,
	}
}

func (e *Emulator) Run() {
	perVblankCycle := Chip8Frequency / VBlankFrequency
	cycle := 0

	for e.running {
		cycle++
		if e.focus {
			e.chip8.step()
		}

		if cycle > perVblankCycle {
			cycle = 0
			e.draw()

			if e.focus {
				e.updateSound()
				e.chip8.decrementTimer()
			}
		}
		e.pollEvents()
	}
}

func (e *Emulator) draw() {
	e.renderer.SetDrawColor(0, 0, 0, 255)
	e.renderer.Clear()

	e.renderer.SetDrawColor(255, 255, 255, 255)
	for y := int32(0); y < Chip8DisplayH; y++ {
		for x := int32(0); x < Chip8DisplayW; x++ {
			if e.chip8.dsp[y*Chip8DisplayW+x] != 0 {
				e.renderer.FillRect(&sdl.Rect{
					X: x * DisplayScale,
					Y: y * DisplayScale,
					W: DisplayScale,
					H: DisplayScale,
				})
			}
		}
	}
	e.renderer.Present()
}

func (e *Emulator) pollEvents() {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch ev := event.(type) {
		case *sdl.QuitEvent:
			e.running = false
		case *sdl.KeyboardEvent:
			switch ev.Type {
			case sdl.KEYDOWN:
				if i, ok := scanCodeToKey[int(ev.Keysym.Scancode)]; ok {
					e.chip8.keys[i] = 1
				} else {
					if ev.Keysym.Scancode == sdl.SCANCODE_L {
						e.chip8 = newChip8(e.rom)
					}
				}
			case sdl.KEYUP:
				if i, ok := scanCodeToKey[int(ev.Keysym.Scancode)]; ok {
					e.chip8.keys[i] = 0
				}
			}
		case *sdl.WindowEvent:
			switch ev.Event {
			case sdl.WINDOWEVENT_FOCUS_LOST:
				e.focus = false
			case sdl.WINDOWEVENT_FOCUS_GAINED:
				e.focus = true
			}
		}
	}
}

func (e *Emulator) updateSound() {
	if e.chip8.soundTtimer > 0 {
		samples := make([]byte, 4*AudioSamples)
		for i := 0; i < len(samples); i += 4 {
			f := 2.0 * math.Pi / 180.0 * float64(360*i/AudioSamples)
			f = math.Sin(f)
			binary.LittleEndian.PutUint32(samples[i:], math.Float32bits(float32(f)))
		}
		err := sdl.QueueAudio(e.audio, samples)
		if err != nil {
			log.Println(err)
		}
	}
}
