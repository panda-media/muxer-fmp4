package ws_fmp4

import (
	"strconv"
	"net/http"
	"github.com/gorilla/websocket"
	"log"
	"github.com/panda-media/muxer-fmp4/codec/H264"
	"github.com/panda-media/muxer-fmp4/format/AVPacket"
	"github.com/panda-media/muxer-fmp4/format/MP4"
	"encoding/json"
	"time"
	"os"
)

type OPT_Play_Info struct {
	Name  string `json:"name"`
	Start int    `json:"start"`
	Len   int    `json:"len"`
	Reset int    `json:"reset"`
	Req   int    `json:"req"`
}

type OPT_Result_Info struct {
	Level string `json:"level"`
	Code  string `json:"code"`
	Req   int    `json:"req"`
}

const(
	msgType_Control=18
	opt_play=1

)

type stClose struct {
	Req int `json:"req"`
}

func WSHandler(w http.ResponseWriter, req *http.Request){
	log.Println("ws req:",req.URL.String())
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	conn,err:=upgrader.Upgrade(w,req,nil)
	if err!=nil{
		log.Println(err.Error())
		return
	}
	handleWSConn(conn)
	defer func(){
		conn.Close()
	}()
}

func RunWS_server(port int,path string){
	strPort:=":"+strconv.Itoa(port)
	http.HandleFunc(path,WSHandler)
	http.ListenAndServe(strPort,nil)
}

//not care the param,just send file
func handleWSConn(conn *websocket.Conn){

	for {
		messageType,data,err:=conn.ReadMessage()
		if err!=nil{
			log.Println(err.Error())
			return
		}
		switch messageType {
		case websocket.BinaryMessage:
			msgType:=int(data[0])
			//for this sample,just play
			if msgType==msgType_Control{
				optType:=(int(data[1])<<16)|(int(data[2])<<8)|(int(data[3])<<0)
				switch optType {
				case opt_play:
					handle_opt_play(conn,data[4:])
				default:
					log.Println("opt type not processed",optType)
				}
			}else{
				//
				log.Println("msg type not processed ",messageType)
			}
		default:
			log.Println("ws msg type processed ",messageType)
			return
		}
	}


}

func handle_opt_play(conn *websocket.Conn,data []byte){
	play :=&OPT_Play_Info{}
	err:=json.Unmarshal(data, play)
	if err!=nil{
		log.Println(err.Error())
		return
	}
	//send play result,just send ok,now
	//"status""NetStream.Play.Start"
	ret:=&OPT_Result_Info{"status",
		"NetStream.Play.Start",
		play.Req}
	retData,err:=json.Marshal(ret)
	if err!=nil{
		log.Println(err.Error())
		return
	}
	dataSend:=make([]byte,len(retData)+4)
	dataSend[0] = msgType_Control
	dataSend[1] = 0
	dataSend[2] = 0
	dataSend[3] = 0
	copy(dataSend[4:], retData)
	err = conn.WriteMessage(websocket.BinaryMessage, dataSend)
	if err!=nil{
		log.Println(err.Error())
		return
	}

	//send video data here
	SendNvrMediaData(conn)
}

