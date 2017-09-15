package example

import (
	"os"
	"log"
)

/*
this code write for imitate h264 and aac stream from FLV file
*/


const (
	FLV_TAG_Audio      = 8
	FLV_TAG_Video      = 9
	FLV_TAG_ScriptData = 18
)

const (
	SoundFormat_LinearPCM_platformEndian = 0
	SoundFormat_ADPCM                    = 1
	SoundFormat_MP3                      = 2
	SoundFormat_LinearPCM_littleEndian   = 3
	SoundFormat_Nellymoser16KHzMono      = 4
	SoundFormat_Nellymoser8KHzMono       = 5
	SoundFormat_Nellymoser               = 6
	SoundFormat_G711ALaw_PCM             = 7
	SoundFormat_G711muLaw_PCM            = 8
	SoundFormat_reserved                 = 9
	SoundFormat_AAC                      = 10
	SoundFormat_Speex                    = 11
	SoundFormat_MP3_8KHz                 = 14
	SoundFormat_DeviceSpecific_sound     = 15
)

const (
	SoundRate_5_5K = 0
	SoundRate_11K  = 1
	SoundRate_22K  = 2
	SoundRate_44K  = 3
)

const (
	SoundSize_8Bit  = 0
	SoundSize_16Bit = 1
)

const (
	SndMono   = 0
	SndStereo = 1
)

const (
	AACSequenceHeader = 0
	AACRaw            = 1
)

const (
	FrameType_Keyframe             = 1
	FrameType_InterFrame           = 2
	FrameType_DisposableInterFrame = 3 //H263 only
	FrameType_GeneratedKeyframe    = 4 //server user only
	FrameType_videoInfoCmdFrame    = 5
)

const (
	CodecID_JPEG               = 1
	CodecID_SorenSonH263       = 2
	CodecID_ScreenVideo        = 3
	CodecID_On2VP6             = 4
	CodecID_On2Vp6AlphaChannel = 5
	CodecID_ScreenVideoV2      = 6
	CodecID_AVC                = 7
)

const (
	AVC_Header = 0
	AVC_NALU   = 1
)

type FlvTag struct {
	TagType   uint8
	Timestamp uint32
	StreamID  uint32
	Data      []byte
}

type AudioTag struct {
}

type VideoTag struct {
}

func GetAudioTag(flvTag *FlvTag) (result *AudioTag, err error) {
	return
}

func GetVideoTag(flvTag *FlvTag) (result *VideoTag, err error) {
	return
}

func (this *FlvTag) Copy() (dst *FlvTag) {
	dst = &FlvTag{}
	dst.StreamID = this.StreamID
	dst.TagType = this.TagType
	dst.Timestamp = this.Timestamp
	if len(this.Data) > 0 {
		dst.Data = make([]byte, len(this.Data))
		copy(dst.Data, this.Data)
	}
	return
}


type FlvFileReader struct {
	fp *os.File
}

func (this *FlvFileReader) Init(name string) error {
	var err error
	this.fp, err = os.Open(name)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	tmp := make([]byte, 13)
	_, err = this.fp.Read(tmp)
	return err
}

func (this *FlvFileReader) GetNextTag() (tag *FlvTag, err error) {
	buf := make([]byte, 11)
	_, err = this.fp.Read(buf)
	if err != nil {
		return
	}
	tag = &FlvTag{}
	tag.TagType = buf[0]
	dataSize := (int(int(buf[1])<<16) | (int(buf[2]) << 8) | (int(buf[3])))
	tag.Timestamp = uint32(int(int(buf[7])<<24) | (int(buf[4]) << 16) | (int(buf[5]) << 8) | (int(buf[6])))

	tag.Data = make([]byte, dataSize)
	_, err = this.fp.Read(tag.Data)
	if err != nil {
		return
	}
	buf = make([]byte, 4)
	this.fp.Read(buf)
	return
}

func (this *FlvFileReader) Close() {
	if this.fp != nil {
		this.fp.Close()
	}
}