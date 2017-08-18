package commonBoxes

import (
	"errors"
	"fmt"
	"github.com/panda-media/muxer-fmp4/codec/AAC"
	"github.com/panda-media/muxer-fmp4/codec/H264"
	"github.com/panda-media/muxer-fmp4/format/AVPacket"
	"strconv"
	"time"
)

func moovBox(durationMS uint64, audioHeader, videoHeader *AVPacket.MediaPacket, arraysAudio, arraysVideo *MOOV_ARRAYS) (box *MP4Box, err error) {

	timestamp := uint64(time.Now().Unix())
	timestamp += 0x83aaef00 //1900 to 1970
	box, err = NewMP4Box("moov")
	if err != nil {
		return
	}
	//mvhd
	param_mvhd := &mvhdPram{version: 1,
		creation_time:     timestamp,
		modification_time: timestamp,
		duration:          durationMS,
		timescale:         VIDE_TIME_SCALE_Millisecond,
		next_track_ID:     TRACK_NEXT}
	mvhd, err := mvhdBox(param_mvhd)
	if err != nil {
		return
	}
	box.PushBox(mvhd)
	var audioSampleRate uint32
	var audioSampleSize uint32
	if audioHeader != nil {
		audioSampleRate, audioSampleSize, err = GetAudioSampleRateSampleSize(audioHeader)
		if err != nil {
			return
		}
	}
	//mvex
	var param_trex_audio *trexParam
	var param_trex_video *trexParam
	if audioHeader != nil {
		param_trex_audio = &trexParam{
			TRACK_AUDIO,
			1,
			audioSampleSize,
			0,
			0x1001,
		}
	}
	if videoHeader != nil {
		param_trex_video = &trexParam{
			TRACK_VIDEO,
			1,
			0x3e8,
			0,
			0x10001,
		}
	}
	mvex, err := mvexBox(param_trex_audio, param_trex_video)
	if err != nil {
		return
	}
	box.PushBox(mvex)
	//track
	if audioHeader != nil {
		duration := durationMS * uint64(audioSampleRate) / VIDE_TIME_SCALE_Millisecond
		var trak *MP4Box
		trak, err = trakBox(audioHeader, arraysAudio, timestamp, duration)
		if err != nil {
			return
		}
		box.PushBox(trak)
	}

	if videoHeader != nil {
		duration := durationMS * VIDE_TIME_SCALE / VIDE_TIME_SCALE_Millisecond
		var trak *MP4Box
		trak, err = trakBox(videoHeader, arraysVideo, timestamp, duration)
		if err != nil {
			return
		}
		box.PushBox(trak)
	}

	return
}

func Box_moov_Data(durationMS uint64, audioHeader, videoHeader *AVPacket.MediaPacket, arraysAudio, arraysVideo *MOOV_ARRAYS) (data []byte, err error) {
	if nil == audioHeader && nil == videoHeader {
		err = errors.New("no audio and video	 header")
		return
	}
	box, err := moovBox(durationMS, audioHeader, videoHeader, arraysAudio, arraysVideo)
	if err != nil {
		return
	}
	data = box.Flush()
	return
}

func GetAudioSampleRateSampleSize(audioHeader *AVPacket.MediaPacket) (sampleRate, sampleSize uint32, err error) {
	//aac only
	soundFormat := audioHeader.Data[0] >> 4
	switch soundFormat {
	case AVPacket.SoundFormat_AAC:
		asc := AAC.AACGetConfig(audioHeader.Data[2:])
		sampleRate = uint32(asc.SampleRate())
		sampleSize = AAC.SAMPLE_SIZE
	default:
		err = errors.New(fmt.Sprintf("in moov ,not support soundformat %d", int(soundFormat)))
		return
	}
	return
}

func getVideoWidthHeight(videoHeader *AVPacket.MediaPacket) (width, height int, err error) {
	videoCodec := videoHeader.Data[0] & 0xf
	switch videoCodec {
	case AVPacket.CodecID_AVC:
		FrameType := videoHeader.Data[0] >> 4
		if FrameType != 1 {
			err = errors.New("not a key frame avc")
			return
		}
		var avc *H264.AVCDecoderConfigurationRecord
		avc, err = H264.DecodeAVC(videoHeader.Data[5:])
		if err != nil {
			return
		}
		if avc.SPS != nil && avc.SPS.Len() > 0 {
			sps := avc.SPS.Front().Value.([]byte)
			width, height, _, _, _, _ = H264.DecodeSPS(sps)
		}
		return
	default:
		err = errors.New("not support video type" + strconv.Itoa(int(videoCodec)))
		return
	}
	return
}
