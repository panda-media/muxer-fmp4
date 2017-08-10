package commonBoxes

import (
	"github.com/panda-media/muxer-fmp4/format/AVPacket"
	"github.com/panda-media/muxer-fmp4/format/MP4"
	"errors"
	"github.com/panda-media/muxer-fmp4/format/H264"
)

func stblBox(media *AVPacket.MediaPacket)(box *MP4.MP4Box,err error){
	box,err=MP4.NewMP4Box("stbl")
	if err!=nil{
		return
	}
	//stsd
	stsd,err:=stbl_stsdBox(media)
	if err!=nil{
		return
	}
	box.PushBox(stsd)
	//stts
	stts,err:=sttsBox(nil)
	if err!=nil{
		return
	}
	box.PushBox(stts)
	//stsc
	stsc,err:=stscBox(nil)
	if err!=nil{
		return
	}
	box.PushBox(stsc)
	//stsz
	stsz,err:=stszBox(0,0,nil)
	if err!=nil{
		return
	}
	box.PushBox(stsz)
	//stco
	stco,err:=stco_co64(nil,false)
	if err!=nil{
		return
	}
	box.PushBox(stco)
	return
}

func stbl_stsdBox(media *AVPacket.MediaPacket)(box *MP4.MP4Box,err error){
	switch media.PacketType {
	case AVPacket.AV_PACKET_TYPE_AUDIO:
		soundFormat:=media.Data[0]>>4
		switch soundFormat {
		case AVPacket.SoundFormat_AAC:
			var sampleRate uint32
			sampleRate,_,err=getAudioSampleRateSampleSize(media)
			if err!=nil{
				return
			}
			box,err=stsdAAC(sampleRate,media.Data[2:])
			if err!=nil{
				return
			}
			return box,nil
		default:
			err=errors.New("only aac now")
			return
		}
	case AVPacket.AV_PACKET_TYPE_VIDEO:
		videoCodec:=media.Data[0]&0xf
		switch videoCodec {
		case AVPacket.CodecID_AVC:
			FrameType:=media.Data[0]>>4
			if FrameType!=1{
				err=errors.New("not a key frame avc")
				return
			}
			var avc *H264.AVCDecoderConfigurationRecord
			var width,height int
			avc,err=H264.DecodeAVC(media.Data[5:])
			if err!=nil{
				return
			}
			if avc.SPS!=nil&&avc.SPS.Len()>0{
				sps:=avc.SPS.Front().Value.([]byte)
				width,height,_,_,_,_=H264.DecodeSPS(sps)
			}
			box,err=stsdH264(avc,width,height)
			if err!=nil{
				return
			}
			return
		default:
			err=errors.New("not h264 for stsd")
			return
		}
	default:
		err=errors.New("not audio and video media for stsd")
	}
return
}