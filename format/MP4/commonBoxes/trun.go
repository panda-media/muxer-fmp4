package commonBoxes

import (
	"container/list"
	"errors"
	"github.com/panda-media/muxer-fmp4/format/MP4"
)

type TRUN_ARRAY_FIELDS struct {
	Sample_duration                uint32
	Sample_size                    uint32
	Sample_flags                   uint32
	Sample_composition_time_offset uint32
}

func trunBox(version byte, tr_flags int, sample_count, data_offset, first_sample_flags uint32, lst *list.List) (box *MP4.MP4Box, err error) {
	box, err = MP4.NewMP4Box("trun")
	box.SetVersionFlags(version, tr_flags)
	box.Push4Bytes(sample_count)
	if (tr_flags & 0x1) != 0 {
		box.Push4Bytes(data_offset)
	}
	if (tr_flags & 0x4) != 0 {
		box.Push4Bytes(first_sample_flags)
	}

	if lst == nil || sample_count != uint32(lst.Len()) {
		err = errors.New("invalid param in trun sample_count not equal data")
		return
	}

	for e := lst.Front(); e != nil; e = e.Next() {
		v := e.Value.(*TRUN_ARRAY_FIELDS)
		if (tr_flags & 0x100) != 0 {
			box.Push4Bytes(v.Sample_duration)
		}
		if (tr_flags & 0x200) != 0 {
			box.Push4Bytes(v.Sample_size)
		}
		if (tr_flags & 0x400) != 0 {
			box.Push4Bytes(v.Sample_flags)
		}
		if (tr_flags & 0x800) != 0 {
			box.Push4Bytes(v.Sample_composition_time_offset)
		}
	}

	return
}
