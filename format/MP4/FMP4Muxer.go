package MP4

import (
	"bytes"
	"github.com/panda-media/muxer-fmp4/format/AVPacket"
	"github.com/panda-media/muxer-fmp4/format/MP4/commonBoxes"
)

const (
	MEDIA_AV = iota
	MEDIA_Audio_Only
	MEDIA_Video_Only
)

type FMP4Muxer struct {
	audioHeader     *AVPacket.MediaPacket
	videoHeader     *AVPacket.MediaPacket
	sequence_number uint32 //1 base
	trunAudio       *commonBoxes.TRUN
	trunVideo       *commonBoxes.TRUN
	audio_data      *bytes.Buffer
	video_data      *bytes.Buffer
	moof_mdat_buf   *bytes.Buffer
	sidx            *commonBoxes.SIDX
	timescale       uint32
	timescaleAudio  uint32
	timescaleVideo  uint32
	timeBeginMS     uint32
	timeLastMS      uint32
	timeLastVideo   uint32
	timeLastAudio   uint32
	timeSlicedMS    uint32
	timeSidxMS      uint32
	mdat_size       int
}
