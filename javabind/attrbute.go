package eragoj

import (
	"local/erago/attribute"
)

const (
	ALIGNMENT_LEFT int8 = iota
	ALIGNMENT_CENTER
	ALIGNMENT_RIGHT
)

var toAlignmentMap = map[int8]attribute.Alignment{
	ALIGNMENT_LEFT:   attribute.AlignmentLeft,
	ALIGNMENT_CENTER: attribute.AlignmentCenter,
	ALIGNMENT_RIGHT:  attribute.AlignmentRight,
}

func toAlignment(b int8) attribute.Alignment {
	if a, ok := toAlignmentMap[b]; ok {
		return a
	}
	return attribute.AlignmentLeft // not match
}

var toAlignmentInt8Map = map[attribute.Alignment]int8{
	attribute.AlignmentLeft:   ALIGNMENT_LEFT,
	attribute.AlignmentCenter: ALIGNMENT_CENTER,
	attribute.AlignmentRight:  ALIGNMENT_RIGHT,
}

func toAlignmentInt8(a attribute.Alignment) int8 {
	if s, ok := toAlignmentInt8Map[a]; ok {
		return s
	}
	return ALIGNMENT_LEFT
}

func ColorOfName(cname string) int32 {
	c, _ := attribute.ColorOfName(cname)
	return int32(c)
}
