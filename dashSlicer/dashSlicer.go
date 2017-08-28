package dashSlicer

import (
	"errors"
	"github.com/panda-media/muxer-fmp4/codec/AAC"
	"github.com/panda-media/muxer-fmp4/format/AVPacket"
	"github.com/panda-media/muxer-fmp4/format/MP4"
	"github.com/panda-media/muxer-fmp4/mpd"
	"logger"
	"os"
	"strconv"
)

const(

	saveAV=true
)

var vidx=0
var aidx=0
type DASHSlicer struct {
	minSliceDuration    int
	maxSliceDuration    int
	maxSliceDataCounter int
	lastBeginTime       int
	h264Processer       dashH264
	aacProcesser        dashAAC
	audioHeaderMuxed    bool
	videoHeaderMuxed    bool
	muxerV              *MP4.FMP4Muxer //video only
	muxerA              *MP4.FMP4Muxer //audio only
	audioFrameCount     int
	mpd                 *mpd.MPDDynamic
}

func NEWSlicer(avSeparate bool, minLengthMS, maxLengthMS, maxSliceDataCounter int) (slicer *DASHSlicer) {
	slicer = &DASHSlicer{}
	slicer.minSliceDuration = minLengthMS
	slicer.maxSliceDuration = maxLengthMS
	slicer.maxSliceDataCounter = maxSliceDataCounter
	slicer.init()
	return
}

func (this *DASHSlicer) init() {
	this.muxerV = MP4.NewMP4Muxer()
	this.muxerA = MP4.NewMP4Muxer()
	this.mpd = mpd.NewDynamicMPDCreater(this.minSliceDuration, this.maxSliceDataCounter)
}

func (this *DASHSlicer) newslice(timestamp uint32) bool {
	if int(timestamp)-this.lastBeginTime >= this.minSliceDuration {
		this.lastBeginTime = int(timestamp)
		return true
	}
	return false
}

//one or more nal
func (this *DASHSlicer) AddH264Nals(data []byte) (err error) {
	tags := this.h264Processer.addNals(data)
	if tags == nil || tags.Len() == 0 {
		return
	}
	for e := tags.Front(); e != nil; e = e.Next() {
		tag := e.Value.(*AVPacket.MediaPacket)
		if this.videoHeaderMuxed == false && tag.Data[0] == 0x17 && tag.Data[1] == 0 {
			err = this.muxerV.SetVideoHeader(tag)
			if err != nil {
				err = errors.New("set video header :" + err.Error())
				return
			}
			this.mpd.SetVideoInfo(1000, this.h264Processer.width, this.h264Processer.height, this.h264Processer.fps,
				1, this.h264Processer.codec)
			this.videoHeaderMuxed = true

			if saveAV{
				fp,_:=os.Create("initV.mp4")
				defer fp.Close()
				initV,_:=this.muxerV.GetInitSegment()
				fp.Write(initV)
			}
			continue
		}
		mediadurationout:=false
		if this.muxerV!=nil{
			if this.muxerV.Duration()>this.maxSliceDuration{
				mediadurationout=true
			}
		}
		if this.muxerA!=nil&&false==mediadurationout{
			if this.muxerA.Duration()>this.maxSliceDuration{
				mediadurationout=true
			}
		}
		if (tag.Data[0] == 0x17 && tag.Data[1] == 1){
			//logger.LOGD("slice key frame")
		}
		if mediadurationout{
			//logger.LOGD("slice duration")
		}
		//mediadurationout=false
		if (tag.Data[0] == 0x17 && tag.Data[1] == 1)||mediadurationout {
			if this.newslice(tag.TimeStamp) {
				_, moofmdat, duration, bitrate, err := this.muxerV.Flush()
				if err != nil {
					return err
				}
				this.mpd.SetVideoBitrate(bitrate)
				this.mpd.AddVideoSlice(duration, moofmdat)
				if saveAV&&vidx<10{

					fp,_:=os.Create("v"+strconv.Itoa(vidx)+".mp4")
					vidx++
					defer fp.Close()
					fp.Write(moofmdat)
				}
				if this.audioHeaderMuxed {
					_, moofmdat, _, bitrate, er := this.muxerA.Flush()
					if er != nil {
						return er
					}

					this.mpd.SetAudioBitrate(bitrate)
					this.mpd.AddAudioSlice(this.audioFrameCount, moofmdat)
					this.audioFrameCount = 0
					if saveAV&&aidx<10{
						fp,_:=os.Create("a"+strconv.Itoa(aidx)+".mp4")
						aidx++
						defer fp.Close()
						fp.Write(moofmdat)
					}
				}
				mpd,err:=this.mpd.GetMPDXML()
				if err!=nil{
					logger.LOGF(err.Error())
				}
				fp,err:=os.Create("mpd.xml")
				defer fp.Close()
				fp.Write(mpd)
			}
		}
		err = this.muxerV.AddPacket(tag)
		if err != nil {
			return
		}

	}
	return
}

//one frame
func (this *DASHSlicer) AddAACFrame(data []byte) (err error) {
	tag := this.aacProcesser.addFrame(data)
	if tag == nil {
		err = errors.New("invalid aac data")
		logger.LOGD(err.Error())
		return
	}
	if false == this.audioHeaderMuxed {
		this.muxerA.SetAudioHeader(tag)
		this.audioHeaderMuxed = true
		this.mpd.SetAudioInfo(this.aacProcesser.asc.SampleRate(),
			this.aacProcesser.asc.SampleRate(),
			16,
			this.aacProcesser.asc.Channel(),
			AAC.SAMPLE_SIZE,
			this.aacProcesser.codec)
		if saveAV{
			fp,_:=os.Create("InitA.mp4")
			defer fp.Close()
			initA,_:=this.muxerA.GetInitSegment()
			fp.Write(initA)
		}
	} else {
		this.muxerA.AddPacket(tag)
		this.audioFrameCount++
	}
	return
}

func (this *DASHSlicer) GetLastedMPD() (data []byte, err error) {
	data,err=this.mpd.GetMPDXML()
	return
}

func (this *DASHSlicer) GetMediaDataByIndex(idx int, audio bool) (data []byte, err error) {

	return
}

func (this *DASHSlicer) GetInitA() (data []byte, err error) {
	if this.audioHeaderMuxed {

		data, err = this.muxerA.GetInitSegment()
		return
	} else {
		err = errors.New("audio not founded")
	}
	return
}
func (this *DASHSlicer) GetInitV() (data []byte, err error) {
	if this.videoHeaderMuxed {
		data, err = this.muxerV.GetInitSegment()
		return
	} else {
		err = errors.New("video header not founded")
	}
	return
}
