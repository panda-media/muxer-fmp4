package commonBoxes

import (
	"github.com/panda-media/muxer-fmp4/format/AVPacket"
	"github.com/panda-media/muxer-fmp4/format/MP4"
)

func trakBox(packet *AVPacket.MediaPacket, arrays *MP4.MOOV_ARRAYS, timestamp, duration uint64) (box *MP4.MP4Box, err error) {
	box, err = MP4.NewMP4Box("trak")
	if err != nil {
		return
	}
	//tkhd
	param_tkhd := &tkhdParam{}
	param_tkhd.version = 1
	param_tkhd.creation_time = timestamp
	param_tkhd.modification_time = timestamp
	param_tkhd.duration = duration
	if AVPacket.AV_PACKET_TYPE_AUDIO == packet.PacketType {
		param_tkhd.track_ID = MP4.TRACK_AUDIO
		param_tkhd.bAudio = true
	} else {
		param_tkhd.track_ID = MP4.TRACK_VIDEO
		var w, h int
		w, h, err = getVideoWidthHeight(packet)
		if err != nil {
			return
		}
		param_tkhd.width = uint32(w)
		param_tkhd.height = uint32(h)
	}
	tkhd, err := tkhdBox(param_tkhd)
	if err != nil {
		return
	}
	box.PushBox(tkhd)
	//mdia
	var timeScale uint32
	timeScale = MP4.VIDE_TIME_SCALE
	if AVPacket.AV_PACKET_TYPE_AUDIO == packet.PacketType {
		timeScale, _, err = getAudioSampleRateSampleSize(packet)
		if err != nil {
			return
		}
	}
	mdia, err := mdiaBox(packet, arrays, timestamp, timeScale)
	if err != nil {
		return
	}
	box.PushBox(mdia)
	return
}
