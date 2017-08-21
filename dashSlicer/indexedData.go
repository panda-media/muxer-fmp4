package dashSlicer

import (
	"sync"
	"errors"
)

type indexedData struct {
	maxSliceCounter int
	avSeparated     bool
	videoHeader     []byte
	audioHeader     []byte
	startNumber_av  int
	idx_av          int
	startNumber_a   int
	idx_a           int
	dataAV          map[int][]byte
	muxAV           sync.RWMutex
	dataA           map[int][]byte
	muxA            sync.RWMutex
}

func (this *indexedData) init(avSeparate bool, maxSliceDataCounter int) {
	this.maxSliceCounter = maxSliceDataCounter
	this.avSeparated = avSeparate
	this.dataAV = make(map[int][]byte)
	if avSeparate {
		this.dataA = make(map[int][]byte)
	}
}

func (this *indexedData) SetVideoHeader(header []byte) {
	this.videoHeader = make([]byte, len(header))
	copy(this.videoHeader, header)
}

func (this *indexedData) SetAudioHeader(header []byte) {
	this.audioHeader = make([]byte, len(header))
	copy(this.audioHeader, header)
}

func (this *indexedData)GetVideoHeader()(header []byte){
	return  this.videoHeader
}

func (this *indexedData)GetAudioHeader()(header []byte){
	return this.audioHeader
}

func (this *indexedData) AddAudioSlice(data []byte) {
	if nil == data || len(data) == 0 {
		return
	}
	this.muxA.Lock()
	defer this.muxA.Unlock()
	if this.avSeparated {
		this.dataA[this.idx_a]=data
		this.idx_a++
		if len(this.dataA) > this.maxSliceCounter {
			delete(this.dataA,this.startNumber_a)
			this.startNumber_a++
		}
	}
}

func (this *indexedData) AddVideoSlice(data []byte) {
	if nil == data || len(data) == 0 {
		return
	}
	this.muxAV.Lock()
	defer this.muxAV.Unlock()
	this.dataAV[this.idx_av]=data
	this.idx_av++
	if len(this.dataAV) > this.maxSliceCounter {
		delete(this.dataAV,this.startNumber_av)
		this.startNumber_av++
	}
}

func (this *indexedData) MediaDataByIndex(idx int, audio bool) (data []byte, err error) {
	if audio{
		if false==this.avSeparated{
			return  nil,errors.New("audio and video are in the same track")
		}else{
			this.muxA.Lock()
			defer this.muxA.Unlock()
			ok:=false
			data,ok=this.dataA[idx]
			if false==ok{
				err=errors.New("audio data index out of range")
				return
			}
			return
		}
	}else{
		this.muxAV.Lock()
		defer this.muxAV.RUnlock()
		ok:=false
		data,ok=this.dataAV[idx]
		if false==ok{
			err=errors.New("av data index out of range")
			return
		}
		return
	}
	return
}
