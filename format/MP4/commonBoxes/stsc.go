package commonBoxes


func stscBox(param *STSC) (box *MP4Box, err error) {
	box, err = NewMP4Box("stsc")
	if err != nil {
		return
	}
	entry_count := uint32(0)
	if param!=nil&&param.Values != nil {
		entry_count = uint32(param.Values.Len())
	}
	box.Push4Bytes(entry_count)
	if entry_count > 0 {
		for e := param.Values.Front(); e != nil; e = e.Next() {
			v := e.Value.(*SampleToChunkVal)
			box.Push4Bytes(v.First_chunk)
			box.Push4Bytes(v.Samples_per_chunk)
			box.Push4Bytes(v.Sample_description_index)
		}
	}
	return
}
