package sf2abst

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	sf2raw "github.com/plasticgaming99/pg99pro/synth/sf2"
	"golang.org/x/image/riff"
)

type SF2Abst struct {
	Info InfoStruct
	Sdta sf2raw.SdtaStruct
	Pdta PdtaStruct
}

func ParseSF2Abst(rd io.Reader) (SF2Abst, error) {
	sf2 := SF2Abst{}
	fcc, data, err := riff.NewReader(rd)
	if err != nil {
		return SF2Abst{}, err
	}

	if fcc != sf2raw.Sfbk {
		return SF2Abst{}, fmt.Errorf("sf2: not a soundfont")
	}

	for {
		chunkid, chunklen, chunkreader, err := data.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return SF2Abst{}, err
		}

		if chunkid == riff.LIST {
			chunkid, lsread, err := riff.NewListReader(chunklen, chunkreader)
			if err != nil {
				fmt.Println(err)
			}
			switch chunkid {
			case sf2raw.INFO:
				for {
					chunkid, _, chunkdata, err := lsread.Next()
					if err == io.EOF {
						break
					} else if err != nil {
						return SF2Abst{}, err
					}
					b, err := make([]byte, 0), error(nil)

					switch chunkid {
					case sf2raw.Ifil:
						b, err = io.ReadAll(chunkdata)
						sf2.Info.Ifil = ParseIfil(b)
					case sf2raw.Isng:
						b, err = io.ReadAll(chunkdata)
						sf2.Info.Isng = ParseIsng(b)
					case sf2raw.INAM:
						b, err = io.ReadAll(chunkdata)
						sf2.Info.INAM = ParseINAM(b)
					case sf2raw.Irom:
						b, err = io.ReadAll(chunkdata)
						sf2.Info.Irom = ParseIrom(b)
					case sf2raw.Iver:
						b, err = io.ReadAll(chunkdata)
						sf2.Info.Iver = ParseIver(b)
					case sf2raw.ICRD:
						b, err = io.ReadAll(chunkdata)
						sf2.Info.ICRD = ParseICRD(b)
					case sf2raw.IENG:
						b, err = io.ReadAll(chunkdata)
						sf2.Info.IENG = ParseICRD(b)
					case sf2raw.IPRD:
						b, err = io.ReadAll(chunkdata)
						sf2.Info.IPRD = ParseIPRD(b)
					case sf2raw.ICOP:
						b, err = io.ReadAll(chunkdata)
						sf2.Info.ICOP = ParseICOP(b)
					case sf2raw.ICMT:
						b, err = io.ReadAll(chunkdata)
						sf2.Info.ICMT = ParseICMT(b)
					case sf2raw.ISFT:
						b, err = io.ReadAll(chunkdata)
						sf2.Info.ISFT = ParseISFT(b)
					}
					if err != nil {
						return SF2Abst{}, err
					}
				}
			case sf2raw.Sdta:
				for {
					chunkid, _, chunkdata, err := lsread.Next()
					if err == io.EOF {
						break
					} else if err != nil {
						return SF2Abst{}, err
					}

					switch chunkid {
					case sf2raw.Smpl:
						sf2.Sdta.Smpl, err = io.ReadAll(chunkdata)
					case sf2raw.Sm24:
						sf2.Sdta.Sm24, err = io.ReadAll(chunkdata)
					}
				}
			case sf2raw.Pdta:
				for {
					chunkid, _, chunkdata, err := lsread.Next()
					if err == io.EOF {
						break
					} else if err != nil {
						return SF2Abst{}, err
					}
					b, err := make([]byte, 0), error(nil)

					switch chunkid {
					case sf2raw.Phdr:
						b, err = io.ReadAll(chunkdata)
						sf2.Pdta.Phdr = ParsePhdr(b)
					case sf2raw.Pbag:
						b, err = io.ReadAll(chunkdata)
						sf2.Pdta.Pbag = ParsePbag(b)
					case sf2raw.Pmod:
						b, err = io.ReadAll(chunkdata)
						sf2.Pdta.Pmod = ParsePmod(b)
					case sf2raw.Pgen:
						b, err = io.ReadAll(chunkdata)
						sf2.Pdta.Pgen = ParsePgen(b)
					case sf2raw.Inst:
						b, err = io.ReadAll(chunkdata)
						sf2.Pdta.Inst = ParseInst(b)
					case sf2raw.Ibag:
						b, err = io.ReadAll(chunkdata)
						sf2.Pdta.Ibag = ParseIbag(b)
					case sf2raw.Imod:
						b, err = io.ReadAll(chunkdata)
						sf2.Pdta.Imod = ParseImod(b)
					case sf2raw.Igen:
						b, err = io.ReadAll(chunkdata)
						sf2.Pdta.Igen = ParseIgen(b)
					case sf2raw.Shdr:
						b, err = io.ReadAll(chunkdata)
						sf2.Pdta.Shdr = ParseShdr(b)
					}
				}
			}
		}
	}

	return sf2, nil
}

