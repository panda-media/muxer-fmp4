package commonBoxes

import (
	"container/list"
	"errors"
	"github.com/panda-media/muxer-fmp4/format/MP4"
)

type SidxReference struct {
	Reference_type      byte
	Referenced_size     uint32
	Subsegment_duration uint32
	Starts_with_SAP     byte
	SAP_type            byte
	SAP_delta_time      uint32
}

type SidxParam struct {
	Version                    byte
	Reference_ID               uint32
	TimeScale                  uint32
	Earliest_presentation_time uint64
	First_offset               uint64
	Reference_count            uint16
	References                 *list.List
}

func SidxBox(param *SidxParam) (box *MP4.MP4Box, err error) {
	if nil == param {
		err = errors.New("nil pointer for init sidx")
		return
	}
	box, err = MP4.NewMP4Box("sid	x")
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
		v := e.Value.(*SidxReference)
		box.Push4Bytes(((uint32(v.Reference_type) << 31) | v.Referenced_size))
		box.Push4Bytes(v.Subsegment_duration)
		box.Push4Bytes((uint32(v.Starts_with_SAP) << 31) | ((uint32(v.SAP_type)) << 28) | v.SAP_delta_time)
	}

	return
}
