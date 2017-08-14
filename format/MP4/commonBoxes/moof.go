package commonBoxes

import "errors"

func moofBox(sequence_number uint32, trackIDA uint32, earlierDurationA uint64, paramTrunA *TRUN, trackIDV uint32, earlierDurationV uint64, paramTrunV *TRUN) (box *MP4Box, err error) {
	box, err = NewMP4Box("moof")
	if err != nil {
		return
	}
	mfhd, err := mfhdBox(sequence_number)
	if err != nil {
		return
	}
	box.PushBox(mfhd)

	//traf audio
	if nil != paramTrunA {
		trafA, err := trafBox(trackIDA, earlierDurationA, paramTrunA)
		if err != nil {
			return nil, err
		}
		box.PushBox(trafA)
	}
	//traf video
	if nil != paramTrunV {
		trafV, err := trafBox(trackIDV, earlierDurationV, paramTrunV)
		if err != nil {
			return nil, err
		}
		box.PushBox(trafV)
	}
	return
}

func Box_moof_Data(sequence_number uint32, earlierDurationA uint64, paramTrunA *TRUN, earlierDurationV uint64, paramTrunV *TRUN) (data []byte, err error) {
	if nil == paramTrunA && nil == paramTrunV {
		return nil, errors.New("no audio and video trun param data")
	}
	moof, err := moofBox(sequence_number, TRACK_AUDIO, earlierDurationA, paramTrunA, TRACK_VIDEO, earlierDurationV, paramTrunV)
	if err != nil {
		return
	}
	data = moof.Flush()
	return
}
