package commonBoxes

func mfhdBox(sequence_number uint32) (box *MP4Box, err error) {
	box, err = NewMP4Box("mfhd")
	if err != nil {
		return
	}
	box.SetVersionFlags(0, 0)
	box.Push4Bytes(sequence_number)
	return
}
