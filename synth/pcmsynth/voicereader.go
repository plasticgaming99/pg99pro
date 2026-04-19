package pcmsynth

import (
	"encoding/binary"
)

func NewVoiceReader(v *Voice) VoiceReader {
	vr := VoiceReader{
		voice:       v,
		currentStep: 0,
	}
	return vr
}

type VoiceReader struct {
	voice       *Voice
	currentStep int // 1 per 2 byte
}

/*func (v *VoiceReader) Read(b []byte) (n int, err error) {
	i := v.currentStep
	for ; i < len(b)/2; i++ {
		if len(v.voice.Sample) < i || int(v.voice.LoopEnd) < i {
			fmt.Println(v.voice.Name)
			i = int(v.voice.LoopStart)
		}

		buf := make([]byte, 2)
		binary.LittleEndian.AppendUint16(buf, uint16(v.voice.Sample[i]))
		b = append(b, buf...)
	}
	v.currentStep = i
	return i * 2, err
}*/

func (v *VoiceReader) Read(b []byte) (n int, err error) {
	maxSamples := len(b) / 2 // 16bit sample

	for i := 0; i < maxSamples; i++ {
		if v.currentStep >= len(v.voice.Sample) || v.currentStep >= int(v.voice.LoopEnd) {
			v.currentStep = int(v.voice.LoopStart)
		}

		sample := uint16(v.voice.Sample[v.currentStep])

		binary.LittleEndian.PutUint16(b[i*2:], sample)

		v.currentStep++
		n += 2
	}

	return n, nil
}
