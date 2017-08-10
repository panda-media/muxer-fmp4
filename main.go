package main

import (
	"github.com/panda-media/muxer-fmp4/format/AVPacket"
	"github.com/panda-media/muxer-fmp4/format/MP4"
	"mediaTypes/flv"
	"os"
	"logger"
)

func main() {
	TestFMP4FromFlvFile("muxer-fmp4/111.flv")
	return
}

func TestFMP4FromFlvFile(fileName string) {
	reader := flv.FlvFileReader{}
	reader.Init(fileName)
	defer reader.Close()
	tag, _ := reader.GetNextTag()
	pkt := TagToAVPacket(tag)
	for pkt.PacketType!=AVPacket.AV_PACKET_TYPE_AUDIO{
		tag,_=reader.GetNextTag()
		pkt=TagToAVPacket(tag)
	}
	if pkt.PacketType == AVPacket.AV_PACKET_TYPE_AUDIO {
		mux := &MP4.FMP4Muxer{}
		mux.SetAudioHeader(pkt)
		data, err := mux.GetInitSegment()
		if err != nil {
			return
		}
		fp, err := os.Create("ah.mp4")
		if err!=nil{
			logger.LOGE(err.Error())
		}
		defer fp.Close()
		fp.Write(data)
	}
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
