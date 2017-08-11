package commonBoxes

import (
	"github.com/panda-media/muxer-fmp4/format/AVPacket"
)

func mdiaBox(packet *AVPacket.MediaPacket, arrays *MOOV_ARRAYS, timestamp uint64, timescale uint32) (box *MP4Box, err error) {
	box, err = NewMP4Box("mdia")
	if err != nil {
		return
	}
	//mdhd
	param_mdhd := &mdhdParam{0,
		timestamp,
		timestamp,
		timescale,
		0}
	mdhd, err := mdhdBox(param_mdhd)
	if err != nil {
		return
	}
	box.PushBox(mdhd)
	//hdlr
	hdlr, err := hdlrBox(packet.PacketType == AVPacket.AV_PACKET_TYPE_AUDIO)
	if err != nil {
		return
	}
	box.PushBox(hdlr)
	//minf
	minf, err := minfBox(packet, arrays)
	if err != nil {
		return
	}
	box.PushBox(minf)

	return
}
