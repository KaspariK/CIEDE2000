package ciede2000

import (
	"image/color"
	"math"
)

// Implementation Notes:
//
// 1. In the formulas above, angles are expressed in degrees, not radians.
//
// 2. In computing h′1 and h′2, be careful with the inverse tangent since a′1
// and/or a′2 could be zero. Instead, use special math functions to do this. In
// the Standard C library and most other math libraries, this function is called
// atan2 and is used calling atan2(b, a). In Microsoft Excel, it is called ATAN2
// and its use is ATAN2(a, b). Note the argument reversal! These special
// functions will compute the proper inverse tangents without needing to worry
// about "divide by zero" conditions. Additionally, atan2 will also handle the
// quadrants for you.
//
// 3. Math libraries generally return inverse tangents in radians, not degrees.
// In Excel, use the DEGREES function to convert radians to degrees.
//
// 4. Math libraries generally require the input to the sine and cosine functions
// to be in radians, not degrees. In Excel, use the RADIANS function to convert
// degrees to radians.

// kL=kC=kH=1 under reference conditions
// Illumination: D65 source
func Distance(c1, c2 color.Color) float64 {
	l1 := toLAB(c1)
	l2 := toLAB(c2)

	// Calculate C'_i, h'_i
	cStar1 := math.Sqrt((l1.a * l1.a) + (l1.b * l1.b))
	cStar2 := math.Sqrt((l2.a * l2.a) + (l2.b * l2.b))

	cBar := (cStar1 + cStar2) / 2

	g := 0.5 * (1 - math.Sqrt(math.Pow(cBar, 7)/(math.Pow(cBar, 7)+math.Pow(25, 7))))

	aPrime1 := (1 + g) * l1.a
	aPrime2 := (1 + g) * l2.a

	cPrime1 := math.Sqrt((aPrime1 * aPrime1) + (l1.b * l1.b))
	cPrime2 := math.Sqrt((aPrime2 * aPrime2) + (l2.b * l2.b))

	var hPrime1 float64

	if l1.b == 0 && aPrime1 == 0 {
		hPrime1 = 0
	} else {
		hPrime1 = math.Atan2(l1.b, aPrime1) // are these in the right order?
	}

	var hPrime2 float64

	if l2.b == 0 && aPrime2 == 0 {
		hPrime2 = 0
	} else {
		hPrime2 = math.Atan2(l2.b, aPrime2)
	}

	deltaL := l2.l - l1.l
	deltaC := cPrime2 - cPrime1

	var deltaH float64

	if cPrime1*cPrime2 == 0 {
		deltaH = 0
	} else if math.Abs(hPrime2-hPrime1) <= 180 {
		deltaH = hPrime2 - hPrime1
	} else if hPrime2-hPrime1 > 180 {
		deltaH = (hPrime2 - hPrime1) - 360
	} else {
		deltaH = (hPrime2 - hPrime1) + 360
	}

	deltaH = 2 * math.Sqrt(cPrime1*cPrime2) * math.Sin(deltaH/2)

	lBarPrime := (l1.l + l2.l) / 2
	cBarPrime := (cPrime1 + cPrime2) / 2

	var hBarPrime float64

	if cPrime1*cPrime2 == 0 {
		hBarPrime = hPrime1 + hPrime2
	} else if math.Abs(hPrime1-hPrime2) <= 180 {
		hBarPrime = (hPrime1 + hPrime2) / 2
	} else if math.Abs(hPrime1-hPrime2) > 180 && math.Abs(hPrime1-hPrime2) < 360 {
		hBarPrime = ((hPrime1 + hPrime2) + 360) / 2
	} else {
		hBarPrime = ((hPrime1 + hPrime2) - 360) / 2
	}

	t := 1 - (0.17 * math.Cos(hBarPrime-30)) + (0.24 * math.Cos(2*hBarPrime)) + (0.32 * math.Cos(3*hBarPrime+6)) - (0.20 * math.Cos(4*hBarPrime-63))

	deltaTheta := 30 * math.Exp(-(((hBarPrime - 275) / 25) * ((hBarPrime - 275) / 25)))

	rC := 2 * math.Sqrt(math.Pow(cBarPrime, 7)/(math.Pow(cBarPrime, 7)*math.Pow(25, 7)))

	// Positional corrections to the lack of uniformity
	sL := 1 + (0.015 * ((lBarPrime - 50) * (lBarPrime - 50))) / (math.Sqrt(20 + ((lBarPrime - 50) * (lBarPrime - 50))))
	sC := 1 + (0.045*cBarPrime)
	sH := 1 + (0.015*cBarPrime*t)
	rT := math.Asin(2 * deltaTheta) * rC

	// Corrections accounting for the influence of experimental viewing conditions
	kL := 1.0
	kC := 1.0
	kH := 1.0

	deltaL /= kL * sL
	deltaC /= kC * sC
	deltaH /= kH * sH

	deltaE := math.Sqrt((deltaL * deltaL) + (deltaC * deltaC) + (deltaH * deltaH) + (rT * deltaC * deltaH))

	return deltaE
}

type xyz struct {
	x float64
	y float64
	z float64
}

// TODO: explain what XYZ is
// http://www.easyrgb.com/en/math.php
func toXYZ(c color.Color) xyz {
	sR, sG, sB, _ := c.RGBA()
	r, g, b := float64(sR), float64(sG), float64(sB)

	r /= 255 // not 255, but 65k?
	g /= 255
	b /= 255

	// TODO: breakout into function? What even is this?
	if r > 0.04045 {
		r = math.Pow((r+0.055)/1.055, 2.4)
	} else {
		r /= 12.92
	}

	if g > 0.04045 {
		g = math.Pow((g+0.055)/1.055, 2.4)
	} else {
		g /= 12.92
	}

	if b > 0.04045 {
		b = math.Pow((b+0.055)/1.055, 2.4)
	} else {
		b /= 12.92
	}

	r *= 100
	g *= 100
	b *= 100

	return xyz{
		x: (r * 0.4124) + (g * 0.3576) + (b * 0.1805),
		y: (r * 0.2126) + (g * 0.7152) + (b * 0.0722),
		z: (r * 0.0193) + (g * 0.1192) + (b * 0.9505),
	}
}

type lab struct {
	l float64
	a float64
	b float64
}

func toLAB(c color.Color) lab {
	xyz := toXYZ(c)

	// using D65 illuminant
	x := xyz.x / 95.047
	y := xyz.y / 100.000
	z := xyz.z / 108.883

	// TODO: breakout into function
	if x > 0.008856 {
		x = math.Pow(x, 1/3)
	} else {
		x = (x * 7.787) + (16 / 116)
	}

	if y > 0.008856 {
		y = math.Pow(y, 1/3)
	} else {
		y = (y * 7.787) + (16 / 116)
	}

	if z > 0.008856 {
		z = math.Pow(z, 1/3)
	} else {
		z = (z * 7.787) + (16 / 116)
	}

	return lab{
		l: (116 * y) - 16,
		a: 500 * (x - y),
		b: 200 * (y - z),
	}
}
