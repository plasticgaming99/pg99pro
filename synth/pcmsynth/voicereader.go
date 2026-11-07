package pcmsynth

import (
	"io"
	"math"

	"github.com/plasticgaming99/pg99pro/synth/sf2abst"
)

type VoiceReader interface {
	ReadFloat32(f32 []float32) (n int, err error)
	NoteOff()
}

// generate synthesizable voice
// with some envelope parameters
func NewVoice(v *VoiceSample, g *sf2abst.GeneratorParam) *Voice {
	gain := float32(1)
	if g != nil && g.Amp.InitialAttenuation > 0 {
		gain = float32(math.Pow(10, -float64(g.Amp.InitialAttenuation)/20))
	}
	vr := Voice{
		gain:                 gain,
		voice:                v,
		generator:            g,
		currentStep:          0,
		step:                 1.,
		envelopeEnabled:      true,
		panEnabled:           true,
		interpolationEnabled: true,
	}
	return &vr
}

type Voice struct {
	voice        *VoiceSample
	generator    *sf2abst.GeneratorParam
	currentStep  float64 // 1 per 2 byte
	monoScratchL []float32
	monoScratchR []float32

	gain               float32
	envStage           envelopeStage
	envElapsed         int
	envLevel           float32
	envReleaseFrom     float32
	envDelayFrames     int
	envAttackFrames    int
	envHoldFrames      int
	envDecayFrames     int
	envReleaseFrames   int
	envReleaseDuration int
	envSustainDB       float64
	envSustain         float32
	envDecayStep       float32
	envReleaseLevel    float32
	envReleaseStep     float32
	envPrepared        bool

	baseStep   float64
	step       float64 // pitch
	outputRate float64
	pitchBend  int

	key                  int
	vel                  int
	playKey              int
	playVel              int
	playKeySet           bool
	channelPan           float64 // -1 left, 0 center, +1 right
	monoOutput           bool
	envelopeEnabled      bool
	panEnabled           bool
	interpolationEnabled bool
	monoUpmix            bool
	drumOneShot          bool
	stopped              bool

	decay bool
}

type envelopeStage uint8

type panGains struct {
	left  float32
	right float32
}

var (
	monoPanTable   = makeMonoPanTable()
	stereoPanTable = makeStereoPanTable()
)

func makeMonoPanTable() [128]panGains {
	var table [128]panGains
	for pan := range table {
		var angle float64
		if pan <= 64 {
			angle = float64(pan) / 64 * math.Pi / 4
		} else {
			angle = math.Pi/4 + float64(pan-64)/63*math.Pi/4
		}
		table[pan] = panGains{left: float32(math.Cos(angle)), right: float32(math.Sin(angle))}
	}
	return table
}

func makeStereoPanTable() [128]panGains {
	var table [128]panGains
	for pan := range table {
		if pan <= 64 {
			angle := float64(64-pan) / 64 * math.Pi / 2
			table[pan] = panGains{left: 1, right: float32(math.Cos(angle))}
		} else {
			angle := float64(pan-64) / 63 * math.Pi / 2
			table[pan] = panGains{left: float32(math.Cos(angle)), right: 1}
		}
	}
	return table
}

const (
	envDelay envelopeStage = iota
	envAttack
	envHold
	envDecay
	envSustain
	envRelease
	envFinished
)

func (v *Voice) ReadFloat32(b []float32) (n int, err error) {
	if cap(v.monoScratchL) < len(b) {
		v.monoScratchL = make([]float32, len(b))
		v.monoScratchR = make([]float32, len(b))
	} else {
		v.monoScratchL = v.monoScratchL[:len(b)]
		v.monoScratchR = v.monoScratchR[:len(b)]
	}
	clear(v.monoScratchL)
	clear(v.monoScratchR)
	n, err = v.ReadFloat32Stereo(v.monoScratchL, v.monoScratchR)
	for i := 0; i < n; i++ {
		b[i] = (v.monoScratchL[i] + v.monoScratchR[i]) * 0.5
	}
	clear(b[n:])
	return n, err
}

