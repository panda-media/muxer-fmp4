package commonBoxes

//true for audio,false for video
func hdlrBox(bAudio bool) (box *MP4Box, err error) {
	box, err = NewMP4Box("hdlr")
	if err != nil {
		return
	}
	box.SetVersionFlags(0, 0)
	box.Push4Bytes(0)
	if bAudio {
		box.PushBytes([]byte("soun"))
		box.Push8Bytes(0)
		box.Push4Bytes(0)
		box.PushBytes([]byte("SoundHandler"))
		box.PushByte(0)
	} else {
		box.PushBytes([]byte("vide"))
		box.Push8Bytes(0)
		box.Push4Bytes(0)
		box.PushBytes([]byte("VideoHandler"))
		box.PushByte(0)
	}
	return
}
