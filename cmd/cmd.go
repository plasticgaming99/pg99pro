package cmd

import (
	"log"
	"os"

	"github.com/plasticgaming99/pg99pro/synth/sf2abst"
)

func Execute(args []string) {
	f, err := os.Open(args[0])
	if err != nil {
		log.Fatal(err)
	}
	sf2, err := sf2abst.ParseSF2Abst(f)
	if err != nil {
		log.Fatal(err)
	}
	PrintlnSF2Bulk(os.Stdout, sf2)

	/*op := &oto.NewContextOptions{
		SampleRate:   31000,
		ChannelCount: 2,
		Format:       oto.FormatSignedInt16LE,
		BufferSize:   10 * time.Millisecond,
	}
	otoCtx, readyChan, err := oto.NewContext(op)
	if err != nil {
		panic("oto.NewContext failed: " + err.Error())
	}
	<-readyChan

	fmt.Println("start synthesizer")
	synth.Synthesis(&sf2, otoCtx)*/
}
