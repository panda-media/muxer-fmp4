package commonBoxes

func ftypBox() (box *MP4Box, err error) {
	box, err = NewMP4Box("ftyp")
	//box.PushBytes([]byte("iso5"))
	//box.Push4Bytes(1)
	//box.PushBytes([]byte("iso5"))
	//box.PushBytes([]byte("dash"))

	box.PushBytes([]byte("isom"))
	box.Push4Bytes(0x00)
	box.PushBytes([]byte("isom"))
	box.PushBytes([]byte("avc1"))
	box.PushBytes([]byte("mp42"))
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
