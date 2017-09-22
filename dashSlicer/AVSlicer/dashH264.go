package AVSlicer

import (
	"container/list"
	"fmt"
	"github.com/panda-media/muxer-fmp4/codec/H264"
	"github.com/panda-media/muxer-fmp4/format/AVPacket"
	"errors"
)


type SlicerH264 struct {
	sps      []byte
	pps      []byte
	sei      []byte
	spsext   []byte
	avcGeted bool
	dp_datas *list.List
	nalTimer H264.H264TimeCalculator
	width    int
	height   int
	fps      int
	custom_time bool
	codec    string
}

func (this *SlicerH264)Init(custom_time bool,fps int){
	this.custom_time=custom_time
	this.fps=fps
}

func (this *SlicerH264) AddNals(data []byte,timestamp int64) (tags *list.List,err error) {
	nals := this.separateNalus(data)
	if nals == nil || nals.Len() == 0 {
		return
	}
	tags = list.New()
	for e := nals.Front(); e != nil; e = e.Next() {
		nal := e.Value.([]byte)

		zero := nal[0] >> 7
		if 0 != zero {
			continue
		}
		var tag *AVPacket.MediaPacket
		tag ,err= this.AddNal(nal,timestamp)
		if nil != tag {
			tags.PushBack(tag)
		}
	}
	return
}

func (this *SlicerH264)AddNal(nal []byte,timestamp int64)(tag *AVPacket.MediaPacket,err error){
	nalType := nal[0] & 0x1f
	switch nalType {
	case H264.NAL_SPS:
		this.sps = make([]byte, len(nal))
		copy(this.sps, nal)
		var fps int
		this.width, this.height, fps,  _, _,_ = H264.DecodeSPS(this.sps)

		if fps<=0{
			if this.fps<=0{
				err=errors.New("sps  timing_info not found")
				return
			}
		}else{
			if false==this.custom_time{
				this.fps=fps
			}
		}

		this.codec = fmt.Sprintf("avc1.%02x%02x%02x", int(this.sps[1]), int(this.sps[2]), int(this.sps[3]))
	case H264.NAL_PPS:
		this.pps = make([]byte, len(nal))
		copy(this.pps, nal)
		if false==this.avcGeted{
			tag = this.createAVCTag()
			if nil == tag {
				break
			}
			this.avcGeted=true
		}
	case H264.NAL_SEI:
		this.sei = make([]byte, len(nal))
		copy(this.sei, nal)
	case H264.NAL_SPS_EXT:
		this.spsext = make([]byte, len(nal))
		copy(this.spsext, nal)
	case H264.NAL_IDR_SLICE:
		if this.avcGeted{
			tag = this.createIdrSliceTag(nal,timestamp)
		}
	case H264.NAL_SLICE:
		if this.avcGeted{
			tag = this.createIdrSliceTag(nal,timestamp)
		}
	case H264.NAL_DPA:
		this.dp_datas = list.New()
		this.dp_datas.PushBack(nal)
	case H264.NAL_DPB:
		if nil == this.dp_datas || this.dp_datas.Len() != 1 {
			break
		}
		this.dp_datas.PushBack(nal)
	case H264.NAL_DPC:
		if nil == this.dp_datas || this.dp_datas.Len() != 1 {
			break
		}
		if this.avcGeted {
			this.dp_datas.PushBack(nal)
			tag = this.createDPTag(this.dp_datas,timestamp)
		}
	}
	return
}

func (this *SlicerH264) separateNalus(data []byte) (nals *list.List) {
	nalData, dataCur := this.getOneNal1(data[0:])
	nals = list.New()
	for nalData != nil && len(nalData) > 0 {

		nals.PushBack(nalData)
		if dataCur == len(data) {
			break
		}
		nalData, dataCur = this.getOneNal1(data[dataCur:])
	}
	return
}

//separated by 0x00000001 or 0x000001
func (this *SlicerH264) getOneNal1(data []byte) (nalData []byte, dataCur int) {
	//read start
	nalStart := 0
	nalEnd := 0
	for {
		if nalStart+4 >= len(data) {
			return
		}
		if data[nalStart] == 0x00 &&
			data[nalStart+1] == 0x00 {
			if data[nalStart+2] == 0x01 {
				nalStart += 3
				break
			} else if data[nalStart+2] == 0x00 {
				if data[nalStart+3] == 0x01 {
					nalStart += 4
					break
				}
			} else {
				nalStart += 2
			}
		}
		nalStart++
	}
	//read end
	nalEnd = nalStart
	for nalEnd < len(data) {
		if nalEnd+4 > len(data) {
			nalEnd = len(data)
			break
		}
		if data[nalEnd] == 0x00 &&
			data[nalEnd+1] == 0x00 {
			if data[nalEnd+2] == 0x01 {
				break
			} else if data[nalEnd+2] == 0x00 && data[nalEnd+3] == 0x01 {
				break
			} else {
				nalEnd += 2
			}

		}
		nalEnd++
	}
	//nal data
	nalData = data[nalStart:nalEnd]
	dataCur = nalEnd
	return
}

