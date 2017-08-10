package commonBoxes

import (
	"github.com/panda-media/muxer-fmp4/format/MP4"
)

func stscBox(param *MP4.STSC) (box *MP4.MP4Box, err error) {
	box, err = MP4.NewMP4Box("stsc")
	if err != nil {
		return
	}
	entry_count := uint32(0)
	if param.Values != nil {
		entry_count = uint32(param.Values.Len())
	}
	box.Push4Bytes(entry_count)
	if entry_count > 0 {
		for e := param.Values.Front(); e != nil; e = e.Next() {
			v := e.Value.(*MP4.SampleToChunkVal)
			box.Push4Bytes(v.First_chunk)
			box.Push4Bytes(v.Samples_per_chunk)
			box.Push4Bytes(v.Sample_description_index)
		}
	}
	return
}
