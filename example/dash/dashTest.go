package dash

import (
	"github.com/panda-media/muxer-fmp4/codec/H264"
	"github.com/panda-media/muxer-fmp4/dashSlicer"
	"log"
	"os"
)

type testReceiver struct {
	fpVideo *os.File
	fpAudio *os.File
}

func (this *testReceiver)Init(saveAudio,saveVideo bool){
	if saveAudio{
		var err error
		this.fpAudio, err = os.Create("audio.mp4")
		if err!=nil{
			log.Fatal(err.Error())
			return
		}
	}
	if saveVideo{
		var err error
		this.fpVideo,err=os.Create("video.mp4")
		if err!=nil{
			log.Fatal(err.Error())
			return
		}

	}
}

func (this *testReceiver)Close(){
	if this.fpAudio!=nil{
		this.fpAudio.Close()
		this.fpAudio=nil
	}
	if this.fpVideo!=nil{
		this.fpVideo.Close()
		this.fpVideo=nil
	}
}

func (this *testReceiver) VideoHeaderGenerated(videoHeader []byte) {
	log.Println("get videoHeader,length:", len(videoHeader))
	if this.fpVideo!=nil{
		this.fpVideo.Write(videoHeader)
	}
}
func (this *testReceiver) VideoSegmentGenerated(videoSegment []byte, timestamp int64, duration int) {
	log.Println("get video segment,length:", len(videoSegment), "\ttimestamp:", timestamp, "\tduration:", duration)
	if this.fpVideo!=nil{
		this.fpVideo.Write(videoSegment)
	}
}
func (this *testReceiver) AudioHeaderGenerated(audioHeader []byte) {
	log.Println("get audioHeader,length:", len(audioHeader))
	if this.fpAudio!=nil{
		this.fpAudio.Write(audioHeader)
	}
}
func (this *testReceiver) AudioSegmentGenerated(audioSegment []byte, timestamp int64, duration int) {
	log.Println("get audio segment,length:", len(audioSegment), "\ttimestamp:", timestamp, "\tduration:", duration)
	if this.fpAudio!=nil{
		this.fpAudio.Write(audioSegment)
	}
}

func FlvFileToFMP4(flvFileName string) {
	receiver := &testReceiver{}
	receiver.Init(true,true)
	defer receiver.Close()
	slicer, err := dashSlicer.NEWSlicer(25, 1000, 1000,1000, 5000, 5, receiver)
	if err != nil {
		log.Println(err.Error())
	}
	reader := &FlvFileReader{}
	err = reader.Init(flvFileName)
	if err != nil {
		log.Println(err.Error())
		return
	}

	defer func() {
		mpd, err := slicer.GetMPD()
		if err != nil {
			log.Println(err)
		} else {
			log.Println("the last MPD sample:\n", string(mpd))
		}
	}()
	count := 0
	audioCount := 0
	videoCount := 0
	tag, err := reader.GetNextTag()
	for tag != nil && err == nil {
		if count > 4000 {
			return
		}
		count++
		if tag.TagType == FLV_TAG_Video {
			if tag.Data[0] == 0x17 && tag.Data[1] == 0 {
				avc, err := H264.DecodeAVC(tag.Data[5:])
				if err != nil {
					log.Println(err.Error())
					return
				}
				for e := avc.SPS.Front(); e != nil; e = e.Next() {
					nal := make([]byte, 3+len(e.Value.([]byte)))
					nal[0] = 0
					nal[1] = 0
					nal[2] = 1
					copy(nal[3:], e.Value.([]byte))
					slicer.AddH264Nals(nal, 0)
				}
				for e := avc.PPS.Front(); e != nil; e = e.Next() {
					nal := make([]byte, 3+len(e.Value.([]byte)))
					nal[0] = 0
					nal[1] = 0
					nal[2] = 1
					copy(nal[3:], e.Value.([]byte))
					slicer.AddH264Nals(nal, 0)
				}
			} else {
				cts := 0
				cts = int(tag.Data[2]) << 16
				cts |= int(tag.Data[3]) << 8
				cts |= int(tag.Data[4]) << 0
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
					slicer.AddH264Nals(nal, int64(tag.Timestamp))
					cur += size
				}

			}
			videoCount++
		} else if tag.TagType == FLV_TAG_Audio {
			//audio timestamp ,translate to AudioSampeHz timescale
			timestamp:=int64(tag.Timestamp*48000/1000)
			timestamp=int64(tag.Timestamp)
			slicer.AddAACFrame(tag.Data[2:], timestamp)
			audioCount++
		}
		tag, err = reader.GetNextTag()
	}
}
