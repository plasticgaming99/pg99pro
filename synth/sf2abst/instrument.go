package sf2abst

type Zone struct {
	KeyMin, KeyMax uint8
	VelMin, VelMax uint8
	Voice          *Voice
}

type Instrument struct {
	Zones []Zone
}

func (ins *Instrument) GetVoice(key uint8, vel uint8) *Voice {
	for i := 0; i < len(ins.Zones); i++ {
		if ins.Zones[i].KeyMin < key && key < ins.Zones[i].KeyMax && ins.Zones[i].VelMin < vel && vel < ins.Zones[i].VelMax {
			return ins.Zones[i].Voice
		}
	}
	return nil
}

type Voice struct {
}
