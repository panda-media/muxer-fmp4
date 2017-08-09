package commonBoxes

import (
	"github.com/panda-media/muxer-fmp4/format/MP4"
	"github.com/panda-media/muxer-fmp4/format/AVPacket"
	"errors"
	"time"
)

func moovBox(audioHeader,videoHeader *AVPacket.MediaPacket) (box *MP4.MP4Box, err error) {
	timestamp:=uint64(time.Now().Unix()+0x7c0f4700)
	box,err=MP4.NewMP4Box("moov")
	if err!=nil{
		return
	}
	//mvhd
	param_mvhd:=&mvhdPram{version:0,
	creation_time:timestamp,
	modification_time:timestamp,
	duration:0,
	timescale:MP4.VIDE_TIME_SCALE,
	next_track_ID:MP4.TRACK_NEXT}
	mvhd,err:=mvhdBox(param_mvhd)
	if err!=nil{
		return
	}
	box.PushBox(mvhd)



	return
}


func Box_moov_Data(audioHeader,videoHeader *AVPacket.MediaPacket)(data []byte,err error)  {
	if nil==audioHeader&&nil==videoHeader{
		err=errors.New("no audio and audio header")
		return
	}
	box,err:=moovBox(audioHeader,videoHeader)
	if err!=nil{
		return
	}
	data=box.Flush()
return
}
