package cmd

import (
	"fmt"
	"io"
	"strconv"

	"github.com/plasticgaming99/pg99pro/synth/sf2abst"
)

func PrintlnSF2Bulk(w io.Writer, sf2 *sf2abst.SF2Abst) {
	fmt.Fprintln(w, "--- information(INFO) chunk")
	fmt.Fprintln(w, "soundfont version(ifil) :", strconv.FormatUint(uint64(sf2.Info.Ifil.Major), 16)+"."+strconv.FormatUint(uint64(sf2.Info.Ifil.Minor), 16))
	fmt.Fprintln(w, "target chip(isng)       :", sf2.Info.Isng)
	fmt.Fprintln(w, "soundfont name(INAM)    :", sf2.Info.INAM)
	fmt.Fprintln(w, "wavetable rom chip(irom):", sf2.Info.Irom)
	fmt.Fprintln(w, "wavetable rom ver(iver) :", strconv.FormatUint(uint64(sf2.Info.Iver.Major), 16)+"."+strconv.FormatUint(uint64(sf2.Info.Iver.Minor), 16))
	fmt.Fprintln(w, "creation date(ICRD)     :", sf2.Info.ICRD)
	fmt.Fprintln(w, "creator name(IENG)      :", sf2.Info.IENG)
	fmt.Fprintln(w, "target soundcard(IPRD)  :", sf2.Info.IPRD)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "--- sample(sdta) chunk")
	fmt.Fprintln(w, "smpl size:", len(sf2.Sdta.Smpl))
	fmt.Fprintln(w, "sm24 size:", len(sf2.Sdta.Sm24))
}
