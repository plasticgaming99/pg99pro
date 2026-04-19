// this package converts sf2 meaning to pcm-rompler meaning and use them
package pcmsynth

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/plasticgaming99/pg99pro/synth/sf2abst"
	resampler "github.com/tphakala/go-audio-resampler"
)

func MergeSm24ToFloat32(smpl int16, sm24 uint8) float32 {
	return float32(int32(smpl)<<8 | int32(sm24))
}

func SmplToFloat32(smpl int16) float32 {
	return float32(smpl)
}

type Voice struct {
	Name        string
	SampleType  sf2abst.SampleType
	Sample      []float32 // left/mono sample
	RSample     []float32 // nil slice with mono
	SampleRate  uint32
	RSampleRate *uint32 // if nil, same as samplerate
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

func GenerateVoices(sf2 *sf2abst.SF2Abst, op *GenerateVoicesOptions, sf2File *os.File) (voices []Voice, err error) {
	voices = make([]Voice, 0)
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

		voice := Voice{
			Name:        smpls[i].Name,
			SampleType:  smpls[i].SampleType,
			Sample:      nil,
			RSample:     nil,
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
		n, e := sr.Read(buf)
		if e != nil {
			spew.Dump(e)
		}
		fmt.Println("read", n)

		if op.ResamplerEnabled {
			buf2 := make([]float32, 0)
			for i := 0; i < len(buf); i += 2 {
				buf2 = append(buf2, SmplToFloat32(int16(binary.LittleEndian.Uint16(buf[i:i+2]))))
			}
			fmt.Println("buf2", len(buf2))
			resampled, err := smplr.ProcessFloat32(buf2)
			if err != nil {
				return nil, err
			}
			fmt.Println("resampled", len(resampled))
			sample = append(sample, resampled...)
		}
		voice.Sample = sample

		voices = append(voices, voice)
	}

	return voices, nil
}
