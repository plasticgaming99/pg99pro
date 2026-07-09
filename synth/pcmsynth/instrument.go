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

			//fmt.Println("-- voiceavailable")
			newvoice := &Voice{}
			*newvoice = *ins.IVoices[i].voice
			newvoice.key = key
			newvoice.vel = vel
			newvoice.baseStep = float64(uint8(key)-newvoice.voice.OriginalKey) * float64(newvoice.generator.Etc.ScaleTuning)
			newvoice.baseStep += float64(newvoice.generator.Sample.CoarseTune) * 100
			newvoice.baseStep += float64(newvoice.generator.Sample.FineTune)
			newvoice.baseStep = math.Exp2(newvoice.baseStep / 1200)
			newvoice.step = newvoice.baseStep
			vs = append(vs, newvoice)
		}
	}
	return vs
}

func calcStep(key, original uint8) float64 {
	semitone := float64(int(key) - int(original))
	pitch := math.Exp2(semitone / 1200.0)
	return pitch
}

func StepToCent(key, original uint8) float64 {
	cents := float64(int(key)-int(original)) * 100
	return cents
}

type Instrument struct {
	Name          string
	ProgramNumber int
	IVoices       []instrumentVoice
}

type instrumentVoice struct {
	keyMin, keyMax uint8
	velMin, velMax uint8
	voice          *Voice
}

func (i *instrumentVoice) Dump() (s string) {
	s += fmt.Sprintln("keyMin:", i.keyMin, "keyMax:", i.keyMax)
	s += fmt.Sprintln("velMin:", i.velMin, "velMax:", i.velMax)
	s += fmt.Sprintln(i.voice.generator.Etc.KeyRange)
	s += fmt.Sprintln(i.voice.generator.Etc.VelRange)
	return
}

func PackInst(msb, lsb, prog uint8) uint32 {
	return uint32(msb)<<16 | uint32(lsb)<<8 | uint32(prog)
}

func UnpackInst(u uint32) (msb, lsb, prog uint8) {
	prog = uint8(u)
	lsb = uint8(u >> 8)
	msb = uint8(u >> 16)
	return
}

// instrument will be packed
func NewInstruments(s2 *sf2abst.SF2Abst, voices []VoiceSample) map[uint32]*Instrument {
	rt := make(map[uint32]*Instrument, 128) // yes preallocate bank 1

	presets := sf2abst.PresetFromSF2Abst(s2)
	insts := sf2abst.InstrumentFromSF2Abst(s2)

	for i := range presets {
		pgens := presets[i].Generators
		fmt.Println("process", i)
		for ii := range pgens {
			if pgens[ii].Etc.Instrument == -1 {
				fmt.Println(" omgg   ---", ii)
				spew.Dump(pgens[ii])
				continue
			}

			/*fmt.Printf("Bank=%d Program=%d\n",
				s2.Pdta.Phdr[i].Bank,
				s2.Pdta.Phdr[i].PresetNo,
			)*/

			igens := insts[pgens[ii].Etc.Instrument].Generators
			mIgenParams := make([]sf2abst.GeneratorParam, 0, len(igens))
			for iii := range igens {
				mg := sf2abst.MergeGenerator(pgens[ii], igens[iii])
				if s2.Pdta.Phdr[i].PresetNo == 0 && s2.Pdta.Phdr[i].Bank == 0 {
					//spew.Dump(pgens[iii].ToParam())
				}
				mIgenParams = append(mIgenParams, mg.ToParam())
				//mIgenParams = append(mIgenParams, igens[iii].ToParam())
			}

			ivs := make([]instrumentVoice, 0, len(mIgenParams))
			for iii := range mIgenParams {
				iv := instrumentVoice{}
				iv.keyMax = mIgenParams[iii].Etc.KeyRange.Max
				iv.keyMin = mIgenParams[iii].Etc.KeyRange.Min
				iv.velMax = mIgenParams[iii].Etc.VelRange.Max
				iv.velMin = mIgenParams[iii].Etc.VelRange.Min
				v := VoiceSample{}
				if mIgenParams[iii].Etc.SampleID == -1 {
					continue
				}
				v = voices[mIgenParams[iii].Etc.SampleID]
				iv.voice = NewVoice(&v, &mIgenParams[iii])
				ivs = append(ivs, iv)
			}

			in := Instrument{
				Name:          s2.Pdta.Phdr[i].Name,
				ProgramNumber: int(s2.Pdta.Phdr[i].PresetNo),
				IVoices:       ivs,
			}
			pk := PackInst(uint8(s2.Pdta.Phdr[i].Bank), uint8(s2.Pdta.Phdr[i].Bank>>8), uint8(s2.Pdta.Phdr[i].PresetNo))
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
