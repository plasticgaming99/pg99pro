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

	evQueue := make(chan pcmsynth.MidiEv, 1024)

	insts := pcmsynth.NewInstruments(&sf2, v)
	synth := pcmsynth.NewPCMSynth(insts, evQueue)

	/*for v, ins := range insts {
		fmt.Println(ins.ProgramNumber)
		fmt.Println(pcmsynth.UnpackInst(v))
	}*/

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
				evQueue <- pcmsynth.MidiEv{
					EvType:  pcmsynth.EvTypeProgramChange,
					Channel: ch,
					Prog:    pg,
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
				evQueue <- pcmsynth.MidiEv{
					EvType:  pcmsynth.EvTypeControlChange,
					Channel: ch,
					Ctrl:    ctrl,
					CCval:   val,
				}
			case msg.GetSysEx(&bt):
				fmt.Printf("got sysex: %X\n", bt)
				btnew := make([]byte, len(bt))
				copy(btnew, bt)

				evQueue <- pcmsynth.MidiEv{
					EvType: pcmsynth.EvTypeSysEx,
					SysEx:  btnew,
				}
			case msg.GetNoteOn(&ch, &key, &vel):
				if vel != 0 {
					evQueue <- pcmsynth.MidiEv{
						EvType:  pcmsynth.EvTypeNoteOn,
						Channel: ch,
						Key:     key,
						Vel:     vel,
					}
					fmt.Printf("got noteon %s (key %v) on channel %v with velocity %v\n", midi.Note(key), key, ch, vel)
				} else {
					evQueue <- pcmsynth.MidiEv{
						EvType:  pcmsynth.EvTypeNoteOff,
						Channel: ch,
						Key:     key,
					}
					fmt.Printf("got noteoff %s (key %v) on channel %v\n", midi.Note(key), key, ch)
				}
			case msg.GetNoteOff(&ch, &key, nil):
				evQueue <- pcmsynth.MidiEv{
					EvType:  pcmsynth.EvTypeNoteOff,
					Channel: ch,
					Key:     key,
				}
				fmt.Printf("got noteoff %s (key %v) on channel %v\n", midi.Note(key), key, ch)
			case msg.GetPitchBend(&ch, &bend, nil):
				evQueue <- pcmsynth.MidiEv{
					EvType:  pcmsynth.EvTypePitchBend,
					Bendval: bend,
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

	//ins := synth.Instruments[pcmsynth.PackInst(0, 0, 0)]
	//fmt.Println(len(ins.IVoices))
	//fmt.Println(ins.ProgramNumber)
	/*for i := 0; i < len(ins.IVoices); i++ {
		fmt.Print(ins.IVoices[i].Dump())
	}*/

	br := pcmsynth.NewFloat32ToByteReader(synth, pcmsynth.BitrateSignedInt16LE, 512)
	otop := otoCtx.NewPlayer(&br)
	otop.SetBufferSize(128)
	otop.Play()

	go func() {
		<-sig
		os.Exit(1)
	}()

	for otop.IsPlaying() {
		time.Sleep(time.Millisecond * 500)
	}

	log.Fatal(otop.Err())
}
