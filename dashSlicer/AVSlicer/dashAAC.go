package AVSlicer

import (
	"github.com/panda-media/muxer-fmp4/codec/AAC"
	"github.com/panda-media/muxer-fmp4/format/AVPacket"
	"strconv"
)

type SlicerAAC struct {
	headerDecode bool
	asc          *AAC.AACAudioSpecificConfig
	frameCount   int64
	codec        string
}

func (this *SlicerAAC) AddFrame(data []byte, timestamp int64,timescale int) (tag *AVPacket.MediaPacket) {
	if data == nil || len(data) == 0||timescale==0 {
		return
	}
	if false == this.headerDecode {
		this.asc = AAC.AACGetConfig(data)
		if this.asc.ObjectType() == 0 || this.asc.SampleRate() == 0 ||
			this.asc.Channel() == 0 {
			return
		}

		timestamp=timestamp*int64(this.asc.SampleRate())/int64(timescale)
		this.headerDecode = true
		tag = &AVPacket.MediaPacket{}
		tag.PacketType = AVPacket.AV_PACKET_TYPE_AUDIO
		tag.TimeStamp = timestamp
		tag.Data = make([]byte, 2+len(data))
		tag.Data[0] = 0xaf
		tag.Data[1] = 0
		copy(tag.Data[2:], data)
		this.codec = "mp4a.40." + strconv.Itoa(this.asc.ObjectType())
	} else {

		timestamp=timestamp*int64(this.asc.SampleRate())/int64(timescale)
		tag = &AVPacket.MediaPacket{}
		tag.PacketType = AVPacket.AV_PACKET_TYPE_AUDIO
		tag.Data = make([]byte, 2+len(data))
		tag.TimeStamp = timestamp
		tag.Data[0] = 0xaf
		tag.Data[1] = 1
		copy(tag.Data[2:], data)
	}
	return
}

func (this *SlicerAAC) calNextTimeStamp() (timestamp int64) {
	this.frameCount++
	timestamp = this.frameCount * 1000 * AAC.SAMPLE_SIZE / int64(this.asc.SampleRate())
	return
}

func (this *SlicerAAC) SampleRate() int {
	return this.asc.SampleRate()
}

func (this *SlicerAAC) Channel() int {
	return this.asc.Channel()
}

func (this *SlicerAAC) Codec() string {
	return this.codec
}
