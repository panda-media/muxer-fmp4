package dashSlicer

import (
	"errors"
	"sync"
)

type sliceData struct {
	duration int
	bitrate  int
	data     []byte
}

type sliceDataContainer struct {
	maxSliceCounter int
	avSeparated     bool
	videoHeader     []byte
	audioHeader     []byte
	startNumber_av  int
	idx_av          int
	startNumber_a   int
	idx_a           int
	dataAV          map[int]*sliceData
	muxAV           sync.RWMutex
	dataA           map[int]*sliceData
	muxA            sync.RWMutex
}

func (this *sliceDataContainer) init(avSeparate bool, maxSliceDataCounter int) {
	this.maxSliceCounter = maxSliceDataCounter
	this.avSeparated = avSeparate
	this.dataAV = make(map[int]*sliceData)
	if avSeparate {
		this.dataA = make(map[int]*sliceData)
	}
}

func (this *sliceDataContainer) SetVideoHeader(header []byte) {
	this.videoHeader = make([]byte, len(header))
	copy(this.videoHeader, header)
}

func (this *sliceDataContainer) SetAudioHeader(header []byte) {
	this.audioHeader = make([]byte, len(header))
	copy(this.audioHeader, header)
}

func (this *sliceDataContainer) GetVideoHeader() (header []byte) {
	return this.videoHeader
}

func (this *sliceDataContainer) GetAudioHeader() (header []byte) {
	return this.audioHeader
}

func (this *sliceDataContainer) AddAudioSlice(data []byte, duration, bitrate int) {

	this.muxA.Lock()
	defer this.muxA.Unlock()
	if this.avSeparated {
		slice_data := &sliceData{
			duration,
			bitrate,
			data,
		}
		this.dataA[this.idx_a] = slice_data
		this.idx_a++
		if len(this.dataA) > this.maxSliceCounter {
			delete(this.dataA, this.startNumber_a)
			this.startNumber_a++
		}
	}
}

func (this *sliceDataContainer) AddVideoSlice(data []byte, duration, bitrate int) {
	if nil == data || len(data) == 0 {
		return
	}
	this.muxAV.Lock()
	defer this.muxAV.Unlock()
	slice_data := &sliceData{
		duration,
		bitrate,
		data,
	}
	this.dataAV[this.idx_av] = slice_data
	this.idx_av++
	if len(this.dataAV) > this.maxSliceCounter {
		delete(this.dataAV, this.startNumber_av)
		this.startNumber_av++
	}
}

func (this *sliceDataContainer) MediaDataByIndex(idx int, audio bool) (slice_data *sliceData, err error) {
	if audio {
		if false == this.avSeparated {
			return nil, errors.New("audio and video are in the same track")
		} else {
			this.muxA.Lock()
			defer this.muxA.Unlock()
			ok := false
			slice_data, ok = this.dataA[idx]
			if false == ok {
				err = errors.New("audio data index out of range")
				return
			}
			return
		}
	} else {
		this.muxAV.Lock()
		defer this.muxAV.RUnlock()
		ok := false
		slice_data, ok = this.dataAV[idx]
		if false == ok {
			err = errors.New("av data index out of range")
			return
		}
		return
	}
	return
}
