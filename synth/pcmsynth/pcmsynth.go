package pcmsynth

import (
	"fmt"

	"gitlab.com/gomidi/midi/v2"
)

type PCMSynth struct {
	Channels    []MIDIChannel
	Instruments map[uint32]*Instrument

	evchannel <-chan MidiEv
	buf       []float32

	samplingRate int
	bitDepth     int
}

func NewPCMSynth(insts map[uint32]*Instrument, ev <-chan MidiEv) *PCMSynth {
	ps := &PCMSynth{
		Channels:    make([]MIDIChannel, 16),
		Instruments: insts,

		evchannel: ev,

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

type EventType int

const (
	EvTypeProgramChange = EventType(iota)
	EvTypeControlChange
	EvTypeSysEx
	EvTypeNoteOn
	EvTypeNoteOff
	EvTypePitchBend
)

type MidiEv struct {
	EvType  EventType
	Channel uint8
	Key     uint8
	Vel     uint8
	Prog    uint8
	Ctrl    uint8
	CCval   uint8
	Bendval int16
	SysEx   []byte
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

func (syn *PCMSynth) PartLevel(channel, level int) {
	syn.Channels[channel].PartLevel(level)
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
drain:
	for {
		select {
		case ev := <-syn.evchannel:
			switch ev.EvType {
			case EvTypeProgramChange:
				if ev.Channel == 9 {
					fmt.Println("drum")
					syn.BankSelectMSB(int(ev.Channel), 128)
					syn.BankSelectLSB(int(ev.Channel), 0)
				}
				syn.ProgramChange(int(ev.Channel), ev.Prog)
			case EvTypeNoteOn:
				syn.NoteOn(int(ev.Channel), int(ev.Key), int(ev.Vel))
			case EvTypeNoteOff:
				syn.NoteOff(int(ev.Channel), int(ev.Key))
			case EvTypeControlChange:
				switch ev.Ctrl {
				case midi.BankSelectMSB:
					syn.BankSelectMSB(int(ev.Channel), ev.CCval)
				case midi.BankSelectLSB:
					syn.BankSelectLSB(int(ev.Channel), ev.CCval)
				}
			case EvTypeSysEx:
			}
		default:
			break drain
		}
	}

	tmp := make([]float32, len(f32))
	for i2 := range f32 {
		f32[i2] = 0 // 初期化
	}
	for i := range syn.Channels {
		syn.Channels[i].ReadFloat32(tmp)
		for i2 := range tmp {
			f32[i2] += tmp[i2] / 100
		}
	}
	//fmt.Println("read", len(f32))
	return len(f32), nil
}
