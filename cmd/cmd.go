package cmd

import (
	"fmt"
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

	ph := sf2.Pdta.Phdr[0]
	index := sf2.Pdta.Pbag[ph.BagIndex].GenIndex
	u := uint16(32768)
	i := int16(u)
	fmt.Println(u, i)
	fmt.Println("generator", sf2.Pdta.Ibag[1].GenIndex)
	for {
		fmt.Println(sf2.Pdta.Pgen[index])
		if int(sf2.Pdta.Pgen[index].GenOper) == int(sf2abst.Op_instrument) {
			g := sf2abst.PGenToGenerator(nil, sf2.Pdta.Pgen[0:index-1])
			fmt.Println(g.ToParam())
			break
		}
		index++
	}

	/*for i := range 5 {
		fmt.Println(sf2.Pdta.Pgen[i])
	}*/

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
