package example

import (
	"github.com/panda-media/muxer-fmp4/codec/H264"
	"github.com/panda-media/muxer-fmp4/dashSlicer"
	"logger"
	"mediaTypes/flv"
)

func TestDash(inputFileName string) {
	slicer := dashSlicer.NEWSlicer(true, 1000, 10000, 5)
	reader := &flv.FlvFileReader{}
	err := reader.Init(inputFileName)
	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	tag, err := reader.GetNextTag()
	for tag != nil && err == nil {
		if tag.TagType == flv.FLV_TAG_Video {
			if tag.Data[0] == 0x17 && tag.Data[1] == 0 {
				avc, err := H264.DecodeAVC(tag.Data[5:])
				if err != nil {
					logger.LOGE(err.Error())
					return
				}
				for e := avc.SPS.Front(); e != nil; e = e.Next() {
					nal := make([]byte, 3+len(e.Value.([]byte)))
					nal[0] = 0
					nal[1] = 0
					nal[2] = 1
					copy(nal[3:], e.Value.([]byte))
					slicer.AddH264Nals(nal)
				}
				for e := avc.PPS.Front(); e != nil; e = e.Next() {
					nal := make([]byte, 3+len(e.Value.([]byte)))
					nal[0] = 0
					nal[1] = 0
					nal[2] = 1
					copy(nal[3:], e.Value.([]byte))
					slicer.AddH264Nals(nal)
				}
			} else {
				cur := 5
				for cur < len(tag.Data) {
					size := int(tag.Data[cur]) << 24
					size |= int(tag.Data[cur+1]) << 16
					size |= int(tag.Data[cur+2]) << 8
					size |= int(tag.Data[cur+3]) << 0
					cur += 4
					nal := make([]byte, 3+size)
					nal[0] = 0
					nal[1] = 0
					nal[2] = 1
					copy(nal[3:], tag.Data[cur:cur+size])
					slicer.AddH264Nals(nal)
					cur += size
				}

			}

		} else if tag.TagType == flv.FLV_TAG_Audio {
			slicer.AddAACFrame(tag.Data[2:])
		}
		tag, err = reader.GetNextTag()
	}
}
