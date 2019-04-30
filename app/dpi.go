package app

import (
	"github.com/mzki/erago/app/internal/devicescale"
	"golang.org/x/exp/shiny/unit"
)

func DPI(pixelsPerPt float32) float64 {
	dpi := float64(pixelsPerPt) * unit.PointsPerInch
	scale := devicescale.GetAt(0, 0) // this works fine on the assumeption it's single monitor.
	return scale * dpi
}
