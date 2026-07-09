package sf2abst

import (
	"math"
)

func MapGenerator(g []Pgen) []Generator {
	rtGen := make([]Generator, 0)
	for i := 0; i < len(g); i++ {
	}
	return rtGen
}

func GetGlobalGenerator(g []Pgen, terminator GeneratorId) Generator {
	gen := Generator{}
	for i := 0; i < len(g); i++ {
		if g[i].GenOper == terminator {
			gen = PGenToGenerator(nil, g[:i])
			break
		}
	}
	return gen
}

// init new generator with default value
func NewGenerator() Generator {
	return Generator{
		Sample: GenSample{
			StartAddrssOffset:          0,
			EndAddrsOffset:             0,
			StartLoopAddrsOffset:       0,
			EndLoopAddrsOffset:         0,
			StartLoopAddrsCoarseOffset: 0,
			EndLoopAddrsCoarseOffset:   0,
			CoarseTune:                 0,
			FineTune:                   0,
			OverridingRootKey:          -1,
		},
		Filter: GenFilter{
			EnvToPitch:          0,
			InitialFilterFc:     13500,
			InitialFilterQ:      0,
			EnvToFilterFc:       0,
			DelayModEnv:         -12000,
			AttackModEnv:        -12000,
			HoldModEnv:          -12000,
			DecayModEnv:         -12000,
			SustainModEnv:       0,
			ReleaseModEnv:       -12000,
			KeynumToModEnvHold:  0,
			KeynumToModEnvDecay: 0,
		},
		Effect: GenEffect{
			ChorusEffectsSend: 0,
			ReverbEffectsSend: 0,
		},
		Amp: GenAmp{
			Pan:                 0,
			DelayVolEnv:         -12000,
			AttackVolEnv:        -12000,
			HoldVolEnv:          -12000,
			DecayVolEnv:         -12000,
			SustainVolEnv:       0,
			ReleaseVolEnv:       -12000,
			KeynumToVolEnvHold:  0,
			KeynumToVolEnvDecay: 0,
			InitialAttenuation:  0,
		},
		LFO: GenLFO{
			ModLfoToPitch:    0,
			ModLfoToFilterFc: 0,
			ModLfoToVolume:   0,
			DelayMod:         0,
			FreqMod:          0,
		},
		Etc: GenEtc{
			Instrument:     -1,
			KeyRange:       0x7F00, // 0-127
			VelRange:       0x7F00, // 0-127
			Keynum:         -1,
			Velocity:       -1,
			SampleID:       -1,
			SampleModes:    0,
			ScaleTuning:    100,
			ExclusiveClass: 0,
		},
	}
}

