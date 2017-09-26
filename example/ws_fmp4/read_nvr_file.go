package ws_fmp4

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/panda-media/muxer-fmp4/codec/H264"
)

//Gop group frame
type Gop struct {
	TimeScaleAudio int               //sps frame rtmp 1000
	TimeScaleVideo int               //sps frame rtmp 1000
	SPS            []byte            //sps frame
	PPS            []byte            //pps frame
	AU             *Mpeg4AudioConfig //audio codec info
	Samples        []Frame           //frame payload data
	ATS            int64             //all time
}

//Mpeg4AudioConfig Obj Struct
type Mpeg4AudioConfig struct {
	ObjectType      uint
	ChannelConf     uint
	SampleRateIndex uint
	SizeLength      int
	IndexLength     int
}

//Frame struct
type Frame struct {
	AUDIO  bool    //is audio
	Data   *[]byte //payload
	Sync   bool    //if key
	RtpTS  int64   //t2
	RtpDts int64   //t3
	RtpLts int64   //t4
	GTS    int64   //td rtp
	GDTS   int64   //t rtp
	ALARM  bool    //-
	ALARMR []byte  //-
}


func read_nvr() {
	//saveFile()
	//return
	log.SetFlags(log.Lshortfile)

	data := NvrReadGopTime("8") //file name only num dir need .nvr and .idx
	sps := data.SPS
	pps := data.PPS
	//103 77 0 42 150 53 64 240 4 79 203 55 1 1 1 2
	//103 77 0 42 150 53 64 240 4 79 203 55 1 1 1 2
	width, height, fps, _, _, _ := H264.DecodeSPS(sps)
	log.Println(width, height, fps)
	VRate := data.TimeScaleVideo
	ARate := data.TimeScaleAudio
	log.Println(sps, pps, VRate, ARate, len(data.Samples))



	for _, sample := range data.Samples {
		//sample.Sync key frame
		//sample.AUDIO audio frame
		if !sample.AUDIO {
			if sample.Sync {
				//log.Println("key", (*sample.Data)[:10])

			} else {
				//log.Println("not key frame", (*sample.Data)[:10])
			}
		} else {
		}

	}
}

//NvrReadGopTime New GOP reader
func NvrReadGopTime(name string) *Gop {
	var AAallSamplesSS Gop

	file, err := ioutil.ReadFile(name + ".idx")
	if err != nil {
		return nil
	}
	for _, gopaddress := range bytes.Split(file, []byte("\r\n")) {
		if len(gopaddress) == 40 {
			//Gtime := ByteToInt64(gopaddress[0:8])
			fdata, err := os.Open(name + ".nvr")
			if err != nil {
				continue
			}
			_, err = fdata.Seek(ByteToInt64(gopaddress[8:16]), 0)
			if err != nil {
				fdata.Close()
				continue
			}
			dec := gob.NewDecoder(fdata)
			var allSamplesss Gop
			err = dec.Decode(&allSamplesss)
			fdata.Close()
			if err != nil {
				continue
			}
			if AAallSamplesSS.Samples == nil {
				AAallSamplesSS = allSamplesss
			} else {
				AAallSamplesSS.Samples = append(AAallSamplesSS.Samples, allSamplesss.Samples...)
			}
		}
	}
	return &AAallSamplesSS
}

//StringToInt64  to int64 conver
func StringToInt64(val string) int64 {
	i, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0
	}
	return i
}

//ByteToInt64 to int64 conver
func ByteToInt64(val []byte) int64 {
	return int64(binary.BigEndian.Uint64(val))
}

