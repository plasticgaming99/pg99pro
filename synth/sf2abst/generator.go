package sf2abst

import "math"

func MapGenerator() Generator {
	return
}

type Generator struct {
	Sample Sample
	Filter Filter
	Amp    Amp
	LFO    LFO
	Etc    Etc
}

type Sample struct {
	StartAddrssOffset      int16 // start offset(sample)
	EndAddrsOffset         int16 // end offset(sample)
	StartLoopAddrsOffset   int16 // loop start offset(sample)
	EndLoopAddrsOffset     int16 // loop end offset(sample)
	StartAddrsCoarseOffset int16 //
	EndAddrsCoarseOffset   int16 //
	CoarseTune             int16 // half-note tune
	FineTune               int16 // cent tune
	OverridingRootKey      int16 // root key overriding
}

type Filter struct {
	EnvToPitch          int16 // envelope to pitch
	InitialFilterFc     int16 // initial filter frequency cutoff
	InitialFilterQ      int16 // initial filter Q
	DelayModEnv         int16 // filter/pitch envelope delay
	AttackModEnv        int16 // filter/pitch envelope attack time
	HoldModEnv          int16 // filter/pitch envelope hold time
	DecayModEnv         int16 // filter/pitch envelope decay time
	SustainModEnv       int16 // filter/pitch envelope sustain amount
	ReleaseModEnv       int16 // filter/pitch envelope release time
	KeynumToModEnvHold  int16 // keynum effect to decay hold
	KeynumToModEnvDecay int16 // keynum effect to decay time
}

type Amp struct {
	Pan                 int16   // pan
	DelayVolEnv         int16   // envelope delay
	AttackVolEnv        int16   // envelope attack time
	HoldVolEnv          int16   // envelope hold time
	DecayVolEnv         int16   // envelope decay time
	SustainVolEnv       int16   // envelope sustain amount
	ReleaseVolEnv       int16   // envelope release time
	KeynumToVolEnvHold  int16   // key effect to hold time
	KeynumToVolEnvDecay float32 // key effect to decay time
	InitialAttenuation  float32 // volume
}

type LFO struct {
	ModLfoToPitch    int16 // pitch modulation
	ModLfoToFilterFc int16 // freq cutoff
	ModLfoToVolume   int16 // volume threshold
	DelayMod         int16 // modulation start
	FreqMod          int16 // modulation frequency
}

type Etc struct {
	Instrument     int16 // instrument
	KeyRange       int16 // key range
	VelRange       int16 // vel range
	Keynum         int16 // force keynum
	Velocity       int16 // force velocity
	SampleID       int16 // sample id
	SampleModes    int16 // loop
	ScaleTuning    int16 // cent per key++
	ExclusiveClass int16 // one sound per time
}

// Param
type GeneratorParam struct {
	Sample SampleParam
	Filter FilterParam
	Amp    AmpParam
	LFO    LFOParam
	Etc    EtcParam
}

