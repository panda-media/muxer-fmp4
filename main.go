package main

import (
	"github.com/panda-media/muxer-fmp4/example/dash"
	"log"
	"github.com/panda-media/muxer-fmp4/example/ws_fmp4"
)

func main() {
	log.SetFlags(log.Lshortfile)
	dash.FlvFileToFMP4("111.flv")
	ws_fmp4.RunWS_server(8080,"/ws/")

	return
}
