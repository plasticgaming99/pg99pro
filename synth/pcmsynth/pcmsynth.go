package pcmsynth

type PCMSynth struct {
	Channels    []MIDIChannel
	Instruments map[uint32]*Instrument

	buf []float32

	samplingRate int
	bitDepth     int
}

func NewPCMSynth(insts map[uint32]*Instrument) *PCMSynth {
	ps := &PCMSynth{
		Channels:    make([]MIDIChannel, 16),
		Instruments: insts,

		samplingRate: 48000,
		bitDepth:     16,
	}
	for i := range ps.Channels {
		ps.Channels[i] = NewMIDIChannel(ps)
		ps.Channels[i].InitializeChannel()
		if i == 9 {
			ps.Channels[i].BankSelectMSB(128)
			ps.Channels[i].BankSelectLSB(0)
			ps.Channels[i].ProgramChange(0)
		}
	}
	return ps
}

func (syn *PCMSynth) NoteOn(channel, key, vel int) {
	if len(syn.Channels) < channel {
		return
	}
	syn.Channels[channel].NoteOn(key, vel)
}

func (syn *PCMSynth) NoteOff(channel, key int) {
	if len(syn.Channels) < channel {
		return
	}
	syn.Channels[channel].NoteOff(key)
}

func (syn *PCMSynth) BankSelectMSB(channel int, msb uint8) {
	if len(syn.Channels) < channel {
		return
	}
	syn.Channels[channel].BankSelectMSB(msb)
}

func (syn *PCMSynth) BankSelectLSB(channel int, lsb uint8) {
	if len(syn.Channels) < channel {
		return
	}
	syn.Channels[channel].BankSelectLSB(lsb)
}

func (syn *PCMSynth) ProgramChange(channel int, prog uint8) {
	if len(syn.Channels) < channel {
		return
	}
	syn.Channels[channel].ProgramChange(prog)
}

func (syn *PCMSynth) ReadFloat32(f32 []float32) (n int, err error) {
	tmp := make([]float32, len(f32))
	for i := range syn.Channels {
		syn.Channels[i].ReadFloat32(tmp)
		for i2 := range tmp {
			f32[i2] = tmp[i2] / 16
		}
	}
	//fmt.Println("read", len(f32))
	return len(f32), nil
}
