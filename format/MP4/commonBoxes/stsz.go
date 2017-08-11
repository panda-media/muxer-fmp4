package commonBoxes

func stszBox(param *STSZ) (box *MP4Box, err error) {
	box, err = NewMP4Box("stsz")
	if err != nil {
		return
	}
	box.SetVersionFlags(0, 0)
	if param != nil {
		box.Push4Bytes(param.SampleSize)
		box.Push4Bytes(uint32(param.Vaules.Len()))
		if param.SampleSize == 0 {
			for e := param.Vaules.Front(); e != nil; e = e.Next() {
				box.Push4Bytes(e.Value.(uint32))
			}
		}
	} else {
		box.Push4Bytes(0)
		box.Push4Bytes(0)
	}
	return
}
