package H264

import (
	"bytes"
	"github.com/panda-media/muxer-fmp4/utils"
)

const (
	NAL_SLICE           = 1
	NAL_DPA             = 2
	NAL_DPB             = 3
	NAL_DPC             = 4
	NAL_IDR_SLICE       = 5
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

	bit := &utils.BitReader{}
	bit.Init(data)
	bit.ReadBits(8)
	profile_idc := bit.ReadBits(8)
	bit.ReadBits(16)
	bit.ReadExponentialGolombCode()

	if profile_idc == 100 || profile_idc == 110 ||
		profile_idc == 122 || profile_idc == 244 ||
		profile_idc == 44 || profile_idc == 83 ||
		profile_idc == 86 || profile_idc == 118 {
		chroma_format_idc = byte(bit.ReadExponentialGolombCode())
		if chroma_format_idc == 3 {
			bit.ReadBit()
		}
		bit_depth_luma_minus8 = byte(bit.ReadExponentialGolombCode())   //bit_depth_luma_minus
		bit_depth_chroma_minus8 = byte(bit.ReadExponentialGolombCode()) //bit_depth_chroma_minus8
		bit.ReadBit()
		seq_scaling_matrix_present_flag := bit.ReadBit()
		if seq_scaling_matrix_present_flag != 0 {
			for i := 0; i < 8; i++ {
				seq_scaling_list_present_flag := bit.ReadBit()
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
						delta_scale := bit.ReadSE()
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
	}

	bit.ReadExponentialGolombCode()
	pic_order_cnt_type := bit.ReadExponentialGolombCode()
	if 0 == pic_order_cnt_type {
		bit.ReadExponentialGolombCode()
	} else if 1 == pic_order_cnt_type {
		bit.ReadBit()
		bit.ReadSE()
		bit.ReadSE()
		num_ref_frames_in_pic_order_cnt_cycle := bit.ReadExponentialGolombCode()
		for i := 0; i < num_ref_frames_in_pic_order_cnt_cycle; i++ {
			bit.ReadSE()
		}
	}

	bit.ReadExponentialGolombCode()
	bit.ReadBit()
	pic_width_in_mbs_minus1 := bit.ReadExponentialGolombCode()
	pic_height_in_map_units_minus1 := bit.ReadExponentialGolombCode()
	frame_mbs_only_flag := bit.ReadBit()
	if frame_mbs_only_flag == 0 {
		bit.ReadBit()
	}
	bit.ReadBit()
	frame_cropping_flag := bit.ReadBit()
	var frame_crop_left_offset int
	var frame_crop_right_offset int
	var frame_crop_top_offset int
	var frame_crop_bottom_offset int
	if frame_cropping_flag != 0 {
		frame_crop_left_offset = bit.ReadExponentialGolombCode()
		frame_crop_right_offset = bit.ReadExponentialGolombCode()
		frame_crop_top_offset = bit.ReadExponentialGolombCode()
		frame_crop_bottom_offset = bit.ReadExponentialGolombCode()
	}

	width = ((pic_width_in_mbs_minus1 + 1) * 16) - frame_crop_bottom_offset*2 - frame_crop_top_offset*2
	height = ((2 - frame_mbs_only_flag) * (pic_height_in_map_units_minus1 + 1) * 16) - (frame_crop_right_offset * 2) - (frame_crop_left_offset * 2)

	vui_parameters_present_flag := bit.ReadBit()
	if vui_parameters_present_flag != 0 {
		aspect_ratio_info_present_flag := bit.ReadBit()
		if aspect_ratio_info_present_flag != 0 {
			aspect_ratio_idc := bit.ReadBits(8)
			if aspect_ratio_idc == 255 {
				bit.ReadBits(16)
				bit.ReadBits(16)
			}
		}
		overscan_info_present_flag := bit.ReadBit()
		if 0 != overscan_info_present_flag {
			bit.ReadBit()
		}
		video_signal_type_present_flag := bit.ReadBit()
		if video_signal_type_present_flag != 0 {
			bit.ReadBits(3)
			bit.ReadBit()
			colour_description_present_flag := bit.ReadBit()
			if colour_description_present_flag != 0 {
				bit.ReadBits(8)
				bit.ReadBits(8)
				bit.ReadBits(8)
			}
		}
		chroma_loc_info_present_flag := bit.ReadBit()
		if chroma_loc_info_present_flag != 0 {
			bit.ReadExponentialGolombCode()
			bit.ReadExponentialGolombCode()
		}

		timing_info_present_flag := bit.ReadBit()
		if 0 != timing_info_present_flag {
			num_units_in_tick := bit.ReadBits(32)
			time_scale := bit.ReadBits(32)
			fps = time_scale / (2 * num_units_in_tick)
		}
	}
	return
}