// if no global generator, you can use nil
func PGenToGenerator(global *Generator, p []Pgen) Generator {
	g := NewGenerator()
	if global != nil {
		g = *global
	}
	for i := 0; i < len(p); i++ {
		switch p[i].GenOper {
		case Op_startAddrsOffset:
			g.Sample.StartAddrssOffset += p[i].GenAmount
		case Op_endAddrsOffset:
			g.Sample.EndAddrsOffset += p[i].GenAmount
		case Op_startloopAddrsOffset:
			g.Sample.StartLoopAddrsOffset += p[i].GenAmount
		case Op_endloopAddrsOffset:
			g.Sample.EndLoopAddrsOffset += p[i].GenAmount
		case Op_startAddrsCoarseOffset:
			g.Sample.StartLoopAddrsCoarseOffset += p[i].GenAmount
		case Op_modLfoToPitch:
			g.LFO.ModLfoToPitch += p[i].GenAmount
		case Op_vibLfoToPitch:
			g.Wheel.VibLfoToPitch += p[i].GenAmount
		case Op_modEnvToPitch:
			g.Filter.EnvToPitch += p[i].GenAmount
		case Op_initialFilterFc:
			g.Filter.InitialFilterFc += p[i].GenAmount
		case Op_initialFilterQ:
			g.Filter.InitialFilterQ += p[i].GenAmount
		case Op_modLfoToFilterFc:
			g.LFO.ModLfoToFilterFc += p[i].GenAmount
		case Op_modEnvToFilterFc:
			g.Filter.EnvToFilterFc += p[i].GenAmount
		case Op_endAddrsCoarseOffset:
			g.Sample.EndLoopAddrsCoarseOffset += p[i].GenAmount
		case Op_modLfoToVolume:
			g.LFO.ModLfoToVolume += p[i].GenAmount
		case Op_unused1:
			// unused1 is unused!
		case Op_chorusEffectsSend:
			g.Effect.ChorusEffectsSend += p[i].GenAmount
		case Op_reverbEffectsSend:
			g.Effect.ReverbEffectsSend += p[i].GenAmount
		case Op_pan:
			g.Amp.Pan += p[i].GenAmount
		case Op_unused2:
			// unused 2 is unused!!
		case Op_unused3:
			// unused 3 is unused!!!
		case Op_unused4:
			// unused 4 is unused!!!!
		case Op_delayModLFO:
			g.LFO.DelayMod += p[i].GenAmount
		case Op_freqModLFO:
			g.LFO.FreqMod += p[i].GenAmount
		case Op_delayVibLFO:
			g.Wheel.DelayVibLfo += p[i].GenAmount
		case Op_freqVibLFO:
			g.Wheel.FreqVibLfo += p[i].GenAmount
		case Op_delayModEnv:
			g.Filter.DelayModEnv += p[i].GenAmount
		case Op_attackModEnv:
			g.Filter.AttackModEnv += p[i].GenAmount
		case Op_holdModEnv:
			g.Filter.HoldModEnv += p[i].GenAmount
		case Op_decayModEnv:
			g.Filter.DecayModEnv += p[i].GenAmount
		case Op_sustainModEnv:
			g.Filter.SustainModEnv += p[i].GenAmount
		case Op_releaseModEnv:
			g.Filter.ReleaseModEnv += p[i].GenAmount
		case Op_keynumToModEnvHold:
			g.Filter.KeynumToModEnvHold += p[i].GenAmount
		case Op_keynumToModEnvDecay:
			g.Filter.KeynumToModEnvDecay += p[i].GenAmount
		case Op_delayVolEnv:
			g.Amp.DelayVolEnv += p[i].GenAmount
		case Op_attackVolEnv:
			g.Amp.AttackVolEnv += p[i].GenAmount
		case Op_holdVolEnv:
			g.Amp.HoldVolEnv += p[i].GenAmount
		case Op_decayVolEnv:
			g.Amp.DecayVolEnv += p[i].GenAmount
		case Op_sustainVolEnv:
			g.Amp.SustainVolEnv += p[i].GenAmount
		case Op_releaseVolEnv:
			g.Amp.ReleaseVolEnv += p[i].GenAmount
		case Op_keynumToVolEnvHold:
			g.Amp.KeynumToVolEnvHold += p[i].GenAmount
		case Op_keynumToVolEnvDecay:
			g.Amp.KeynumToVolEnvDecay += p[i].GenAmount
		case Op_instrument:
			g.Etc.Instrument = p[i].GenAmount
		case Op_reserved1:
			// unused
		case Op_keyRange:
			g.Etc.KeyRange = uint16(p[i].GenAmount)
		case Op_velRange:
			g.Etc.VelRange = uint16(p[i].GenAmount)
		case Op_startloopAddrsCoarseOffset:
			g.Sample.StartLoopAddrsCoarseOffset += p[i].GenAmount
		case Op_keynum:
			g.Etc.Keynum = p[i].GenAmount
		case Op_velocity:
			g.Etc.Velocity = p[i].GenAmount
		case Op_initialAttenuation:
			g.Amp.InitialAttenuation += p[i].GenAmount
		case Op_reserved2:
			// unused
		case Op_endloopAddrsCoarseOffset:
			g.Sample.EndLoopAddrsCoarseOffset += p[i].GenAmount
		case Op_coarseTune:
			g.Sample.CoarseTune += p[i].GenAmount
		case Op_fineTune:
			g.Sample.FineTune += p[i].GenAmount
		case Op_sampleID:
			g.Etc.SampleID = p[i].GenAmount
		case Op_sampleModes:
			g.Etc.SampleModes = p[i].GenAmount
		case Op_reserved3:
			// unused
		case Op_scaleTuning:
			g.Etc.ScaleTuning = p[i].GenAmount
		case Op_exclusiveClass:
			g.Etc.ExclusiveClass = p[i].GenAmount
		case Op_overridingRootKey:
			g.Sample.OverridingRootKey = p[i].GenAmount
		case Op_unused5:
			// unused 5 is unused!!!!!
		case Op_endoper:
			// end
		}
	}
	return g
}

