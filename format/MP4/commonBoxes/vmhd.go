package commonBoxes

import "github.com/panda-media/muxer-fmp4/format/MP4"

func vmhdBox()(box *MP4.MP4Box,err error)  {
	box,err=MP4.NewMP4Box("vmhd")
	if err!=nil{
		return
	}
	box.SetVersionFlags(0,1)
	box.Push8Bytes(0)
	return
}
