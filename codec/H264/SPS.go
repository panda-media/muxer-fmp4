package H264

import (
	"github.com/panda-media/muxer-fmp4/utils"
)

type SPS struct {
	profile_idc                           int
	constraint_set_flags                  int
	level_idc                             int
	chroma_format_idc                     int
	separate_colour_plane_flag            int
	bit_depth_luma_minus8                 int
	bit_depth_chroma_minus8               int
	qpprime_y_zero_transform_bypass_flag  int
	seq_scaling_matrix_present_flag       int
	seq_parameter_set_id                  int
	log2_max_frame_num_minus4             int
	pic_order_cnt_type                    int
	log2_max_pic_order_cnt_lsb_minus4     int
	delta_pic_order_always_zero_flag      int
	offset_for_non_ref_pic                int
	offset_for_top_to_bottom_field        int
	num_ref_frames_in_pic_order_cnt_cycle int
	offset_for_ref_frame                  []int
	max_num_ref_frames                    int
	gaps_in_frame_num_value_allowed_flag  int
	pic_width_in_mbs_minus1               int
	pic_height_in_map_units_minus1        int
	frame_mbs_only_flag                   int
	mb_adaptive_frame_field_flag          int
	direct_8x8_inference_flag             int
	frame_cropping_flag                   int
	frame_crop_left_offset                int
	frame_crop_right_offset               int
	frame_crop_top_offset                 int
	frame_crop_bottom_offset              int
	width                                 int
	height                                int
	vui_parameters_present_flag           int
	vui                                   *VUI
}

type VUI struct {
	aspect_ratio_info_present_flag          int
	aspect_ratio_idc                        int
	sar_width                               int
	sar_height                              int
	overscan_info_present_flag              int
	overscan_appropriate_flag               int
	video_signal_type_present_flag          int
	video_format                            int
	video_full_range_flag                   int
	colour_description_present_flag         int
	colour_primaries                        int
	transfer_characteristics                int
	matrix_coefficients                     int
	chroma_loc_info_present_flag            int
	chroma_sample_loc_type_top_field        int
	chroma_sample_loc_type_bottom_field     int
	timing_info_present_flag                int
	num_units_in_tick                       int
	time_scale                              int
	fixed_frame_rate_flag                   int
	nal_hrd_parameters_present_flag         int
	hrd_nal                                 *HRD
	vcl_hrd_parameters_present_flag         int
	hrd_vcl                                 *HRD
	low_delay_hrd_flag                      int
	pic_struct_present_flag                 int
	bitstream_restriction_flag              int
	motion_vectors_over_pic_boundaries_flag int
	max_bytes_per_pic_denom                 int
	max_bits_per_mb_denom                   int
	log2_max_mv_length_horizontal           int
	log2_max_mv_length_vertical             int
	max_num_reorder_frames                  int
	max_dec_frame_buffering                 int
}

type HRD struct {
	cpb_cnt_minus1 int
	bit_rate_scale int
	cpb_size_scale int

	bit_rate_value_minus1 []int
	cpb_size_value_minus1 []int
	cbr_flag              []int

	initial_cpb_removal_delay_length_minus1 int
	cpb_removal_delay_length_minus1         int
	dpb_output_delay_length_minus1          int
	time_offset_length                      int
}

