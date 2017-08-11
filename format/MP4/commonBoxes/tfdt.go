package commonBoxes

func tfdt(version byte, baseMediaDecodeTime uint64) (box *MP4Box, err error) {
	box, err = NewMP4Box("tfdt")
	if err != nil {
		return
	}
	box.SetVersionFlags(version, 0)
	if version == 1 {
		box.Push8Bytes(baseMediaDecodeTime)
	} else {
		box.Push4Bytes(uint32(baseMediaDecodeTime & 0xffffffff))
	}
	return
}
