package MP4

import (
	"container/list"
	"errors"
	"github.com/panda-media/muxer-fmp4/codec/H264"
	"github.com/panda-media/muxer-fmp4/format/AVPacket"
	"github.com/panda-media/muxer-fmp4/format/MP4/commonBoxes"
	"strconv"
	"bytes"
)

func (this *FMP4Muxer) AddPacket(packet *AVPacket.MediaPacket) (err error) {
	if nil == packet {
		err = errors.New("nil packet into FMP4 muxer")
		return
	}
	switch packet.PacketType {
	case AVPacket.AV_PACKET_TYPE_AUDIO:
		soundFormat := packet.Data[0] >> 4
		switch soundFormat {
		case AVPacket.SoundFormat_AAC:
			err = this.addAAC(packet)
			if err != nil {
				return
			}
		default:
			err = errors.New("not support audio format " + strconv.Itoa(int(soundFormat)))
			return
		}
	case AVPacket.AV_PACKET_TYPE_VIDEO:
		videoFormat := packet.Data[0] & 0xf
		switch videoFormat {
		case AVPacket.CodecID_AVC:
			err = this.addH264(packet)
			if err != nil {
				return
			}
		default:
			err = errors.New("not support video codec " + strconv.Itoa(int(videoFormat)))
			return
		}
	default:
		err = errors.New("invalid packet type" + strconv.Itoa(packet.PacketType))
		return
	}
	if 0 == this.timeBeginMS {
		this.timeBeginMS = packet.TimeStamp
		this.timeSlicedMS = packet.TimeStamp
		this.timeSidxMS = packet.TimeStamp
	}
	this.timeLastMS = packet.TimeStamp
	return
}

func (this *FMP4Muxer) Flush() (sidx, moof_mdats []byte, duration, bitrate int, err error) {
	defer func() {
		this.timeSidxMS = this.timeLastMS
		this.moof_mdat_buf.Reset()
		this.moof_mdat_buf=new(bytes.Buffer)
		this.mdat_size = 0
	}()
	if this.audio_data.Len() > 0 || this.video_data.Len() > 0 {
		err = this.sliceKeyFrame()
		if err != nil {
			err = errors.New("slice fmp4 in flush failed:" + err.Error())
			return
		}
	}
	moof_mdats = this.moof_mdat_buf.Bytes()
	sidx, err = commonBoxes.Box_sidx_data(this.sidx)
	if err != nil {
		err = errors.New("flush fmp4,sidx err" + err.Error())
		return
	}

	//duration
	duration = int(this.timeLastMS - this.timeSidxMS)
	if duration == 0 {
		duration = 1
	}
	//bitrate
	bitrate = 1000 * this.mdat_size * 8 / duration
	return
}

func (this *FMP4Muxer)Duration()int{
	return int(this.timeLastMS-this.timeSidxMS)
}

func (this *FMP4Muxer) sliceKeyFrame() (err error) {
	defer func() {
		this.timeSlicedMS = this.timeLastMS
		this.sidx.Reference_count = 0
		this.sidx.References = list.New()
		this.audio_data.Reset()
		this.video_data.Reset()
		this.sequence_number++
	}()
	//moof
	moofData, err := this.moof(0, false)
	if err != nil {
		return
	}
	moofSize := uint32(len(moofData))
	moofData, err = this.moof(moofSize, true)
	if err != nil {
		return
	}
	this.moof_mdat_buf.Write(moofData)
	//mdat
	mdat, err := commonBoxes.NewMP4Box("mdat")
	if err != nil {
		return
	}

	if this.audioHeader != nil && this.audio_data.Len() > 0 {
		mdat.PushBytes(this.audio_data.Bytes())
	}
	if this.videoHeader != nil && this.video_data.Len() > 0 {
		mdat.PushBytes(this.video_data.Bytes())
	}
	mdatData := mdat.Flush()
	this.moof_mdat_buf.Write(mdatData)
	this.mdat_size += len(mdatData)
	refSize := uint32(len(mdatData) + len(moofData))
	//sidx
	this.addSIDX(refSize, this.timeTranslate(uint64(this.timeSidxMS-this.timeBeginMS), commonBoxes.VIDE_TIME_SCALE_Millisecond, this.timescale))
	return
}

func (this *FMP4Muxer) moof(moofSize uint32, reset bool) (moofData []byte, err error) {
	var earlierDurationA, earlierDurationV uint64
	var paramTrunA, paramTrunV *commonBoxes.TRUN
	data_offset := moofSize + 8
	if this.audioHeader != nil {
		earlierDurationA = this.timeTranslate(uint64(this.timeSlicedMS-this.timeBeginMS), commonBoxes.VIDE_TIME_SCALE_Millisecond, this.timescaleAudio)
		paramTrunA = this.trunAudio.Copy()
		paramTrunA.Data_offset = data_offset
		data_offset += uint32(this.audio_data.Len())
		if reset {
			this.trunAudio.Sample_count = 0
			this.trunAudio.Vals = list.New()
		}
	}
	if this.videoHeader != nil {
		earlierDurationV = this.timeTranslate(uint64(this.timeSlicedMS-this.timeBeginMS), commonBoxes.VIDE_TIME_SCALE_Millisecond, this.timescaleVideo)
		paramTrunV = this.trunVideo.Copy()
		paramTrunV.Data_offset = data_offset
		if reset {
			this.trunVideo.Sample_count = 0
			this.trunVideo.Vals = list.New()
		}
	}
	moofData, err = commonBoxes.Box_moof_Data(this.sequence_number, earlierDurationA, paramTrunA, earlierDurationV, paramTrunV)

	return
}

