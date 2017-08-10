package commonBoxes

import (
	"errors"
	"github.com/panda-media/muxer-fmp4/format/AVPacket"
	"github.com/panda-media/muxer-fmp4/format/MP4"
)

func minfBox(packet *AVPacket.MediaPacket) (box *MP4.MP4Box, err error) {
	box, err = MP4.NewMP4Box("minf")
	if err != nil {
		return
	}
	if packet.PacketType == AVPacket.AV_PACKET_TYPE_AUDIO {
		var smhd *MP4.MP4Box
		smhd, err = smhdBox()
		if err != nil {
			return
		}
		box.PushBox(smhd)
	} else if packet.PacketType == AVPacket.AV_PACKET_TYPE_VIDEO {
		var vmhd *MP4.MP4Box
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

	stbl, err := stblBox(packet)
	if err != nil {
		return
	}
	box.PushBox(stbl)

	return
}
