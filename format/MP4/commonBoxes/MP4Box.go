package commonBoxes

import (
	"bytes"
	"encoding/binary"
	"errors"
)

//unixtime + 0x7c0f4700
const (
	TRACK_VIDEO = iota + 1
	TRACK_AUDIO
	TRACK_NEXT
)

const (
	VIDE_TIME_SCALE             = 1000 //90000
	VIDE_TIME_SCALE_Millisecond = 1000
	MP3_SAMPLE_SIZE             = 1152
)

type MP4Box struct {
	boxType      []byte
	version      byte
	flags        int
	fullBox      bool
	boxSize      uint32
	boxSizeLarge uint64
	writer       *bytes.Buffer
}

func NewMP4Box(boxType string) (box *MP4Box, err error) {
	if len(boxType) != 4 {
		err = errors.New("invalid this type:" + boxType)
		return
	}

	box = &MP4Box{}
	box.writer = new(bytes.Buffer)
	box.boxType = []byte(boxType)
	return
}

func (this *MP4Box) SetVersionFlags(version byte, flags int) {
	this.version = version
	this.flags = flags
	this.fullBox = true
}

func (this *MP4Box) Flush() []byte {
	defer func() {
		this.writer.Reset()
		this.fullBox = false
	}()
	this.boxSize = 8
	if this.fullBox {
		this.boxSize += 4
	}
	if this.writer.Len() >= int(0xffffffff-this.boxSize) {
		this.boxSizeLarge = uint64(int(this.boxSize) + 8 + this.writer.Len())
		this.boxSize = 1
	} else {
		this.boxSize += uint32(this.writer.Len())
	}
	writer := bytes.Buffer{}
	binary.Write(&writer, binary.BigEndian, &this.boxSize)
	writer.Write(this.boxType)

	if 1 == this.boxSize {
		binary.Write(&writer, binary.BigEndian, &this.boxSizeLarge)
	}
	if this.fullBox {
		writer.WriteByte(byte(this.version))
		var tmp8 byte
		tmp8 = byte((this.flags >> 16) & 0xff)
		writer.WriteByte(tmp8)
		tmp8 = byte((this.flags >> 8) & 0xff)
		writer.WriteByte(tmp8)
		tmp8 = byte((this.flags >> 0) & 0xff)
		writer.WriteByte(tmp8)
	}

	writer.Write(this.writer.Bytes())
	return writer.Bytes()
}

func (this *MP4Box) Push8Bytes(data uint64) {
	binary.Write(this.writer, binary.BigEndian, &data)
}

func (this *MP4Box) Push4Bytes(data uint32) {
	binary.Write(this.writer, binary.BigEndian, &data)
}

func (this *MP4Box) Push2Bytes(data uint16) {
	binary.Write(this.writer, binary.BigEndian, &data)
}

func (this *MP4Box) PushByte(data byte) {
	this.writer.WriteByte(data)
}

func (this *MP4Box) PushBytes(data []byte) {
	this.writer.Write(data)
}

func (this *MP4Box) PushBox(inBox *MP4Box) {
	data := inBox.Flush()
	this.writer.Write(data)
}