func decodeSPS_RBSP(nal []byte) (sps *SPS) {
	reader := &utils.BitReader{}
	reader.Init(nal)

	sps = &SPS{}
	sps.profile_idc = reader.ReadBits(8)
	sps.constraint_set_flags = reader.ReadBit() << 0
	sps.constraint_set_flags = reader.ReadBit() << 1
	sps.constraint_set_flags = reader.ReadBit() << 2
	sps.constraint_set_flags = reader.ReadBit() << 3
	sps.constraint_set_flags = reader.ReadBit() << 4
	sps.constraint_set_flags = reader.ReadBit() << 5
	reader.ReadBits(2)

	sps.level_idc = reader.ReadBits(8)
	sps.seq_parameter_set_id = reader.ReadUE_GolombCode()

	if sps.profile_idc == 100 ||
		sps.profile_idc == 110 ||
		sps.profile_idc == 122 ||
		sps.profile_idc == 244 ||
		sps.profile_idc == 44 ||
		sps.profile_idc == 83 ||
		sps.profile_idc == 86 ||
		sps.profile_idc == 118 ||
		sps.profile_idc == 128 ||
		sps.profile_idc == 138 ||
		sps.profile_idc == 144 {
		sps.chroma_format_idc = reader.ReadUE_GolombCode()
		if 3 == sps.chroma_format_idc {
			sps.separate_colour_plane_flag = reader.ReadBit()
		}
		sps.bit_depth_luma_minus8 = reader.ReadUE_GolombCode()
		sps.bit_depth_chroma_minus8 = reader.ReadUE_GolombCode()

		sps.qpprime_y_zero_transform_bypass_flag = reader.ReadBit()
		sps.seq_scaling_matrix_present_flag = reader.ReadBit()
		if sps.seq_scaling_matrix_present_flag == 1 {
			len := 8
			if sps.chroma_format_idc == 3 {
				len = 12
			}
			for i := 0; i < len; i++ {
				seq_scaling_list_present_flag := reader.ReadBit()
				if seq_scaling_list_present_flag != 0 {
					var sizeOfScalingList int
					if i < 6 {
						sizeOfScalingList = 16
					} else {
						sizeOfScalingList = 64
					}
					lastScale := 8
					nextScale := 8
					for j := 0; j < sizeOfScalingList; j++ {
						delta_scale := reader.ReadSE()
						nextScale = (lastScale + delta_scale + 256) % 256
					}
					if nextScale == 0 {
						lastScale = lastScale
					} else {
						lastScale = nextScale
					}
				}
			}
		}
	} else {
		sps.chroma_format_idc = 1
		sps.bit_depth_luma_minus8 = 0
		sps.bit_depth_chroma_minus8 = 0
	}

	sps.log2_max_frame_num_minus4 = reader.ReadUE_GolombCode()
	sps.pic_order_cnt_type = reader.ReadUE_GolombCode()

	switch sps.pic_order_cnt_type {
	case 0:
		sps.log2_max_pic_order_cnt_lsb_minus4 = reader.ReadUE_GolombCode()
	case 1:
		sps.delta_pic_order_always_zero_flag = reader.ReadBit()
		sps.offset_for_non_ref_pic = reader.ReadSE()
		sps.offset_for_top_to_bottom_field = reader.ReadSE()
		sps.num_ref_frames_in_pic_order_cnt_cycle = reader.ReadUE_GolombCode()
		sps.offset_for_ref_frame = make([]int, sps.num_ref_frames_in_pic_order_cnt_cycle)
		for i := 0; i < sps.num_ref_frames_in_pic_order_cnt_cycle; i++ {
			sps.offset_for_ref_frame[i] = reader.ReadSE()
		}
	}

	sps.max_num_ref_frames = reader.ReadUE_GolombCode()
	sps.gaps_in_frame_num_value_allowed_flag = reader.ReadBit()
	sps.pic_width_in_mbs_minus1 = reader.ReadUE_GolombCode()
	sps.pic_height_in_map_units_minus1 = reader.ReadUE_GolombCode()
	sps.frame_mbs_only_flag = reader.ReadBit()
	if 0 == sps.frame_mbs_only_flag {
		sps.mb_adaptive_frame_field_flag = reader.ReadBit()
	}
	sps.direct_8x8_inference_flag = reader.ReadBit()
	sps.frame_cropping_flag = reader.ReadBit()

	if sps.frame_cropping_flag != 0 {
		sps.frame_crop_left_offset = reader.ReadUE_GolombCode()
		sps.frame_crop_right_offset = reader.ReadUE_GolombCode()
		sps.frame_crop_top_offset = reader.ReadUE_GolombCode()
		sps.frame_crop_bottom_offset = reader.ReadUE_GolombCode()
	}

	sps.width = ((sps.pic_width_in_mbs_minus1 + 1) * 16) - (sps.frame_crop_right_offset * 2) - (sps.frame_crop_left_offset * 2)
	sps.height = ((2 - sps.frame_mbs_only_flag) * (sps.pic_height_in_map_units_minus1 + 1) * 16) - sps.frame_crop_bottom_offset*2 - sps.frame_crop_top_offset*2

	sps.vui_parameters_present_flag = reader.ReadBit()
	if sps.vui_parameters_present_flag != 0 {
		sps.vui = decodeVUI(reader)
	}

	return
}

