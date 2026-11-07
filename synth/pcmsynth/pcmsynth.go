package pcmsynth

import (
	"io"
	"math"

	"gitlab.com/gomidi/midi/v2"
)

type PCMSynth struct {
	Channels    []MIDIChannel
	Instruments map[uint32]*Instrument

	evchannel            <-chan MidiEv
	buf                  []float32
	bufferL              []float32
	bufferR              []float32
	monoScratch          []float32
	peakEnvelope         float32
	outputChannels       int
	legacyMono           bool
	envelopeEnabled      bool
	panEnabled           bool
	interpolationEnabled bool
	monoUpmix            bool
	debugDrums           bool
	debugNotes           bool

	samplingRate int
	bitDepth     int
}

func NewPCMSynth(insts map[uint32]*Instrument, ev <-chan MidiEv) *PCMSynth {
	return NewPCMSynthWithChannels(insts, ev, 2)
}

func NewPCMSynthWithChannels(insts map[uint32]*Instrument, ev <-chan MidiEv, channels int) *PCMSynth {
	if channels != 1 {
		channels = 2
	}
	ps := &PCMSynth{
		Channels:    make([]MIDIChannel, 16),
		Instruments: insts,

		evchannel: ev,

		samplingRate:         48000,
		bitDepth:             16,
		outputChannels:       channels,
		legacyMono:           channels == 1,
		envelopeEnabled:      true,
		panEnabled:           true,
		interpolationEnabled: true,
	}
	for i := range ps.Channels {
		ps.Channels[i] = NewMIDIChannel(ps)
		ps.Channels[i].channelNumber = i
		ps.Channels[i].SetDrumMode(i == 9)
		ps.Channels[i].InitializeChannel()
		if i == 9 {
			ps.Channels[i].BankSelectMSB(128)
			ps.Channels[i].BankSelectLSB(0)
			ps.Channels[i].ProgramChange(0)
		}
	}
	return ps
}

func (syn *PCMSynth) SetDrumDebug(enabled bool) {
	syn.debugDrums = enabled
}

func (syn *PCMSynth) SetNoteDebug(enabled bool) {
	syn.debugNotes = enabled
}

// SetDrumChannel toggles percussion-bank behavior independently for each MIDI channel.
func (syn *PCMSynth) SetDrumChannel(channel int, enabled bool) {
	if channel < 0 || channel >= len(syn.Channels) {
		return
	}
	syn.Channels[channel].SetDrumMode(enabled)
}

// SetLegacyMonoOutput selects the previous mono voice path while keeping the
// requested device channel count. Disable it to render stereo voice processing
// and downmix the result to a mono device.
func (syn *PCMSynth) SetLegacyMonoOutput(enabled bool) {
	syn.legacyMono = enabled
	for i := range syn.Channels {
		for _, voice := range syn.Channels[i].voices {
			voice.SetMonoOutput(enabled)
		}
	}
}

// SetEnvelopeEnabled toggles the SoundFont volume envelope for current and
// future voices. Disabled envelopes retain the legacy note-off fade.
func (syn *PCMSynth) SetEnvelopeEnabled(enabled bool) {
	syn.envelopeEnabled = enabled
	for i := range syn.Channels {
		for _, voice := range syn.Channels[i].voices {
			voice.SetEnvelopeEnabled(enabled)
		}
	}
}

// SetPanEnabled toggles generator and MIDI channel pan for current and future voices.
func (syn *PCMSynth) SetPanEnabled(enabled bool) {
	syn.panEnabled = enabled
	for i := range syn.Channels {
		for _, voice := range syn.Channels[i].voices {
			voice.SetPanEnabled(enabled)
		}
	}
}

// SetInterpolationEnabled toggles linear sample interpolation for current and future voices.
func (syn *PCMSynth) SetInterpolationEnabled(enabled bool) {
	syn.interpolationEnabled = enabled
	for i := range syn.Channels {
		for _, voice := range syn.Channels[i].voices {
			voice.SetInterpolationEnabled(enabled)
		}
	}
}

// SetMonoUpmix duplicates mono samples to both output channels without applying pan.
func (syn *PCMSynth) SetMonoUpmix(enabled bool) {
	syn.monoUpmix = enabled
	for i := range syn.Channels {
		for _, voice := range syn.Channels[i].voices {
			voice.SetMonoUpmix(enabled)
		}
	}
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
	if channel < 0 || channel >= len(syn.Channels) {
		return
	}
	if vel <= 0 {
		syn.NoteOff(channel, key)
		return
	}
	syn.Channels[channel].NoteOn(key, vel)
}

func (syn *PCMSynth) NoteOff(channel, key int) {
	if channel < 0 || channel >= len(syn.Channels) || syn.Channels[channel].drumMode {
		return
	}
	syn.Channels[channel].NoteOff(key)
}

func (syn *PCMSynth) PartLevel(channel, level int) {
	if channel < 0 || channel >= len(syn.Channels) {
		return
	}
	syn.Channels[channel].PartLevel(level)
}

func (syn *PCMSynth) BankSelectMSB(channel int, msb uint8) {
	if channel < 0 || channel >= len(syn.Channels) {
		return
	}
	syn.Channels[channel].BankSelectMSB(msb)
}

func (syn *PCMSynth) BankSelectLSB(channel int, lsb uint8) {
	if channel < 0 || channel >= len(syn.Channels) {
		return
	}
	syn.Channels[channel].BankSelectLSB(lsb)
}

