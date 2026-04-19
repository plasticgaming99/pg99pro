package sf2abst

func PresetToGenerators(index int, sf2 SF2Abst) []Generator {
	pgs := GetPresetGlobalZone(index, sf2)
	global := NewGenerator()
	localStart := index
	if pgs != nil {
		global = PGenToGenerator(nil, pgs)
		localStart = int(sf2.Pdta.Phdr[index].BagIndex)
	}

	rtGen := make([]Generator, 0)

	endLen := 0

	if len(sf2.Pdta.Phdr)-1 == index {
		endLen = len(sf2.Pdta.Pbag) - 1
	} else {
		endLen = int(sf2.Pdta.Phdr[index+1].BagIndex)
	}

	for i := localStart; i < endLen; i++ {
		gi := sf2.Pdta.Pbag[i].GenIndex
		ngi := 0
		if i+1 < len(sf2.Pdta.Pbag) {
			ngi = int(sf2.Pdta.Pbag[i+1].GenIndex)
		} else {
			ngi = len(sf2.Pdta.Pgen)
		}
		rtGen = append(rtGen, PGenToGenerator(&global, sf2.Pdta.Pgen[gi:ngi]))
	}
	return rtGen
}

func InstToGenerators(index int, sf2 SF2Abst) []Generator {
	pgs := GetInstrumentGlobalZone(index, sf2)
	global := NewGenerator()
	localStart := index
	if pgs != nil {
		global = PGenToGenerator(nil, pgs)
		localStart = int(sf2.Pdta.Inst[index].BagIndex)
	}

	rtGen := make([]Generator, 0)

	endLen := 0

	if len(sf2.Pdta.Inst)-1 == index {
		endLen = len(sf2.Pdta.Ibag) - 1
	} else {
		endLen = int(sf2.Pdta.Inst[index+1].BagIndex)
	}

	for i := localStart; i < endLen; i++ {
		gi := sf2.Pdta.Ibag[i].GenIndex
		ngi := 0
		if i+1 < len(sf2.Pdta.Ibag) {
			ngi = int(sf2.Pdta.Ibag[i+1].GenIndex)
		} else {
			ngi = len(sf2.Pdta.Igen)
		}
		rtGen = append(rtGen, PGenToGenerator(&global, sf2.Pdta.Igen[gi:ngi]))
	}
	return rtGen
}

func GetPresetGlobalZone(presetIndex int, sf2 SF2Abst) []Pgen {
	// get bag length for global
	bagStart := int(sf2.Pdta.Phdr[presetIndex].BagIndex)
	bagEnd := 0
	if len(sf2.Pdta.Phdr)-1 == presetIndex {
		bagEnd = len(sf2.Pdta.Phdr) - 1
	} else {
		bagEnd = int(sf2.Pdta.Phdr[presetIndex+1].BagIndex)
	}

	// no bags, don't do anything
	if bagStart >= bagEnd {
		return nil
	}

	genStart := sf2.Pdta.Pbag[bagStart].GenIndex
	genEnd := sf2.Pdta.Pbag[bagStart+1].GenIndex

	if genStart >= genEnd {
		return nil
	}

	lastGen := sf2.Pdta.Pgen[genEnd-1]

	if lastGen.GenOper != Op_instrument {
		return sf2.Pdta.Pgen[genStart:genEnd]
	}

	// no global zone
	return nil
}

func GetInstrumentGlobalZone(presetIndex int, sf2 SF2Abst) []Pgen {
	// get bag length for global
	bagStart := int(sf2.Pdta.Inst[presetIndex].BagIndex)
	bagEnd := 0
	if len(sf2.Pdta.Inst)-1 == presetIndex {
		bagEnd = len(sf2.Pdta.Inst) - 1
	} else {
		bagEnd = int(sf2.Pdta.Inst[presetIndex+1].BagIndex)
	}

	// no bags, don't do anything
	if bagStart >= bagEnd {
		return nil
	}

	genStart := sf2.Pdta.Ibag[bagStart].GenIndex
	genEnd := sf2.Pdta.Ibag[bagStart+1].GenIndex

	if genStart >= genEnd {
		return nil
	}

	lastGen := sf2.Pdta.Igen[genEnd-1]

	if lastGen.GenOper != Op_sampleID {
		return sf2.Pdta.Igen[genStart:genEnd]
	}

	// no global zone
	return nil
}