func (this *SlicerH264) createAVCTag() (tag *AVPacket.MediaPacket) {
	if nil == this.sps || nil == this.pps {
		return nil
	}
	this.nalTimer.SetSPS(this.sps,this.fps)
	avc := H264.AVCDecoderConfigurationRecord{}
	avc.AddSPS(this.sps)
	avc.AddPPS(this.pps)
	avc.AddSPSExt(this.spsext)
	tag = &AVPacket.MediaPacket{}
	tag.PacketType = AVPacket.AV_PACKET_TYPE_VIDEO
	tag.TimeStamp = 0
	avcData := avc.AVCData()
	tag.Data = make([]byte, len(avcData)+5)
	tag.Data[0] = 0x17
	tag.Data[1] = 0
	tag.Data[2] = 0
	tag.Data[3] = 0
	tag.Data[4] = 0
	copy(tag.Data[5:], avcData)
	this.avcGeted = true
	return
}

func (this *SlicerH264) createIdrSliceTag(data []byte,timestamp int64) (tag *AVPacket.MediaPacket) {
	tag = &AVPacket.MediaPacket{}
	tag.PacketType = AVPacket.AV_PACKET_TYPE_VIDEO
	pts, cts, _ := this.nalTimer.AddNal(data,timestamp)
	tag.TimeStamp = pts

	tag.Data = make([]byte, len(data)+5+4)

	if H264.NAL_IDR_SLICE == data[0]&0x1f {
		tag.Data[0] = 0x17
	} else {
		tag.Data[0] = 0x27
	}
	tag.Data[1] = 1
	tag.Data[2] = byte((cts >> 16) & 0xff)
	tag.Data[3] = byte((cts >> 8) & 0xff)
	tag.Data[4] = byte((cts >> 0) & 0xff)
	nalSize := len(data)
	tag.Data[5] = byte((nalSize >> 24) & 0xff)
	tag.Data[6] = byte((nalSize >> 16) & 0xff)
	tag.Data[7] = byte((nalSize >> 8) & 0xff)
	tag.Data[8] = byte((nalSize >> 0) & 0xff)
	copy(tag.Data[9:], data)
	return
}

func (this *SlicerH264) createDPTag(data_dps *list.List,timestamp int64) (tag *AVPacket.MediaPacket) {
	if nil == data_dps || data_dps.Len() != 3 {
		return
	}
	tag = &AVPacket.MediaPacket{}
	tag.PacketType = AVPacket.AV_PACKET_TYPE_VIDEO
	pts, cts, _ := this.nalTimer.AddNal(data_dps.Front().Value.([]byte),timestamp)
	tag.TimeStamp = pts
	datasize := 5
	for e := data_dps.Front(); e != nil; e = e.Next() {
		datasize += len(e.Value.([]byte)) + 4
	}
	tag.Data = make([]byte, datasize)
	tag.Data[0] = 0x27
	tag.Data[1] = 1
	tag.Data[2] = byte((cts >> 16) & 0xff)
	tag.Data[3] = byte((cts >> 8) & 0xff)
	tag.Data[4] = byte((cts >> 0) & 0xff)
	cur := 5
	for e := data_dps.Front(); e != nil; e = e.Next() {
		v := e.Value.([]byte)
		nalSize := len(v)
		tag.Data[cur] = byte((nalSize >> 24) & 0xff)
		tag.Data[cur+1] = byte((nalSize >> 16) & 0xff)
		tag.Data[cur+2] = byte((nalSize >> 8) & 0xff)
		tag.Data[cur+4] = byte((nalSize >> 0) & 0xff)
		cur += 4
		copy(tag.Data[cur:], v)
		cur += nalSize
	}
	return
}

func (this *SlicerH264) Width()int{
	return this.width
}

func (this *SlicerH264)Height()int{
	return this.height
}

func (this *SlicerH264)FPS()int{
	return this.fps
}

func (this *SlicerH264)Codec()string{
	return this.codec
}