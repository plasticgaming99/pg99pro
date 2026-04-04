package synth

import (
	"math"

	"github.com/ebitengine/oto/v3"
	"github.com/plasticgaming99/pg99pro/synth/sf2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

func Synthesis(sf2 *sf2.SF2Raw, otoctx *oto.Context) {
	/*defer midi.CloseDriver()

	in, err := midi.FindInPort("Midi Through Port-0")
	if err != nil {
		fmt.Println("can't find port")
		return
	}

	player := otoctx.NewPlayer(synth)
	player.SetBufferSize(2048)
	player.Play()

	stop, err := midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
		var bt []byte
		var ch, key, vel uint8
		switch {
		case msg.GetSysEx(&bt):
			fmt.Printf("got sysex: % X\n", bt)
		case msg.GetNoteStart(&ch, &key, &vel):
			fmt.Printf("starting note %s on channel %v with velocity %v\n", midi.Note(key), ch, vel)
			synth.NoteOn(ch, key, vel, sample)
		case msg.GetNoteEnd(&ch, &key):
			fmt.Printf("ending note %s on channel %v\n", midi.Note(key), ch)
			synth.NoteOff(ch, key)
		default:
			// ignore
		}
	}, midi.UseSysEx())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT)
	fmt.Println("might be ready")

	<-sig
	stop()
	os.Exit(0)*/
}

func calcStep(key, original uint8, sampleRate, outRate int) float64 {
	semitone := float64(int(key) - int(original))
	pitch := math.Pow(2.0, semitone/12.0)
	return pitch * float64(sampleRate) / float64(outRate)
}
