package commonBoxes

import (
	"container/list"
	"github.com/panda-media/muxer-fmp4/format/MP4"
)

type SampleToChunkBox struct {
	First_chunk              uint32
	Samples_per_chunk        uint32
	Sample_description_index uint32
}

func stscBox(sampletochunkboxex *list.List) (box *MP4.MP4Box, err error) {
	box, err = MP4.NewMP4Box("stsc")
	if err != nil {
		return
	}
	entry_count := uint32(0)
	if sampletochunkboxex != nil {
		entry_count = uint32(sampletochunkboxex.Len())
	}
	box.Push4Bytes(entry_count)
	if entry_count > 0 {
		for e := sampletochunkboxex.Front(); e != nil; e = e.Next() {
			v := e.Value.(*SampleToChunkBox)
			box.Push4Bytes(v.First_chunk)
			box.Push4Bytes(v.Samples_per_chunk)
			box.Push4Bytes(v.Sample_description_index)
		}
	}
	return
}
