package commonBoxes

import "github.com/panda-media/muxer-fmp4/format/MP4"

type tfhdParam struct {
	Tf_flags                 int
	Track_ID                 uint32
	Base_data_offset         uint64
	Sample_description_index uint32
	Default_sample_duration  uint32
	Default_sample_size      uint32
	Default_sample_flags     uint32
}

func tfhdBox(param *tfhdParam) (box *MP4.MP4Box, err error) {
	box, err = MP4.NewMP4Box("tfhd")
	if err != nil {
		return
	}
	box.SetVersionFlags(0, param.Tf_flags)
	box.Push4Bytes(param.Track_ID)
	if (param.Tf_flags & 0x000001) != 0 {
		box.Push8Bytes(param.Base_data_offset)
	}
	if (param.Tf_flags & 0x2) != 0 {
		box.Push4Bytes(param.Sample_description_index)
	}
	if (param.Tf_flags & 0x8) != 0 {
		box.Push4Bytes(param.Default_sample_duration)
	}
	if (param.Tf_flags & 0x10) != 0 {
		box.Push4Bytes(param.Default_sample_size)
	}
	if (param.Tf_flags & 0x20) != 0 {
		box.Push4Bytes(param.Default_sample_flags)
	}
	return
}