func (syn *PCMSynth) ProgramChange(channel int, prog uint8) {
	if channel < 0 || channel >= len(syn.Channels) {
		return
	}
	syn.Channels[channel].ProgramChange(prog)
}

func (syn *PCMSynth) ReadFloat32(f32 []float32) (n int, err error) {
	if syn.outputChannels == 1 {
		return syn.readMono(f32)
	}
	if len(f32)%2 != 0 {
		return 0, io.ErrShortBuffer
	}
drain:
	for {
		select {
		case ev := <-syn.evchannel:
			switch ev.EvType {
			case EvTypeProgramChange:
				syn.ProgramChange(int(ev.Channel), ev.Prog)
			case EvTypeNoteOn:
				syn.NoteOn(int(ev.Channel), int(ev.Key), int(ev.Vel))
			case EvTypeNoteOff:
				syn.NoteOff(int(ev.Channel), int(ev.Key))
			case EvTypePitchBend:
				if int(ev.Channel) < len(syn.Channels) {
					syn.Channels[ev.Channel].PitchBend(int(ev.Bendval))
				}
			case EvTypeControlChange:
				switch ev.Ctrl {
				case midi.BankSelectMSB:
					syn.BankSelectMSB(int(ev.Channel), ev.CCval)
				case midi.BankSelectLSB:
					syn.BankSelectLSB(int(ev.Channel), ev.CCval)
				case 7: // MIDI channel volume
					if int(ev.Channel) < len(syn.Channels) {
						syn.Channels[ev.Channel].Volume(int(ev.CCval))
					}
				case 11: // MIDI expression
					if int(ev.Channel) < len(syn.Channels) {
						syn.Channels[ev.Channel].Expression(int(ev.CCval))
					}
				case 10: // MIDI pan controller
					if int(ev.Channel) < len(syn.Channels) {
						syn.Channels[ev.Channel].SetPanpot(int(ev.CCval))
					}
				}
			case EvTypeSysEx:
			}
		default:
			break drain
		}
	}

	frames := len(f32) / 2
	if cap(syn.bufferL) < frames {
		syn.bufferL = make([]float32, frames)
		syn.bufferR = make([]float32, frames)
	} else {
		syn.bufferL = syn.bufferL[:frames]
		syn.bufferR = syn.bufferR[:frames]
		clear(syn.bufferL)
		clear(syn.bufferR)
	}
	for i := range syn.Channels {
		syn.Channels[i].ReadFloat32Stereo(syn.bufferL, syn.bufferR)
	}
	// Keep the existing make-up gain, but detect the linked stereo peak sample by
	// sample. Smoothing both attack and release avoids gain jumps at audio-block
	// boundaries when many notes overlap.
	const preGain = float32(1.5)
	const peakCeiling = float32(0.85)
	attack := float32(math.Exp(-1 / (float64(syn.samplingRate) * 0.005)))
	release := float32(math.Exp(-1 / (float64(syn.samplingRate) * 0.15)))
	for i := 0; i < frames; i++ {
		peak := max(float32(math.Abs(float64(syn.bufferL[i]))), float32(math.Abs(float64(syn.bufferR[i])))) * preGain
		coefficient := release
		if peak > syn.peakEnvelope {
			coefficient = attack
		}
		syn.peakEnvelope = peak + (syn.peakEnvelope-peak)*coefficient
		gain := float32(1)
		if syn.peakEnvelope > peakCeiling {
			gain = peakCeiling / syn.peakEnvelope
		}
		f32[i*2] = softLimit(syn.bufferL[i] * preGain * gain)
		f32[i*2+1] = softLimit(syn.bufferR[i] * preGain * gain)
	}
	//fmt.Println("read", len(f32))
	return len(f32), nil
}

func (syn *PCMSynth) readMono(f32 []float32) (n int, err error) {
	// Handle MIDI events through the common event-drain path by rendering one
	// interleaved stereo frame buffer, then downmixing using the legacy mono law.
	if cap(syn.monoScratch) < len(f32)*2 {
		syn.monoScratch = make([]float32, len(f32)*2)
	} else {
		syn.monoScratch = syn.monoScratch[:len(f32)*2]
	}
	oldChannels := syn.outputChannels
	syn.outputChannels = 2
	n, err = syn.ReadFloat32(syn.monoScratch)
	syn.outputChannels = oldChannels
	for i := range f32 {
		f32[i] = clampSample((syn.monoScratch[i*2] + syn.monoScratch[i*2+1]) * 0.5)
	}
	return len(f32), err
}

func clampSample(sample float32) float32 {
	if math.IsNaN(float64(sample)) || math.IsInf(float64(sample), 0) {
		return 0
	}
	if sample > 1 {
		return 1
	}
	if sample < -1 {
		return -1
	}
	return sample
}

// softLimit is a final safety net while the smoothed peak detector catches a
// sudden transient. The curve is continuous at the knee and approaches the
// ceiling asymptotically, avoiding hard-clipped samples.
func softLimit(sample float32) float32 {
	if math.IsNaN(float64(sample)) || math.IsInf(float64(sample), 0) {
		return 0
	}
	const knee = float32(0.85)
	const ceiling = float32(0.98)
	magnitude := float32(math.Abs(float64(sample)))
	if magnitude <= knee {
		return sample
	}
	limited := ceiling - (ceiling-knee)*float32(math.Exp(-float64(magnitude-knee)/float64(ceiling-knee)))
	if sample < 0 {
		return -limited
	}
	return limited
}
