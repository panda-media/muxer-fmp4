package commonBoxes

import (
	"container/list"
	"github.com/panda-media/muxer-fmp4/format/MP4"
)

func stco_co64(chunkoffsets *list.List, co64 bool) (box *MP4.MP4Box, err error) {
	if co64 {
		box, err = MP4.NewMP4Box("co64")
		if err != nil {
			return
		}
	} else {
		box, err = MP4.NewMP4Box("stco")
		if err != nil {
			return
		}
	}
	box.SetVersionFlags(0, 0)
	entry_count := uint32(0)
	if chunkoffsets != nil {
		entry_count = uint32(chunkoffsets.Len())
	}
	box.Push4Bytes(entry_count)

	if entry_count > 0 {
		if co64 {
			for e := chunkoffsets.Front(); e != nil; e = e.Next() {
				box.Push8Bytes(e.Value.(uint64))
			}
		} else {
			for e := chunkoffsets.Front(); e != nil; e = e.Next() {
				box.Push4Bytes(e.Value.(uint32))
			}
		}
	}

	return
}
