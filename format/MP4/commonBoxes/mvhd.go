package commonBoxes

import (
	"github.com/panda-media/muxer-fmp4/format/MP4"
	"errors"
)

type mvhdPram struct {
	version byte
	creation_time uint64
	modification_time uint64
	duration uint64
	timescale uint32//time-scale for the entire presentation
	next_track_ID uint32
}

func mvhdBox(param *mvhdPram)(box *MP4.MP4Box,err error){
	if nil==param{
		return  nil,errors.New("nil pointer for mvhd")
	}
	box,err=MP4.NewMP4Box("mvhd")
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

	box.Push4Bytes(0x10000)
	box.Push2Bytes(0x100)
	box.Push2Bytes(0x00)
	box.Push8Bytes(0x00)

	box.Push4Bytes(0x10000)
	box.Push4Bytes(0)
	box.Push4Bytes(0)

	box.Push4Bytes(0)
	box.Push4Bytes(0x10000)
	box.Push4Bytes(0)

	box.Push4Bytes(0)
	box.Push4Bytes(0)
	box.Push4Bytes(0x40000000)

	for i:=0;i<6;i++{
		box.Push4Bytes(0)
	}
	box.Push4Bytes(param.next_track_ID)
	return
}
