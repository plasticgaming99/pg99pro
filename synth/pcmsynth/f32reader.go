package pcmsynth

import (
	"fmt"
	"io"
)

type Float32Reader interface {
	ReadFloat32(f32 []float32) (n int, err error)
}

type Float32StereoMixReader struct {
	leftReader  Float32Reader
	rightReader Float32Reader
	bufferL     []float32
	bufferR     []float32
}

// NewFloat32StereoMixReader combines two independent mono streams. Do not pass
// two readers backed by the same stateful voice or synth; each reader advances
// its own stream once per call.
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
	frames := len(f32) / 2
	if cap(f32sr.bufferL) < frames {
		f32sr.bufferL = make([]float32, frames)
		f32sr.bufferR = make([]float32, frames)
	} else {
		f32sr.bufferL = f32sr.bufferL[:frames]
		f32sr.bufferR = f32sr.bufferR[:frames]
	}
	clear(f32sr.bufferL)
	clear(f32sr.bufferR)

	nl, errl := f32sr.leftReader.ReadFloat32(f32sr.bufferL)
	nr, errr := f32sr.rightReader.ReadFloat32(f32sr.bufferR)

	minn := min(nl, nr)

	for i := 0; i < minn; i++ {
		f32[i*2] = f32sr.bufferL[i]
		f32[i*2+1] = f32sr.bufferR[i]
	}
	clear(f32[minn*2:])

	if errl != nil && errl != io.EOF {
		return minn * 2, errl
	}
	if errr != nil && errr != io.EOF {
		return minn * 2, errr
	}
	if errl != nil || errr != nil {
		return minn * 2, io.EOF
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
	reqFloats := len(b) / 2
	if cap(ir.f32buf) < reqFloats {
		ir.f32buf = make([]float32, reqFloats)
	} else {
		ir.f32buf = ir.f32buf[:reqFloats]
	}
	clear(ir.f32buf)
	nf, readErr := ir.fr.ReadFloat32(ir.f32buf)
	if nf < 0 {
		nf = 0
	} else if nf > reqFloats {
		nf = reqFloats
	}
	var in int16
	for i := 0; i < nf; i++ {
		sample := ir.f32buf[i]
		if sample > 1 {
			sample = 1
		} else if sample < -1 {
			sample = -1
		}
		in = int16(sample * 32767.0)
		b[i*2] = byte(in & 0xFF)
		b[i*2+1] = byte((in >> 8) & 0xFF)
	}
	return nf * 2, readErr
}

func (ir *Float32ToByteReader) readFloat32(b []byte) (n int, err error) {
	return len(b), nil
}