func MergeGenerator(pgen Generator, igen Generator) Generator {
	//ik := ParseSFRange(igen.Etc.KeyRange)
	//pk := ParseSFRange(pgen.Etc.KeyRange)
	//fmt.Println("before:", ik, pk)
	//mergedKeyRange := rangeMinMax{Min: max(ik.Min, pk.Min), Max: min(ik.Max, pk.Max)}.ToUint16()
	//fmt.Println("after:", ParseSFRange(mergedKeyRange))
	//iv := ParseSFRange(igen.Etc.VelRange)
	//pv := ParseSFRange(pgen.Etc.VelRange)
	//fmt.Println("before:", iv, pv)
	//mergedVelRange := rangeMinMax{Min: max(iv.Min, pv.Min), Max: min(iv.Max, pv.Max)}.ToUint16()
	//fmt.Println("after:", ParseSFRange(mergedVelRange))

	ret := Generator{
		Sample: GenSample{
			StartAddrssOffset:          pgen.Sample.StartAddrssOffset + igen.Sample.StartAddrssOffset,
			EndAddrsOffset:             pgen.Sample.EndAddrsOffset + igen.Sample.EndAddrsOffset,
			StartLoopAddrsOffset:       pgen.Sample.StartLoopAddrsOffset + igen.Sample.StartLoopAddrsOffset,
			EndLoopAddrsOffset:         pgen.Sample.EndLoopAddrsOffset + pgen.Sample.EndLoopAddrsOffset,
			StartLoopAddrsCoarseOffset: pgen.Sample.StartLoopAddrsCoarseOffset + igen.Sample.StartLoopAddrsCoarseOffset,
			EndLoopAddrsCoarseOffset:   pgen.Sample.EndLoopAddrsCoarseOffset + igen.Sample.EndLoopAddrsCoarseOffset,
			CoarseTune:                 pgen.Sample.CoarseTune + igen.Sample.CoarseTune,
			FineTune:                   pgen.Sample.FineTune + igen.Sample.FineTune,
			OverridingRootKey:          igen.Sample.OverridingRootKey, // override
		},
		Filter: GenFilter{
			EnvToPitch:          pgen.Filter.EnvToPitch + igen.Filter.EnvToPitch,
			InitialFilterFc:     pgen.Filter.InitialFilterFc + igen.Filter.InitialFilterFc,
			InitialFilterQ:      pgen.Filter.InitialFilterQ + igen.Filter.InitialFilterQ,
			EnvToFilterFc:       pgen.Filter.EnvToFilterFc + igen.Filter.EnvToFilterFc,
			DelayModEnv:         pgen.Filter.DelayModEnv + igen.Filter.DelayModEnv,
			AttackModEnv:        pgen.Filter.AttackModEnv + igen.Filter.AttackModEnv,
			HoldModEnv:          pgen.Filter.HoldModEnv + igen.Filter.HoldModEnv,
			DecayModEnv:         pgen.Filter.DecayModEnv + igen.Filter.DecayModEnv,
			SustainModEnv:       pgen.Filter.SustainModEnv + igen.Filter.SustainModEnv,
			ReleaseModEnv:       pgen.Filter.ReleaseModEnv + igen.Filter.ReleaseModEnv,
			KeynumToModEnvHold:  pgen.Filter.KeynumToModEnvHold + pgen.Filter.KeynumToModEnvHold,
			KeynumToModEnvDecay: pgen.Filter.KeynumToModEnvDecay + pgen.Filter.KeynumToModEnvDecay,
		},
		Effect: GenEffect{
			ChorusEffectsSend: pgen.Effect.ChorusEffectsSend + igen.Effect.ChorusEffectsSend,
			ReverbEffectsSend: pgen.Effect.ReverbEffectsSend + igen.Effect.ReverbEffectsSend,
		},
		Wheel: GenWheel{
			VibLfoToPitch: pgen.Wheel.VibLfoToPitch + igen.Wheel.VibLfoToPitch,
			DelayVibLfo:   pgen.Wheel.DelayVibLfo + igen.Wheel.DelayVibLfo,
			FreqVibLfo:    pgen.Wheel.FreqVibLfo + igen.Wheel.FreqVibLfo,
		},
		Amp: GenAmp{
			Pan:                 pgen.Amp.Pan + igen.Amp.Pan,
			DelayVolEnv:         pgen.Amp.DelayVolEnv + igen.Amp.DelayVolEnv,
			AttackVolEnv:        pgen.Amp.AttackVolEnv + igen.Amp.AttackVolEnv,
			HoldVolEnv:          pgen.Amp.HoldVolEnv + igen.Amp.HoldVolEnv,
			DecayVolEnv:         pgen.Amp.DecayVolEnv + igen.Amp.DecayVolEnv,
			SustainVolEnv:       pgen.Amp.SustainVolEnv + igen.Amp.SustainVolEnv,
			ReleaseVolEnv:       pgen.Amp.ReleaseVolEnv + igen.Amp.ReleaseVolEnv,
			KeynumToVolEnvHold:  pgen.Amp.KeynumToVolEnvHold + igen.Amp.KeynumToVolEnvHold,
			KeynumToVolEnvDecay: pgen.Amp.KeynumToVolEnvDecay + igen.Amp.KeynumToVolEnvDecay,
			InitialAttenuation:  pgen.Amp.InitialAttenuation + igen.Amp.InitialAttenuation,
		},
		LFO: GenLFO{
			ModLfoToPitch:    pgen.LFO.ModLfoToPitch + igen.LFO.ModLfoToPitch,
			ModLfoToFilterFc: pgen.LFO.ModLfoToFilterFc + igen.LFO.ModLfoToFilterFc,
			ModLfoToVolume:   pgen.LFO.ModLfoToVolume + igen.LFO.ModLfoToVolume,
			DelayMod:         pgen.LFO.DelayMod + igen.LFO.DelayMod,
			FreqMod:          pgen.LFO.FreqMod + igen.LFO.FreqMod,
		},
		Etc: GenEtc{
			Instrument: pgen.Etc.Instrument,
			KeyRange:/*mergedKeyRange*/ igen.Etc.KeyRange,
			VelRange:/*mergedVelRange*/ igen.Etc.VelRange,
			Keynum:         igen.Etc.Keynum,
			Velocity:       igen.Etc.Velocity,
			SampleID:       igen.Etc.SampleID,
			SampleModes:    igen.Etc.SampleModes,
			ScaleTuning:    igen.Etc.ScaleTuning,
			ExclusiveClass: igen.Etc.ExclusiveClass,
		},
	}
	return ret
}

