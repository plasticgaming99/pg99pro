// this package converts sf2 meaning to pcm-rompler meaning and use them
package pcmsynth

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/plasticgaming99/pg99pro/synth/sf2abst"
	resampler "github.com/tphakala/go-audio-resampler"
)

const sm24scale = 1 / 8388607.

func MergeSm24ToFloat32(smpl int16, sm24 uint8) float32 {
	return float32(int32(smpl)<<8|int32(sm24)) * sm24scale
}

const sm16scale = 1 / 32767.

func SmplToFloat32(smpl int16) float32 {
	return float32(smpl) * sm16scale
}

type VoiceSample struct {
	Name        string
	SampleType  sf2abst.SampleType
	Sample      []float32 // left/mono sample
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
	ResamplerEnabled    bool // resample all voices to specified rate
	ResamplerRate       uint // resample rate, default 48000 hz
	Use16bitSamples     bool // ignore sm24 and use 16bit only
	KeepOriginalOrder   bool // when true, it keeps original index
	UseBytesFromSF2Abst bool // force using smpl chunk from sf2 even sf2File is not nil
}

// VoiceSamples are ordered by SoundFont's Order
func GenerateVoiceSamples(sf2 *sf2abst.SF2Abst, op *GenerateVoicesOptions, sf2File *os.File) (voices []VoiceSample, err error) {
	voices = make([]VoiceSample, 0)
	smpls := make([]sf2abst.Sample, 0, len(sf2.Pdta.Shdr))
	for i := 0; i < len(sf2.Pdta.Shdr); i++ {
		smpls = append(smpls, sf2abst.ShdrToSample(sf2.Pdta.Shdr[i]))
	}
	sf2File.Seek(0, 0)
	offset, size, err := FindSmplToOffset(sf2File)
	fmt.Println("offset", offset, "size", size)

	for i := 0; i < len(smpls); i++ {
		cfg := &resampler.Config{}
		cfg.EnableSIMD = true
		cfg.Quality.Precision = 24
		cfg.Channels = 1
		cfg.InputRate = float64(smpls[i].SampleRate)
		cfg.OutputRate = float64(op.ResamplerRate)

		voice := VoiceSample{
			Name:        smpls[i].Name,
			SampleType:  smpls[i].SampleType,
			Sample:      nil,
			SampleRate:  smpls[i].SampleRate,
			LoopStart:   smpls[i].LoopStart - smpls[i].Start,
			LoopEnd:     smpls[i].LoopEnd - smpls[i].Start,
			OriginalKey: smpls[i].OriginalKey,
			PitchCorr:   smpls[i].PitchCorr,
			ShdrOrigin:  i,
		}

		sample := make([]float32, 0)

		sr := io.NewSectionReader(sf2File, offset+int64(smpls[i].Start*2), int64((smpls[i].End-smpls[i].Start)*2))

		smplr, err := resampler.New(cfg)
		if err != nil {
			return nil, err
		}
		buf := make([]byte, int((smpls[i].End-smpls[i].Start)*2))
		//io.ReadAtLeast(sampleReader, buf, int((smpls[i].End-smpls[i].Start)*2))
		_, e := io.ReadFull(sr, buf)
		if e != nil {
			return nil, e
		}

		if op.ResamplerEnabled {
			buf2 := make([]float32, 0)
			for i := 0; i < len(buf); i += 2 {
				buf2 = append(buf2, SmplToFloat32(int16(binary.LittleEndian.Uint16(buf[i:i+2]))))
			}
			resampled, err := smplr.ProcessFloat32(buf2)
			if err != nil {
				return nil, err
			}
			sample = append(sample, resampled...)
		}
		voice.Sample = sample

		voices = append(voices, voice)
	}

	return voices, nil
}
