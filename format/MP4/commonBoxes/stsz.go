package commonBoxes

import (
	"github.com/panda-media/muxer-fmp4/format/MP4"
)

func stszBox(param *MP4.STSZ) (box *MP4.MP4Box, err error) {
	box, err = MP4.NewMP4Box("stsz")
	if err != nil {
		return
	}
	box.SetVersionFlags(0, 0)
	box.Push4Bytes(param.SampleSize)
	box.Push4Bytes(uint32(param.Vaules.Len()))
	if param.SampleSize == 0 {
		for e := param.Vaules.Front(); e != nil; e = e.Next() {
			box.Push4Bytes(e.Value.(uint32))
		}
	}
	return
}
