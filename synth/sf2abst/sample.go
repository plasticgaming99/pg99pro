// not sample but wave sample

package sf2abst

type SampleType int

const (
	SampleTypeMono        = 1
	SampleTypeStereoRight = 2
	SampleTypeStereoLeft  = 4
	SampleTypeLinked      = 8
)

type Sample struct {
	Name        string
	Start       uint32
	End         uint32
	LoopStart   uint32
	LoopEnd     uint32
	SampleRate  uint32
	OriginalKey uint8
	PitchCorr   int8
	SampleLink  uint16
	SampleType  SampleType
}

func ShdrToSample(s Shdr) Sample {
	return Sample{
		Name:        s.Name,
		Start:       s.Start,
		End:         s.End,
		LoopStart:   s.LoopStart,
		LoopEnd:     s.LoopEnd,
		SampleRate:  s.SampleRate,
		OriginalKey: s.OriginalKey,
		PitchCorr:   s.PitchCorr,
		SampleLink:  s.SampleLink,
		SampleType:  SampleType(s.SampleType),
	}
}
