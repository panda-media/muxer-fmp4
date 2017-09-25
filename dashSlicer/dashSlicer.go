package dashSlicer

import (
	"errors"
	"fmt"
	"strings"

	"github.com/panda-media/muxer-fmp4/codec/AAC"
	"github.com/panda-media/muxer-fmp4/dashSlicer/AVSlicer"
	"github.com/panda-media/muxer-fmp4/format/AVPacket"
	"github.com/panda-media/muxer-fmp4/format/MP4"
	"github.com/panda-media/muxer-fmp4/mpd"
)

type DASHSlicer struct {
	minSegmentDuration   int   //ms
	maxSegmentDuration   int   //valid when audio only
	maxSegmentCountInMPD int   //ms
	lastVideoStartTime   int64 //in pts
	h264Transfer         AVSlicer.SlicerH264
	videoTimescale       int
	audioTimescale int
	aacTransfer          AVSlicer.SlicerAAC
	audioHeaderMuxed     bool
	adtsHeaderEncoed     bool
	videoHeaderMuxed     bool
	muxerV               *MP4.FMP4Muxer //video only
	muxerA               *MP4.FMP4Muxer //audio only
	lastAudioStartTime   int64          //in pts
	mpd                  *mpd.MPDDynamic
	receiver             FMP4Receiver
}
//videoTimescale and audioTimescale for input
func NEWSlicer(fps, videoTimescale,audioTimescale, minLengthMS, maxLengthMS, maxSegmentCountInMPD int, receiver FMP4Receiver) (slicer *DASHSlicer, err error) {
	slicer = &DASHSlicer{}

	if maxSegmentCountInMPD < 2 || nil == receiver {
		err = errors.New("invalid param ")
		return nil, err
	}
	if minLengthMS < 1000 {
		minLengthMS = 1000
	}
	if maxLengthMS < minLengthMS {
		maxLengthMS = minLengthMS
	}
	if videoTimescale <= 60 {
		videoTimescale = 90000
	}

	slicer.minSegmentDuration = minLengthMS
	slicer.maxSegmentDuration = maxLengthMS
	slicer.maxSegmentCountInMPD = maxSegmentCountInMPD
	slicer.receiver = receiver
	slicer.videoTimescale = videoTimescale
	slicer.audioTimescale=audioTimescale

	slicer.h264Transfer.Init(fps)
	slicer.init()

	return
}

//add nal,separate 0x 00 00 01 or 0x 00 00 00 01
func (this *DASHSlicer) AddH264Nals(data []byte, timestamp int64) (err error) {
	tags, err := this.h264Transfer.AddNals(data, timestamp)
	if err != nil || tags == nil || tags.Len() == 0 {
		return
	}
	for e := tags.Front(); e != nil; e = e.Next() {
		tag := e.Value.(*AVPacket.MediaPacket)

		err = this.appendH264Tag(tag)
		if err != nil {
			err = errors.New("AddH264Nals failed:" + err.Error())
			return
		}
	}
	return
}

//add nal,four bytes network byte order size+ nal data......
func (this *DASHSlicer) AddH264Frame(nal []byte, timestamp int64,compositionTime int) (err error) {
	tags,err:=this.h264Transfer.AddFrame(nal,timestamp,compositionTime)
	if err != nil || tags == nil || tags.Len() == 0 {
		return
	}
	for e := tags.Front(); e != nil; e = e.Next() {
		tag := e.Value.(*AVPacket.MediaPacket)

		err = this.appendH264Tag(tag)
		if err != nil {
			err = errors.New("AddH264Nals failed:" + err.Error())
			return
		}
	}
	return
}

func (this *DASHSlicer) appendH264Tag(tag *AVPacket.MediaPacket) (err error) {

	if this.videoHeaderMuxed == false && tag.Data[0] == 0x17 && tag.Data[1] == 0 {
		this.lastVideoStartTime = tag.TimeStamp
		err = this.muxerV.SetVideoHeader(tag,uint32(this.videoTimescale))
		if err != nil {
			err = errors.New("set video header :" + err.Error())
			return
		}
		this.mpd.SetVideoInfo(this.videoTimescale, this.h264Transfer.Width(), this.h264Transfer.Height(), this.h264Transfer.FPS(),
			1, this.h264Transfer.Codec())
		this.videoHeaderMuxed = true
		var videoHeader []byte
		videoHeader, err = this.muxerV.GetInitSegment()
		this.receiver.VideoHeaderGenerated(videoHeader)
		return
	}

	if tag.Data[0] == 0x17 && tag.Data[1] == 1 {
		if this.needNewSegment(tag.TimeStamp) {
			_, moofmdat, duration, bitrate, err := this.muxerV.Flush()
			if err != nil {
				return err
			}
			if duration==0{
				return err
			}
			this.mpd.SetVideoBitrate(bitrate)

			var timestamp int64
			timestamp, err = this.mpd.AddVideoSlice(duration, moofmdat)
			this.receiver.VideoSegmentGenerated(moofmdat, timestamp, duration)
			if this.audioHeaderMuxed {
				_, moofmdat, duration, bitrate, er := this.muxerA.Flush()
				if er != nil {
					return er
				}
				this.mpd.SetAudioBitrate(bitrate)

				timestamp, _ := this.mpd.AddAudioSlice(duration, moofmdat)
				this.receiver.AudioSegmentGenerated(moofmdat, timestamp, duration)
			}

		}
	}
	err = this.muxerV.AddPacket(tag)
	if err != nil {
		return
	}
	return
}

