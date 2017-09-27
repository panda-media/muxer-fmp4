package ws_fmp4

import "time"

/*
in this demo,we read file loop to simulate living stream
*/
func WSFMP4Demo(){

	data := NvrReadGopTime("9")

	wsInstance:=&WS_FMP4_DEMO{}

	wsInstance.Init(data.SPS,data.PPS,data.TimeScaleVideo,8080,"/ws/")
	var timestamp int64
	//just loop
	for{
		for _, sample := range data.Samples {
			if !sample.AUDIO {
				timestamp+=sample.RtpDts
				wsInstance.AddH264Frame(*sample.Data,timestamp)
				//for time sync
				durationMS:=int(sample.RtpDts)*1000/data.TimeScaleVideo
				time.Sleep(time.Duration(durationMS)*time.Millisecond)
			}
		}
	}
}
