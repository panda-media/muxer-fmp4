package dashSlicer

import (
	"github.com/panda-media/muxer-fmp4/format/MP4"
	"github.com/panda-media/muxer-fmp4/format/AVPacket"
	"errors"
)

type DASHSlicer struct {
	audioVideoSeparated bool
	minSliceDuration int
	maxSliceDuration int
	maxSliceDataCounter int
	H264Processer dashH264
	AACProcesser dashAAC
	audioHeaderMuxed bool
	videoHeaderMuxed bool
	avMuxer *MP4.FMP4Muxer	//audio and video or video only
	aMuxer *MP4.FMP4Muxer	//audio only
	avData *indexedData
}

func NEWSlicer(avSeparate bool,minLengthMS,maxLengthMS,maxSliceDataCounter int)(slicer *DASHSlicer){
	slicer=&DASHSlicer{}
	slicer.audioVideoSeparated=avSeparate
	slicer.minSliceDuration=minLengthMS
	slicer.maxSliceDuration=maxLengthMS
	slicer.maxSliceDataCounter=maxSliceDataCounter
	slicer.init()
	return
}

func (this *DASHSlicer)init(){
	this.avMuxer=MP4.NewMP4Muxer()
	if this.audioVideoSeparated{
		this.aMuxer=MP4.NewMP4Muxer()
	}
	this.avData=&indexedData{}
	this.avData.init(this.audioVideoSeparated,this.maxSliceDataCounter)
}

//one or more nal
func (this *DASHSlicer)AddH264Nals(data []byte)(err error){
	tags:=this.H264Processer.addNals(data)
	if tags==nil||tags.Len()==0{
		return
	}
	for e:=tags.Front();e!=nil;e=e.Next(){
		tag:=e.Value.(*AVPacket.MediaPacket)
		if this.videoHeaderMuxed==false&&tag.Data[0]==0x1f&&tag.Data[1]==0{
			err=this.avMuxer.SetVideoHeader(tag)
			if err!=nil{
				err=errors.New("set video header :"+err.Error())
				return
			}
			this.videoHeaderMuxed=true
			continue
		}
		err=this.avMuxer.AddPacket(tag)
		if err!=nil{
			return
		}
		check for slice
	}
	return
}
//one frame
func (this *DASHSlicer)AddAACFrame(data []byte)(err error){
	return
}

func (this *DASHSlicer)GetLastedMPD()(data []byte,err error){
return
}

func (this *DASHSlicer)GetMediaDataByIndex(idx int,audio bool)(data []byte,err error){
	data,err=this.avData.MediaDataByIndex(idx,audio)
	return
}

func (this *DASHSlicer)GetInitA()(data []byte){
	if this.audioVideoSeparated{
		data=this.avData.GetAudioHeader()
	}
	return
}
func (this *DASHSlicer)GetInitV()(data []byte){
	return this.avData.GetVideoHeader()
}