type InfoStruct struct {
	Ifil Ifil
	Isng Isng
	INAM INAM
	Irom Irom
	Iver Iver
	ICRD ICRD
	IENG IENG
	IPRD IPRD
	ICOP ICOP
	ICMT ICMT
	ISFT ISFT
}

type Ifil struct {
	Major uint16
	Minor uint16
}

func ParseIfil(b []byte) Ifil {
	return Ifil{
		Major: binary.LittleEndian.Uint16(b[0:]),
		Minor: binary.LittleEndian.Uint16(b[2:]),
	}
}

type Isng = string

func ParseIsng(b []byte) Isng {
	return string(bytes.TrimRight(b, "\x00"))
}

type INAM = string

func ParseINAM(b []byte) INAM {
	return string(bytes.TrimRight(b, "\x00"))
}

type Irom = string

func ParseIrom(b []byte) Irom {
	return string(bytes.TrimRight(b, "\x00"))
}

type Iver = Ifil

func ParseIver(b []byte) Iver {
	return Iver(ParseIfil(b))
}

type ICRD = string

func ParseICRD(b []byte) ICRD {
	return string(bytes.TrimRight(b, "\x00"))
}

type IENG = string

func ParseIENG(b []byte) IENG {
	return string(bytes.TrimRight(b, "\x00"))
}

type IPRD = string

func ParseIPRD(b []byte) IPRD {
	return string(bytes.TrimRight(b, "\x00"))
}

type ICOP = string

func ParseICOP(b []byte) ICOP {
	return string(bytes.TrimRight(b, "\x00"))
}

type ICMT = string

func ParseICMT(b []byte) ICMT {
	return string(bytes.TrimRight(b, "\x00"))
}

type ISFT = string

func ParseISFT(b []byte) ISFT {
	return string(bytes.TrimRight(b, "\x00"))
}

type PdtaStruct struct {
	Phdr []Phdr
	Pbag []Pbag
	Pmod []Pmod
	Pgen []Pgen
	Inst []Inst
	Ibag []Ibag
	Imod []Imod
	Igen []Igen
	Shdr []Shdr
}

type Phdr struct {
	Name     string
	PresetNo uint16
	Bank     uint16
	BagIndex uint16
	Library  uint32
	Genre    uint32
	Morph    uint32
}

func ParsePhdr(b []byte) []Phdr {
	const size = 38
	n := len(b)/size - 1 // drop EOS

	out := make([]Phdr, 0, n)
	for i := 0; i < n; i++ {
		o := i * size

		nameBytes := b[o : o+20]
		name := string(bytes.TrimRight(nameBytes, "\x00"))

		out = append(out, Phdr{
			Name:     name,
			PresetNo: binary.LittleEndian.Uint16(b[o+20:]),
			Bank:     binary.LittleEndian.Uint16(b[o+22:]),
			BagIndex: binary.LittleEndian.Uint16(b[o+24:]),
			Library:  binary.LittleEndian.Uint32(b[o+26:]),
			Genre:    binary.LittleEndian.Uint32(b[o+38:]),
			Morph:    binary.LittleEndian.Uint32(b[o+42:]),
		})
	}

	return out
}

