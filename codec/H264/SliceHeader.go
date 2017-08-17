package H264

import "github.com/panda-media/muxer-fmp4/utils"

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

func decodeSliceHeader(reader *utils.BitReader, sps *SPS, nalType int) (header *slice_header) {
	header = &slice_header{}
	header.first_mb_in_slice = reader.ReadUE_GolombCode()
	header.slice_type = reader.ReadUE_GolombCode()
	header.pic_parameter_set_id = reader.ReadUE_GolombCode()
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
