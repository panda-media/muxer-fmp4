package dashSlicer

type FMP4Receiver interface {
	VideoHeaderGenerated(videoHeader []byte)
	VideoSegmentGenerated(videoSegment []byte,timestamp int64,duration int)//in video timescale
	AudioHeaderGenerated(audioHeader []byte)
	AudioSegmentGenerated(audioSegment []byte,timestamp int64,duration int)//in audio timescale
}