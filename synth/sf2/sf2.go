package sf2

import (
	"fmt"
	"io"

	"golang.org/x/image/riff"
)

var (
	sfbk = riff.FourCC{'s', 'f', 'b', 'k'}

	INFO = riff.FourCC{'I', 'N', 'F', 'O'}

	Ifil = riff.FourCC{'i', 'f', 'i', 'l'}
	Isng = riff.FourCC{'i', 's', 'n', 'g'}
	INAM = riff.FourCC{'I', 'N', 'A', 'M'}
	Irom = riff.FourCC{'i', 'r', 'o', 'm'}
	Iver = riff.FourCC{'i', 'v', 'e', 'r'}
	ICRD = riff.FourCC{'I', 'C', 'R', 'D'}
	IENG = riff.FourCC{'I', 'E', 'N', 'G'}
	IPRD = riff.FourCC{'I', 'P', 'R', 'D'}
	ICOP = riff.FourCC{'I', 'C', 'O', 'P'}
	ICMT = riff.FourCC{'I', 'C', 'M', 'T'}
	ISFT = riff.FourCC{'I', 'S', 'F', 'T'}

	Sdta = riff.FourCC{'s', 'd', 't', 'a'}

	Smpl = riff.FourCC{'s', 'm', 'p', 'l'}
	Sm24 = riff.FourCC{'s', 'm', '2', '4'}

	Pdta = riff.FourCC{'p', 'd', 't', 'a'}

	Phdr = riff.FourCC{'p', 'h', 'd', 'r'}
	Pbag = riff.FourCC{'p', 'b', 'a', 'g'}
	Pmod = riff.FourCC{'p', 'm', 'o', 'd'}
	Pgen = riff.FourCC{'p', 'g', 'e', 'n'}
	Inst = riff.FourCC{'i', 'n', 's', 't'}
	Ibag = riff.FourCC{'i', 'b', 'a', 'g'}
	Imod = riff.FourCC{'i', 'm', 'o', 'd'}
	Igen = riff.FourCC{'i', 'g', 'e', 'n'}
	Shdr = riff.FourCC{'s', 'h', 'd', 'r'}
)

type SF2Raw struct {
	Info InfoStruct
	Stda StdaStruct
	Pdta PdtaStruct
}

type InfoStruct struct {
	Ifil []byte
	Isng []byte
	INAM []byte
	Irom []byte
	Iver []byte
	ICRD []byte
	IENG []byte
	IPRD []byte
	ICOP []byte
	ICMT []byte
	ISFT []byte
}

type StdaStruct struct {
	Smpl []byte
	Sm24 []byte
}

type PdtaStruct struct {
	Phdr []byte
	Pbag []byte
	Pmod []byte
	Pgen []byte
	Inst []byte
	Ibag []byte
	Imod []byte
	Igen []byte
	Shdr []byte
}

/*func ParseSF2(rd io.Reader) error {
	return ParseSF2Raw(rd)
}*/

func ParseSF2Raw(rd io.Reader) (SF2Raw, error) {
	sf2raw := SF2Raw{}
	fcc, data, err := riff.NewReader(rd)
	if err != nil {
		return SF2Raw{}, err
	}

	if fcc != sfbk {
		return SF2Raw{}, fmt.Errorf("sf2: not a soundfont")
	}

	for {
		chunkid, chunklen, chunkreader, err := data.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return SF2Raw{}, err
		}

		if chunkid == riff.LIST {
			chunkid, lsread, err := riff.NewListReader(chunklen, chunkreader)
			if err != nil {
				fmt.Println(err)
			}
			switch chunkid {
			case INFO:
				for {
					chunkid, _, chunkdata, err := lsread.Next()
					if err == io.EOF {
						break
					} else if err != nil {
						return SF2Raw{}, err
					}

					switch chunkid {
					case Ifil:
						sf2raw.Info.Ifil, err = io.ReadAll(chunkdata)
					case Isng:
						sf2raw.Info.Isng, err = io.ReadAll(chunkdata)
					case INAM:
						sf2raw.Info.INAM, err = io.ReadAll(chunkdata)
					case Irom:
						sf2raw.Info.Irom, err = io.ReadAll(chunkdata)
					case Iver:
						sf2raw.Info.Iver, err = io.ReadAll(chunkdata)
					case ICRD:
						sf2raw.Info.ICRD, err = io.ReadAll(chunkdata)
					case IENG:
						sf2raw.Info.IENG, err = io.ReadAll(chunkdata)
					case IPRD:
						sf2raw.Info.IPRD, err = io.ReadAll(chunkdata)
					case ICOP:
						sf2raw.Info.ICOP, err = io.ReadAll(chunkdata)
					case ICMT:
						sf2raw.Info.ICMT, err = io.ReadAll(chunkdata)
					case ISFT:
						sf2raw.Info.ISFT, err = io.ReadAll(chunkdata)
					}
					if err != nil {
						return SF2Raw{}, err
					}
				}
			case Sdta:
				for {
					chunkid, _, chunkdata, err := lsread.Next()
					if err == io.EOF {
						break
					} else if err != nil {
						return SF2Raw{}, err
					}

					switch chunkid {
					case Smpl:
						sf2raw.Stda.Smpl, err = io.ReadAll(chunkdata)
					case Sm24:
						sf2raw.Stda.Sm24, err = io.ReadAll(chunkdata)
					}
				}
			case Pdta:
				for {
					chunkid, _, chunkdata, err := lsread.Next()
					if err == io.EOF {
						break
					} else if err != nil {
						return SF2Raw{}, err
					}

					switch chunkid {
					case Phdr:
						sf2raw.Pdta.Phdr, err = io.ReadAll(chunkdata)
					case Pbag:
						sf2raw.Pdta.Pbag, err = io.ReadAll(chunkdata)
					case Pmod:
						sf2raw.Pdta.Pmod, err = io.ReadAll(chunkdata)
					case Pgen:
						sf2raw.Pdta.Pgen, err = io.ReadAll(chunkdata)
					case Inst:
						sf2raw.Pdta.Inst, err = io.ReadAll(chunkdata)
					case Ibag:
						sf2raw.Pdta.Ibag, err = io.ReadAll(chunkdata)
					case Imod:
						sf2raw.Pdta.Imod, err = io.ReadAll(chunkdata)
					case Igen:
						sf2raw.Pdta.Igen, err = io.ReadAll(chunkdata)
					case Shdr:
						sf2raw.Pdta.Shdr, err = io.ReadAll(chunkdata)
					}
				}
			}
		}
	}

	return sf2raw, nil
}
