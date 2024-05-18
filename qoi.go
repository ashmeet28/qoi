package main

import "encoding/binary"

var ColorspaceSRGB uint8 = 0
var ColorspaceLinear uint8 = 1

var (
	qoiOpIndex uint8 = 0b00000000
	qoiOpDiff  uint8 = 0b01000000
	qoiOpLuma  uint8 = 0b10000000
	qoiOpRun   uint8 = 0b11000000
	qoiOpRGB   uint8 = 0b11111110
	qoiOpRGBA  uint8 = 0b11111111
)

var qoiMagic []uint8 = []byte("qoif")

var qoiPadding []uint8 = []byte{0, 0, 0, 0, 0, 0, 0, 1}

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
	return (c.r*3 + c.g*5 + c.b*7 + c.a*11) % 64
}

func Encode(data []byte, desc qoiHeader) []byte {
	var pxRun uint8

	var pxIndex []qoiRGBA = make([]qoiRGBA, 64)

	var px qoiRGBA
	var pxPrev qoiRGBA

	var pxChannels uint8

	var encData []byte

	var pxPos uint32
	var pxLen uint32
	var pxEnd uint32

	encData = append(encData, qoiMagic...)
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

	for i := range pxIndex {
		pxIndex[i].r = 0
		pxIndex[i].g = 0
		pxIndex[i].b = 0
		pxIndex[i].a = 0
	}

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
				encData = append(encData, qoiOpRun|pxRun)
				pxRun = 0
			}
		} else {
			if pxRun > 0 {
				encData = append(encData, qoiOpRun|pxRun)
				pxRun = 0
			}

			pxIndexPos := qoiColorHash(px)

			if pxIndex[pxIndexPos] == px {
				encData = append(encData, qoiOpIndex|pxIndexPos)
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
						encData = append(encData, qoiOpDiff|((vr+2)<<4)|((vg+2)<<2)|(vb+2))
					} else if vgr > 247 && vgr < 8 &&
						vg > 223 && vg < 32 &&
						vgb > 247 && vgb < 8 {
						encData = append(encData, qoiOpLuma|(vg+32), ((vgr+8)<<4)|(vgb+8))
					} else {
						encData = append(encData, qoiOpRGB, px.r, px.g, px.b)
					}
				} else {
					encData = append(encData, qoiOpRGBA, px.r, px.g, px.b, px.a)
				}
			}
		}
		pxPrev = px
	}

	encData = append(encData, qoiPadding...)

	return encData
}

func Decode() {}
