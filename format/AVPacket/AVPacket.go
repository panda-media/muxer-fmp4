package AVPacket

const (
	AV_PACKET_TYPE_AUDIO    = 8
	AV_PACKET_TYPE_VIDEO    = 9
	AV_PACKET_TYPE_METADATA = 18
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

type MediaPacket struct {
	PacketType int
	TimeStamp  uint32
	StreamID   int
	Data       []byte
}

func (this *MediaPacket) Copy() (out *MediaPacket) {
	out = &MediaPacket{
		PacketType: this.PacketType,
		TimeStamp:  this.TimeStamp,
		StreamID:   this.StreamID,
	}
	if len(this.Data) > 0 {
		out.Data = make([]byte, len(this.Data))
		copy(out.Data, this.Data)
	}
	return
}