func (v *Voice) ReadFloat32Stereo(left, right []float32) (n int, err error) {
	if len(left) != len(right) {
		return 0, io.ErrShortBuffer
	}
	if len(left) < 1 {
		return 0, io.ErrShortBuffer
	}
	if v.stopped {
		clear(left)
		clear(right)
		return 0, io.EOF
	}
	if v.voice == nil || len(v.voice.Sample) == 0 {
		clear(left)
		clear(right)
		return 0, io.EOF
	}
	if v.monoOutput {
		return v.readLegacyMono(left, right)
	}
	leftGain, rightGain := v.panGains()
	velocity := v.playVel
	if velocity <= 0 {
		velocity = v.vel
	}
	voiceLevel := v.gain * (float32(velocity) / 127)
	for i := range left {
		envelope := float32(1)
		if !v.envelopeEnabled {
			if v.decay {
				v.gain *= 0.9995
				if v.gain < 0.0001 {
					clear(left[i:])
					clear(right[i:])
					return i, io.EOF
				}
			}
		} else {
			envelope = v.advanceEnvelope()
		}
		if v.envStage == envFinished {
			clear(left[i:])
			clear(right[i:])
			return i, io.EOF
		}
		if int(v.currentStep) >= len(v.voice.Sample) {
			if v.shouldLoop() && int(v.voice.LoopEnd) <= len(v.voice.Sample) {
				v.currentStep = float64(v.voice.LoopStart)
			} else {
				clear(left[i:])
				clear(right[i:])
				return i, io.EOF
			}
		}
		if v.shouldLoop() && int(v.currentStep) >= int(v.voice.LoopEnd) {
			loopStart := float64(v.voice.LoopStart)
			loopLength := float64(v.voice.LoopEnd - v.voice.LoopStart)
			v.currentStep = loopStart + math.Mod(v.currentStep-loopStart, loopLength)
		}

		index := int(v.currentStep)
		fraction := float32(v.currentStep - float64(index))
		nextIndex := index + 1
		if v.shouldLoop() && nextIndex >= int(v.voice.LoopEnd) {
			nextIndex = int(v.voice.LoopStart)
		} else if nextIndex >= len(v.voice.Sample) {
			nextIndex = index
		}
		sampleL := v.voice.Sample[index]
		if v.interpolationEnabled {
			sampleL = lerp(sampleL, v.voice.Sample[nextIndex], fraction)
		}
		sampleR := sampleL
		if len(v.voice.RightSample) == len(v.voice.Sample) {
			sampleR = v.voice.RightSample[index]
			if v.interpolationEnabled {
				sampleR = lerp(sampleR, v.voice.RightSample[nextIndex], fraction)
			}
		}
		level := voiceLevel * envelope
		left[i] = sampleL * level * leftGain
		right[i] = sampleR * level * rightGain
		v.currentStep += v.step
		n++
	}
	return n, nil
}

// readLegacyMono preserves the pre-stereo sampler path: nearest-sample reads,
// legacy release fade, and identical samples in both scratch channels so the
// caller's mono downmix leaves the signal unchanged.
func (v *Voice) readLegacyMono(left, right []float32) (n int, err error) {
	velocity := v.playVel
	if velocity <= 0 {
		velocity = v.vel
	}
	voiceLevel := v.gain * (float32(velocity) / 127)
	for i := range left {
		if v.decay {
			v.gain *= 0.9995
			if v.gain < 0.0001 {
				clear(left[i:])
				clear(right[i:])
				return i, io.EOF
			}
		}
		if int(v.currentStep) >= len(v.voice.Sample) || int(v.currentStep) >= int(v.voice.LoopEnd) {
			v.currentStep = float64(v.voice.LoopStart)
		}
		index := int(v.currentStep)
		if index < 0 || index >= len(v.voice.Sample) {
			clear(left[i:])
			clear(right[i:])
			return i, io.EOF
		}
		sample := v.voice.Sample[index] * voiceLevel
		left[i], right[i] = sample, sample
		v.currentStep += v.step
		n++
	}
	return n, nil
}

