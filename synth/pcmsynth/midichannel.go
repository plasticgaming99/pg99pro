package pcmsynth

import (
	"io"
	"slices"
)

type MIDIChannel struct {
	instrument *Instrument // current instrument
	voices     []*Voice
	pcmsynth   *PCMSynth // to get voices, parent
	level      int       // 0-127
	defPanpot  int       // 0-127
	reverb     int       // 0-127
	chorus     int       // 0-127

	bankMSB uint8
	bankLSB uint8
}

func NewMIDIChannel(pc *PCMSynth) (syn MIDIChannel) {
	return MIDIChannel{
		voices:    make([]*Voice, 0),
		pcmsynth:  pc,
		defPanpot: 63,
		reverb:    40,
		chorus:    0,
		bankMSB:   0,
		bankLSB:   0,
	}
}

func (syn *MIDIChannel) InitializeChannel() {
	syn.PartLevel(100)
	syn.BankSelectMSB(0)
	syn.BankSelectLSB(0)
	syn.ProgramChange(0)
}

func (syn *MIDIChannel) NoteOn(key int, vel int) {
	v := syn.instrument.GetVoice(key, vel)
	/*if len(v) > 0 {
		fmt.Println(v[0].voice.Name)
	}*/
	syn.voices = append(syn.voices, v...)
}

func (syn *MIDIChannel) NoteOff(key int) {
	for i := range syn.voices {
		if syn.voices[i].key == key {
			syn.voices[i].NoteOff()
		}
	}
}

func (syn *MIDIChannel) PitchBend(pitch int) {
	for i := range syn.voices {
		syn.voices[i].PitchBend(pitch)
	}
}

func (syn *MIDIChannel) Expression(exp int) {

}

func (syn *MIDIChannel) PartLevel(lvl int) {
	if lvl < 0 && 127 < lvl {
		return
	}
	syn.level = lvl
}

func (syn *MIDIChannel) BankSelectMSB(msb uint8) {
	syn.bankMSB = msb
}

func (syn *MIDIChannel) BankSelectLSB(lsb uint8) {
	syn.bankLSB = lsb
}

func (syn *MIDIChannel) ProgramChange(program uint8) {
	//fmt.Println("msb:", syn.bankMSB, "lsb:", syn.bankLSB, "prog:", program)
	syn.instrument = GetInstrument(syn.pcmsynth.Instruments, syn.bankMSB, syn.bankLSB, program)
	/*for _, v := range syn.instrument.IVoices {
		fmt.Println(v.voice.generator.Etc.KeyRange)
	}*/
}

func (syn *MIDIChannel) ReadFloat32(f32 []float32) (n int, err error) {
	delvoice := make([]int, 0)
	for i := range syn.voices {
		tmp := make([]float32, len(f32))
		_, err := syn.voices[i].ReadFloat32(tmp)
		if err == io.EOF {
			delvoice = append(delvoice, i)
			continue
		}
		for i2 := range tmp {
			f32[i2] += tmp[i2] * (float32(syn.level) / 127)
		}
	}
	for i, vi := range delvoice {
		syn.voices = slices.Delete(syn.voices, vi-i, vi-i+1)
	}
	return len(f32), nil
}
