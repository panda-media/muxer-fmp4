package commonBoxes

import "github.com/panda-media/muxer-fmp4/format/MP4"

func smhdBox()(box *MP4.MP4Box,err error)  {
	box,err=MP4.NewMP4Box("smhd")
	if err!=nil{
		return
	}
	box.SetVersionFlags(0,0)
	box.Push4Bytes(0)
	return
}
