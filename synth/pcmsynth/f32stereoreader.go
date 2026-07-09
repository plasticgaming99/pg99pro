package pcmsynth

type Float32StereoReader interface {
	ReadFloat32Stereo(f32l, f32r []float32) (l, r int, err error)
}

type Float32PanReader struct {
	source Float32Reader
	panpot int
}
