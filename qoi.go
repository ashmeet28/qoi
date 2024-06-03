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

	for pxPos = 0; pxPos < pxLen; pxPos += uint32(desc.channels) {
		px.r = data[pxPos]
		px.g = data[pxPos+1]
		px.b = data[pxPos+2]

		if desc.channels == 4 {
			px.a = data[pxPos+3]
		}

		if px == pxPrev {
			pxRun += 1
			if pxRun == 62 || pxPos == pxEnd {
				encData = append(encData, opRun|(pxRun-1))
				pxRun = 0
			}
		} else {
			if pxRun > 0 {
				encData = append(encData, opRun|(pxRun-1))
				pxRun = 0
			}

			pxIndexPos := pxHash(px)

			if pxIndex[pxIndexPos] == px {

				encData = append(encData, opIndex|pxIndexPos)

			} else {
				pxIndex[pxIndexPos] = px

				if px.a == pxPrev.a {
					var vr int8 = int8(px.r - pxPrev.r)
					var vg int8 = int8(px.g - pxPrev.g)
					var vb int8 = int8(px.b - pxPrev.b)

					var vgr int8 = int8((px.r - pxPrev.r) - (px.g - pxPrev.g))
					var vgb int8 = int8((px.b - pxPrev.b) - (px.g - pxPrev.g))

					if vr > -3 && vr < 2 &&
						vg > -3 && vg < 2 &&
						vb > -3 && vb < 2 {

						encData = append(encData,
							opDiff|(uint8(vr+2)<<4)|(uint8(vg+2)<<2)|uint8(vb+2))

					} else if vgr > -9 && vgr < 8 &&
						vg > -33 && vg < 32 &&
						vgb > -9 && vgb < 8 {

						encData = append(encData,
							opLuma|uint8(vg+32), (uint8(vgr+8)<<4)|uint8(vgb+8))

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

func Decode(data []byte, channels uint8) (Header, []byte) {
	var desc Header

	var pxRun uint8

	var pxIndex []pxRGBA = make([]pxRGBA, 64)

	var px pxRGBA

	var decData []byte

	data = data[len(magic) : len(data)-len(padding)]

	desc.width = binary.BigEndian.Uint32(data[:4])
	data = data[4:]
	desc.height = binary.BigEndian.Uint32(data[:4])
	data = data[4:]
	desc.channels = data[0]
	data = data[1:]
	desc.colorspace = data[0]
	data = data[1:]

	if channels == 0 {
		channels = desc.channels
	}

	px.r = 0
	px.g = 0
	px.b = 0
	px.a = 255

	for len(decData) < int(desc.width*desc.height*uint32(channels)) {

		if pxRun > 0 {

			pxRun--

		} else {

			b1 := data[0]
			data = data[1:]

			if b1 == opRGB {

				px.r = data[0]
				px.g = data[1]
				px.b = data[2]
				data = data[3:]

			} else if b1 == opRGBA {

				px.r = data[0]
				px.g = data[1]
				px.b = data[2]
				px.a = data[3]
				data = data[4:]

			} else if (b1 & 0b11000000) == opIndex {

				px = pxIndex[b1]

			} else if (b1 & 0b11000000) == opDiff {

				px.r += ((b1 >> 4) & 0x03) - 2
				px.g += ((b1 >> 2) & 0x03) - 2
				px.b += (b1 & 0x03) - 2

			} else if (b1 & 0b11000000) == opLuma {

				b2 := data[0]
				data = data[1:]

				var vg uint8 = (b1 & 0x3f) - 32
				px.r += (vg - 8) + ((b2 >> 4) & 0x0f)
				px.g += vg
				px.b += (vg - 8) + (b2 & 0x0f)

			} else if (b1 & 0b11000000) == opRun {

				pxRun = b1 & 0x3f

			}

			pxIndex[pxHash(px)] = px
		}

		decData = append(decData, px.r, px.g, px.b)

		if channels == 4 {
			decData = append(decData, px.a)
		}

	}

	return desc, decData
}
