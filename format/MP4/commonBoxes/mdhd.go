package commonBoxes

import (
	"github.com/panda-media/muxer-fmp4/format/MP4"
	"errors"
)

//time scale for this media

type mdhdParam struct {
	version byte
	creation_time uint64
	modification_time uint64
	timescale uint32
	duration uint64
}

func mdhdBox(param *mdhdParam)(box *MP4.MP4Box,err error){
	if nil==param{
		return  nil,errors.New("nil pointer for mdhd")
	}
	box,err=MP4.NewMP4Box("mdhd")
	if err!=nil{
		return
	}
	box.SetVersionFlags(param.version,0)
	if 1==param.version{
		box.Push8Bytes(param.creation_time)
		box.Push8Bytes(param.modification_time)
		box.Push4Bytes(param.timescale)
		box.Push8Bytes(param.duration)
	}else {
		box.Push4Bytes(uint32(param.creation_time&0xffffffff))
		box.Push4Bytes(uint32(param.modification_time&0xffffffff))
		box.Push4Bytes(param.timescale)
		box.Push4Bytes(uint32(param.duration))
	}
	box.Push4Bytes(0x55c40000)

	return
}
