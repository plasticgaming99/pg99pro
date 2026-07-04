// this package parses inst and phdr easier to use

package sf2abst

// sample type does not have samples
func SampleFromSF2Abst(s *SF2Abst) []Sample {
	smpl := make([]Sample, 0, len(s.Pdta.Shdr))

	for i := range s.Pdta.Shdr {
		sample := ShdrToSample(s.Pdta.Shdr[i])
		smpl = append(smpl, sample)
	}

	return smpl
}

type Zone struct {
	KeyMin, KeyMax uint8
	VelMin, VelMax uint8
}

type Instrument struct {
	Generators []Generator
	Zones      []Zone
}

func InstrumentFromSF2Abst(s *SF2Abst) []Instrument {
	inst := make([]Instrument, 0, len(s.Pdta.Inst))

	for i := range s.Pdta.Inst {
		gens := InstToGenerators(i, *s)
		zones := make([]Zone, 0)
		for i := range len(gens) {
			gp := gens[i].ToParam()
			z := Zone{
				KeyMin: gp.Etc.KeyRange.Min,
				KeyMax: gp.Etc.KeyRange.Max,
				VelMin: gp.Etc.VelRange.Min,
				VelMax: gp.Etc.VelRange.Max,
			}
			zones = append(zones, z)
		}
		inst = append(inst, Instrument{Generators: gens, Zones: zones})
	}

	return inst
}

type Preset struct {
	Generators []Generator
	Zones      []Zone
}

func PresetFromSF2Abst(s *SF2Abst) []Preset {
	prst := make([]Preset, 0, len(s.Pdta.Phdr))

	for i := range s.Pdta.Phdr {
		gens := PresetToGenerators(i, *s)
		zones := make([]Zone, 0)
		for i := range len(gens) {
			gp := gens[i].ToParam()
			z := Zone{
				KeyMin: gp.Etc.KeyRange.Min,
				KeyMax: gp.Etc.KeyRange.Max,
				VelMin: gp.Etc.VelRange.Min,
				VelMax: gp.Etc.VelRange.Max,
			}
			zones = append(zones, z)
		}
		prst = append(prst, Preset{Generators: gens, Zones: zones})
	}

	return prst
}
