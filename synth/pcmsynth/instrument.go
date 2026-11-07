package pcmsynth

import (
	"fmt"
	"math"

	"github.com/davecgh/go-spew/spew"
	"github.com/plasticgaming99/pg99pro/synth/sf2abst"
)

// might be stereo, or multi-layered
func (ins *Instrument) GetVoice(key, vel int) []*Voice {
	return ins.GetVoicesInto(nil, key, vel)
}

// GetVoicesInto fills a caller-owned slice so MIDI channels can reuse the
// temporary slice used to collect the sample layers for each note.
func (ins *Instrument) GetVoicesInto(dst []*Voice, key, vel int) []*Voice {
	vs := dst[:0]
	if ins == nil {
		return vs
	}
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
			newvoice.playKey = key
			newvoice.playVel = vel
			newvoice.playKeySet = true
			if newvoice.generator.Etc.Keynum >= 0 {
				newvoice.playKey = int(newvoice.generator.Etc.Keynum)
			}
			if newvoice.generator.Etc.Velocity >= 0 {
				newvoice.playVel = int(newvoice.generator.Etc.Velocity)
			}
			rootKey := int(newvoice.voice.OriginalKey)
			// SF2 uses 255 to mean that the sample's root key is unspecified;
			// the format defines MIDI note 60 as the fallback.
			if rootKey > 127 {
				rootKey = 60
			}
			if newvoice.generator.Sample.OverridingRootKey >= 0 {
				rootKey = int(newvoice.generator.Sample.OverridingRootKey)
			}
			// Key distance is in semitones; scale tuning is cents per semitone.
			newvoice.baseStep = float64(newvoice.playKey-rootKey) * float64(newvoice.generator.Etc.ScaleTuning)
			newvoice.baseStep += float64(newvoice.generator.Sample.CoarseTune) * 100
			newvoice.baseStep += float64(newvoice.generator.Sample.FineTune)
			newvoice.baseStep += float64(newvoice.voice.PitchCorr)
			newvoice.baseStep = math.Exp2(newvoice.baseStep / 1200)
			newvoice.updateStep()
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
				sampleID := int(mIgenParams[iii].Etc.SampleID)
				if sampleID < 0 || sampleID >= len(voices) {
					continue
				}
				left, right, stereo := linkedStereoSamples(voices, sampleID)
				if stereo {
					left.StereoSide = -1
					right.StereoSide = 1
					appendInstrumentVoice(&ivs, iv, left, &mIgenParams[iii], s2)
					appendInstrumentVoice(&ivs, iv, right, &mIgenParams[iii], s2)
				} else {
					v = voices[sampleID]
					v.StereoSide = 0
					appendInstrumentVoice(&ivs, iv, v, &mIgenParams[iii], s2)
				}
			}

			pk := PackInst(uint8(s2.Pdta.Phdr[i].Bank), uint8(s2.Pdta.Phdr[i].Bank>>8), uint8(s2.Pdta.Phdr[i].PresetNo))
			in := rt[pk]
			if in == nil {
				in = &Instrument{
					Name:          s2.Pdta.Phdr[i].Name,
					ProgramNumber: int(s2.Pdta.Phdr[i].PresetNo),
				}
				rt[pk] = in
			}
			in.IVoices = append(in.IVoices, ivs...)
		}
	}

	return rt
}

func linkedStereoSamples(voices []VoiceSample, sampleID int) (left, right VoiceSample, ok bool) {
	if sampleID < 0 || sampleID >= len(voices) {
		return VoiceSample{}, VoiceSample{}, false
	}
	sample := voices[sampleID]
	switch sample.SampleType & 0x7fff {
	case sf2abst.SampleTypeStereoLeft:
		partnerID := int(sample.SampleLink)
		if partnerID < 0 || partnerID >= len(voices) ||
			voices[partnerID].SampleType&0x7fff != sf2abst.SampleTypeStereoRight {
			return VoiceSample{}, VoiceSample{}, false
		}
		return sample, voices[partnerID], true
	case sf2abst.SampleTypeStereoRight:
		partnerID := int(sample.SampleLink)
		if partnerID < 0 || partnerID >= len(voices) ||
			voices[partnerID].SampleType&0x7fff != sf2abst.SampleTypeStereoLeft {
			return VoiceSample{}, VoiceSample{}, false
		}
		return voices[partnerID], sample, true
	default:
		return VoiceSample{}, VoiceSample{}, false
	}
}

func appendInstrumentVoice(voices *[]instrumentVoice, base instrumentVoice, sample VoiceSample, generator *sf2abst.GeneratorParam, sf2 *sf2abst.SF2Abst) {
	if sample.ShdrOrigin >= 0 && sample.ShdrOrigin < len(sf2.Pdta.Shdr) {
		applyGeneratorSampleOffsets(&sample, generator.Sample, sf2.Pdta.Shdr[sample.ShdrOrigin].SampleRate)
	}
	base.voice = NewVoice(&sample, generator)
	*voices = append(*voices, base)
}

func applyGeneratorSampleOffsets(v *VoiceSample, p sf2abst.SampleParam, sourceRate uint32) {
	if v == nil || len(v.Sample) == 0 || sourceRate == 0 || v.SampleRate == 0 {
		return
	}
	ratio := float64(v.SampleRate) / float64(sourceRate)
	resampled := func(fine, coarse int16) int {
		return int(math.Round(float64(int64(fine)+int64(coarse)*32768) * ratio))
	}
	start := resampled(p.StartAddrssOffset, p.StartAddrsCoarseOffset)
	endAdjustment := resampled(p.EndAddrsOffset, p.EndAddrsCoarseOffset)
	loopStartAdjustment := resampled(p.StartLoopAddrsOffset, p.StartLoopAddrsCoarseOffset)
	loopEndAdjustment := resampled(p.EndLoopAddrsOffset, p.EndLoopAddrsCoarseOffset)
	oldLength := len(v.Sample)
	if start < 0 {
		start = 0
	}
	if start > oldLength {
		start = oldLength
	}
	end := oldLength + endAdjustment
	if end < start {
		end = start
	}
	if end > oldLength {
		end = oldLength
	}
	oldLoopStart, oldLoopEnd := int(v.LoopStart), int(v.LoopEnd)
	v.Sample = v.Sample[start:end]
	loopStart := oldLoopStart + loopStartAdjustment - start
	loopEnd := oldLoopEnd + loopEndAdjustment - start
	if loopStart < 0 {
		loopStart = 0
	}
	if loopStart > len(v.Sample) {
		loopStart = len(v.Sample)
	}
	if loopEnd < loopStart {
		loopEnd = loopStart
	}
	if loopEnd > len(v.Sample) {
		loopEnd = len(v.Sample)
	}
	v.LoopStart, v.LoopEnd = uint32(loopStart), uint32(loopEnd)
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
