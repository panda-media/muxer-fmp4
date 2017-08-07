package H264

import "container/list"

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

type AVCDecoderConfigurationRecord struct {
	ConfigurationVersion       byte //aways 1
	AVCProfileIndication       byte //sps[1]
	Profile_compatibility      byte //sps[2]
	AVCLevelIndication         byte //sps[3]
	LengthSizeMinusOne         byte //
	NumOfSequenceParameterSets int
	SPS                        *list.List
	NumOfPictureParameterSets int
	PPS *list.List
	Chroma_format byte
	Bit_depth_luma_minus8 byte
	Bit_depth_chroma_minus8 byte
	NumOfSequenceParameterSetExt int
	SequenceParameterSetExt *list.List
}
