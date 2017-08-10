package commonBoxes

import "container/list"

//stts

type TimeToSampleVal struct {
	SampleCount uint32
	SampleDelta uint32
}
type STTS struct {
	Values *list.List //*TimeToSampleVal
}

//stsc

type SampleToChunkVal struct {
	First_chunk              uint32
	Samples_per_chunk        uint32
	Sample_description_index uint32
}
type STSC struct {
	Values *list.List //*SampleToChunkVal
}

//stsz  uint32
type STSZ struct {
	SampleSize uint32
	Vaules     *list.List //uint32
}

//stco uint32 co64 uint64
type STCO_CO64 struct {
	USE_64       bool
	Chunk_offset *list.List //
}

//moov needed
type MOOV_ARRAYS struct {
	Stts     *STTS
	Stsc     *STSC
	Stsz     *STSZ
	StcoCo64 *STCO_CO64
}

//sidx
//trun

type TRUN_ARRAY_FIELDS struct {
	Sample_duration                uint32
	Sample_size                    uint32
	Sample_flags                   uint32
	Sample_composition_time_offset uint32
}

type TRUN struct {
	Version            byte
	Tr_flags           int
	Sample_count       uint32
	Data_offset        uint32
	First_sample_flags uint32
	Vals               *list.List //*TRUN_ARRAY_FIELDS
}
