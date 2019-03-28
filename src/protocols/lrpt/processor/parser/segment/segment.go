package segment

import (
	"encoding/binary"
	"fmt"
	"weather-dump/src/protocols/lrpt"
	"weather-dump/src/protocols/lrpt/processor/parser/segment/jpeg"
)

const segmentDataMinimum = 13

type Data struct {
	time  lrpt.Time
	MCUN  uint8
	QT    uint8
	DC    uint8
	AC    uint8
	QFM   uint16
	QF    uint8
	valid bool
	Lines [8][14 * 8]uint8
}

func NewFiller() *Data {
	return &Data{}
}

func New(buf []byte) *Data {
	e := Data{}
	e.FromBinary(buf)
	return &e
}

func (e *Data) FromBinary(dat []byte) {
	if len(dat) < segmentDataMinimum {
		return
	}

	e.time.FromBinary(dat[0:])
	e.MCUN = uint8(dat[8])
	e.QT = uint8(dat[9])
	e.DC = uint8(dat[10]) & 0xF0 >> 4
	e.AC = uint8(dat[10]) & 0x0F
	e.QFM = binary.BigEndian.Uint16(dat[11:])
	e.QF = uint8(dat[13])
	e.valid = true

	e.Decode(dat[14:])
}

func (e Data) GetMCUNumber() uint8 {
	return e.MCUN
}

func (e Data) GetDate() lrpt.Time {
	return e.time
}

func (e Data) GetID() uint32 {
	return e.GetDate().GetMilliseconds()
}

func (e Data) IsValid() bool {
	if e.valid && e.QT == 0x00 && e.DC == 0x00 && e.AC == 0x00 && e.QFM == 0xFFF0 && e.time.IsValid() {
		return true
	}
	return false
}

func (e Data) Print() {
	fmt.Println("### LRPT Segment Frame")
	fmt.Printf("MCU Number: %d\n", e.MCUN)
	fmt.Printf("Quantization Table: %08b\n", e.QT)
	fmt.Printf("Huffman Table DC: %04b\n", e.DC)
	fmt.Printf("Huffman Table AC: %04b\n", e.AC)
	fmt.Printf("Quality Factor Marker: %16b\n", e.QFM)
	fmt.Printf("Quality Factor: %08b\n", e.QF)
	fmt.Println()
	e.time.Print()
}

func (e *Data) Decode(data []byte) {
	buf := jpeg.ConvertToArray(data, len(data))
	qTable := jpeg.GetQuantizationTable(float64(e.QF))
	lastDC := int64(0)

	for i := 0; i < 14; i++ {
		var block [64]int64
		index := 1

		val := jpeg.FindDC(buf)
		if val == jpeg.CFC[0] {
			e.valid = false
			return
		}

		lastDC += val
		block[0] = lastDC

		for j := 0; j < 63; {
			vals := jpeg.FindAC(buf)
			j += len(vals)

			if vals[0] == jpeg.CFC[0] {
				e.valid = false
			}
			if vals[0] != jpeg.EOB[0] && index+len(vals) < len(block) {
				copy(block[index:], vals)
				index += len(vals)
			} else {
				break
			}
		}

		var idctBlock [64]int64
		for x := 0; x < 64; x++ {
			idctBlock[x] = block[jpeg.Zigzag[x]] * qTable[x]
		}

		jpeg.Idct(&idctBlock)

		for x := 0; x < 64; x++ {
			normalizedPixel := idctBlock[x] + 128

			if normalizedPixel > 255 {
				normalizedPixel = 255
			}
			if normalizedPixel < 0 {
				normalizedPixel = 0
			}

			e.Lines[x/8][(i*8)+(x%8)] = uint8(normalizedPixel)
		}
	}
}