func SendNvrMediaData(conn *websocket.Conn){

	data := NvrReadGopTime("9")
	sps:=data.SPS
	pps:=data.PPS

	//package sps pps into media tag
	avc:=H264.AVCDecoderConfigurationRecord{}
	avc.AddSPS(sps)
	avc.AddPPS(pps)
	avc_data:=avc.AVCData()
	tag:=&AVPacket.MediaPacket{}
	tag.PacketType=AVPacket.AV_PACKET_TYPE_VIDEO
	tag.TimeStamp=0
	tag.Data=make([]byte,len(avc_data)+5)
	tag.Data[0] = 0x17
	tag.Data[1] = 0
	tag.Data[2] = 0
	tag.Data[3] = 0
	tag.Data[4] = 0
	copy(tag.Data[5:], avc_data)
	//send packet
	fmp4:=MP4.NewMP4Muxer()
	fmp4.SetVideoHeader(tag,uint32(data.TimeScaleVideo))
	log.Println("video timescale:",data.TimeScaleVideo)
	keyframed:=false
	//just video data now
	var lasttimestamp int64
	for _, sample := range data.Samples {
		//sample.Sync key frame
		//sample.AUDIO audio frame
		if !sample.AUDIO {
			tag=createTag(sample,lasttimestamp)
			lasttimestamp=tag.TimeStamp
			//duration:=int(sample.RtpDts)*1000/data.TimeScaleVideo
			//log.Println(duration)
			//time.Sleep(time.Duration(duration) *time.Millisecond)
			time.Sleep(time.Millisecond*40)
			if sample.Sync {
				//log.Println("key", (*sample.Data)[:10])
				fmp4.AddPacket(tag)
				if false==keyframed{
					//send header
					header,err:=fmp4.GetInitSegment()
					if err!=nil{
						log.Println(err.Error())
						return
					}
					err=sendVideo(conn,header)
					log.Println("send init segment")
					if err!=nil{
						return
					}
					keyframed=true
				}
				_,seg,_,_,err:=fmp4.Flush()
				if err!=nil{
					log.Println(err.Error())
					return
				}
				err=sendVideo(conn,seg)
				log.Println("send idr")
				if err!=nil{
					log.Println(err.Error())
					return
				}
			} else {
				//log.Println("not key frame", (*sample.Data)[:10])
				if keyframed{
					fmp4.AddPacket(tag)
					_,seg,_,_,err:=fmp4.Flush()
					if err!=nil{
						log.Println(err.Error())
						return
					}
					err=sendVideo(conn,seg)
					if err!=nil{
						log.Println(err.Error())
						return
					}
					log.Println("send slice")
				}
			}
		}

	}
}

func createTag(sample Frame,lasttimestamp int64)(tag *AVPacket.MediaPacket){
	tag = &AVPacket.MediaPacket{}
	tag.PacketType = AVPacket.AV_PACKET_TYPE_VIDEO
	tag.TimeStamp = sample.RtpDts+lasttimestamp
	tag.Data = make([]byte, len(*sample.Data)+5+4)
	if sample.Sync{
		tag.Data[0] = 0x17
	}else{
		tag.Data[0] = 0x27
	}
	tag.Data[1] = 1
	cts := 0
	tag.Data[2] = byte((cts >> 16) & 0xff)
	tag.Data[3] = byte((cts >> 8) & 0xff)
	tag.Data[4] = byte((cts >> 0) & 0xff)
	nalSize := len(*sample.Data)
	tag.Data[5] = byte((nalSize >> 24) & 0xff)
	tag.Data[6] = byte((nalSize >> 16) & 0xff)
	tag.Data[7] = byte((nalSize >> 8) & 0xff)
	tag.Data[8] = byte((nalSize >> 0) & 0xff)
	copy(tag.Data[9:], *sample.Data)
	return
}

func sendVideo(conn *websocket.Conn,data []byte)(err error){
	dataSend := make([]byte, len(data)+1)
	dataSend[0] = AVPacket.AV_PACKET_TYPE_VIDEO
	copy(dataSend[1:],data)
	err=conn.WriteMessage(websocket.BinaryMessage,dataSend)
	return
}

func SaveSegment(idx string,count int){
	data := NvrReadGopTime(idx)
	sps:=data.SPS
	pps:=data.PPS

	//package sps pps into media tag
	avc:=H264.AVCDecoderConfigurationRecord{}
	avc.AddSPS(sps)
	avc.AddPPS(pps)
	avc_data:=avc.AVCData()
	tag:=&AVPacket.MediaPacket{}
	tag.PacketType=AVPacket.AV_PACKET_TYPE_VIDEO
	tag.TimeStamp=0
	tag.Data=make([]byte,len(avc_data)+5)
	tag.Data[0] = 0x17
	tag.Data[1] = 0
	tag.Data[2] = 0
	tag.Data[3] = 0
	tag.Data[4] = 0
	copy(tag.Data[5:], avc_data)
	//send packet
	fmp4:=MP4.NewMP4Muxer()
	fmp4.SetVideoHeader(tag,uint32(data.TimeScaleVideo))
	header,_:=fmp4.GetInitSegment()
	fp,_:=os.Create("vide/0.mp4")
	fp.Write(header)
	fp.Close()
	ic:=0
	var lastTimestamp int64
	for _, sample := range data.Samples {
		if !sample.AUDIO {
			if ic>count{
				return
			}
			ic++
			tag=createTag(sample,lastTimestamp)
			lastTimestamp=tag.TimeStamp
			fmp4.AddPacket(tag)
			_,moofmdat,_,_,_:=fmp4.Flush()
			fp,_=os.Create("vide/"+strconv.Itoa(ic)+".mp4")
			fp.Write(moofmdat)
			fp.Close()
		}
	}
}