func decodeVUI(reader *utils.BitReader) (vui *VUI) {
	vui = &VUI{}
	vui.aspect_ratio_info_present_flag = reader.ReadBit()
	if vui.aspect_ratio_info_present_flag != 0 {
		vui.aspect_ratio_idc = reader.ReadBits(8)
		if 255 == vui.aspect_ratio_idc {
			vui.sar_width = reader.ReadBits(16)
			vui.sar_height = reader.ReadBits(16)
		}
	}
	vui.overscan_info_present_flag = reader.ReadBit()
	if 0 != vui.overscan_info_present_flag {
		vui.overscan_appropriate_flag = reader.ReadBit()
	}
	vui.video_signal_type_present_flag = reader.ReadBit()
	if 0 != vui.video_signal_type_present_flag {
		vui.video_format = reader.ReadBits(3)
		vui.video_full_range_flag = reader.ReadBits(1)
		vui.colour_description_present_flag = reader.ReadBit()
		if 0 != vui.colour_description_present_flag {
			vui.colour_primaries = reader.ReadBits(8)
			vui.transfer_characteristics = reader.ReadBits(8)
			vui.matrix_coefficients = reader.ReadBits(8)
		}
	}

	vui.chroma_loc_info_present_flag = reader.ReadBit()
	if vui.chroma_loc_info_present_flag != 0 {
		vui.chroma_sample_loc_type_top_field = reader.ReadUE_GolombCode()
		vui.chroma_sample_loc_type_bottom_field = reader.ReadUE_GolombCode()
	}

	vui.timing_info_present_flag = reader.ReadBit()
	if 0 != vui.timing_info_present_flag {
		vui.num_units_in_tick = reader.ReadBits(32)
		vui.time_scale = reader.ReadBits(32)
		vui.fixed_frame_rate_flag = reader.ReadBit()
	}
	vui.nal_hrd_parameters_present_flag = reader.ReadBit()
	if 0 != vui.nal_hrd_parameters_present_flag {
		vui.hrd_nal = decodeHRD(reader)
	}
	vui.vcl_hrd_parameters_present_flag = reader.ReadBit()
	if 0 != vui.vcl_hrd_parameters_present_flag {
		vui.hrd_vcl = decodeHRD(reader)
	}
	if vui.nal_hrd_parameters_present_flag != 0 ||
		vui.vcl_hrd_parameters_present_flag != 0 {
		vui.low_delay_hrd_flag = reader.ReadBit()
	}
	vui.pic_struct_present_flag = reader.ReadBit()
	if reader.BitsLeft() <= 0 {
		return
	}
	vui.bitstream_restriction_flag = reader.ReadBit()
	if 0 != vui.bitstream_restriction_flag {
		vui.motion_vectors_over_pic_boundaries_flag = reader.ReadBit()
		vui.max_bytes_per_pic_denom = reader.ReadUE_GolombCode()
		vui.max_bits_per_mb_denom = reader.ReadUE_GolombCode()
		vui.log2_max_mv_length_horizontal = reader.ReadUE_GolombCode()
		vui.log2_max_mv_length_vertical = reader.ReadUE_GolombCode()
		vui.max_num_reorder_frames = reader.ReadUE_GolombCode()
		vui.max_dec_frame_buffering = reader.ReadUE_GolombCode()
	}
	return
}

func decodeHRD(reader *utils.BitReader) (hrd *HRD) {
	hrd = &HRD{}
	hrd.cpb_cnt_minus1 = reader.ReadUE_GolombCode()
	hrd.bit_rate_scale = reader.ReadBits(4)
	hrd.cpb_size_scale = reader.ReadBits(4)

	hrd.bit_rate_value_minus1 = make([]int, hrd.cpb_cnt_minus1+1)
	hrd.cpb_size_value_minus1 = make([]int, hrd.cpb_cnt_minus1+1)
	hrd.cbr_flag = make([]int, hrd.cpb_cnt_minus1+1)

	for i := 0; i <= hrd.cpb_cnt_minus1; i++ {
		hrd.bit_rate_value_minus1[i] = reader.ReadUE_GolombCode()
		hrd.cpb_size_value_minus1[i] = reader.ReadUE_GolombCode()
		hrd.cbr_flag[i] = reader.ReadBit()
	}

	hrd.initial_cpb_removal_delay_length_minus1 = reader.ReadBits(5)
	hrd.cpb_removal_delay_length_minus1 = reader.ReadBits(5)
	hrd.dpb_output_delay_length_minus1 = reader.ReadBits(5)
	hrd.time_offset_length = reader.ReadBits(5)
	return
}
