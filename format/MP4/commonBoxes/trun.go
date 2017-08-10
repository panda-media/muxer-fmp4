package commonBoxes

import (
	"errors"
	"github.com/panda-media/muxer-fmp4/format/MP4"
)

func trunBox(param *MP4.TRUN) (box *MP4.MP4Box, err error) {
	box, err = MP4.NewMP4Box("trun")
	box.SetVersionFlags(param.Version, param.Tr_flags)
	box.Push4Bytes(param.Sample_count)
	if (param.Tr_flags & 0x1) != 0 {
		box.Push4Bytes(param.Data_offset)
	}
	if (param.Tr_flags & 0x4) != 0 {
		box.Push4Bytes(param.First_sample_flags)
	}

	if param.Vals == nil || param.Sample_count != uint32(param.Vals.Len()) {
		err = errors.New("invalid param in trun sample_count not equal data")
		return
	}

	for e := param.Vals.Front(); e != nil; e = e.Next() {
		v := e.Value.(*MP4.TRUN_ARRAY_FIELDS)
		if (param.Tr_flags & 0x100) != 0 {
			box.Push4Bytes(v.Sample_duration)
		}
		if (param.Tr_flags & 0x200) != 0 {
			box.Push4Bytes(v.Sample_size)
		}
		if (param.Tr_flags & 0x400) != 0 {
			box.Push4Bytes(v.Sample_flags)
		}
		if (param.Tr_flags & 0x800) != 0 {
			box.Push4Bytes(v.Sample_composition_time_offset)
		}
	}

	return
}
