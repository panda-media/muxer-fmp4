package commonBoxes

import (
	"errors"
)


func SidxBox(param *SIDX) (box *MP4Box, err error) {
	if nil == param {
		err = errors.New("nil pointer for init sidx")
		return
	}
	box, err = NewMP4Box("sid	x")
	if err != nil {
		return
	}
	box.SetVersionFlags(param.Version, 0)
	box.Push4Bytes(param.Reference_ID)
	box.Push4Bytes(param.TimeScale)

	if param.Version == 0 {
		box.Push4Bytes(uint32(param.Earliest_presentation_time))
		box.Push4Bytes(uint32(param.First_offset))
	} else {
		box.Push8Bytes(param.Earliest_presentation_time)
		box.Push8Bytes(param.First_offset)
	}
	box.Push2Bytes(0)
	box.Push2Bytes(param.Reference_count)

	for e := param.References.Front(); e != nil; e = e.Next() {
		v := e.Value.(*SIDX_REFERENCE)
		box.Push4Bytes(((uint32(v.Reference_type) << 31) | v.Referenced_size))
		box.Push4Bytes(v.Subsegment_duration)
		box.Push4Bytes((uint32(v.Starts_with_SAP) << 31) | ((uint32(v.SAP_type)) << 28) | v.SAP_delta_time)
	}

	return
}