//WriteBitsADTSFrameHeader Make New Header
func WriteBitsADTSFrameHeader(w io.Writer, config Mpeg4AudioConfig, size uint) {
	bw := &BitWriter{W: w}
	fullFrameSize := 7 + size
	bw.WriteBits(12, 0xfff)              /* syncword */
	bw.WriteBits(1, 0)                   /* ID */
	bw.WriteBits(2, 0)                   /* layer */
	bw.WriteBits(1, 1)                   /* protection_absent */
	bw.WriteBits(2, config.ObjectType-1) /* profile_objecttype */
	bw.WriteBits(4, config.SampleRateIndex)
	bw.WriteBits(1, 0)                  /* private_bit */
	bw.WriteBits(3, config.ChannelConf) /* channel_configuration */
	bw.WriteBits(1, 0)                  /* original_copy */
	bw.WriteBits(1, 0)                  /* home */
	/* adts_variable_header */
	bw.WriteBits(1, 0)              /* copyright_identification_bit */
	bw.WriteBits(1, 0)              /* copyright_identification_start */
	bw.WriteBits(13, fullFrameSize) /* aac_frame_length */
	bw.WriteBits(11, 0x7ff)         /* adts_buffer_fullness */
	bw.WriteBits(2, 0)              /* number_of_raw_data_blocks_in_frame */
	bw.FlushBits()
}

//ParseAUFrame Parse frame
func ParseAUFrame(frame []byte, config Mpeg4AudioConfig) (data []byte) {
	br := bytes.NewReader(frame)
	r := &BitReader{R: br}
	headersLength, _ := r.ReadBits(16)
	headersLengthBytes := int(headersLength+7) / 8
	headerSize := config.IndexLength + config.SizeLength
	nHeaders := int(headersLength) / headerSize
	if nHeaders > 0 {
		size, _ := r.ReadBits(config.SizeLength)
		skip := 2 + headersLengthBytes
		if len(frame) < skip+int(size) {
			return
		}
		data = frame[skip : skip+int(size)]
		return
	}
	return
}

//ParseAudioSpecificConfig config
func ParseAudioSpecificConfig(r io.Reader, config *Mpeg4AudioConfig) (err error) {
	br := &BitReader{R: r}
	if config.ObjectType, err = br.ReadBits(5); err != nil {
		return
	}
	if config.ObjectType == 31 {
		if config.ObjectType, err = br.ReadBits(6); err != nil {
			return
		}
		config.ObjectType += 32
	}
	if config.SampleRateIndex, err = br.ReadBits(4); err != nil {
		if config.SampleRateIndex == 0xf {
			if config.SampleRateIndex, err = br.ReadBits(24); err != nil {
				return
			}
		}
		return
	}
	if config.ChannelConf, err = br.ReadBits(4); err != nil {
		return
	}
	return
}

//BitWriter bit writer
type BitWriter struct {
	W       io.Writer
	buf     [1]byte
	written byte
}

//WriteBits weite single bit
func (obj *BitWriter) WriteBits(n int, val uint) (err error) {
	for i := n - 1; i >= 0; i-- {
		obj.buf[0] <<= 1
		if val&(1<<uint(i)) != 0 {
			obj.buf[0] |= 1
		}
		obj.written++
		if obj.written == 8 {
			if _, err = obj.W.Write(obj.buf[:]); err != nil {
				return
			}
			obj.buf[0] = 0
			obj.written = 0
		}
	}
	return
}

//FlushBits flush bit
func (obj *BitWriter) FlushBits() (err error) {
	if obj.written > 0 {
		obj.buf[0] <<= 8 - obj.written
		if _, err = obj.W.Write(obj.buf[:]); err != nil {
			return
		}
		obj.written = 0
	}
	return
}

//BitReader bit reader
type BitReader struct {
	R    io.Reader
	buf  [1]byte
	left byte
}

//ReadBit read one bit
func (obj *BitReader) ReadBit() (res uint, err error) {
	if obj.left == 0 {
		if _, err = obj.R.Read(obj.buf[:]); err != nil {
			return
		}
		obj.left = 8
	}
	obj.left--
	res = uint(obj.buf[0]>>obj.left) & 1
	return
}

//ReadBits read n bit
func (obj *BitReader) ReadBits(n int) (res uint, err error) {
	for i := 0; i < n; i++ {
		var bit uint
		if bit, err = obj.ReadBit(); err != nil {
			return
		}
		res |= bit << uint(n-i-1)
	}
	return
}
