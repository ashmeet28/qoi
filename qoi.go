package main

const qoiSRGB = 0
const qoiLinear = 1

const (
	qoiOpIndex = 0b00000000
	qoiOpDiff  = 0b01000000
	qoiOpLuma  = 0b10000000
	qoiOpRun   = 0b11000000
	qoiOpRGB   = 0b11111110
	qoiOpRGBA  = 0b11111111
)

var qoiMagic []uint8 = []byte("qoif")

type qoiHeader struct {
	width      uint32
	height     uint32
	channels   uint8
	colorspace uint8
}

type qoiRGBA struct {
	r uint8
	g uint8
	b uint8
	a uint8
}

func qoiColorHash(c qoiRGBA) uint8 {
	return uint8((uint32(c.r)*3 + uint32(c.g)*5 + uint32(c.b)*7 + uint32(c.a)*11) % 64)
}
func main() {
	print(qoiMagic)
}
