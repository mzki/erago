package model

import (
	"math"

	"golang.org/x/image/math/fixed"
)

var i26_6FracAtom = math.Pow(2, -6) // 6digits for fraction,

func floatToFixedInt(x float64) fixed.Int26_6 {
	i, frac := math.Modf(x)
	fracI32 := int32(frac / i26_6FracAtom) // may have float error.
	return fixed.Int26_6(int32(i)<<6 | fracI32)
}
