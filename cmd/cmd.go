package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"slices"
	"time"

	"github.com/ebitengine/oto/v3"
	"github.com/plasticgaming99/pg99pro/gui"
	"github.com/plasticgaming99/pg99pro/synth/pcmsynth"
	"github.com/plasticgaming99/pg99pro/synth/sf2abst"
	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

type eventType int

const (
	evTypeProgramChange = eventType(iota)
	evTypeControlChange
	evTypeSysEx
	evTypeNoteOn
	evTypeNoteOff
	evTypePitchBend
)

type midiEv struct {
	evType  eventType
	channel uint8
	key     uint8
	vel     uint8
	prog    uint8
	ctrl    uint8
	ccval   uint8
	bendval int16
	sysEx   []byte
}

func Execute(args []string) {
	f, err := os.Open(args[0])
	defer f.Close()
	if err != nil {
		log.Fatal(err)
	}
	op := sf2abst.NewParseSF2RawOptions()
	op.ReadSdta = false
	sf2, err := sf2abst.ParseSF2Abst(f, op)
	if err != nil {
		log.Fatal(err)
	}
	PrintlnSF2Bulk(os.Stdout, &sf2)

	i := sf2abst.InstrumentFromSF2Abst(&sf2)

	p := sf2abst.PresetFromSF2Abst(&sf2)

	fmt.Println()
	fmt.Println("sf2 samples    :", len(sf2.Pdta.Shdr))
	fmt.Println("sf2 instruments:", len(i))
	fmt.Println("sf2 presets    :", len(p))
	new, err := os.Open(args[0])
	if err != nil {
		log.Fatal(err)
	}
	gvop := pcmsynth.NewGenerateVoicesOptions()
	gvop.ResamplerEnabled = true
	gvop.Use16bitSamples = true
	gvop.UseBytesFromSF2Abst = false
	gvop.ResamplerRate = 48000
	v, err := pcmsynth.GenerateVoiceSamples(&sf2, gvop, new)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(v))

	runtime.GC()

	if slices.Contains(args[1:], "--gui") {
		gui.Execute()
	}

	otoop := &oto.NewContextOptions{
		SampleRate:   48000,
		ChannelCount: 1,
		Format:       oto.FormatSignedInt16LE,
		BufferSize:   1 * time.Millisecond,
	}
	otoCtx, readyChan, err := oto.NewContext(otoop)
	if err != nil {
		panic("oto.NewContext failed: " + err.Error())
	}
	_ = otoCtx
	<-readyChan

	defer midi.CloseDriver()

	in, err := midi.FindInPort("Midi Through Port-0")

	if err != nil {
		log.Fatal(err)
	}

	insts := pcmsynth.NewInstruments(&sf2, v)
	synth := pcmsynth.NewPCMSynth(insts)

	evQueue := make(chan midiEv)

	const gm, gs, xg = 0, 1, 2
	midimode := gm
	{
		_, err = midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
			var bt []byte
			var ch, key, vel, pg uint8
			var ctrl, val uint8
			var bend int16
			switch {
			case msg.GetProgramChange(&ch, &pg):
				fmt.Printf("got programchange on channel %v with program %v\n", ch, pg)
				//synth.ProgramChange(int(ch), pg)
				evQueue <- midiEv{
					evType:  evTypeProgramChange,
					channel: ch,
					prog:    pg,
				}
				switch midimode {
				case gm:
					//synth.ProgramChange(ch, msb, pg)
				case gs:
					//synth.ProgramChange(ch, msb, pg)
				case xg:
					//synth.ProgramChange(ch, lsb, pg)
				}
			case msg.GetControlChange(&ch, &ctrl, &val):
				fmt.Printf("got %s on channel %v with value %v\n", midi.ControlChangeName[ctrl], ch, val)
				evQueue <- midiEv{
					evType:  evTypeControlChange,
					channel: ch,
					ctrl:    ctrl,
					ccval:   val,
				}
			case msg.GetSysEx(&bt):
				fmt.Printf("got sysex: %X\n", bt)
				btnew := make([]byte, len(bt))
				copy(btnew, bt)

				evQueue <- midiEv{
					evType: evTypeSysEx,
					sysEx:  btnew,
				}
			case msg.GetNoteOn(&ch, &key, &vel):
				if vel != 0 {
					evQueue <- midiEv{
						evType:  evTypeNoteOn,
						channel: ch,
						key:     key,
						vel:     vel,
					}
					fmt.Printf("got noteon %s on channel %v with velocity %v\n", midi.Note(key), ch, vel)
				} else {
					evQueue <- midiEv{
						evType:  evTypeNoteOff,
						channel: ch,
						key:     key,
					}
					fmt.Printf("got noteoff %s on channel %v\n", midi.Note(key), ch)
				}
			case msg.GetNoteOff(&ch, &key, nil):
				evQueue <- midiEv{
					evType:  evTypeNoteOff,
					channel: ch,
					key:     key,
				}
				fmt.Printf("got noteoff %s on channel %v\n", midi.Note(key), ch)
			case msg.GetPitchBend(&ch, &bend, nil):
				evQueue <- midiEv{
					evType:  evTypePitchBend,
					bendval: bend,
				}
				fmt.Printf("got pitchbend on %v and bend %v\n", ch, bend)
			default:
				// ignore
			}
		}, midi.UseSysEx())
	}
	//fmt.Println("start synthesizer")
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)

	fmt.Println("successfly read soundfont")

	ins := synth.Instruments[pcmsynth.PackInst(0, 0, 0)]
	for i := 0; i < len(ins.IVoices); i++ {
		fmt.Printf("ins.IVoices[%v]: %v\n", i, ins.IVoices[i])
	}

	br := pcmsynth.NewFloat32ToByteReader(synth, pcmsynth.BitrateSignedInt16LE, 512)
	otop := otoCtx.NewPlayer(&br)
	otop.SetBufferSize(128)
	otop.Play()

	go func() {
		for ev := range evQueue {
			switch ev.evType {
			case evTypeProgramChange:
				if ev.channel == 9 {
					synth.BankSelectMSB(int(ev.channel), 128)
				}
				synth.ProgramChange(int(ev.channel), ev.prog)
			case evTypeNoteOn:
				synth.NoteOn(int(ev.channel), int(ev.key), int(ev.vel))
			case evTypeNoteOff:
				synth.NoteOff(int(ev.channel), int(ev.key))
			case evTypeControlChange:
				switch ev.ctrl {
				case midi.BankSelectMSB:
					synth.BankSelectMSB(int(ev.channel), ev.ccval)
				case midi.BankSelectLSB:
					synth.BankSelectLSB(int(ev.channel), ev.ccval)
				}
			case evTypeSysEx:
				continue

			}
		}
	}()

	go func() {
		<-sig
		os.Exit(1)
	}()

	for otop.IsPlaying() {
		time.Sleep(time.Millisecond * 500)
	}

	log.Fatal(otop.Err())
}
