package commonBoxes

import (
	"github.com/panda-media/muxer-fmp4/format/AVPacket"
	"github.com/panda-media/muxer-fmp4/format/MP4"
)

func mdiaBox(packet *AVPacket.MediaPacket,timestamp uint64,timescale uint32)(box *MP4.MP4Box,err error){
	box,err=MP4.NewMP4Box("mdia")
	if err!=nil{
		return
	}
	//mdhd
	param_mdhd:=&mdhdParam{0,
	timestamp,
	timestamp,
	timescale,
	0}
	mdhd,err:=mdhdBox(param_mdhd)
	if err!=nil{
		return
	}
	box.PushBox(mdhd)

	hdlr,err:=hdlrBox(packet.PacketType==AVPacket.AV_PACKET_TYPE_AUDIO)
	if err!=nil{
		return
	}
	box.PushBox(hdlr)

	stbl,err:=stblBox(packet)
	if err!=nil{
		return
	}

	box.PushBox(stbl)
	return
}