type Generator struct {
	Sample GenSample
	Filter GenFilter
	Effect GenEffect
	Wheel  GenWheel
	Amp    GenAmp
	LFO    GenLFO
	Etc    GenEtc
}

type GenSample struct {
	StartAddrssOffset          int16 // start offset(sample)
	EndAddrsOffset             int16 // end offset(sample)
	StartLoopAddrsOffset       int16 // loop start offset(sample)
	EndLoopAddrsOffset         int16 // loop end offset(sample)
	StartAddrsCoarseOffset     int16 // (32000 sample)
	EndAddrsCoarseOffset       int16 // (32000 sample)
	StartLoopAddrsCoarseOffset int16 //
	EndLoopAddrsCoarseOffset   int16 //
	CoarseTune                 int16 // half-note tune
	FineTune                   int16 // cent tune
	OverridingRootKey          int16 // root key overriding
}

type GenFilter struct {
	EnvToPitch          int16 // envelope to pitch
	InitialFilterFc     int16 // initial filter frequency cutoff
	InitialFilterQ      int16 // initial filter Q
	EnvToFilterFc       int16 // envelope filter cutoff
	DelayModEnv         int16 // filter/pitch envelope delay
	AttackModEnv        int16 // filter/pitch envelope attack time
	HoldModEnv          int16 // filter/pitch envelope hold time
	DecayModEnv         int16 // filter/pitch envelope decay time
	SustainModEnv       int16 // filter/pitch envelope sustain amount
	ReleaseModEnv       int16 // filter/pitch envelope release time
	KeynumToModEnvHold  int16 // keynum effect to decay hold
	KeynumToModEnvDecay int16 // keynum effect to decay time
}

