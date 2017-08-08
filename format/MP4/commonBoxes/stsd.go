package commonBoxes

import (
	"bytes"
	"github.com/panda-media/muxer-fmp4/format/AAC"
	"github.com/panda-media/muxer-fmp4/format/H264"
	"github.com/panda-media/muxer-fmp4/format/MP4"
)

const (
	MP4ESDescrTag          = 0x03
	MP4DecConfigDescrTag   = 0x04
	MP4DecSpecificDescrTag = 0x05
)

func stsdH264(avc *H264.AVCDecoderConfigurationRecord, width, height int) (box *MP4.MP4Box, err error) {
	box, err = MP4.NewMP4Box("stsd")
	if err != nil {
		return
	}
	box.SetVersionFlags(0, 0)
	box.Push4Bytes(1)
	avc1, err := MP4.NewMP4Box("avc1")
	if err != nil {
		return
	}
	avc1.Push8Bytes(1)
	avc1.Push8Bytes(0)
	avc1.Push8Bytes(0)

	avc1.Push2Bytes(uint16(width))
	avc1.Push2Bytes(uint16(height))

	avc1.Push4Bytes(0x480000)
	avc1.Push4Bytes(0x480000)
	avc1.Push4Bytes(0)

	avc1.Push2Bytes(1)
	avc1.PushByte(byte(len("AVC Coding")))
	avc1.PushBytes([]byte("AVC Coding"))
	spaceEnd := make([]byte, 31-len("AVC Coding"))
	avc1.PushBytes(spaceEnd)
	avc1.Push2Bytes(0x18)
	avc1.Push2Bytes(0xffff)

	avcC, err := MP4.NewMP4Box("avcC")
	if err != nil {
		return
	}
	avcC.PushBytes(avc.AVCData())

	avc1.PushBox(avcC)
	box.PushBox(avc1)

	return
}

func stsdAAC(sampleRate uint32, AudioSpecificConfig []byte) (box *MP4.MP4Box, err error) {
	box, err = MP4.NewMP4Box("stsd")
	if err != nil {
		return
	}
	box.SetVersionFlags(0, 0)
	box.Push4Bytes(1)

	mp4a, err := MP4.NewMP4Box("mp4a")
	if err != nil {
		return
	}
	mp4a.Push8Bytes(1)
	mp4a.Push8Bytes(0)
	mp4a.Push2Bytes(1)
	mp4a.Push2Bytes(16)
	mp4a.Push4Bytes(0)
	mp4a.Push4Bytes(sampleRate << 16)

	esds, err := MP4.NewMP4Box("esds")
	if err != nil {
		return
	}
	esds.SetVersionFlags(0, 0)
	esds.PushByte(MP4ESDescrTag)
	esdesc := esdDescrData(AudioSpecificConfig)
	esds.PushByte(byte(len(esdesc)))
	esds.PushBytes(esdesc)

	mp4a.PushBox(esds)
	box.PushBox(mp4a)
	return
}

func esdDescrData(AudioSpecificConfig []byte) []byte {
	buf := bytes.Buffer{}

	buf.WriteByte(0x00)
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)
	buf.WriteByte(MP4DecConfigDescrTag)
	asc := AAC.ASCForMP4(AudioSpecificConfig, "")
	esdDescData := esdDesc(asc)
	buf.WriteByte(byte(len(esdDescData)))

	buf.WriteByte(0x06)
	buf.WriteByte(0x01)
	buf.WriteByte(0x02)

	return buf.Bytes()
}

func esdDesc(asc []byte) []byte {
	buf := bytes.Buffer{}
	buf.WriteByte(0x40)
	buf.WriteByte(0x15)
	buf.WriteByte(0)
	buf.WriteByte(6)
	buf.WriteByte(0)
	buf.Write(make([]byte, 8)) //max bitrate avg bitrate
	buf.WriteByte(MP4DecSpecificDescrTag)
	if len(asc) > 0 {
		buf.WriteByte(byte(len(asc)))
		buf.Write(asc)
	}

	return buf.Bytes()
}
