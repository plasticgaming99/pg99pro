package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/plasticgaming99/pg99pro/synth"
	"github.com/plasticgaming99/pg99pro/synth/sf2"
)

func Execute(args []string) {
	f, err := os.Open(args[0])
	if err != nil {
		log.Fatal(err)
	}
	sf2, err := sf2.ParseSF2Raw(f)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("soundfont name: ", string(sf2.Info.INAM))

	fmt.Println("start synthesizer")
	synth.Synthesis()
}
