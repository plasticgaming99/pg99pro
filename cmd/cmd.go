package cmd

import (
	"fmt"
	"log"
	"os"
	"slices"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/ebitengine/oto/v3"
	"github.com/plasticgaming99/pg99pro/gui"
	"github.com/plasticgaming99/pg99pro/synth/pcmsynth"
	"github.com/plasticgaming99/pg99pro/synth/sf2abst"
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
	spew.Dump(p[0].Zones)
	new, err := os.Open(args[0])
	spew.Dump(err)
	gvop := pcmsynth.NewGenerateVoicesOptions()
	gvop.ResamplerEnabled = true
	gvop.Use16bitSamples = true
	gvop.UseBytesFromSF2Abst = false
	gvop.ResamplerRate = 48000
	v, err := pcmsynth.GenerateVoices(&sf2, gvop, new)
	spew.Dump(err)
	fmt.Println(len(v))

	//spew.Dump(sf2abst.GeneratorsToParam(sf2abst.PresetToGenerator(0, sf2)))

	/*index := sf2.Pdta.Pbag[0].GenIndex
	for {
		fmt.Println(sf2.Pdta.Pgen[index])
		if int(sf2.Pdta.Pgen[index].GenOper) == int(sf2abst.Op_instrument) {
			g := sf2abst.PGenToGenerator(nil, sf2.Pdta.Pgen[0:index-1])
			fmt.Println(g.ToParam())
			break
		}
		index++
	}*/

	/*inindex := sf2.Pdta.Ibag[0].GenIndex
	for {
		fmt.Println(sf2.Pdta.Igen[inindex])
		if int(sf2.Pdta.Igen[inindex].GenOper) == int(sf2abst.Op_sampleID) {
			g := sf2abst.PGenToGenerator(nil, sf2.Pdta.Pgen[0:inindex-1])
			fmt.Println(g.ToParam())
			break
		}
		inindex++
	}*/
	//fmt.Println(sf2.Pdta.Igen[len(sf2.Pdta.Igen)-1])

	if slices.Contains(args, "--gui") {
		gui.Execute()
	}

	otoop := &oto.NewContextOptions{
		SampleRate:   44100,
		ChannelCount: 1,
		Format:       oto.FormatSignedInt16LE,
		BufferSize:   10 * time.Millisecond,
	}
	otoCtx, readyChan, err := oto.NewContext(otoop)
	if err != nil {
		panic("oto.NewContext failed: " + err.Error())
	}
	<-readyChan

	fmt.Println("now playing?")

	voicetoplay := &v[60]

	fmt.Println(voicetoplay.Name)

	rd := pcmsynth.NewVoiceReader(voicetoplay)

	/*vtp := v[100]
	fmt.Println(vtp.Name)
	voicebyte := make([]byte, 0)
	for i := 0; i < len(vtp.Sample); i++ {
		b := make([]byte, 2)
		binary.LittleEndian.PutUint16(b, uint16(int16(vtp.Sample[i])))
		voicebyte = append(voicebyte, b...)
	}*/

	//pl := otoCtx.NewPlayer(bytes.NewReader(voicebyte))
	pl := otoCtx.NewPlayer(&rd)
	pl.Play()
	for pl.IsPlaying() {
		time.Sleep(time.Millisecond)
	}

	//fmt.Println("start synthesizer")
}
