package qoi

import (
	"fmt"
	"os"
	"strconv"
	"testing"
)

func TestQOIEncodeOrDecode(t *testing.T) {
	a := os.Args[3:]
	op := a[0]
	w, _ := strconv.ParseInt(a[1], 10, 32)
	h, _ := strconv.ParseInt(a[2], 10, 32)
	c, _ := strconv.ParseInt(a[3], 10, 32)
	f1 := a[4]
	f2 := a[5]

	if op == "encode" {

		d, _ := os.ReadFile(f1)

		var desc Header
		desc.width = uint32(w)
		desc.height = uint32(h)
		desc.channels = uint8(c)
		desc.colorspace = ColorspaceSRGB

		os.WriteFile(f2, Encode(d, desc), 0666)

	} else if op == "decode" {

		encData, _ := os.ReadFile(f1)
		desc, decData := Decode(encData, uint8(c))
		os.WriteFile(f2, decData, 0666)

		fmt.Println(desc.width)
		fmt.Println(desc.height)
		fmt.Println(desc.channels)
		fmt.Println(desc.colorspace)

	}
}