func (v *Voice) advanceEnvelope() float32 {
	if v.generator == nil {
		v.envLevel = 1
		return v.envLevel
	}
	if !v.envPrepared {
		v.prepareEnvelope()
	}

	switch v.envStage {
	case envDelay:
		v.envLevel = 0
		v.envElapsed++
		if v.envElapsed >= v.envDelayFrames {
			v.envStage, v.envElapsed = envAttack, 0
		}
	case envAttack:
		v.envElapsed++
		v.envLevel = min(float32(v.envElapsed)/float32(v.envAttackFrames), 1)
		if v.envElapsed >= v.envAttackFrames {
			v.envStage, v.envElapsed = envHold, 0
		}
	case envHold:
		v.envLevel = 1
		v.envElapsed++
		if v.envElapsed >= v.envHoldFrames {
			v.envStage, v.envElapsed = envDecay, 0
		}
	case envDecay:
		v.envElapsed++
		v.envLevel *= v.envDecayStep
		if v.envElapsed >= v.envDecayFrames {
			v.envLevel = v.envSustain
			v.envStage, v.envElapsed = envSustain, 0
		}
	case envSustain:
		v.envLevel = v.envSustain
	case envRelease:
		v.envElapsed++
		v.envReleaseLevel *= v.envReleaseStep
		v.envLevel = v.envReleaseLevel
		if v.envElapsed >= v.envReleaseDuration {
			v.envStage = envFinished
		}
	}
	return v.envLevel
}

// prepareEnvelope converts generator times to output frames once per voice,
// keeping logarithms and exponentials out of the real-time sample loop.
func (v *Voice) prepareEnvelope() {
	if v.generator == nil {
		v.envPrepared = true
		return
	}
	rate := v.outputRate
	if rate <= 0 && v.voice != nil {
		rate = float64(v.voice.SampleRate)
	}
	if rate <= 0 {
		rate = 48000
	}
	key := v.playKey
	if !v.playKeySet {
		key = v.key
	}
	if v.generator.Etc.Keynum >= 0 {
		key = int(v.generator.Etc.Keynum)
	}
	duration := func(seconds float32, keyScale int16) int {
		if seconds <= 0 {
			seconds = float32(math.Exp2(-10))
		}
		timecents := 1200 * math.Log2(float64(seconds))
		timecents += float64(key-60) * float64(keyScale)
		timecents = min(max(timecents, -12000), 8000)
		seconds = float32(math.Exp2(timecents / 1200))
		frames := int(float64(seconds) * rate)
		if frames < 1 {
			return 1
		}
		return frames
	}
	amp := v.generator.Amp
	v.envDelayFrames = duration(amp.DelayVolEnv, 0)
	v.envAttackFrames = duration(amp.AttackVolEnv, 0)
	v.envHoldFrames = duration(amp.HoldVolEnv, amp.KeynumToVolEnvHold)
	v.envSustainDB = min(max(float64(amp.SustainVolEnv), 0), 144)
	v.envDecayFrames = int(float64(duration(amp.DecayVolEnv, amp.KeynumToVolEnvDecay)) * v.envSustainDB / 100)
	if v.envDecayFrames < 1 {
		v.envDecayFrames = 1
	}
	v.envReleaseFrames = duration(amp.ReleaseVolEnv, 0)
	v.envSustain = float32(math.Pow(10, -v.envSustainDB/20))
	v.envDecayStep = float32(math.Pow(float64(v.envSustain), 1/float64(v.envDecayFrames)))
	v.envPrepared = true
}

func (v *Voice) shouldLoop() bool {
	if v.drumOneShot || v.voice.LoopEnd <= v.voice.LoopStart || v.generator == nil {
		return false
	}
	mode := v.generator.Etc.SampleModes
	return mode == 1 || (mode == 3 && !v.decay)
}

func lerp(a, b, t float32) float32 { return a + (b-a)*t }