func (g *Generator) ToParam() GeneratorParam {
	gp := GeneratorParam{
		Sample: SampleParam{
			StartAddrssOffset:    g.Sample.StartAddrssOffset,
			EndAddrsOffset:       g.Sample.EndAddrsOffset,
			StartLoopAddrsOffset: g.Sample.StartLoopAddrsOffset,
			EndLoopAddrsOffset:   g.Sample.EndLoopAddrsOffset,
			CoarseTune:           float32(g.Sample.CoarseTune) / 10,
			FineTune:             g.Sample.FineTune,
		},
		Filter: FilterParam{
			EnvToPitch:          g.Filter.EnvToPitch / 100,
			InitialFilterFc:     float32(8.16 * math.Pow(2, float64(g.Filter.InitialFilterFc)/1200)),
			DelayModEnv:         float32(math.Pow(2, float64(g.Filter.InitialFilterFc)/1200)),
			AttackModEnv:        float32(math.Pow(2, float64(g.Filter.AttackModEnv)/1200)),
			HoldModEnv:          float32(math.Pow(2, float64(g.Filter.HoldModEnv)/1200)),
			DecayModEnv:         float32(math.Pow(2, float64(g.Filter.DecayModEnv)/1200)),
			SustainModEnv:       g.Filter.SustainModEnv / 10,
			ReleaseModEnv:       int16(math.Round(math.Pow(2, float64(g.Filter.ReleaseModEnv)/1200))),
			KeynumToModEnvHold:  int16(math.Round(float64(g.Filter.KeynumToModEnvHold / 100))),
			KeynumToModEnvDecay: int16(math.Round(float64(g.Filter.KeynumToModEnvDecay / 100))),
		},
		Amp: AmpParam{
			Pan:                 int16(math.Round(float64(g.Amp.Pan) / 10)),
			DelayVolEnv:         float32(math.Pow(2, float64(g.Amp.DelayVolEnv)/1200)),
			AttackVolEnv:        float32(math.Pow(2, float64(g.Amp.AttackVolEnv)/1200)),
			HoldVolEnv:          float32(math.Pow(2, float64(g.Amp.HoldVolEnv)/1200)),
			DecayVolEnv:         float32(math.Pow(2, float64(g.Amp.DecayVolEnv)/1200)),
			SustainVolEnv:       int16(math.Round(float64(g.Amp.SustainVolEnv) / 10)),
			ReleaseVolEnv:       float32(math.Pow(2, float64(g.Amp.ReleaseVolEnv)/1200)),
			KeynumToVolEnvHold:  int16(math.Round(float64(g.Amp.KeynumToVolEnvHold) / 100)),
			KeynumToVolEnvDecay: int16(math.Round(float64(g.Amp.KeynumToVolEnvDecay) / 100)),
			InitialAttenuation:  int16(math.Round(float64(g.Amp.InitialAttenuation) / 10)),
		},
		LFO: LFOParam{
			ModLfoToPitch:    int16(math.Round(float64(g.LFO.ModLfoToPitch) / 100)),
			ModLfoToFilterFc: int16(math.Round(float64(g.LFO.ModLfoToFilterFc) / 100)),
			ModLfoToVolume:   int16(math.Round(float64(g.LFO.ModLfoToVolume) / 10)),
			DelayModLFO:      float32(math.Pow(2, float64(g.Amp.DelayVolEnv)/1200)),
			FreqModLFO:       float32(math.Pow(2, float64(g.Amp.DelayVolEnv)/1200)),
		},
		Etc: EtcParam{
			Instrument:     g.Etc.Instrument,
			KeyRange:       g.Etc.KeyRange,
			VelRange:       g.Etc.VelRange,
			Keynum:         g.Etc.Keynum,
			Velocity:       g.Etc.Velocity,
			SampleID:       g.Etc.SampleID,
			SampleModes:    g.Etc.SampleModes,
			ExclusiveClass: g.Etc.ExclusiveClass,
		},
	}
	return gp
}

type SampleParam struct {
	StartAddrssOffset      int16   // start offset(sample)
	EndAddrsOffset         int16   // end offset(sample)
	StartLoopAddrsOffset   int16   // (32000 sample) loop start offset(sample)
	EndLoopAddrsOffset     int16   // loop end offset(sample)
	StartAddrsCoarseOffset int16   //
	EndAddrsCoarseOffset   int16   //
	CoarseTune             float32 // half-note tune
	FineTune               int16   // cent tune
	OverridingRootKey      int16   // root key overriding
}

type FilterParam struct {
	EnvToPitch          int16   // (key) envelope to pitch
	InitialFilterFc     float32 // (Hz) initial filter frequency cutoff
	InitialFilterQ      int16   // (dB) initial filter Q
	DelayModEnv         float32 // (sec) filter/pitch envelope delay
	AttackModEnv        float32 // (sec) filter/pitch envelope attack time
	HoldModEnv          float32 // (sec) filter/pitch envelope hold time
	DecayModEnv         float32 // (sec) filter/pitch envelope decay time
	SustainModEnv       int16   // (%) filter/pitch envelope sustain amount
	ReleaseModEnv       int16   // (sec) filter/pitch envelope release time
	KeynumToModEnvHold  int16   // (key) keynum effect to decay hold
	KeynumToModEnvDecay int16   // (key) keynum effect to decay time
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
	ModLfoToPitch    int16   // (key) pitch modulation
	ModLfoToFilterFc int16   // (key) freq cutoff
	ModLfoToVolume   int16   // (key) volume threshold
	DelayModLFO      float32 // (sec) modulation start
	FreqModLFO       float32 // (Hz) modulation frequency
}

type EtcParam struct {
	Instrument     int16 // (instrument id) instrument
	KeyRange       int16 // key range
	VelRange       int16 // vel range
	Keynum         int16 // (gm key) force keynum
	Velocity       int16 // (gm vel) force velocity
	SampleID       int16 // (sample id) sample id
	SampleModes    int16 // (flag) loop
	ScaleTuning    int16 // (key) cent per key++
	ExclusiveClass int16 // one sound per time
}
