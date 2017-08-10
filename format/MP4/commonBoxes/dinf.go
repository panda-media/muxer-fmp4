package commonBoxes

func dinfBox() (box *MP4Box, err error) {
	box, err = NewMP4Box("dinf")
	if err != nil {
		return
	}
	dref, err := drefBox()
	if err != nil {
		return
	}
	box.PushBox(dref)
	return
}

func drefBox() (box *MP4Box, err error) {
	box, err = NewMP4Box("dref")
	if err != nil {
		return
	}
	box.SetVersionFlags(0, 0)
	box.Push4Bytes(1)
	url, err := urlBox()
	if err != nil {
		return
	}
	box.PushBox(url)
	return
}

func urlBox() (box *MP4Box, err error) {
	box, err = NewMP4Box("url ")
	if err != nil {
		return
	}
	box.Push4Bytes(1)
	return
}
