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
	timescaleSIDX   uint32
	timescaleAudio  uint32
	timescaleVideo  uint32

	timeBeginVideo int64
	timeNowVideo int64
	timeBeginAudio int64
	timeNowAudio int64

	timeSlicedAudio int64//segment by video
	timeSidxAudio   int64//flush for sidx
	timeSlicedVideo int64//segment by key frame
	timeSidxVideo   int64//fluse for sidx
	mdat_size       int
}