type Pbag struct {
	GenIndex uint16
	ModIndex uint16
}

func ParsePbag(b []byte) []Pbag {
	const size = 4
	n := len(b)/size - 1

	out := make([]Pbag, 0, n)
	for i := 0; i < n; i++ {
		o := i * size
		out = append(out, Pbag{
			GenIndex: binary.LittleEndian.Uint16(b[o+0:]),
			ModIndex: binary.LittleEndian.Uint16(b[o+2:]),
		})
	}
	return out
}

type Pmod struct {
	SrcOper      uint16
	DestOper     uint16
	ModAmount    int16
	AmdSrcOper   uint16
	ModTransOper uint16
}

func ParsePmod(b []byte) []Pmod {
	const size = 10
	n := len(b)/size - 1

	out := make([]Pmod, 0, n)
	for i := 0; i < n; i++ {
		o := i * size
		out = append(out, Pmod{
			SrcOper:      binary.LittleEndian.Uint16(b[o+0:]),
			DestOper:     binary.LittleEndian.Uint16(b[o+2:]),
			ModAmount:    int16(binary.LittleEndian.Uint16(b[o+4:])),
			AmdSrcOper:   binary.LittleEndian.Uint16(b[o+6:]),
			ModTransOper: binary.LittleEndian.Uint16(b[o+8:]),
		})
	}
	return out
}

type Pgen struct {
	GenOper   uint16
	GenAmount int16
}

func ParsePgen(b []byte) []Pgen {
	const size = 4
	n := len(b)/size - 1

	out := make([]Pgen, 0, n)
	for i := 0; i < n; i++ {
		o := i * size
		out = append(out, Pgen{
			GenOper:   binary.LittleEndian.Uint16(b[o+0:]),
			GenAmount: int16(binary.LittleEndian.Uint16(b[o+2:])),
		})
	}
	return out
}

type Inst struct {
	Name     string
	BagIndex uint16
}

func ParseInst(b []byte) []Inst {
	const size = 22
	n := len(b)/size - 1

	out := make([]Inst, 0, n)
	for i := 0; i < n; i++ {
		o := i * size
		out = append(out, Inst{
			Name:     string(bytes.TrimRight(b[o:o+20], "\x00")),
			BagIndex: binary.LittleEndian.Uint16(b[o+20:]),
		})
	}
	return out
}

type Ibag = Pbag

func ParseIbag(b []byte) []Ibag {
	return []Ibag(ParsePbag(b))
}

type Imod = Pmod

func ParseImod(b []byte) []Imod {
	return []Imod(ParsePmod(b))
}

type Igen = Pgen

func ParseIgen(b []byte) []Igen {
	return []Igen(ParsePgen(b))
}

type Shdr struct {
	Name        string
	Start       uint32
	End         uint32
	LoopStart   uint32
	LoopEnd     uint32
	SampleRate  uint32
	OriginalKey uint8
	PitchCorr   int8
	SampleLink  uint16
	SampleType  uint16
}

func ParseShdr(b []byte) []Shdr {
	const size = 46
	n := len(b)/size - 1 // drop EOS

	out := make([]Shdr, 0, n)
	for i := 0; i < n; i++ {
		o := i * size

		nameBytes := b[o : o+20]
		name := string(bytes.TrimRight(nameBytes, "\x00"))

		out = append(out, Shdr{
			Name:        name,
			Start:       binary.LittleEndian.Uint32(b[o+20:]),
			End:         binary.LittleEndian.Uint32(b[o+24:]),
			LoopStart:   binary.LittleEndian.Uint32(b[o+28:]),
			LoopEnd:     binary.LittleEndian.Uint32(b[o+32:]),
			SampleRate:  binary.LittleEndian.Uint32(b[o+36:]),
			OriginalKey: b[o+40],
			PitchCorr:   int8(b[o+41]),
			SampleLink:  binary.LittleEndian.Uint16(b[o+42:]),
			SampleType:  binary.LittleEndian.Uint16(b[o+44:]),
		})
	}

	return out
}
