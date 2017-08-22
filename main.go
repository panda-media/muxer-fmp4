package main

import (
	"github.com/panda-media/muxer-fmp4/format/AVPacket"
	"github.com/panda-media/muxer-fmp4/format/MP4"
	"logger"
	"mediaTypes/flv"
	"mediaTypes/mp4"
	"os"
	"github.com/panda-media/muxer-fmp4/codec/H264"
	"mediaTypes/amf"
)


func main() {

	var d byte
	d = 0xe1
	logger.LOGD(d)
	//TestFMP4FromFlvFile("muxer-fmp4/111.flv")
	//TestOldFMP4("muxer-fmp4/111.flv")
	TestPTSDTS("muxer-fmp4/111.flv")
	return
}

func TestFMP4FromFlvFile(fileName string) {
	reader := flv.FlvFileReader{}
	reader.Init(fileName)
	defer reader.Close()

	var audioHeader, videoHeader *AVPacket.MediaPacket
	logger.LOGD(audioHeader)
	for /*audioHeader == nil ||*/ videoHeader == nil {
		tag, err := reader.GetNextTag()
		if err != nil {
			return
		}
		pkt := TagToAVPacket(tag)
		if pkt.PacketType == AVPacket.AV_PACKET_TYPE_AUDIO {
			audioHeader = pkt
		}
		if pkt.PacketType == AVPacket.AV_PACKET_TYPE_VIDEO {
			videoHeader = pkt
		}
	}
	var err error
	mux := MP4.NewMP4Muxer()
	//err = mux.SetAudioHeader(audioHeader)
	err = mux.SetVideoHeader(videoHeader)
	if err != nil {
		logger.LOGE(err.Error())
	}

	fp, err := os.Create("fmp4V.mp4")
	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	defer fp.Close()
	initData, err := mux.GetInitSegment()
	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	fp.Write(initData)
	tag, err := reader.GetNextTag()
	for tag != nil && err == nil {

		pkt := TagToAVPacket(tag)
		mux.AddPacket(pkt)
		tag, err = reader.GetNextTag()
	}
	sidx, moofmdat, err := mux.Flush()
	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	if false {
		fp.Write(sidx)
	}
	fp.Write(moofmdat)

}

func TagToAVPacket(tag *flv.FlvTag) (pkt *AVPacket.MediaPacket) {
	pkt = &AVPacket.MediaPacket{
		int(tag.TagType),
		tag.Timestamp,
		int(tag.StreamID),
		tag.Data,
	}
	return
}

func TestOldFMP4(fileName string) {
	reader := flv.FlvFileReader{}
	reader.Init(fileName)
	defer reader.Close()
	tag, _ := reader.GetNextTag()
	pkt := TagToAVPacket(tag)
	for pkt.PacketType != AVPacket.AV_PACKET_TYPE_AUDIO {
		tag, _ = reader.GetNextTag()
		pkt = TagToAVPacket(tag)
	}
	if pkt.PacketType == AVPacket.AV_PACKET_TYPE_AUDIO {
		creater := &mp4.FMP4Creater{}
		slice := creater.AddFlvTag(tag)
		fp, err := os.Create("ahold.mp4")
		if err != nil {
			logger.LOGE(err.Error())
		}
		defer fp.Close()
		fp.Write(slice.Data)
	}
}

func TestPTSDTS(fileName string){
	reader := flv.FlvFileReader{}
	reader.Init(fileName)
	defer reader.Close()
	tag, err := reader.GetNextTag()
	if err != nil {
		return
	}

	htimer:=&H264.H264TimeCalculator{}
	for tag!=nil&&err==nil{
		if tag.TagType==flv.FLV_TAG_Video{
			if tag.Data[1]==0{
				//avc
				avc,err:=H264.DecodeAVC(tag.Data[5:])
				if err!=nil{
					logger.LOGD(tag.Data)
					logger.LOGE(err.Error())
				}
				htimer.SetSPS(avc.SPS.Front().Value.([]byte))
			}else{
				cur:=5
				for cur<len(tag.Data){
					nalSize,err:=amf.AMF0DecodeInt32(tag.Data[cur:])
					cur+=4
					if err!=nil{
						logger.LOGE(err.Error())
						return
					}
					htimer.AddNal(tag.Data[cur:cur+int(nalSize)])
					cur+=int(nalSize)
				}
			}
		}
		tag,err=reader.GetNextTag()
	}
}