package commonBoxes

import "fmt"

func sttsBox(param *STTS) (box *MP4Box, err error) {
	box, err = NewMP4Box("stts")
	box.SetVersionFlags(0, 0)
	entry_count := uint32(0)
	if param != nil && param.Values != nil {
		fmt.Println(param)
		entry_count = uint32(param.Values.Len())
	}
	box.Push4Bytes(entry_count)

	if entry_count > 0 {
		for e := param.Values.Front(); e != nil; e = e.Next() {
			v := e.Value.(*TimeToSampleVal)
			box.Push4Bytes(v.SampleCount)
			box.Push4Bytes(v.SampleDelta)
		}
	}
	return
}
