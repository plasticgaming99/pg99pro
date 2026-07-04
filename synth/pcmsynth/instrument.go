package pcmsynth

import (
	"fmt"
	"math"

	"github.com/davecgh/go-spew/spew"
	"github.com/plasticgaming99/pg99pro/synth/sf2abst"
)

// might be stereo, or multi-layered
func (ins *Instrument) GetVoice(key, vel int) []*Voice {
	if ins == nil {
		return nil
	}
	vs := make([]*Voice, 0, 1)
	for i := range ins.IVoices {
		if uint8(ins.IVoices[i].keyMin) <= uint8(key) &&
			uint8(key) <= uint8(ins.IVoices[i].keyMax) &&
			uint8(ins.IVoices[i].velMin) <= uint8(vel) &&
			uint8(vel) <= uint8(ins.IVoices[i].velMax) {
			fmt.Println("-- voiceavailable")
			newvoice := &Voice{}
			*newvoice = *ins.IVoices[i].voice
			newvoice.key = key
			newvoice.vel = vel
			newvoice.baseStep = calcStep(uint8(key), ins.IVoices[i].voice.voice.OriginalKey)
			newvoice.step = newvoice.baseStep
			vs = append(vs, newvoice)
		}
	}
	return vs
}

func calcStep(key, original uint8) float64 {
	semitone := float64(int(key) - int(original))
	pitch := math.Pow(2.0, semitone/12.0)
	return pitch
}

type Instrument struct {
	ProgramNumber int
	IVoices       []instrumentVoice
}

type instrumentVoice struct {
	keyMin, keyMax uint8
	velMin, velMax uint8
	voice          *Voice
}

func PackInst(msb, lsb, prog uint8) uint32 {
	return uint32(msb&0x7F)<<14 | uint32(lsb&0x7F)<<7 | uint32(prog&0x7F)
}

func UnpackInst(u uint32) (msb, lsb, prog uint8) {
	prog = uint8(u & 0x7F)
	lsb = uint8((u >> 7) & 0x7F)
	msb = uint8((u >> 14) & 0x7F)
	return
}

// instrument will be packed
func NewInstruments(s2 *sf2abst.SF2Abst, voices []VoiceSample) map[uint32]*Instrument {
	rt := make(map[uint32]*Instrument, 128) // yes preallocate bank 1

	presets := sf2abst.PresetFromSF2Abst(s2)
	insts := sf2abst.InstrumentFromSF2Abst(s2)

	for i := range presets {
		pgens := presets[i].Generators
		for ii := range pgens {
			if pgens[ii].Etc.Instrument == -1 {
				fmt.Println(" omgg   ---", ii)
				spew.Dump(pgens[ii])
				continue
			}
			igens := insts[pgens[ii].Etc.Instrument].Generators
			mIgenParams := make([]sf2abst.GeneratorParam, 0, len(igens))
			for iii := range igens {
				mg := sf2abst.MergeGenerator(pgens[ii], igens[iii])
				mIgenParams = append(mIgenParams, mg.ToParam())
			}

			ivs := make([]instrumentVoice, 0, len(mIgenParams))
			for iii := range mIgenParams {
				iv := instrumentVoice{}
				iv.keyMax = mIgenParams[iii].Etc.KeyRange.Max
				iv.keyMin = mIgenParams[iii].Etc.KeyRange.Min
				iv.velMax = mIgenParams[iii].Etc.VelRange.Max
				iv.velMin = mIgenParams[iii].Etc.VelRange.Min
				iv.voice = &Voice{
					voice:     &voices[iii],
					generator: &mIgenParams[iii],
				}
				ivs = append(ivs, iv)
			}

			in := Instrument{
				ProgramNumber: int(s2.Pdta.Phdr[i].PresetNo),
				IVoices:       ivs,
			}

			pk := PackInst(0, uint8(s2.Pdta.Phdr[i].Bank), uint8(s2.Pdta.Phdr[i].PresetNo))
			rt[pk] = &in
		}
	}

	return rt
}

func GetInstrument(ins map[uint32]*Instrument, msb, lsb, prog uint8) *Instrument {
	in, ok := ins[PackInst(msb, lsb, prog)]
	if !ok {
		return nil
	}
	return in
}

/*type SynthChannel struct {
	Voices []*Voice // not a pointer to sample
	pitchBend int   // range is -8192 to 8191
	panpot    int   // 0 to 127, -1 to rnd
	reverb    int   // 0 to 127
	chorus    int   // 0 to 127
}*/

// generate note
