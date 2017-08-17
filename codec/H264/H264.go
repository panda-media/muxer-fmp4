package H264

import (
	"bytes"
	"logger"
)

const (
	NAL_SLICE           = 1	//has slice header
	NAL_DPA             = 2		//has slice header
	NAL_DPB             = 3
	NAL_DPC             = 4
	NAL_IDR_SLICE       = 5	//has slice header
	NAL_SEI             = 6
	NAL_SPS             = 7
	NAL_PPS             = 8
	NAL_AUD             = 9
	NAL_END_SEQUENCE    = 10
	NAL_END_STREAM      = 11
	NAL_FILLER_DATA     = 12
	NAL_SPS_EXT         = 13
	NAL_AUXILIARY_SLICE = 19
)


func emulation_prevention(nal []byte) []byte {
	buf := bytes.Buffer{}
	for i := 0; i < len(nal); i++ {
		if i+2 < len(nal) {
			if 0 == (nal[i]^0x00)+(nal[i+1]^0x00)+(nal[i+2]^0x03) {
				buf.WriteByte(nal[i])
				i++
				buf.WriteByte(nal[i])
				i++
				continue
			}
		}
		buf.WriteByte(nal[i])
	}
	return buf.Bytes()
}

func DecodeSPS(sps []byte) (width, height, fps int, chroma_format_idc, bit_depth_luma_minus8, bit_depth_chroma_minus8 byte) {

	data := emulation_prevention(sps)

	spsv := decodeSPS_RBSP(data[1:])
	width=spsv.width
	height=spsv.height
	if spsv.vui!=nil{
		fps= spsv.vui.time_scale/(spsv.vui.num_units_in_tick*2)
	}
	chroma_format_idc=byte(spsv.chroma_format_idc)
	bit_depth_chroma_minus8=byte(spsv.bit_depth_chroma_minus8)
	bit_depth_luma_minus8=byte(spsv.bit_depth_luma_minus8)

	return
}

type H264TimeCalculator struct {
	sps                *SPS
	last_i_frame_counter   int64
	frame_counter      int64
	fps                int64
	frame_duration     int64
	duration_remainder int64
}

func (this *H264TimeCalculator)SetSPS(sps []byte){
	data := emulation_prevention(sps)
	this.sps=decodeSPS_RBSP(data[1:])
	if this.sps.time_scale>0{
		this.fps=int64(this.sps.time_scale/(this.sps.num_units_in_tick*2))
		this.frame_duration=1000/this.fps
		this.duration_remainder=1000%this.fps
	}
}

func (this *H264TimeCalculator)AddNal(data []byte)(pts, dts int64,bFrame bool){
	nalType:=data[0]&0x1f
	if this.sps==nil||this.sps.time_scale==0{
		return 0,0,false
	}
	if nalType==NAL_SLICE||nalType==NAL_DPA||
	nalType==NAL_IDR_SLICE{
		pts =this.frame_counter*this.frame_duration
		header:=decodeNalSliceHeader(data,this.sps)
		if NAL_IDR_SLICE==nalType{
			this.last_i_frame_counter = this.frame_counter
		}
		dts=this.frame_duration*(int64(header.pic_order_cnt_lsb/2)+this.last_i_frame_counter)
加一个80?
		没有修正
		logger.LOGD(dts,header.pic_order_cnt_lsb/2,pic_type(header.slice_type))

		this.frame_counter++
	}else{
		logger.LOGD(nalType)
		return 0,0,false
	}
	return
}
