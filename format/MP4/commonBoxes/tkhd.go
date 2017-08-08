package commonBoxes

import (
	"errors"
	"github.com/panda-media/muxer-fmp4/format/MP4"
)

type tkhdParam struct {
	version           byte //flags aways 1
	creation_time     uint64
	modification_time uint64
	duration          uint64
	track_ID          uint32
	bAudio            bool
	width             uint32
	height            uint32
}

func tkhdBox(param *tkhdParam) (box *MP4.MP4Box, err error) {
	if nil == param {
		return nil, errors.New("nil pointer for tkhd")
	}
	box, err = MP4.NewMP4Box("tkhd")
	if err != nil {
		return
	}
	box.SetVersionFlags(param.version, 1)
	if param.version == 1 {
		box.Push8Bytes(param.creation_time)
		box.Push8Bytes(param.modification_time)
		box.Push4Bytes(param.track_ID)
		box.Push4Bytes(0)
		box.Push8Bytes(param.duration)
	} else {
		box.Push4Bytes(uint32(param.creation_time & 0xffffffff))
		box.Push4Bytes(uint32(param.modification_time & 0xffffffff))
		box.Push4Bytes(param.track_ID)
		box.Push4Bytes(0)
		box.Push4Bytes(uint32(param.duration))
	}

	box.Push8Bytes(0)
	box.Push4Bytes(0)
	if param.bAudio {
		box.Push2Bytes(0x100)
	} else {
		box.Push2Bytes(0x00)
	}
	box.Push2Bytes(0x00)

	box.Push4Bytes(0x10000)
	box.Push4Bytes(0)
	box.Push4Bytes(0)

	box.Push4Bytes(0)
	box.Push4Bytes(0x10000)
	box.Push4Bytes(0)

	box.Push4Bytes(0)
	box.Push4Bytes(0)
	box.Push4Bytes(0x40000000)

	box.Push4Bytes(param.width)
	box.Push4Bytes(param.height)

	return
}
