package MP4

import (
	"github.com/panda-media/muxer-fmp4/format/AVPacket"
	"errors"
	"fmt"
	"bytes"
	"github.com/panda-media/muxer-fmp4/format/MP4/commonBoxes"
)

func NewMP4Muxer()*FMP4Muxer {
	muxer:=new(FMP4Muxer)
	return muxer
}


func (this *FMP4Muxer)SetAudioHeader(packet *AVPacket.MediaPacket)(err error){
	if nil==packet{
		return errors.New("nil audio header")
	}
	if AVPacket.AV_PACKET_TYPE_AUDIO!=packet.PacketType{
		return errors.New("invalid packet type")
	}
	SoundFormat:=packet.Data[0]>>4
	switch SoundFormat {
	case AVPacket.SoundFormat_AAC:
		this.audioHeader=packet.Copy()
	default:
		return errors.New(fmt.Sprintf("sound format %d not support now",int(SoundFormat)))

	}
	return
}


func (this *FMP4Muxer)SetVideoHeader(packet *AVPacket.MediaPacket)(err error){
	if nil==packet{
		return errors.New("nil video header")
	}
	if packet.PacketType!=AVPacket.AV_PACKET_TYPE_VIDEO{
		return errors.New("invalid packet type")
	}
	codecID:=packet.Data[0]&0xf
	switch codecID {
	case AVPacket.CodecID_AVC:
		FrameType:=packet.Data[0]>>4
		AVCPacketType:=packet.Data[1]
		if FrameType!=1||AVCPacketType!=0{
			return  errors.New("for AVC .no AVCDecoderConfigurationRecord")
		}
		this.videoHeader=packet.Copy()
	default:
		return errors.New(fmt.Sprintf("codeciID %d not support now",int(codecID)))
	}
	return
}

func (this *FMP4Muxer)GetInitSegment()(segData []byte,err error){
	buf:=bytes.Buffer{}
	//ftyp
	ftyp,err:=commonBoxes.Box_ftyp_Data()
	if err!=nil{
		err=errors.New(fmt.Sprintf("create ftyp failed:%s",err.Error()))
		return
	}
	buf.Write(ftyp)
	//moov

	return
}