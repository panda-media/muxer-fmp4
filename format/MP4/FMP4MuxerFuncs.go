package MP4

import (
	"github.com/panda-media/muxer-fmp4/format/AVPacket"
	"errors"
	"strconv"
	"github.com/panda-media/muxer-fmp4/format/H264"
	"github.com/panda-media/muxer-fmp4/format/MP4/commonBoxes"
)

func (this *FMP4Muxer) AddPacket(packet *AVPacket.MediaPacket) (err error) {
	if nil==packet{
		err=errors.New("nil packet into FMP4 muxer")
		return
	}
	switch packet.PacketType {
	case AVPacket.AV_PACKET_TYPE_AUDIO:
		soundFormat:=packet.Data[0]>>4
		switch soundFormat {
		case AVPacket.SoundFormat_AAC:
			err=this.addAAC(packet)
			if err!=nil{
				return
			}
		default:
			err=errors.New("not support audio format "+strconv.Itoa(int(soundFormat)))
			return
		}
	case AVPacket.AV_PACKET_TYPE_VIDEO:
		videoFormat:=packet.Data[0]&0xf
		switch videoFormat {
		case AVPacket.CodecID_AVC:
			err=this.addH264(packet)
			if err!=nil{
				return
			}
		default:
			err=errors.New("not support video codec "+strconv.Itoa(int(videoFormat)))
			return
		}
	default:
		err=errors.New("invalid packet type"+strconv.Itoa(packet.PacketType))
		return
	}
	if 0==this.timeBeginMS{
		this.timeBeginMS=packet.TimeStamp
	}
	this.timeLastMS=packet.TimeStamp
	return
}

func (this *FMP4Muxer) Flush() (sidx, moof_mdats []byte, err error) {
	//no sidx now
	return
}

func (this *FMP4Muxer)sliceKeyFrame(){
	this.sidx.Version=0
	this.sidx.Reference_ID=1
	this.sidx.TimeScale=this.timescale
	this.sidx.Earliest_presentation_time=0x3e8

	sidxData:=commonBoxes.SIDX_REFERENCE{}
	sidxData.Reference_type=0;
	//sidxData.Referenced_size;//moof+mdat?
	sidxData.Subsegment_duration=0
	if this.sidx.Reference_count==0{
		sidxData.Starts_with_SAP=1
	}
	sidxData.SAP_type=1
	//sidxData.SAP_delta_time=?
	this.sidx.References.PushBack(sidxData)
	this.sidx.Reference_count=uint16(this.sidx.References.Len())
}

func (this *FMP4Muxer) addH264(packet *AVPacket.MediaPacket)(err error){
	if this.trunVideo==nil{
		return errors.New("video track not inited")
	}
	if packet.Data[1]==1{
		//AVC  parse,continue
		return
	}

	nalType:=packet.Data[5]&0x1f
	sampleSize :=0
	if nalType==H264.NAL_IDR_SLICE{
		//add sps pps
		avc,err:=H264.DecodeAVC(this.videoHeader.Data[5:])
		if err!=nil{
			sampleSize +=this.AddSPSPPS(avc)
		}
	}
	this.media_data.Write(packet.Data[5:])
	sampleSize +=len(packet.Data[5:])
	compositionTime:=uint32(0)
	compositionTime|=uint32(packet.Data[2])<<16
	compositionTime|=uint32(packet.Data[3])<<8
	compositionTime|=uint32(packet.Data[4])<<0

	trunData:=&commonBoxes.TRUN_ARRAY_FIELDS{}
	trunData.Sample_size= uint32(sampleSize)
	trunData.Sample_composition_time_offset=compositionTime
	this.trunVideo.Vals.PushBack(trunData)

	if nalType==H264.NAL_IDR_SLICE&&this.trunVideo.Vals.Len()>1{
		this.sliceKeyFrame()
	}

	return
}

func (this *FMP4Muxer)addAAC(packet *AVPacket.MediaPacket)(err error){
	if this.trunAudio==nil{
		return  errors.New("audio track not inited")
	}

	this.media_data.Write(packet.Data[1:])
	trunData:=&commonBoxes.TRUN_ARRAY_FIELDS{}
	trunData.Sample_size=uint32(len(packet.Data[1:]))
	this.trunAudio.Vals.PushBack(trunData)
	return
}

func (this *FMP4Muxer)AddSPSPPS(avc *H264.AVCDecoderConfigurationRecord)(size int){
	for e:=avc.SPS.Front();e!=nil ;  e=e.Next(){
		frameSize:=len(e.Value.([]byte))
		this.media_data.WriteByte(byte((frameSize>>24)&0xff))
		this.media_data.WriteByte(byte((frameSize>>16)&0xff))
		this.media_data.WriteByte(byte((frameSize>>8)&0xff))
		this.media_data.WriteByte(byte((frameSize>>0)&0xff))
		this.media_data.Write(e.Value.([]byte))
		size+=frameSize+4
	}
	for e:=avc.PPS.Front();e!=nil;e=e.Next(){
		frameSize:=len(e.Value.([]byte))
		this.media_data.WriteByte(byte((frameSize>>24)&0xff))
		this.media_data.WriteByte(byte((frameSize>>16)&0xff))
		this.media_data.WriteByte(byte((frameSize>>8)&0xff))
		this.media_data.WriteByte(byte((frameSize>>0)&0xff))
		this.media_data.Write(e.Value.([]byte))
		size+=frameSize+4
	}

	return
}
