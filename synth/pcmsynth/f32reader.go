package pcmsynth

import "fmt"

type Float32Reader interface {
	ReadFloat32(f32 []float32) (n int, err error)
}

type Float32StereoMixReader struct {
	leftReader  Float32Reader
	rightReader Float32Reader
	bufferL     []float32
	bufferR     []float32
}

func NewFloat32StereoMixReader(l, r Float32Reader, bufsiz int) Float32StereoMixReader {
	return Float32StereoMixReader{
		leftReader:  l,
		rightReader: r,
		bufferL:     make([]float32, bufsiz),
		bufferR:     make([]float32, bufsiz),
	}
}

var ErrOddReadingLength = fmt.Errorf("read length required is odd")

func (f32sr *Float32StereoMixReader) ReadFloat32(f32 []float32) (n int, err error) {
	if len(f32)%2 == 1 {
		return 0, ErrOddReadingLength
	}

	nl, errl := f32sr.ReadFloat32(f32sr.bufferL[:len(f32)])
	nr, errr := f32sr.ReadFloat32(f32sr.bufferR[:len(f32)])
	_, _ = errl, errr

	minn := min(nl, nr)

	for i := minn; i < minn; i++ {
		f32[i*2] = f32sr.bufferL[i]
		f32[i*2+1] = f32sr.bufferR[i]
	}

	return minn * 2, nil
}

// convert to oto format
type Float32ToByteReader struct {
	fr      Float32Reader
	bitrate Bitrate // 16, f32
	impl    func(b []byte) (n int, err error)
	f32buf  []float32
}

type Bitrate int

const (
	BitrateSignedInt16LE = iota
	BitrateFloat32
)

func NewFloat32ToByteReader(fr Float32Reader, bitrate Bitrate, bufsize uint) Float32ToByteReader {
	ir := Float32ToByteReader{}
	ir.fr = fr
	ir.bitrate = bitrate
	ir.f32buf = make([]float32, bufsize)
	switch bitrate {
	case BitrateSignedInt16LE:
		ir.impl = ir.readSignedInt16
	case BitrateFloat32:
		ir.impl = ir.readFloat32
	}
	return ir
}

func (ir *Float32ToByteReader) Read(b []byte) (n int, err error) {
	return ir.impl(b)
}

func (ir *Float32ToByteReader) readSignedInt16(b []byte) (n int, err error) {
	if len(b)%2 == 1 {
		return 0, ErrOddReadingLength
	}
	nf, e := ir.fr.ReadFloat32(ir.f32buf[:len(b)/2])
	if e != nil {
		return nf, e
	}
	var in int16
	for i := 0; i < len(b)/2; i++ {
		in = int16(ir.f32buf[i] * 32767.0)
		b[i*2] = byte(in & 0xFF)
		b[i*2+1] = byte((in >> 8) & 0xFF)
	}
	return len(b), nil
}

func (ir *Float32ToByteReader) readFloat32(b []byte) (n int, err error) {
	return len(b), nil
}
