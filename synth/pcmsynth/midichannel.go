package pcmsynth

import (
	"fmt"
	"io"
)

type MIDIChannel struct {
	instrument       *Instrument // current instrument
	voices           []*Voice
	noteVoiceScratch []*Voice
	voiceBufL        []float32
	voiceBufR        []float32
	monoScratchL     []float32
	monoScratchR     []float32
	pcmsynth         *PCMSynth // to get voices, parent
	channelNumber    int
	program          uint8
	level            int // 0-127
	expression       int // 0-127
	drumMode         bool
	defPanpot        int // 0-127
	reverb           int // 0-127
	chorus           int // 0-127
	pitchBend        int

	bankMSB uint8
	bankLSB uint8
}

func NewMIDIChannel(pc *PCMSynth) (syn MIDIChannel) {
	return MIDIChannel{
		voices:           make([]*Voice, 0, 64),
		noteVoiceScratch: make([]*Voice, 0, 16),
		pcmsynth:         pc,
		defPanpot:        64,
		expression:       127,
		reverb:           40,
		chorus:           0,
		bankMSB:          0,
		bankLSB:          0,
	}
}

func (syn *MIDIChannel) InitializeChannel() {
	syn.PartLevel(100)
	syn.BankSelectMSB(0)
	syn.BankSelectLSB(0)
	syn.ProgramChange(0)
}

func (syn *MIDIChannel) NoteOn(key int, vel int) {
	debugNote := syn.pcmsynth.debugNotes || (syn.drumMode && syn.pcmsynth.debugDrums)
	debugTag := ""
	if debugNote {
		debugTag = fmt.Sprintf("[midi channel %d]", syn.channelNumber+1)
		name := "<missing>"
		if syn.instrument != nil {
			name = syn.instrument.Name
		}
		fmt.Printf("%s note-on key=%d velocity=%d bank=%d:%d program=%d instrument=%q\n", debugTag,
			key, vel, syn.bankMSB, syn.bankLSB, syn.program, name)
	}
	v := syn.instrument.GetVoicesInto(syn.noteVoiceScratch, key, vel)
	syn.noteVoiceScratch = v
	if debugNote {
		fmt.Printf("%s matched voices=%d\n", debugTag, len(v))
	}
	for _, voice := range v {
		voice.SetOutputRate(float64(syn.pcmsynth.samplingRate))
		voice.PitchBend(syn.pitchBend)
		voice.SetMonoOutput(syn.pcmsynth.legacyMono)
		voice.SetEnvelopeEnabled(syn.pcmsynth.envelopeEnabled)
		voice.SetPanEnabled(syn.pcmsynth.panEnabled)
		voice.SetInterpolationEnabled(syn.pcmsynth.interpolationEnabled)
		voice.SetMonoUpmix(syn.pcmsynth.monoUpmix)
		voice.SetDrumOneShot(syn.drumMode)
		voice.SetChannelPan((float64(syn.defPanpot) - 64) / 63)
		if debugNote && voice.generator != nil && voice.voice != nil {
			g := voice.generator
			fmt.Printf("%s sample=%q frames=%d rate=%d key-range=%d..%d vel-range=%d..%d modes=%d env(delay/attack/hold/decay/sustain/release)=%.4f/%.4f/%.4f/%.4f/%d/%.4f attenuation=%ddB\n", debugTag,
				voice.voice.Name, len(voice.voice.Sample), voice.voice.SampleRate,
				g.Etc.KeyRange.Min, g.Etc.KeyRange.Max, g.Etc.VelRange.Min, g.Etc.VelRange.Max,
				g.Etc.SampleModes, g.Amp.DelayVolEnv, g.Amp.AttackVolEnv, g.Amp.HoldVolEnv,
				g.Amp.DecayVolEnv, g.Amp.SustainVolEnv, g.Amp.ReleaseVolEnv, g.Amp.InitialAttenuation)
		}
	}
	for _, next := range v {
		if next.generator == nil || next.generator.Etc.ExclusiveClass <= 0 {
			continue
		}
		for _, active := range syn.voices {
			if active.generator != nil && active.generator.Etc.ExclusiveClass == next.generator.Etc.ExclusiveClass {
				active.Stop()
			}
		}
	}
	/*if len(v) > 0 {
		fmt.Println(v[0].voice.Name)
	}*/
	syn.voices = append(syn.voices, v...)
	clear(v)
	syn.noteVoiceScratch = v[:0]
}

