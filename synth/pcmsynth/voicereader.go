package pcmsynth

import (
	"io"

	"github.com/plasticgaming99/pg99pro/synth/sf2abst"
)

type VoiceReader interface {
	ReadFloat32(f32 []float32) (n int, err error)
	NoteOff()
}

// generate synthesizable voice
// with some envelope parameters
func NewVoice(v *VoiceSample, g *sf2abst.GeneratorParam) *Voice {
	vr := Voice{
		voice:       v,
		generator:   g,
		currentStep: 0,
		step:        1.,
	}
	return &vr
}

type Voice struct {
	voice       *VoiceSample
	generator   *sf2abst.GeneratorParam
	currentStep float64 // 1 per 2 byte

	baseStep float64
	step     float64 // pitch

	key          int
	vel          int
	panpotOffset int // mainly for rendering stereo samples

	decay bool
}

func (v *Voice) ReadFloat32(b []float32) (n int, err error) {
	if len(b) < 1 {
		return 0, io.ErrShortBuffer
	}

	for i := 0; i < len(b); i++ {
		if int(v.currentStep) >= len(v.voice.Sample) || int(v.currentStep) >= int(v.voice.LoopEnd) {
			if v.decay {
				return i, io.EOF
			}
			v.currentStep = float64(v.voice.LoopStart)
		}

		b[i] = v.voice.Sample[int(v.currentStep)]

		v.currentStep += v.step
		n++
	}

	return n, nil
}

func (v *Voice) PitchBend(pitch int) {
	v.step = v.baseStep * (float64(pitch) / 8192)
}

func (v *Voice) NoteOff() {
	v.decay = true
}
