package main

import (
	"encoding/hex"
	"fmt"
	"math"
	"strings"
)

// ConvertRGBToCIE changes an RGB value to a CIE 1931 value
// Reference: https://gist.github.com/popcorn245/30afa0f98eea1c2fd34d
func ConvertRGBToCIE(redInt int, greenInt int, blueInt int) (float64, float64) {

	// Convert RGB values to values between 0-1
	red := float64(redInt / 255)
	green := float64(greenInt / 255)
	blue := float64(blueInt / 255)

	// Perform color correction for red
	red = RGBColorCorrect(red)
	green = RGBColorCorrect(green)
	blue = RGBColorCorrect(blue)

	// Convert RGB to XYZ
	x := (red * 0.649926) + (green * 0.103455) + (blue * 0.197109)
	y := (red * 0.234327) + (green * 0.743075) + (blue * 0.022598)
	z := (red * 0) + (green * 0.053077) + (blue * 1.035763)

	// Convert XYZ to XY
	x = x / (x + y + z)
	y = y / (x + y + z)

	return x, y
}

// RGBColorCorrect performs color correction on a converted RGB value
func RGBColorCorrect(input float64) float64 {
	if input > 0.04045 {
		return math.Pow((input+0.0055)/1.055, 2.4)
	}
	return input / 12.92
}

// SetLights changes the brightness
func (l *Lightshow) SetLights(lights []int, red int, green int, blue int, brightness float32) {
	if brightness < 0 || brightness > 1 {
		l.Logger.Error("Cannot set brightness outside of range 0-1")
		return
	}

	if red < 0 || red > 255 {
		l.Logger.Error("Cannot set red value outside of range 0-255")
		return
	}

	if green < 0 || green > 255 {
		l.Logger.Error("Cannot set green value outside of range 0-255")
		return
	}

	if blue < 0 || blue > 255 {
		l.Logger.Error("Cannot set blue value outside of range 0-255")
		return
	}

	xVal, yVal := ConvertRGBToCIE(red, green, blue)

	xValInt := int32(xVal * 0xffff)
	yValInt := int32(yVal * 0xffff)
	brightnessInt := int32(brightness * 0xffff)

	xValHighBit, xValLowBit := uint8(xValInt>>8), uint8(xValInt&0xff)
	yValHighBit, yValLowBit := uint8(yValInt>>8), uint8(yValInt&0xff)
	brightnessHighBit, brightnessLowBit := uint8(brightnessInt>>8), uint8(brightnessInt&0xff)

	message := []byte{}
	for _, light := range lights {
		message = append(message, []byte{
			0x00, 0x00, byte(light),
			xValHighBit, xValLowBit, yValHighBit, yValLowBit, brightnessHighBit, brightnessLowBit,
		}...)
	}

	l.MessageBytes = message

	msg := ""
	for _, b := range message {
		msg += fmt.Sprintf("0x%s ", hex.EncodeToString([]byte{b}))
	}
	l.Logger.WithField("bytes", strings.Trim(msg, " ")).Debug("Updated MessageBytes")
}
