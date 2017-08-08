package commonBoxes

import (
	"github.com/panda-media/muxer-fmp4/format/MP4"
)

type trexParam struct {
	trackId                          uint32
	default_sample_description_index uint32 //1
	default_sample_duration          uint32 //audio hz,video 1000
	default_sample_size              uint32
	default_sample_flags             uint32
}

func mvexBox(audioParam, videoParam *trexParam) (box *MP4.MP4Box, err error) {
	box, err = MP4.NewMP4Box("mvex")
	if err != nil {
		return
	}
	if nil != audioParam {
		audioTrex, err := trexBox(audioParam)
		if err != nil {
			return
		}
		box.PushBox(audioTrex)
	}
	if nil != videoParam {
		videoTrex, err := trexBox(videoParam)
		if err != nil {
			return
		}
		box.PushBox(videoTrex)
	}
	return
}

func trexBox(param *trexParam) (box *MP4.MP4Box, err error) {
	box, err = MP4.NewMP4Box("trex")
	if err != nil {
		return
	}
	box.SetVersionFlags(0, 0)
	box.Push4Bytes(param.trackId)
	box.Push4Bytes(param.default_sample_description_index)
	box.Push4Bytes(param.default_sample_duration)
	box.Push4Bytes(param.default_sample_size)
	box.Push4Bytes(param.default_sample_flags)
	return
}
