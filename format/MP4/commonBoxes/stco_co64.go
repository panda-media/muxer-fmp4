package commonBoxes

import (
	"github.com/panda-media/muxer-fmp4/format/MP4"
)

func stco_co64(param *MP4.STCO_CO64) (box *MP4.MP4Box, err error) {
	if param.USE_64 {
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
	if param.Chunk_offset != nil {
		entry_count = uint32(param.Chunk_offset.Len())
	}
	box.Push4Bytes(entry_count)

	if entry_count > 0 {
		if param.USE_64 {
			for e := param.Chunk_offset.Front(); e != nil; e = e.Next() {
				box.Push8Bytes(e.Value.(uint64))
			}
		} else {
			for e := param.Chunk_offset.Front(); e != nil; e = e.Next() {
				box.Push4Bytes(e.Value.(uint32))
			}
		}
	}

	return
}
