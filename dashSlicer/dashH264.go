package dashSlicer

import (
	"container/list"
	"fmt"
	"github.com/panda-media/muxer-fmp4/codec/H264"
	"github.com/panda-media/muxer-fmp4/format/AVPacket"
)

type dashH264 struct {
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
	codec    string
}

func (this *dashH264) addNals(data []byte) (tags *list.List) {
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
		nalType := nal[0] & 0x1f
		switch nalType {
		case H264.NAL_SPS:
			this.sps = make([]byte, len(nal))
			copy(this.sps, nal)
			this.width, this.height, this.fps, _, _, _ = H264.DecodeSPS(this.sps)
			this.codec = fmt.Sprintf("avc1.%02x%02x%02x", int(this.sps[1]), int(this.sps[2]), int(this.sps[3]))
		case H264.NAL_PPS:
			this.pps = make([]byte, len(nal))
			copy(this.pps, nal)
		case H264.NAL_SEI:
			this.sei = make([]byte, len(nal))
			copy(this.sei, nal)
		case H264.NAL_SPS_EXT:
			this.spsext = make([]byte, len(nal))
			copy(this.spsext, nal)
		case H264.NAL_IDR_SLICE:
			if false == this.avcGeted {
				tag := this.createAVCTag()
				if nil == tag {
					continue
				}
				tags.PushBack(tag)
			}
			tag := this.createIdrSliceTag(nal)
			if nil != tag {
				tags.PushBack(tag)
			}
		case H264.NAL_SLICE:
			if false == this.avcGeted {
				continue
			}
			tag := this.createIdrSliceTag(nal)
			if nil != tag {
				tags.PushBack(tag)
			}
		case H264.NAL_DPA:
			this.dp_datas = list.New()
			this.dp_datas.PushBack(nal)
		case H264.NAL_DPB:
			if nil == this.dp_datas || this.dp_datas.Len() != 1 {
				continue
			}
			this.dp_datas.PushBack(nal)
		case H264.NAL_DPC:
			if nil == this.dp_datas || this.dp_datas.Len() != 1 {
				continue
			}
			if true == this.avcGeted {
				this.dp_datas.PushBack(nal)
				tag := this.createDPTag(this.dp_datas)
				if nil != tag {
					tags.PushBack(tag)
				}
			}
		}
	}
	return
}

func (this *dashH264) separateNalus(data []byte) (nals *list.List) {
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
func (this *dashH264) getOneNal1(data []byte) (nalData []byte, dataCur int) {
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

func (this *dashH264) createAVCTag() (tag *AVPacket.MediaPacket) {
	if nil == this.sps || nil == this.pps {
		return nil
	}
	this.nalTimer.SetSPS(this.sps)
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

func (this *dashH264) createIdrSliceTag(data []byte) (tag *AVPacket.MediaPacket) {
	tag = &AVPacket.MediaPacket{}
	tag.PacketType = AVPacket.AV_PACKET_TYPE_VIDEO
	pts, cts, _ := this.nalTimer.AddNal(data)
	tag.TimeStamp = uint32(pts & 0xffffffff)

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

func (this *dashH264) createDPTag(data_dps *list.List) (tag *AVPacket.MediaPacket) {
	if nil == data_dps || data_dps.Len() != 3 {
		return
	}
	tag = &AVPacket.MediaPacket{}
	tag.PacketType = AVPacket.AV_PACKET_TYPE_VIDEO
	pts, cts, _ := this.nalTimer.AddNal(data_dps.Front().Value.([]byte))
	tag.TimeStamp = uint32(pts & 0xffffffff)
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