func (v *Voice) SetChannelPan(pan float64) {
	if pan < -1 {
		pan = -1
	}
	if pan > 1 {
		pan = 1
	}
	v.channelPan = pan
}

func (v *Voice) SetMonoOutput(mono bool) {
	v.monoOutput = mono
}

func (v *Voice) SetEnvelopeEnabled(enabled bool) {
	v.envelopeEnabled = enabled
}

func (v *Voice) SetPanEnabled(enabled bool) {
	v.panEnabled = enabled
}

func (v *Voice) SetInterpolationEnabled(enabled bool) {
	v.interpolationEnabled = enabled
}

func (v *Voice) SetMonoUpmix(enabled bool) {
	v.monoUpmix = enabled
}

func (v *Voice) SetDrumOneShot(enabled bool) {
	v.drumOneShot = enabled
}

func (v *Voice) Stop() {
	v.stopped = true
}

func (v *Voice) panGains() (float32, float32) {
	if v.monoOutput || v.monoUpmix {
		return 1, 1
	}
	var generatorPan float64
	if v.panEnabled && v.generator != nil {
		generatorPan = float64(v.generator.Amp.Pan) / 50
	}
	pan := generatorPan
	if v.panEnabled {
		pan += v.channelPan
	}
	if pan < -1 {
		pan = -1
	}
	if pan > 1 {
		pan = 1
	}
	// GM2 defines MIDI CC10 value 64 as center. Keep the asymmetric integer
	// ranges on either side so index 64 maps to an exact center entry.
	panIndex := 64
	if pan < 0 {
		panIndex += int(math.Round(pan * 64))
	} else if pan > 0 {
		panIndex += int(math.Round(pan * 63))
	}
	if panIndex < 0 {
		panIndex = 0
	} else if panIndex > 127 {
		panIndex = 127
	}
	if len(v.voice.RightSample) == len(v.voice.Sample) {
		gains := stereoPanTable[panIndex]
		return gains.left, gains.right
	}
	gains := monoPanTable[panIndex]
	return gains.left, gains.right
}

func (v *Voice) PitchBend(pitch int) {
	// MIDI pitch bend is centered at zero, with a conventional +/-2 semitone range.
	if pitch < -8192 {
		pitch = -8192
	}
	if pitch > 8191 {
		pitch = 8191
	}
	v.pitchBend = pitch
	v.updateStep()
}

func (v *Voice) SetOutputRate(rate float64) {
	if rate <= 0 {
		return
	}
	v.outputRate = rate
	v.updateStep()
	v.envPrepared = false
	v.prepareEnvelope()
}

func (v *Voice) updateStep() {
	if v.voice == nil || v.voice.SampleRate == 0 {
		return
	}
	outputRate := v.outputRate
	if outputRate <= 0 {
		outputRate = float64(v.voice.SampleRate)
	}
	// Samples advanced per output frame = source/output rate times note tuning.
	v.step = float64(v.voice.SampleRate) / outputRate * v.baseStep *
		math.Exp2((float64(v.pitchBend)/8192*200)/1200)
}

func (v *Voice) NoteOff() {
	v.decay = true
	if !v.monoOutput && v.envelopeEnabled && v.envStage != envFinished && v.envStage != envRelease {
		if !v.envPrepared {
			v.prepareEnvelope()
		}
		v.envReleaseFrom = v.envLevel
		startDB := -20 * math.Log10(max(float64(v.envReleaseFrom), 1e-5))
		remainingDepth := max(100-startDB, 0)
		v.envReleaseDuration = int(float64(v.envReleaseFrames) * remainingDepth / 100)
		if v.envReleaseDuration < 1 {
			v.envReleaseDuration = 1
		}
		v.envReleaseLevel = float32(math.Pow(10, -startDB/20))
		v.envReleaseStep = float32(math.Pow(10, -remainingDepth/(20*float64(v.envReleaseDuration))))
		v.envElapsed = 0
		v.envStage = envRelease
	}
}
