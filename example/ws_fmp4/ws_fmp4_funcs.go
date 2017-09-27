package ws_fmp4

import (
	"github.com/panda-media/muxer-fmp4/codec/H264"
	"github.com/panda-media/muxer-fmp4/format/AVPacket"
	"github.com/panda-media/muxer-fmp4/format/MP4"
	"strconv"
	"net/http"
	"log"
	"github.com/gorilla/websocket"
	"sync"
	"sync/atomic"
	"encoding/json"
)

type SegmentData struct {
	moofmdat []byte
	seg_type int
}

type WS_FMP4_DEMO struct {
	fmp4 *MP4.FMP4Muxer
	path string
	port int
	sinks map[int]chan*SegmentData
	muxsink sync.RWMutex
	idx_sinks int32
	videoHeader []byte
}

func (this *WS_FMP4_DEMO)Init(sps,pps []byte,timescale int,port int,path string){
	this.sinks=make(map[int]chan *SegmentData)
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
	this.fmp4=MP4.NewMP4Muxer()
	this.fmp4.SetVideoHeader(tag,uint32(timescale))
	var err error
	this.videoHeader,err=this.fmp4.GetInitSegment()
	if err!=nil{
		log.Println(err.Error())
		return
	}
	this.path=path
	this.port=port

	go this.start()
}

func (this *WS_FMP4_DEMO)start(){
	strPort:=":"+strconv.Itoa(this.port)
	http.Handle(this.path,this)
	err:=http.ListenAndServe(strPort,nil)
	if err!=nil{
		log.Println(err.Error())
		return
	}
}

func (this *WS_FMP4_DEMO)ServeHTTP(w http.ResponseWriter,req *http.Request){
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
	this.handleWSConn(conn)
	defer func(){
		conn.Close()
	}()
}

func (this *WS_FMP4_DEMO)AddH264Frame(data []byte,timestamp int64)(err error){
	//package nal to a tag
	nalType:=data[0]&0x1f
	tag := &AVPacket.MediaPacket{}
	tag.PacketType = AVPacket.AV_PACKET_TYPE_VIDEO
	tag.TimeStamp = timestamp
	tag.Data = make([]byte, len(data)+5+4)
	if nalType==H264.NAL_IDR_SLICE{
		tag.Data[0] = 0x17
	}else{
		tag.Data[0] = 0x27
	}
	tag.Data[1] = 1
	cts := 0
	tag.Data[2] = byte((cts >> 16) & 0xff)
	tag.Data[3] = byte((cts >> 8) & 0xff)
	tag.Data[4] = byte((cts >> 0) & 0xff)
	nalSize := len(data)
	tag.Data[5] = byte((nalSize >> 24) & 0xff)
	tag.Data[6] = byte((nalSize >> 16) & 0xff)
	tag.Data[7] = byte((nalSize >> 8) & 0xff)
	tag.Data[8] = byte((nalSize >> 0) & 0xff)
	copy(tag.Data[9:], data)
	//add into fmp4
	err=this.fmp4.AddPacket(tag)
	if err!=nil{
		return
	}
	_,mdat,_,_,err:=this.fmp4.Flush()
	if err!=nil{
		return
	}
	//
	this.muxsink.RLock()
	defer this.muxsink.RUnlock()
	for _,v:=range this.sinks{
		v<-&SegmentData{mdat,int(nalType)}
	}
	return
}

//just support play now
func (this *WS_FMP4_DEMO)handleWSConn(conn *websocket.Conn){
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
					//handle_opt_play(conn,data[4:])
					this.handle_opt_play(conn,data[4:])
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

func (this *WS_FMP4_DEMO)handle_opt_play(conn *websocket.Conn,opt []byte){
	play :=&OPT_Play_Info{}
	err:=json.Unmarshal(opt, play)
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
	key,sink:=this.regSink()
	defer func(){
		this.delSink(key)
		close(sink)
	}()
	keyframed:=false
	for{
		select {
			case segment_data:=<-sink:
				if keyframed==false{
					if segment_data.seg_type==H264.NAL_IDR_SLICE{
						keyframed=true
						//send header
						err=this.sendwsVideo(conn,this.videoHeader)
						if err!=nil{
							return
						}
						//send first idr
						err=this.sendwsVideo(conn,segment_data.moofmdat)
						if err!=nil{
							return
						}
					}
				}else{
					err=this.sendwsVideo(conn,segment_data.moofmdat)
					if err!=nil{
						return
					}
				}
		}
	}
}

func (this *WS_FMP4_DEMO)regSink()(key int,sink chan *SegmentData){
	key=int(atomic.AddInt32(&this.idx_sinks,1))
	sink=make(chan *SegmentData,3)
	this.muxsink.Lock()
	defer this.muxsink.Unlock()
	this.sinks[key]=sink
	return
}

func (this *WS_FMP4_DEMO)delSink(key int){
	this.muxsink.Lock()
	defer this.muxsink.Unlock()
	delete(this.sinks,key)
	return
}

func (this *WS_FMP4_DEMO)sendwsVideo(conn *websocket.Conn,data []byte)(err error){
	dataSend := make([]byte, len(data)+1)
	dataSend[0] = AVPacket.AV_PACKET_TYPE_VIDEO
	copy(dataSend[1:],data)
	err=conn.WriteMessage(websocket.BinaryMessage,dataSend)
	return
}