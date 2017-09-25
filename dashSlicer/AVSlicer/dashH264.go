package AVSlicer

import (
	"container/list"
	"fmt"
	"github.com/panda-media/muxer-fmp4/codec/H264"
	"github.com/panda-media/muxer-fmp4/format/AVPacket"
)

type SlicerH264 struct {
	sps      []byte
	pps      []byte
	sei      []byte
	sps_ext  []byte
	avcGeted bool
	keyframed bool
	dp_data  *list.List
	width    int
	height   int
	fps      int
	codec    string
}

func (this *SlicerH264) Init(fps int) {
	this.fps = fps
}

func (this *SlicerH264) AddNals(data []byte, timestamp int64) (tags *list.List, err error) {
	nals := this.separateNals(data)
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
		tag, err = this.AddNal(nal, timestamp)
		if nil != tag {
			tags.PushBack(tag)
		}
	}
	return
}

func (this *SlicerH264)AddFrame(data []byte,timestamp int64,compositionTime int)(tags *list.List,err error){
	nals:=this.getNalsbySize(data)
	if nals==nil||nals.Len()==0{
		return
	}
	tags=list.New()
	for e := nals.Front(); e != nil; e = e.Next() {
		nal := e.Value.([]byte)

		zero := nal[0] >> 7
		if 0 != zero {
			continue
		}
		var tag *AVPacket.MediaPacket
		tag, err = this.AddNal(nal, timestamp)
		if nil != tag {
			tag.Data[2]=byte((compositionTime>>16)&0xff)
			tag.Data[3]=byte((compositionTime>>8)&0xff)
			tag.Data[4]=byte((compositionTime>>0)&0xff)
			tags.PushBack(tag)
		}
	}
	return
}

func (this *SlicerH264) AddNal(nal []byte, timestamp int64) (tag *AVPacket.MediaPacket, err error) {
	nalType := nal[0] & 0x1f
	switch nalType {
	case H264.NAL_SPS:
		this.sps = make([]byte, len(nal))
		copy(this.sps, nal)
		this.width, this.height, _, _, _, _ = H264.DecodeSPS(this.sps)
		this.codec = fmt.Sprintf("avc1.%02x%02x%02x", int(this.sps[1]), int(this.sps[2]), int(this.sps[3]))
	case H264.NAL_PPS:
		this.pps = make([]byte, len(nal))
		copy(this.pps, nal)
		if false == this.avcGeted {
			tag = this.createAVCTag(timestamp)
			if nil == tag {
				break
			}
			this.avcGeted = true
		}
	case H264.NAL_SEI:
		this.sei = make([]byte, len(nal))
		copy(this.sei, nal)
	case H264.NAL_SPS_EXT:
		this.sps_ext = make([]byte, len(nal))
		copy(this.sps_ext, nal)
	case H264.NAL_IDR_SLICE:
		if this.avcGeted {
			tag = this.createIdrAndSliceTag(nal, timestamp)
			this.keyframed=true
		}
	case H264.NAL_SLICE:
		if this.keyframed {
			tag = this.createIdrAndSliceTag(nal, timestamp)
		}
	case H264.NAL_DPA:
		this.dp_data = list.New()
		this.dp_data.PushBack(nal)
	case H264.NAL_DPB:
		if nil == this.dp_data || this.dp_data.Len() != 1 {
			break
		}
		this.dp_data.PushBack(nal)
	case H264.NAL_DPC:
		if nil == this.dp_data || this.dp_data.Len() != 1 {
			break
		}
		if this.keyframed {
			this.dp_data.PushBack(nal)
			tag = this.createDPTag(this.dp_data, timestamp)
		}
	}
	return
}

func (this *SlicerH264) separateNals(data []byte) (nals *list.List) {
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

func (this *SlicerH264)getNalsbySize(data []byte)(nals *list.List){
	nalCur:=0
	nals=list.New()
	for nalCur+4<len(data){
		nalSize:=int(data[nalCur])<<24
		nalSize|=int(data[nalCur+1])<<16
		nalSize|=int(data[nalCur+2])<<8
		nalSize|=int(data[nalCur+3])<<0
		nalCur+=4
		nals.PushBack(data[nalCur:nalCur+nalSize])
		nalCur+=nalSize
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

func (this *SlicerH264) createAVCTag(timestamp int64) (tag *AVPacket.MediaPacket) {
	if nil == this.sps || nil == this.pps {
		return nil
	}
	avc := H264.AVCDecoderConfigurationRecord{}
	avc.AddSPS(this.sps)
	avc.AddPPS(this.pps)
	avc.AddSPSExt(this.sps_ext)
	tag = &AVPacket.MediaPacket{}
	tag.PacketType = AVPacket.AV_PACKET_TYPE_VIDEO
	tag.TimeStamp = timestamp
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

func (this *SlicerH264) createIdrAndSliceTag(data []byte, timestamp int64) (tag *AVPacket.MediaPacket) {
	tag = &AVPacket.MediaPacket{}
	tag.PacketType = AVPacket.AV_PACKET_TYPE_VIDEO
	tag.TimeStamp = timestamp

	tag.Data = make([]byte, len(data)+5+4)

	if H264.NAL_IDR_SLICE == data[0]&0x1f {
		tag.Data[0] = 0x17
	} else {
		tag.Data[0] = 0x27
	}
	tag.Data[1] = 1
	cts := 0
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

func (this *SlicerH264) createDPTag(data_dps *list.List, timestamp int64) (tag *AVPacket.MediaPacket) {
	if nil == data_dps || data_dps.Len() != 3 {
		return
	}
	tag = &AVPacket.MediaPacket{}
	tag.PacketType = AVPacket.AV_PACKET_TYPE_VIDEO

	tag.TimeStamp = timestamp
	datasize := 5
	for e := data_dps.Front(); e != nil; e = e.Next() {
		datasize += len(e.Value.([]byte)) + 4
	}
	tag.Data = make([]byte, datasize)
	tag.Data[0] = 0x27
	tag.Data[1] = 1
	cts := 0
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

func (this *SlicerH264) Width() int {
	return this.width
}

func (this *SlicerH264) Height() int {
	return this.height
}

func (this *SlicerH264) FPS() int {
	return this.fps
}

func (this *SlicerH264) Codec() string {
	return this.codec
}
