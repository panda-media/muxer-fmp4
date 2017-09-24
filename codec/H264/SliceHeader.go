package H264

import (
	"github.com/panda-media/muxer-fmp4/utils"
)

type slice_header struct {
	first_mb_in_slice                int
	slice_type                       int
	pic_parameter_set_id             int
	colour_plane_id                  int
	frame_num                        int
	field_pic_flag                   int
	bottom_field_flag                int
	idr_pic_id                       int
	pic_order_cnt_lsb                int
	delta_pic_order_cnt_bottom       int
	delta_pic_order_cnt              []int
	redundant_pic_cnt                int
	direct_spatial_mv_pred_flag      int
	num_ref_idx_active_override_flag int
	num_ref_idx_l0_active_minus1     int
	num_ref_idx_l1_active_minus1     int
}

const (
	H264_PICTURE_TYPE_P = iota
	H264_PICTURE_TYPE_B //double
	H264_PICTURE_TYPE_I
	H264_PICTURE_TYPE_SP
	H264_PICTURE_TYPE_SI
)

func pic_type(slice_type int) int {
	switch slice_type {
	case 0:
		return H264_PICTURE_TYPE_P
	case 1:
		return H264_PICTURE_TYPE_B
	case 2:
		return H264_PICTURE_TYPE_I
	case 3:
		return H264_PICTURE_TYPE_SP
	case 4:
		return H264_PICTURE_TYPE_SI
	case 5:
		return H264_PICTURE_TYPE_P
	case 6:
		return H264_PICTURE_TYPE_B
	case 7:
		return H264_PICTURE_TYPE_I
	case 8:
		return H264_PICTURE_TYPE_SP
	case 9:
		return H264_PICTURE_TYPE_SI
	}
	return -1
}

func decodeNalSliceHeader(data []byte, sps *SPS) (header *slice_header) {
	nalType := int(data[0] & 0x1f)
	dataEmulationPreventioned := emulation_prevention(data)
	reader := &utils.BitReader{}
	reader.Init(dataEmulationPreventioned[1:])
	header = decodeSliceHeader(reader, sps, nalType)
	return
}

func decodeSliceHeader(reader *utils.BitReader, sps *SPS, nalType int) (header *slice_header) {
	header = &slice_header{}
	header.first_mb_in_slice = reader.ReadUE_GolombCode()
	header.slice_type = reader.ReadUE_GolombCode()
	header.pic_parameter_set_id = reader.ReadUE_GolombCode()
	if sps.separate_colour_plane_flag == 1 {
		header.colour_plane_id = reader.ReadBits(2)
	}
	header.frame_num = reader.ReadBits(sps.log2_max_frame_num_minus4 + 4)
	if sps.frame_mbs_only_flag == 0 {
		header.field_pic_flag = reader.ReadBit()
		if header.field_pic_flag != 0 {
			header.bottom_field_flag = reader.ReadBit()
		}
	}
	if NAL_IDR_SLICE == nalType {
		header.idr_pic_id = reader.ReadUE_GolombCode()
	}
	if sps.pic_order_cnt_type == 0 {
		header.pic_order_cnt_lsb = reader.ReadBits(sps.log2_max_pic_order_cnt_lsb_minus4 + 4) //poc enough
	}
	return
}
