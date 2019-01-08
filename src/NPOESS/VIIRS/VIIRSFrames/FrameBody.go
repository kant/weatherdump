package VIIRSFrames

import (
	"encoding/binary"
	"fmt"
	"weather-dump/src/NPOESS"
)

type FrameBody struct {
	sequenceCount    uint32
	packetTime       NPOESS.Time
	formatVersion    uint8
	instrumentNumber uint8
	integrityCheck   uint8
	selfTestPattern  uint8
	band             uint8
	detector         uint8
	syncWordPattern  uint32
	detectorData     [6]DetectorData
}

func (e *FrameBody) FromBinary(dat []byte) {
	e.sequenceCount = binary.BigEndian.Uint32(dat[0:])
	e.packetTime.FromBinary(dat[4:12])
	e.formatVersion = uint8(dat[12])
	e.instrumentNumber = uint8(dat[13])
	// Spare 16 bits
	e.integrityCheck = uint8(dat[16]) >> 7
	e.selfTestPattern = uint8(dat[16]&0x80) >> 4
	// Reserved 11 bits
	e.band = uint8(dat[18])
	e.detector = uint8(dat[19])
	e.syncWordPattern = binary.BigEndian.Uint32(dat[20:])
	// Reserved 512 bits
	buf := dat[88:]
	for i, _ := range e.detectorData {
		e.detectorData[i].FromBinary(&buf)
	}
}

func (e FrameBody) Print() {
	fmt.Println("### VIIRS Science Body")
	fmt.Printf("Sequence Count: %032b\n", e.sequenceCount)
	fmt.Printf("Packet Time: %s\n", e.packetTime.GetZulu())
	fmt.Printf("Format Version: %08b\n", e.formatVersion)
	fmt.Printf("Instrument Number: %08b\n", e.instrumentNumber)
	fmt.Println()
	fmt.Printf("Integrity Check: %01b\n", e.integrityCheck)
	fmt.Printf("Self Test Data Pattern: %04b\n", e.selfTestPattern)
	fmt.Printf("Band: %08b\n", e.band)
	fmt.Printf("Detector: %08b\n", e.detector)
	fmt.Printf("Sync Word Pattern: %032b\n", e.syncWordPattern)
	fmt.Println()

	for i, _ := range e.detectorData {
		e.detectorData[i].Print()
	}

	if e.IsValid() {
		fmt.Println("VALID FRAME")
	} else {
		fmt.Println("INVALID FRAME")
	}
	fmt.Println()
}

func (e FrameBody) IsValid() bool {
	for i, detector := range e.detectorData {
		if detector.syncWord != e.syncWordPattern && i != 5 {
			return false
		}
	}
	return true
}

func (e FrameBody) IsFillData(aggregationZone int) bool {
	return e.detectorData[aggregationZone].GetChecksum() == 0x0008
}

func (e FrameBody) GetAggrLen() int {
	return len(e.detectorData)
}

func (e *FrameBody) GetData(zone int, width int, oversample int) []byte {
	return e.detectorData[zone].GetData(width, oversample)
}

func (e *FrameBody) SetData(zone int, dat *[]byte) {
	e.detectorData[zone].SetData(dat)
}

func (e FrameBody) GetDetectorNumber() uint8 {
	return e.detector
}

func (e FrameBody) GetID() uint32 {
	return e.sequenceCount
}