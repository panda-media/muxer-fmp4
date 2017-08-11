package commonBoxes

import (
	"errors"
	"github.com/panda-media/muxer-fmp4/format/AVPacket"
)

func minfBox(packet *AVPacket.MediaPacket, arrays *MOOV_ARRAYS) (box *MP4Box, err error) {
	box, err = NewMP4Box("minf")
	if err != nil {
		return
	}
	if packet.PacketType == AVPacket.AV_PACKET_TYPE_AUDIO {
		var smhd *MP4Box
		smhd, err = smhdBox()
		if err != nil {
			return
		}
		box.PushBox(smhd)
	} else if packet.PacketType == AVPacket.AV_PACKET_TYPE_VIDEO {
		var vmhd *MP4Box
		vmhd, err = vmhdBox()
		if err != nil {
			return
		}
		box.PushBox(vmhd)
	} else {
		err = errors.New("not audio and not video media data")
		return
	}

	dinf, err := dinfBox()
	if err != nil {
		return
	}
	box.PushBox(dinf)

	stbl, err := stblBox(packet, arrays)
	if err != nil {
		return
	}
	box.PushBox(stbl)

	return
}
