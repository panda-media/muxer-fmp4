package commonBoxes

import (
	"container/list"
	"github.com/panda-media/muxer-fmp4/format/MP4"
)

func stszBox(sample_size, sample_count uint32, entry_sizes *list.List) (box *MP4.MP4Box, err error) {
	box, err = MP4.NewMP4Box("stsz")
	if err != nil {
		return
	}
	box.SetVersionFlags(0, 0)
	box.Push4Bytes(sample_size)
	box.Push4Bytes(sample_count)
	if sample_size == 0 {
		for e := entry_sizes.Front(); e != nil; e = e.Next() {
			box.Push4Bytes(e.Value.(uint32))
		}
	}
	return
}
