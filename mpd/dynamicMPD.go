package mpd

import (
	"errors"
	"fmt"
	"strconv"
	"container/list"
)

type videoInfo struct {
	timeScale int
	width     int
	height    int
	frameRate int
	bitrate   int
	codecs    string
	data      []byte
}
type audioInfo struct {
	timescale  int
	sampleRate int
	bandwidth  int
	channel    int
	sampelSize int
	codecs     string
	data       []byte
}

type segmentTimeData struct{
	t int64
	d int
	data []byte
}

type MPDDynamic struct {
	availabilityStartTime      string
	publishTime                string
	minBufferTime              string
	minimumUpdatePeriod        string
	suggestedPresentationDelay string
	minBufferMS                int
	maxSliceCount              int
	vide                       *videoInfo
	audi                       *audioInfo
	videoData map[int64]*segmentTimeData
	videoKeys *list.List
	lastVideoTimestamp int64
	audioData map[int64]*segmentTimeData
	audioKeys *list.List
	lastAudioTimestamp int64
}

func NewDynamicMPDCreater(minBufferDurationMS, maxSliceCount int) (mpd *MPDDynamic) {
	mpd = &MPDDynamic{}
	mpd.init(minBufferDurationMS, maxSliceCount)
	return
}

func (this *MPDDynamic) init(minBufferDurationMS, maxSliceCount int) {
	this.minBufferMS = minBufferDurationMS
	this.maxSliceCount = maxSliceCount
	this.videoKeys=list.New()
	this.audioKeys=list.New()
	this.videoData=make(map[int64]*segmentTimeData)
	this.audioData=make(map[int64]*segmentTimeData)
	//init availabilityStartTime
	//this.availabilityStartTime
}
func (this *MPDDynamic) generatePTime(year, month, day, hour, minute, sec, mill int) (pt string) {
	pt += "P"
	if year > 0 {
		pt += strconv.Itoa(year) + "Y"
	}
	if month > 0 {
		pt += strconv.Itoa(month) + "M"
	}
	if day > 0 {
		pt += strconv.Itoa(day) + "D"
	}
	pt += "T"
	if hour > 0 {
		pt += strconv.Itoa(hour) + "H"
	}
	if minute > 0 {
		pt += strconv.Itoa(minute) + "M"
	}
	pt += fmt.Sprintf("%.3f", float32(float32(sec)+float32(mill)/1000.0)) + "S"
	return
}

func (this *MPDDynamic) SetVideoInfo(timescale, width, height, frameRate, bandwidth int, codecs string, initV []byte) (err error) {
	if this.vide != nil {
		return errors.New("set video info for mpd again")
	}
	this.vide = &videoInfo{timescale,
		width,
		height,
		frameRate,
		bandwidth,
		codecs,
		initV}
	return
}

func (this *MPDDynamic) SetAudioInfo(timescale, sampleRate, bandwidth, channel, sampleSize int, codecs string, initA []byte) (err error) {
	if this.audi != nil {
		return errors.New("set audio info in mpd times")
	}
	this.audi = &audioInfo{timescale,
		sampleRate,
		bandwidth,
		channel,
		sampleSize,
		codecs,
		initA}
	return
}

func (this *MPDDynamic) AddVideoSlice(durationMS int, data []byte)(err error) {
	if nil==this.vide{
		err=errors.New("video info not seted")
		return
	}

	return
}

func (this *MPDDynamic) AddAudioSlice(frameCount int, data []byte)(err error) {
	if nil==this.audi{
		err=errors.New("audio info not seted")
		return
	}
return
}
