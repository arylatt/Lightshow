package lightshow

import (
	"context"
	"fmt"
	"math"
	"time"
)

var (
	// StateOff is a disabled state with no colour or brightness
	StateOff = State{[3]int{0, 0, 0}, 0}
)

// State holds the colour and brightness values
type State struct {
	Colour     [3]int  `json:"colour"`
	Brightness float64 `json:"brightness"`
}

// On indicates whether a light should be on or off
func (s *State) On() bool {
	return s.Brightness > 0 && s.Colour == [3]int{0, 0, 0}
}

// CIE returns the Kelvin XY colour values
// See https://gist.github.com/popcorn245/30afa0f98eea1c2fd34d
func (s *State) CIE() (xy [2]float64) {
	gammaCorrectedRGB := [3]float64{}

	for i, val := range s.Colour {
		gammaCorrectedRGB[i] = GammaCorrection(float64(val) / 255)
	}

	wideRGBD65XYZ := [3]float64{
		(gammaCorrectedRGB[0] * 0.649926) + (gammaCorrectedRGB[1] * 0.103455) + (gammaCorrectedRGB[2] * 0.197109),
		(gammaCorrectedRGB[0] * 0.234327) + (gammaCorrectedRGB[1] * 0.743075) + (gammaCorrectedRGB[2] * 0.022598),
		(gammaCorrectedRGB[0] * 0.000000) + (gammaCorrectedRGB[1] * 0.053077) + (gammaCorrectedRGB[2] * 1.035763),
	}

	xy[0] = wideRGBD65XYZ[0] / (wideRGBD65XYZ[0] + wideRGBD65XYZ[1] + wideRGBD65XYZ[2])
	xy[1] = wideRGBD65XYZ[1] / (wideRGBD65XYZ[0] + wideRGBD65XYZ[1] + wideRGBD65XYZ[2])
	return
}

// StepsTo generates a slice of steps to linearly change from one state to another
func (s *State) StepsTo(steps int, target State) (states []State) {
	if steps == 0 {
		return []State{target}
	}

	diffBri := -((s.Brightness - target.Brightness) / float64(steps))

	startRGB, endRGB := s.Colour, target.Colour
	diffRGB := [3]int{
		-((startRGB[0] - endRGB[0]) / steps),
		-((startRGB[1] - endRGB[1]) / steps),
		-((startRGB[2] - endRGB[2]) / steps),
	}

	states = append(states, *s)

	for i := 1; i < steps; i++ {
		states = append(states, State{
			[3]int{
				startRGB[0] + (diffRGB[0] * i),
				startRGB[1] + (diffRGB[1] * i),
				startRGB[2] + (diffRGB[2] * i),
			},
			s.Brightness + (diffBri * float64(i)),
		})
	}

	return append(states, target)
}

// bytes returns the byte array to send on the DTLS connection (minus header)
func (s *State) bytes(lights []int) (b []byte) {
	kelvin := s.CIE()

	lightData := [3]int32{int32(kelvin[0] * 0xffff), int32(kelvin[1] * 0xffff), int32(s.Brightness * 0xffff)}
	lightBytes := []byte{uint8(lightData[0] >> 8), uint8(lightData[0] & 0xff),
		uint8(lightData[1] >> 8), uint8(lightData[1] & 0xff),
		uint8(lightData[2] >> 8), uint8(lightData[2] & 0xff)}

	for _, light := range lights {
		b = append(b, append([]byte{0x00, 0x00, byte(light)}, lightBytes...)...)
	}

	return
}

// run will send the messages for the state changes
func (s *State) run(ctx context.Context, done context.CancelFunc, show *Lightshow, lights []int, next []State) {
	t := time.Now()

	defer func() {
		fmt.Printf("Elapsed for state run: %f\r\n", time.Since(t).Seconds())
	}()

	show.write(s.bytes(lights))

	if len(next) == 0 {
		done()

		return
	}

	select {
	case <-ctx.Done():
		return
	case <-time.After(EventFrequency - time.Millisecond*3):
		go next[0].run(ctx, done, show, lights, next[1:])
	}
}

// GammaCorrection performs Gamma Correction on an RGB value
// See https://gist.github.com/popcorn245/30afa0f98eea1c2fd34d
func GammaCorrection(input float64) float64 {
	if input > 0.04045 {
		return math.Pow((input+0.0055)/1.055, 2.4)
	}
	return input / 12.92
}