type GenEffect struct {
	ChorusEffectsSend int16 // chorus level
	ReverbEffectsSend int16 // reverb level
}

type GenWheel struct {
	VibLfoToPitch int16 // wheel effect to pitch
	DelayVibLfo   int16 // delay to start wheel vibration
	FreqVibLfo    int16 // frequency of wheel vibration
}

type GenAmp struct {
	Pan                 int16 // pan
	DelayVolEnv         int16 // envelope delay
	AttackVolEnv        int16 // envelope attack time
	HoldVolEnv          int16 // envelope hold time
	DecayVolEnv         int16 // envelope decay time
	SustainVolEnv       int16 // envelope sustain amount
	ReleaseVolEnv       int16 // envelope release time
	KeynumToVolEnvHold  int16 // key effect to hold time
	KeynumToVolEnvDecay int16 // key effect to decay time
	InitialAttenuation  int16 // volume
}

type GenLFO struct {
	ModLfoToPitch    int16 // pitch modulation
	ModLfoToFilterFc int16 // freq cutoff
	ModLfoToVolume   int16 // volume threshold
	DelayMod         int16 // modulation start
	FreqMod          int16 // modulation frequency
}

type GenEtc struct {
	Instrument     int16  // instrument
	KeyRange       uint16 // key range
	VelRange       uint16 // vel range
	Keynum         int16  // force keynum
	Velocity       int16  // force velocity
	SampleID       int16  // sample id
	SampleModes    int16  // loop
	ScaleTuning    int16  // cent per key++
	ExclusiveClass int16  // one sound per time
}

// Param
type GeneratorParam struct {
	Sample SampleParam
	Filter FilterParam
	Effect EffectParam
	Wheel  WheelParam
	Amp    AmpParam
	LFO    LFOParam
	Etc    EtcParam
}

