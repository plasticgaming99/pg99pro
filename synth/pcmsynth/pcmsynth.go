// this package converts sf2 meaning to pcm-rompler meaning and use them
package pcmsynth

import "github.com/plasticgaming99/pg99pro/synth/sf2abst"

type Voice struct {
	Name        string
	SampleType  sf2abst.SampleType
	Sample      []float32 // left/mono sample
	RSample     []float32 // it will be 0 length with mono
	SampleRate  uint32
	LoopStart   uint32
	LoopEnd     uint32
	OriginalKey uint8
	PitchCorr   int8
	ShdrOrigin  int // original shdr index
}

func NewGenerateVoicesOptions() *GenerateVoicesOptions {
	return &GenerateVoicesOptions{
		ResamplerEnabled: true,
		ResamplerRate:    48000,
	}
}

type GenerateVoicesOptions struct {
	ResamplerEnabled  bool // resample
	ResamplerRate     uint // resample rate, default 48000 hz
	KeepOriginalOrder bool // when true, it keeps original index
}

func GenerateVoices(sf2 *sf2abst.SF2Abst, op *GenerateVoicesOptions) []Voice {
	voices := make([]Voice, 0)
	smpls := make([]sf2abst.Sample, 0, len(sf2.Pdta.Shdr))
	for i := 0; i < len(sf2.Pdta.Shdr); i++ {
		smpls = append(smpls, sf2abst.ShdrToSample(sf2.Pdta.Shdr[i]))
	}
	for i := 0; i < len(smpls); i++ {

	}
	return voices
}
