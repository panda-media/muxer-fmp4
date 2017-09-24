package main

import (
	"github.com/panda-media/muxer-fmp4/example"
	"log"
)

func main() {
	log.SetFlags(log.Lshortfile)
	example.FlvFileToFMP4("111.flv")
	return
}
