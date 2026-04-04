package sf2abst

type GeneratorId = uint16

const (
	Op_startAddrsOffset GeneratorId = iota
	Op_endAddrsOffset
	Op_startloopAddrsOffset
	Op_endloopAddrsOffset
	Op_startAddrsCoarseOffset
	Op_modLfoToPitch
	Op_vibLfoToPitch
	Op_modEnvToPitch
	Op_initialFilterFc
	Op_initialFilterQ
	Op_modLfoToFilterFc
	Op_modEnvToFilterFc
	Op_endAddrsCoarseOffset
	Op_modLfoToVolume
	Op_unused1
	Op_chorusEffectsSend
	Op_reverbEffectsSend
	Op_pan
	Op_unused2
	Op_unused3
	Op_unused4
	Op_delayModLFO
	Op_freqModLFO
	Op_delayVibLFO
	Op_freqVibLFO
	Op_delayModEnv
	Op_attackModEnv
	Op_holdModEnv
	Op_decayModEnv
	Op_sustainModEnv
	Op_releaseModEnv
	Op_keynumToModEnvHold
	Op_keynumToModEnvDecay
	Op_delayVolEnv
	Op_attackVolEnv
	Op_holdVolEnv
	Op_decayVolEnv
	Op_sustainVolEnv
	Op_releaseVolEnv
	Op_keynumToVolEnvHold
	Op_keynumToVolEnvDecay
	Op_instrument
	Op_reserved1
	Op_keyRange
	Op_velRange
	Op_startloopAddrsCoarseOffset
	Op_keynum
	Op_velocity
	Op_initialAttenuation
	Op_reserved2
	Op_endloopAddrsCoarseOffset
	Op_coarseTune
	Op_fineTune
	Op_sampleID
	Op_sampleModes
	Op_reserved3
	Op_scaleTuning
	Op_exclusiveClass
	Op_overridingRootKey
	Op_unused5
	Op_endoper
)
