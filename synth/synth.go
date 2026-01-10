package synth

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

func Synthesis() {
	defer midi.CloseDriver()

	in, err := midi.FindInPort("Midi Through Port-0")
	if err != nil {
		fmt.Println("can't find port")
		return
	}

	stop, err := midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
		var bt []byte
		var ch, key, vel uint8
		switch {
		case msg.GetSysEx(&bt):
			fmt.Printf("got sysex: % X\n", bt)
		case msg.GetNoteStart(&ch, &key, &vel):
			fmt.Printf("starting note %s on channel %v with velocity %v\n", midi.Note(key), ch, vel)
		case msg.GetNoteEnd(&ch, &key):
			fmt.Printf("ending note %s on channel %v\n", midi.Note(key), ch)
		default:
			// ignore
		}
	}, midi.UseSysEx())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT)

	<-sig
	stop()
	os.Exit(0)
}
