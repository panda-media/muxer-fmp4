package commonBoxes

import (
	"github.com/panda-media/muxer-fmp4/format/MP4"
)

func sttsBox(param *MP4.STTS) (box *MP4.MP4Box, err error) {
	box, err = MP4.NewMP4Box("stts")
	box.SetVersionFlags(0, 0)
	entry_count := uint32(0)
	if param.Values != nil {
		entry_count = uint32(param.Values.Len())
	}
	box.Push4Bytes(entry_count)

	if entry_count > 0 {
		for e := param.Values.Front(); e != nil; e = e.Next() {
			v := e.Value.(*MP4.TimeToSampleVal)
			box.Push4Bytes(v.SampleCount)
			box.Push4Bytes(v.SampleDelta)
		}
	}
	return
}
