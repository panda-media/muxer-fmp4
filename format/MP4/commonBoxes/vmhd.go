package commonBoxes

func vmhdBox() (box *MP4Box, err error) {
	box, err = NewMP4Box("vmhd")
	if err != nil {
		return
	}
	box.SetVersionFlags(0, 1)
	box.Push8Bytes(0)
	return
}
