package commonBoxes

import "github.com/panda-media/muxer-fmp4/format/MP4"

func ftypBox() (box *MP4.MP4Box, err error) {
	box, err = MP4.NewMP4Box("ftyp")
	box.PushBytes([]byte("iso5"))
	box.Push4Bytes(1)
	box.PushBytes([]byte("iso5"))
	box.PushBytes([]byte("dash"))
	return
}

func Box_ftyp_Data() (data []byte, err error) {
	box, err := ftypBox()
	if err != nil {
		return
	}
	data = box.Flush()
	return
}