func (g Generator) ToParam() GeneratorParam {
	gp := GeneratorParam{
		Sample: SampleParam{
			StartAddrssOffset:      g.Sample.StartAddrssOffset,
			EndAddrsOffset:         g.Sample.EndAddrsOffset,
			StartAddrsCoarseOffset: g.Sample.StartAddrsCoarseOffset,
			EndAddrsCoarseOffset:   g.Sample.EndAddrsCoarseOffset,
			StartLoopAddrsOffset:   g.Sample.StartLoopAddrsOffset,
			EndLoopAddrsOffset:     g.Sample.EndLoopAddrsOffset,
			CoarseTune:             float32(g.Sample.CoarseTune) / 10,
			FineTune:               g.Sample.FineTune,
		},
		Filter: FilterParam{
			EnvToPitch:          g.Filter.EnvToPitch / 100,
			InitialFilterFc:     float32(8.176 * math.Exp2(float64(g.Filter.InitialFilterFc)/1200)),
			InitialFilterQ:      int16(math.Round(float64(g.Filter.InitialFilterQ) / 10)),
			EnvToFilterFc:       int16(math.Round(float64(g.Filter.InitialFilterFc) / 100)),
			DelayModEnv:         float32(math.Exp2(float64(g.Filter.InitialFilterFc) / 1200)),
			AttackModEnv:        float32(math.Exp2(float64(g.Filter.AttackModEnv) / 1200)),
			HoldModEnv:          float32(math.Exp2(float64(g.Filter.HoldModEnv) / 1200)),
			DecayModEnv:         float32(math.Exp2(float64(g.Filter.DecayModEnv) / 1200)),
			SustainModEnv:       g.Filter.SustainModEnv / 10,
			ReleaseModEnv:       int16(math.Round(math.Exp2(float64(g.Filter.ReleaseModEnv) / 1200))),
			KeynumToModEnvHold:  int16(math.Round(float64(g.Filter.KeynumToModEnvHold / 100))),
			KeynumToModEnvDecay: int16(math.Round(float64(g.Filter.KeynumToModEnvDecay / 100))),
		},
		Effect: EffectParam{
			ReverbEffectsSend: int16(math.Round(float64(g.Effect.ReverbEffectsSend) / 10)),
			ChorusEffectsSend: int16(math.Round(float64(g.Effect.ChorusEffectsSend) / 10)),
		},
		Wheel: WheelParam{
			VibLfoToPitch: int16(g.Wheel.VibLfoToPitch / 100),
			DelayVibLfo:   float32(math.Exp2(float64(g.Wheel.DelayVibLfo) / 1200)),
			FreqVibLfo:    float32(8.176 * math.Exp2(float64(g.Wheel.FreqVibLfo)/1200)),
		},
		Amp: AmpParam{
			Pan:                 int16(math.Round(float64(g.Amp.Pan) / 10)),
			DelayVolEnv:         float32(math.Exp2(float64(g.Amp.DelayVolEnv) / 1200)),
			AttackVolEnv:        float32(math.Exp2(float64(g.Amp.AttackVolEnv) / 1200)),
			HoldVolEnv:          float32(math.Exp2(float64(g.Amp.HoldVolEnv) / 1200)),
			DecayVolEnv:         float32(math.Exp2(float64(g.Amp.DecayVolEnv) / 1200)),
			SustainVolEnv:       int16(math.Round(float64(g.Amp.SustainVolEnv) / 10)),
			ReleaseVolEnv:       float32(math.Exp2(float64(g.Amp.ReleaseVolEnv) / 1200)),
			KeynumToVolEnvHold:  int16(math.Round(float64(g.Amp.KeynumToVolEnvHold) / 100)),
			KeynumToVolEnvDecay: int16(math.Round(float64(g.Amp.KeynumToVolEnvDecay) / 100)),
			InitialAttenuation:  int16(math.Round(float64(g.Amp.InitialAttenuation) / 10)),
		},
		LFO: LFOParam{
			ModLfoToPitch:    int16(math.Round(float64(g.LFO.ModLfoToPitch) / 100)),
			ModLfoToFilterFc: int16(math.Round(float64(g.LFO.ModLfoToFilterFc) / 100)),
			ModLfoToVolume:   int16(math.Round(float64(g.LFO.ModLfoToVolume) / 10)),
			DelayModLFO:      float32(math.Exp2(float64(g.Amp.DelayVolEnv) / 1200)),
			FreqModLFO:       float32(8.176 * math.Exp2(float64(g.Amp.DelayVolEnv)/1200)),
		},
		Etc: EtcParam{
			Instrument:     g.Etc.Instrument,
			KeyRange:       ParseSFRange(uint16(g.Etc.KeyRange)),
			VelRange:       ParseSFRange(uint16(g.Etc.VelRange)),
			Keynum:         g.Etc.Keynum,
			Velocity:       g.Etc.Velocity,
			SampleID:       g.Etc.SampleID,
			SampleModes:    g.Etc.SampleModes,
			ScaleTuning:    g.Etc.ScaleTuning,
			ExclusiveClass: g.Etc.ExclusiveClass,
		},
	}
	return gp
}