func (this *DASHSlicer) AddAACADTSFrame(data []byte, timestamp int64) (err error) {
	if !this.adtsHeaderEncoed {
		this.adtsHeaderEncoed = true
		adts, err := AAC.ParseAdts(data)
		if err != nil {
			return err
		}
		headerData := AAC.EncodeAudioSpecificConfig(adts)
		err = this.AddAACFrame(headerData, timestamp)
		if err != nil {
			return err
		}
	}
	frameData := AAC.ReMuxerADTSData(data)
	return this.AddAACFrame(frameData, timestamp)
}

//add one  aac frame
func (this *DASHSlicer) AddAACFrame(data []byte, timestamp int64) (err error) {
	tag := this.aacTransfer.AddFrame(data, timestamp,this.audioTimescale)
	if tag == nil {
		err = errors.New("invalid aac data")
		return
	}
	if false == this.audioHeaderMuxed {
		this.lastAudioStartTime = tag.TimeStamp
		this.muxerA.SetAudioHeader(tag)
		this.audioHeaderMuxed = true
		this.mpd.SetAudioInfo(this.aacTransfer.SampleRate(),
			this.aacTransfer.SampleRate(),
			16,
			this.aacTransfer.Channel(),
			AAC.SAMPLE_SIZE,
			this.aacTransfer.Codec())
		audioHeader, err := this.muxerA.GetInitSegment()
		if err != nil {
			return err
		}
		this.receiver.AudioHeaderGenerated(audioHeader)
	} else {
		this.muxerA.AddPacket(tag)
		if false == this.videoHeaderMuxed {
			timestamp_MS := tag.TimeStamp * 1000 / int64(this.aacTransfer.SampleRate())
			if timestamp_MS > this.lastAudioStartTime+int64(this.maxSegmentDuration) {
				this.lastAudioStartTime = timestamp_MS
				_, moofmdat, duration, bitrate, er := this.muxerA.Flush()
				if er != nil {
					return er
				}
				this.mpd.SetAudioBitrate(bitrate)

				timestamp, _ := this.mpd.AddAudioSlice(duration, moofmdat)
				this.receiver.AudioSegmentGenerated(moofmdat, timestamp, duration)
			}
		}
	}
	return
}

//get the last MPD
func (this *DASHSlicer) GetMPD() (data []byte, err error) {
	data, err = this.mpd.GetMPDXML()
	return
}

func (this *DASHSlicer) init() {
	this.muxerV = MP4.NewMP4Muxer()
	this.muxerA = MP4.NewMP4Muxer()
	this.mpd = mpd.NewDynamicMPDCreater(this.minSegmentDuration, this.maxSegmentCountInMPD)
}

func (this *DASHSlicer) needNewSegment(timestamp int64) bool {
	timestamp_MS := timestamp * 1000 / int64(this.videoTimescale)
	if timestamp_MS >= this.lastVideoStartTime+int64(this.minSegmentDuration) {
		this.lastVideoStartTime = timestamp_MS
		return true
	}
	return false
}

func (this *DASHSlicer) GetVideoData(param string) (data []byte, err error) {
	if strings.Contains(param, "_init_") {
		data, err = this.muxerV.GetInitSegment()
	} else {
		id := int64(0)
		fmt.Sscanf(param, "video_video0_%d_mp4.m4s", &id)
		data, err = this.mpd.GetVideoSlice(id)
	}
	return
}

func (this *DASHSlicer) GetAudioData(param string) (data []byte, err error) {
	if strings.Contains(param, "_init_") {
		data, err = this.muxerA.GetInitSegment()
	} else {
		id := int64(0)
		fmt.Sscanf(param, "audio_audio0_%d_mp4.m4s", &id)
		data, err = this.mpd.GetAudioSlice(id)
	}
	return
}

//notice the slicer stream end
func (this *DASHSlicer) EndofStream() {
	if this.videoHeaderMuxed {
		//video only or av
		_, moofmdat, duration, bitrate, err := this.muxerV.Flush()
		if err != nil {
			return
		}
		this.mpd.SetVideoBitrate(bitrate)

		var timestamp int64
		timestamp, err = this.mpd.AddVideoSlice(duration, moofmdat)
		this.receiver.VideoSegmentGenerated(moofmdat, timestamp, duration)
		if this.audioHeaderMuxed {
			_, moofmdat, duration, bitrate, err := this.muxerA.Flush()
			if err != nil {
				return
			}

			this.mpd.SetAudioBitrate(bitrate)
			timestamp, _ := this.mpd.AddAudioSlice(duration, moofmdat)
			this.receiver.AudioSegmentGenerated(moofmdat, timestamp, duration)

		}
	} else if this.audioHeaderMuxed {
		//audio only
		_, moofmdat, duration, bitrate, err := this.muxerA.Flush()
		if err != nil {
			return
		}

		this.mpd.SetAudioBitrate(bitrate)

		timestamp, _ := this.mpd.AddAudioSlice(duration, moofmdat)
		this.receiver.AudioSegmentGenerated(moofmdat, timestamp, duration)
	}

}
