package MP4

import "github.com/panda-media/muxer-fmp4/format/AVPacket"

const (
	MEDIA_AV=iota
	MEDIA_Audio_Only
	MEDIA_Video_Only
)

type MP4Muxer struct {
	audioHeader *AVPacket.MediaPacket
	videoHeader *AVPacket.MediaPacket
}
