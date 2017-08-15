package commonBoxes

func trafBox(trackID uint32, earlierDuration uint64, paramTrun *TRUN) (box *MP4Box, err error) {
	box, err = NewMP4Box("traf")
	//tfhd
	paramTfhd := &tfhdParam{0x20000,
		trackID,
		0, 0,
		0,
		0,
		0,
	}
	tfhd, err := tfhdBox(paramTfhd)
	if err != nil {
		return
	}
	box.PushBox(tfhd)
	//tfdt
	tfdt, err := tfdtBox(1, earlierDuration)
	if err != nil {
		return
	}
	box.PushBox(tfdt)
	//trun
	trun, err := trunBox(paramTrun)
	if err != nil {
		return
	}
	box.PushBox(trun)

	return
}