func (this *FMP4Muxer) addSIDX(reference_size uint32, earliest_presentation_time uint64) {
	this.sidx.Version = 1
	this.sidx.Reference_ID = 1
	this.sidx.TimeScale = this.timescale
	this.sidx.Earliest_presentation_time = earliest_presentation_time

	sidxData := commonBoxes.SIDX_REFERENCE{}
	sidxData.Reference_type = 0
	//sidxData.Referenced_size;//moof+mdat
	sidxData.Referenced_size = reference_size
	sidxData.Subsegment_duration = 0
	if this.sidx.Reference_count == 0 {
		sidxData.Starts_with_SAP = 1
	}
	//aways 1,key frame start
	sidxData.Starts_with_SAP = 1
	sidxData.SAP_type = 0
	sidxData.SAP_delta_time = 0
	this.sidx.References.PushBack(sidxData)
	this.sidx.Reference_count = uint16(this.sidx.References.Len())
}

func (this *FMP4Muxer) addH264(packet *AVPacket.MediaPacket) (err error) {
	if this.trunVideo == nil {
		return errors.New("video track not inited")
	}
	if packet.Data[1] == 0 {
		//AVC  parse,continue
		return
	}
	if packet.Data[1] == 2 {
		//end
		return
	}

	nalType := packet.Data[9] & 0x1f
	sampleSize := 0
	if nalType == H264.NAL_IDR_SLICE || nalType == H264.NAL_SEI {
		//add sps pps
		avc, err := H264.DecodeAVC(this.videoHeader.Data[5:])
		if err == nil {
			sampleSize += this.addSPSPPS(avc)
		} else {
			//logger.LOGE(err.Error())
		}
	}

	this.video_data.Write(packet.Data[5:])
	sampleSize += len(packet.Data[5:])
	compositionTime := uint32(0)
	compositionTime |= uint32(packet.Data[2]) << 16
	compositionTime |= uint32(packet.Data[3]) << 8
	compositionTime |= uint32(packet.Data[4]) << 0

	//logger.LOGD(packet.TimeStamp,compositionTime+packet.TimeStamp)
	trunData := &commonBoxes.TRUN_ARRAY_FIELDS{}
	trunData.Sample_size = uint32(sampleSize)
	trunData.Sample_flags = 0
	var duration uint32
	duration = 10
	if this.timeLastVideo != 0 {
		duration = uint32(packet.TimeStamp - this.timeLastVideo)
	}
	this.timeLastVideo = packet.TimeStamp
	trunData.Sample_duration = duration
	trunData.Sample_composition_time_offset = compositionTime
	this.trunVideo.Vals.PushBack(trunData)
	this.trunVideo.Sample_count = uint32(this.trunVideo.Vals.Len())

	if nalType == H264.NAL_IDR_SLICE && this.trunVideo.Vals.Len() > 1 {
		err = this.sliceKeyFrame()
		if err != nil {
			return
		}
	}

	return
}

func (this *FMP4Muxer) addAAC(packet *AVPacket.MediaPacket) (err error) {
	if this.trunAudio == nil {
		return errors.New("audio track not inited")
	}

	this.audio_data.Write(packet.Data[2:])
	trunData := &commonBoxes.TRUN_ARRAY_FIELDS{}
	trunData.Sample_size = uint32(len(packet.Data[2:]))
	this.trunAudio.Vals.PushBack(trunData)
	this.trunAudio.Sample_count = uint32(this.trunAudio.Vals.Len())
	return
}

func (this *FMP4Muxer) addSPSPPS(avc *H264.AVCDecoderConfigurationRecord) (size int) {
	for e := avc.SPS.Front(); e != nil; e = e.Next() {
		frameSize := len(e.Value.([]byte))
		this.video_data.WriteByte(byte((frameSize >> 24) & 0xff))
		this.video_data.WriteByte(byte((frameSize >> 16) & 0xff))
		this.video_data.WriteByte(byte((frameSize >> 8) & 0xff))
		this.video_data.WriteByte(byte((frameSize >> 0) & 0xff))
		this.video_data.Write(e.Value.([]byte))
		size += frameSize + 4
	}
	for e := avc.PPS.Front(); e != nil; e = e.Next() {
		frameSize := len(e.Value.([]byte))
		this.video_data.WriteByte(byte((frameSize >> 24) & 0xff))
		this.video_data.WriteByte(byte((frameSize >> 16) & 0xff))
		this.video_data.WriteByte(byte((frameSize >> 8) & 0xff))
		this.video_data.WriteByte(byte((frameSize >> 0) & 0xff))
		this.video_data.Write(e.Value.([]byte))
		size += frameSize + 4
	}

	return
}

func (this *FMP4Muxer) timeTranslate(src uint64, srcScale, dstScale uint32) (dst uint64) {
	dst = src * uint64(dstScale) / uint64(srcScale)
	return
}
