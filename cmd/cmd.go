package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"slices"
	"strconv"
	"strings"
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

	legacyMono := slices.Contains(args[1:], "--mono")
	verbose := slices.Contains(args[1:], "-v") || slices.Contains(args[1:], "--verbose")
	monoDownmix := slices.Contains(args[1:], "--mono-downmix")
	if legacyMono && monoDownmix {
		log.Fatal("choose either --mono or --mono-downmix")
	}
	monoOutput := legacyMono || monoDownmix
	monoUpmix := slices.Contains(args[1:], "--mono-upmix")
	if monoUpmix && monoOutput {
		log.Fatal("--mono-upmix requires stereo output; do not combine it with --mono or --mono-downmix")
	}
	envelopeEnabled := !slices.Contains(args[1:], "--no-envelope")
	panEnabled := !slices.Contains(args[1:], "--no-pan")
	interpolationEnabled := !slices.Contains(args[1:], "--nearest-sample")
	drumDebug := slices.Contains(args[1:], "--debug-drums")
	noteDebug := slices.Contains(args[1:], "--debug-notes")
	bufferProfile, bufferFrames := parseBufferProfile(args[1:])
	channelCount := 2
	if monoOutput {
		channelCount = 1
	}
	bufferDuration := time.Duration(bufferFrames) * time.Second / 48000
	otoop := &oto.NewContextOptions{
		SampleRate:   48000,
		ChannelCount: channelCount,
		Format:       oto.FormatSignedInt16LE,
		BufferSize:   bufferDuration,
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
	synth := pcmsynth.NewPCMSynthWithChannels(insts, evQueue, channelCount)
	// --mono-upmix deliberately uses the old mono voice renderer, then sends
	// that same signal to both Oto channels to isolate the stereo output path.
	synth.SetLegacyMonoOutput(legacyMono || monoUpmix)
	synth.SetEnvelopeEnabled(envelopeEnabled)
	synth.SetPanEnabled(panEnabled)
	synth.SetInterpolationEnabled(interpolationEnabled)
	synth.SetMonoUpmix(monoUpmix)
	synth.SetDrumDebug(drumDebug)
	synth.SetNoteDebug(noteDebug)
	configureDrumChannels(args[1:], synth)
	fmt.Printf("audio path: output=%s envelope=%t pan=%t interpolation=%t mono-upmix=%t legacy-mono=%t\n",
		map[bool]string{true: "mono", false: "stereo"}[monoOutput], envelopeEnabled, panEnabled, interpolationEnabled, monoUpmix, legacyMono)

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
				if verbose {
					fmt.Printf("got programchange on channel %v with program %v\n", ch, pg)
				}
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
				if verbose {
					fmt.Printf("got %s on channel %v with value %v\n", midi.ControlChangeName[ctrl], ch, val)
				}
				evQueue <- pcmsynth.MidiEv{
					EvType:  pcmsynth.EvTypeControlChange,
					Channel: ch,
					Ctrl:    ctrl,
					CCval:   val,
				}
			case msg.GetSysEx(&bt):
				if verbose {
					fmt.Printf("got sysex: %X\n", bt)
				}
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
					if verbose {
						fmt.Printf("got noteon %s (key %v) on channel %v with velocity %v\n", midi.Note(key), key, ch, vel)
					}
				} else {
					evQueue <- pcmsynth.MidiEv{
						EvType:  pcmsynth.EvTypeNoteOff,
						Channel: ch,
						Key:     key,
					}
					if verbose {
						fmt.Printf("got noteoff %s (key %v) on channel %v\n", midi.Note(key), key, ch)
					}
				}
			case msg.GetNoteOff(&ch, &key, nil):
				evQueue <- pcmsynth.MidiEv{
					EvType:  pcmsynth.EvTypeNoteOff,
					Channel: ch,
					Key:     key,
				}
				if verbose {
					fmt.Printf("got noteoff %s (key %v) on channel %v\n", midi.Note(key), key, ch)
				}
			case msg.GetPitchBend(&ch, &bend, nil):
				evQueue <- pcmsynth.MidiEv{
					EvType:  pcmsynth.EvTypePitchBend,
					Channel: ch,
					Bendval: bend,
				}
				if verbose {
					fmt.Printf("got pitchbend on %v and bend %v\n", ch, bend)
				}
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

	br := pcmsynth.NewFloat32ToByteReader(synth, pcmsynth.BitrateSignedInt16LE, uint(bufferFrames*channelCount))
	otop := otoCtx.NewPlayer(&br)
	playerBufferBytes := bufferFrames * channelCount * 2 // int16 bytes per channel frame
	otop.SetBufferSize(playerBufferBytes)
	fmt.Printf("audio buffer: profile=%s frames=%d bytes=%d duration=%s\n",
		bufferProfile, bufferFrames, playerBufferBytes, bufferDuration)
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

func parseBufferProfile(args []string) (string, int) {
	const prefix = "--buffer-size="
	profile := "mid"
	for _, arg := range args {
		if strings.HasPrefix(arg, prefix) {
			profile = strings.TrimPrefix(arg, prefix)
		}
	}
	framesByProfile := map[string]int{
		"min": 128,
		"mid": 256,
		"max": 512,
	}
	frames, ok := framesByProfile[profile]
	if !ok {
		log.Fatalf("invalid --buffer-size %q; choose min, mid, or max", profile)
	}
	return profile, frames
}

func configureDrumChannels(args []string, synth *pcmsynth.PCMSynth) {
	const prefix = "--drum-channels="
	var configured string
	for _, arg := range args {
		if strings.HasPrefix(arg, prefix) {
			configured = strings.TrimPrefix(arg, prefix)
		}
	}
	if configured == "" {
		return // default: MIDI channel 10
	}
	for channel := 0; channel < 16; channel++ {
		synth.SetDrumChannel(channel, false)
	}
	if configured == "none" {
		return
	}
	for _, token := range strings.Split(configured, ",") {
		channel, err := strconv.Atoi(strings.TrimSpace(token))
		if err != nil || channel < 1 || channel > 16 {
			log.Fatalf("invalid --drum-channels value %q; use MIDI channel numbers 1-16 or none", configured)
		}
		synth.SetDrumChannel(channel-1, true)
	}
}
