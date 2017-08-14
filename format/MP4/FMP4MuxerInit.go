package MP4

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/panda-media/muxer-fmp4/format/AVPacket"
	"github.com/panda-media/muxer-fmp4/format/MP4/commonBoxes"
	"container/list"
)

func NewMP4Muxer() *FMP4Muxer {
	muxer := new(FMP4Muxer)
	muxer.sequence_number = 1
	muxer.audio_data =new(bytes.Buffer)
	muxer.video_data =new(bytes.Buffer)
	muxer.sidx=&commonBoxes.SIDX{}
	muxer.sidx.References=list.New()
	return muxer
}

func (this *FMP4Muxer) SetAudioHeader(packet *AVPacket.MediaPacket) (err error) {
	if nil == packet {
		return errors.New("nil audio header")
	}
	if AVPacket.AV_PACKET_TYPE_AUDIO != packet.PacketType {
		return errors.New("invalid packet type")
	}
	SoundFormat := packet.Data[0] >> 4
	switch SoundFormat {
	case AVPacket.SoundFormat_AAC:
		this.audioHeader = packet.Copy()
		this.timescaleAudio,_,_=commonBoxes.GetAudioSampleRateSampleSize(packet)
		if this.videoHeader==nil{
			this.timescale=this.timescaleAudio
		}
	default:
		return errors.New(fmt.Sprintf("sound format %d not support now", int(SoundFormat)))

	}
	this.trunAudio=&commonBoxes.TRUN{}
	this.trunAudio.Version=0
	this.trunAudio.Tr_flags=0x201//offset,samplesize
	this.trunAudio.Vals=list.New()
	return
}

func (this *FMP4Muxer) SetVideoHeader(packet *AVPacket.MediaPacket) (err error) {
	if nil == packet {
		return errors.New("nil video header")
	}
	if packet.PacketType != AVPacket.AV_PACKET_TYPE_VIDEO {
		return errors.New("invalid packet type")
	}
	codecID := packet.Data[0] & 0xf
	switch codecID {
	case AVPacket.CodecID_AVC:
		FrameType := packet.Data[0] >> 4
		AVCPacketType := packet.Data[1]
		if FrameType != 1 || AVCPacketType != 0 {
			return errors.New("for AVC .no AVCDecoderConfigurationRecord")
		}
		this.videoHeader = packet.Copy()
		this.trunVideo=&commonBoxes.TRUN{}
		this.trunVideo.Version=0
		this.trunVideo.Tr_flags=0xa01//offset,samplesize,composition
		this.trunVideo.Vals=list.New()
		this.timescale=commonBoxes.VIDE_TIME_SCALE
		this.timescaleVideo=commonBoxes.VIDE_TIME_SCALE
	default:
		return errors.New(fmt.Sprintf("codeciID %d not support now", int(codecID)))
	}
	return
}

func (this *FMP4Muxer) GetInitSegment() (segData []byte, err error) {
	buf := bytes.Buffer{}
	//ftyp
	ftyp, err := commonBoxes.Box_ftyp_Data()
	if err != nil {
		err = errors.New(fmt.Sprintf("create ftyp failed:%s", err.Error()))
		return
	}
	buf.Write(ftyp)
	//moov
	moov, err := commonBoxes.Box_moov_Data(0, this.audioHeader, this.videoHeader, nil, nil)
	if err != nil {
		err = errors.New(fmt.Sprintf("create moov failed:%s", err.Error()))
		return
	}
	buf.Write(moov)

	segData = buf.Bytes()

	return
}
