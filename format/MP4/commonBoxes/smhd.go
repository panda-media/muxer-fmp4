package commonBoxes

func smhdBox() (box *MP4Box, err error) {
	box, err = NewMP4Box("smhd")
	if err != nil {
		return
	}
	box.SetVersionFlags(0, 0)
	box.Push4Bytes(0)
	return
}