func (syn *MIDIChannel) NoteOff(key int) {
	if syn.drumMode {
		return
	}
	for i := range syn.voices {
		if syn.voices[i].key == key {
			syn.voices[i].NoteOff()
		}
	}
}

func (syn *MIDIChannel) PitchBend(pitch int) {
	syn.pitchBend = pitch
	for i := range syn.voices {
		syn.voices[i].PitchBend(pitch)
	}
}

func (syn *MIDIChannel) Expression(exp int) {
	if exp < 0 {
		exp = 0
	}
	if exp > 127 {
		exp = 127
	}
	syn.expression = exp
}

func (syn *MIDIChannel) PartLevel(lvl int) {
	if lvl < 0 {
		lvl = 0
	} else if lvl > 127 {
		lvl = 127
	}
	syn.level = lvl
}

func (syn *MIDIChannel) Volume(value int) {
	syn.PartLevel(value)
}

func (syn *MIDIChannel) BankSelectMSB(msb uint8) {
	syn.bankMSB = msb
}

func (syn *MIDIChannel) BankSelectLSB(lsb uint8) {
	syn.bankLSB = lsb
}

func (syn *MIDIChannel) ProgramChange(program uint8) {
	//fmt.Println("msb:", syn.bankMSB, "lsb:", syn.bankLSB, "prog:", program)
	if syn.drumMode {
		syn.bankMSB, syn.bankLSB = 128, 0
	}
	syn.instrument = GetInstrument(syn.pcmsynth.Instruments, syn.bankMSB, syn.bankLSB, program)
	syn.program = program
	/*for _, v := range syn.instrument.IVoices {
		fmt.Println(v.voice.generator.Etc.KeyRange)
	}*/
}

func (syn *MIDIChannel) SetDrumMode(enabled bool) {
	syn.drumMode = enabled
	if enabled {
		syn.bankMSB, syn.bankLSB = 128, 0
	} else if syn.bankMSB == 128 && syn.bankLSB == 0 {
		syn.bankMSB, syn.bankLSB = 0, 0
	}
}

func (syn *MIDIChannel) SetPanpot(pan int) {
	if pan < 0 {
		pan = 0
	}
	if pan > 127 {
		pan = 127
	}
	syn.defPanpot = pan
	for _, voice := range syn.voices {
		voice.SetChannelPan((float64(pan) - 64) / 63)
	}
}

func (syn *MIDIChannel) ReadFloat32Stereo(left, right []float32) (n int, err error) {
	if len(left) != len(right) {
		return 0, io.ErrShortBuffer
	}
	if cap(syn.voiceBufL) < len(left) {
		syn.voiceBufL = make([]float32, len(left))
		syn.voiceBufR = make([]float32, len(right))
	} else {
		syn.voiceBufL = syn.voiceBufL[:len(left)]
		syn.voiceBufR = syn.voiceBufR[:len(right)]
	}
	level := float32(syn.level) / 127 * (float32(syn.expression) / 127)
	write := 0
	for i := 0; i < len(syn.voices); i++ {
		clear(syn.voiceBufL)
		clear(syn.voiceBufR)
		voice := syn.voices[i]
		n, err := voice.ReadFloat32Stereo(syn.voiceBufL, syn.voiceBufR)
		for i2 := 0; i2 < n; i2++ {
			left[i2] += syn.voiceBufL[i2] * level
			right[i2] += syn.voiceBufR[i2] * level
		}
		if err != io.EOF {
			syn.voices[write] = voice
			write++
		}
	}
	clear(syn.voices[write:])
	syn.voices = syn.voices[:write]
	return len(left), nil
}

func (syn *MIDIChannel) ReadFloat32(f32 []float32) (n int, err error) {
	if cap(syn.monoScratchL) < len(f32) {
		syn.monoScratchL = make([]float32, len(f32))
		syn.monoScratchR = make([]float32, len(f32))
	} else {
		syn.monoScratchL = syn.monoScratchL[:len(f32)]
		syn.monoScratchR = syn.monoScratchR[:len(f32)]
	}
	clear(syn.monoScratchL)
	clear(syn.monoScratchR)
	n, err = syn.ReadFloat32Stereo(syn.monoScratchL, syn.monoScratchR)
	for i := 0; i < n; i++ {
		f32[i] = (syn.monoScratchL[i] + syn.monoScratchR[i]) * 0.5
	}
	clear(f32[n:])
	return n, err
}
