package qoi

import (
	"encoding/binary"
)

var ColorspaceSRGB uint8 = 0
var ColorspaceLinear uint8 = 1

var (
	opIndex uint8 = 0b00000000
	opDiff  uint8 = 0b01000000
	opLuma  uint8 = 0b10000000
	opRun   uint8 = 0b11000000
	opRGB   uint8 = 0b11111110
	opRGBA  uint8 = 0b11111111
)

var magic []uint8 = []byte("qoif")

var padding []uint8 = []byte{0, 0, 0, 0, 0, 0, 0, 1}

type Header struct {
	width      uint32
	height     uint32
	channels   uint8
	colorspace uint8
}

type pxRGBA struct {
	r uint8
	g uint8
	b uint8
	a uint8
}

func pxHash(px pxRGBA) uint8 {
	return (px.r*3 + px.g*5 + px.b*7 + px.a*11) % 64
}

func Encode(data []byte, desc Header) []byte {
	var pxRun uint8

	var pxIndex []pxRGBA = make([]pxRGBA, 64)

	var px pxRGBA
	var pxPrev pxRGBA

	var pxChannels uint8

	var encData []byte

	var pxPos uint32
	var pxLen uint32
	var pxEnd uint32

	encData = append(encData, magic...)
	encData = binary.BigEndian.AppendUint32(encData, desc.width)
	encData = binary.BigEndian.AppendUint32(encData, desc.height)
	encData = append(encData, desc.channels)
	encData = append(encData, desc.colorspace)

	pxRun = 0

	pxPrev.r = 0
	pxPrev.g = 0
	pxPrev.b = 0
	pxPrev.a = 255

	px = pxPrev

	pxLen = desc.width * desc.height * uint32(desc.channels)
	pxEnd = pxLen - uint32(desc.channels)
	pxChannels = desc.channels

	for pxPos = 0; pxPos < pxLen; pxPos += uint32(pxChannels) {
		px.r = data[pxPos]
		px.g = data[pxPos+1]
		px.b = data[pxPos+2]

		if pxChannels == 4 {
			px.a = data[pxPos+3]
		}

		if px == pxPrev {
			pxRun += 1
			if pxRun == 62 || pxPos == pxEnd {
				encData = append(encData, opRun|pxRun)
				pxRun = 0
			}
		} else {
			if pxRun > 0 {
				encData = append(encData, opRun|pxRun)
				pxRun = 0
			}

			pxIndexPos := pxHash(px)

			if pxIndex[pxIndexPos] == px {
				encData = append(encData, opIndex|pxIndexPos)
			} else {
				pxIndex[pxIndexPos] = px

				if px.a == pxPrev.a {
					var vr uint8 = px.r - pxPrev.r
					var vg uint8 = px.g - pxPrev.g
					var vb uint8 = px.b - pxPrev.b

					var vgr uint8 = vr - vg
					var vgb uint8 = vb - vg

					if vr > 253 && vr < 2 &&
						vg > 253 && vg < 2 &&
						vb > 253 && vb < 2 {
						encData = append(encData, opDiff|((vr+2)<<4)|((vg+2)<<2)|(vb+2))
					} else if vgr > 247 && vgr < 8 &&
						vg > 223 && vg < 32 &&
						vgb > 247 && vgb < 8 {
						encData = append(encData, opLuma|(vg+32), ((vgr+8)<<4)|(vgb+8))
					} else {
						encData = append(encData, opRGB, px.r, px.g, px.b)
					}
				} else {
					encData = append(encData, opRGBA, px.r, px.g, px.b, px.a)
				}
			}
		}
		pxPrev = px
	}

	encData = append(encData, padding...)

	return encData
}

func Decode() {
}