func GeneratorsToParam(g []Generator) []GeneratorParam {
	gp := make([]GeneratorParam, len(g))
	for i := 0; i < len(g); i++ {
		gp[i] = g[i].ToParam()
	}
	return gp
}

type SampleParam struct {
	StartAddrssOffset      int16   // (sample) start offset
	EndAddrsOffset         int16   // (sample) end offset
	StartLoopAddrsOffset   int16   // (32000 sample) loop start offset(sample)
	EndLoopAddrsOffset     int16   // (sample) loop end offset
	StartAddrsCoarseOffset int16   // (32k sample)
	EndAddrsCoarseOffset   int16   // (32k sample)
	CoarseTune             float32 // half-note tune
	FineTune               int16   // (cent) tune
	OverridingRootKey      int16   // root key overriding
}

type FilterParam struct {
	EnvToPitch          int16   // (key) envelope to pitch
	InitialFilterFc     float32 // (Hz) initial filter frequency cutoff
	InitialFilterQ      int16   // (dB) initial filter Q
	EnvToFilterFc       int16   // (key) envelope filter cutoff
	DelayModEnv         float32 // (sec) filter/pitch envelope delay
	AttackModEnv        float32 // (sec) filter/pitch envelope attack time
	HoldModEnv          float32 // (sec) filter/pitch envelope hold time
	DecayModEnv         float32 // (sec) filter/pitch envelope decay time
	SustainModEnv       int16   // (%) filter/pitch envelope sustain amount
	ReleaseModEnv       int16   // (sec) filter/pitch envelope release time
	KeynumToModEnvHold  int16   // (key) keynum effect to decay hold
	KeynumToModEnvDecay int16   // (key) keynum effect to decay time
}

type EffectParam struct {
	ChorusEffectsSend int16 // chorus level
	ReverbEffectsSend int16 // reverb level
}

type WheelParam struct {
	VibLfoToPitch int16   // (semitone) wheel effect to pitch
	DelayVibLfo   float32 // (sec) delay to start wheel vibration
	FreqVibLfo    float32 // (Hz) frequency of wheel vibration
}

type AmpParam struct {
	Pan                 int16   // (?) pan
	DelayVolEnv         float32 // (sec) envelope delay
	AttackVolEnv        float32 // (sec) envelope attack time
	HoldVolEnv          float32 // (sec) envelope hold time
	DecayVolEnv         float32 // (dB) envelope decay time
	SustainVolEnv       int16   // (dB) envelope sustain amount
	ReleaseVolEnv       float32 // (dB) envelope release time
	KeynumToVolEnvHold  int16   // (key) key effect to hold time
	KeynumToVolEnvDecay int16   // (key) key effect to decay time
	InitialAttenuation  int16   // (dB) volume
}

type LFOParam struct {
	ModLfoToPitch    int16   // (semitone) pitch modulation
	ModLfoToFilterFc int16   // (semitone) freq cutoff
	ModLfoToVolume   int16   // (semitone) volume threshold
	DelayModLFO      float32 // (sec) modulation start
	FreqModLFO       float32 // (Hz) modulation frequency
}

type EtcParam struct {
	Instrument     int16    // (instrument id) instrument
	KeyRange       KeyRange // key range
	VelRange       VelRange // vel range
	Keynum         int16    // (gm key) force keynum
	Velocity       int16    // (gm vel) force velocity
	SampleID       int16    // (sample id) sample id
	SampleModes    int16    // (flag) loop
	ScaleTuning    int16    // (key) cent per key++
	ExclusiveClass int16    // one sound per time
}

type rangeMinMax struct {
	Min uint8
	Max uint8
}

type KeyRange = rangeMinMax

type VelRange = rangeMinMax

func ParseSFRange(u uint16) rangeMinMax {
	rt := rangeMinMax{}
	rt.Min = uint8(u & 0x00FF)
	rt.Max = uint8(u >> 8)
	return rt
}

func (r rangeMinMax) ToUint16() uint16 {
	u := uint16(uint16(r.Min) | uint16(r.Max)<<8)
	return u
}
