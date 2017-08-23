package dashSlicer

import (
	"github.com/panda-media/muxer-fmp4/format/MP4"
	"github.com/panda-media/muxer-fmp4/format/AVPacket"
	"errors"
	"logger"
)

type DASHSlicer struct {
	audioVideoSeparated bool
	minSliceDuration    int
	maxSliceDuration    int
	maxSliceDataCounter int
	lastBeginTime       int
	h264Processer       dashH264
	aacProcesser        dashAAC
	audioHeaderMuxed    bool
	videoHeaderMuxed    bool
	avMuxer             *MP4.FMP4Muxer	//audio and video or video only
	aMuxer              *MP4.FMP4Muxer	//audio only
	avData              *sliceDataContainer
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
	this.avData=&sliceDataContainer{}
	this.avData.init(this.audioVideoSeparated,this.maxSliceDataCounter)
}

func (this *DASHSlicer)newslice(timestamp uint32)bool{
	if int(timestamp)-this.lastBeginTime>=this.minSliceDuration{
		this.lastBeginTime=int(timestamp)
		return true
	}
	return false
}

//one or more nal
func (this *DASHSlicer)AddH264Nals(data []byte)(err error){
	tags:=this.h264Processer.addNals(data)
	if tags==nil||tags.Len()==0{
		logger.LOGD("no tag")
		return
	}
	for e:=tags.Front();e!=nil;e=e.Next(){
		tag:=e.Value.(*AVPacket.MediaPacket)
		if this.videoHeaderMuxed==false&&tag.Data[0]==0x17&&tag.Data[1]==0{
			err=this.avMuxer.SetVideoHeader(tag)
			if err!=nil{
				err=errors.New("set video header :"+err.Error())
				return
			}
			this.videoHeaderMuxed=true
			continue
		}
		if tag.Data[0]==0x17&&tag.Data[1]==1{
			if this.newslice(tag.TimeStamp){
					_,moofmdat,duration,bitrate,err:=this.avMuxer.Flush()
				if err!=nil{
					return err
				}
				this.avData.AddVideoSlice(moofmdat,duration,bitrate)
				if this.audioVideoSeparated{
					_,moofmdat,duration,bitrate,err:=this.aMuxer.Flush()
					if err!=nil{
						this.avData.AddAudioSlice(nil,0,0)
					}
					this.avData.AddAudioSlice(moofmdat,duration,bitrate)
				}
			}
		}
		err=this.avMuxer.AddPacket(tag)
		if err!=nil{
			return
		}

	}
	return
}
//one frame
func (this *DASHSlicer)AddAACFrame(data []byte)(err error){
	tag:=this.aacProcesser.addFrame(data)
	if tag==nil{
		err=errors.New("invalid aac data")
		logger.LOGD(err.Error())
		return
	}
	if false==this.audioHeaderMuxed{
		if this.audioVideoSeparated{
			this.aMuxer.SetAudioHeader(tag)
			logger.LOGD("set audio header")
		}else{
			this.avMuxer.SetAudioHeader(tag)
		}
		this.audioHeaderMuxed=true
	}else{
		if this.audioVideoSeparated{
			this.aMuxer.AddPacket(tag)
		}else{
			this.avMuxer.AddPacket(tag)
		}
	}
	return
}

func (this *DASHSlicer)GetLastedMPD()(data []byte,err error){
	//period id,update
return
}

func (this *DASHSlicer)GetMediaDataByIndex(idx int,audio bool)(data []byte,err error){
	slice_data,err:=this.avData.MediaDataByIndex(idx,audio)
	data=slice_data.data
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