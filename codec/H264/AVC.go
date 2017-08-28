package H264

import (
	"bytes"
	"container/list"
	"errors"
)

type AVCDecoderConfigurationRecord struct {
	ConfigurationVersion         byte //aways 1
	AVCProfileIndication         byte
	Profile_compatibility        byte
	AVCLevelIndication           byte
	LengthSizeMinusOne           byte
	NumOfSequenceParameterSets   int
	SPS                          *list.List
	NumOfPictureParameterSets    int
	PPS                          *list.List
	Chroma_format_idc            byte
	Bit_depth_luma_minus8        byte
	Bit_depth_chroma_minus8      byte
	NumOfSequenceParameterSetExt int
	SequenceParameterSetExt      *list.List
}

func (this *AVCDecoderConfigurationRecord) AVCData() []byte {
	writer := &bytes.Buffer{}
	writer.WriteByte(1)
	writer.WriteByte(this.AVCProfileIndication)
	writer.WriteByte(this.Profile_compatibility)
	writer.WriteByte(this.AVCLevelIndication)
	//writer.WriteByte(0xfc | this.LengthSizeMinusOne)
	writer.WriteByte(0xff)
	writer.WriteByte(0xe0 | byte(this.NumOfSequenceParameterSets))
	writeSizeDataList(writer, this.SPS)
	writer.WriteByte(byte(this.NumOfPictureParameterSets))
	writeSizeDataList(writer, this.PPS)
	if this.Chroma_format_idc != 0 || this.NumOfSequenceParameterSetExt != 0 {
		if this.AVCProfileIndication == 100 || this.AVCProfileIndication == 110 ||
			this.AVCProfileIndication == 122 || this.AVCProfileIndication == 144 {
			writer.WriteByte(0xfc | this.Chroma_format_idc)
			writer.WriteByte(0xf8 | this.Bit_depth_luma_minus8)
			writer.WriteByte(0xf8 | this.Bit_depth_chroma_minus8)
			writer.WriteByte(byte(this.NumOfSequenceParameterSetExt))
			writeSizeDataList(writer, this.SequenceParameterSetExt)
		}
	}
	return writer.Bytes()
}

func writeSizeDataList(writer *bytes.Buffer, data *list.List) {
	if data != nil && data.Len() > 0 {
		for e := data.Front(); e != nil; e = e.Next() {
			size := len(e.Value.([]byte))
			writer.WriteByte(byte(size >> 8))
			writer.WriteByte(byte(size & 0xff))
			writer.Write(e.Value.([]byte))
		}
	}
}

func readSizeDataList(reader *bytes.Reader, count int) (lst *list.List, err error) {
	lst = list.New()
	var data []byte
	defer func() {
		if err != nil {
			lst = nil
		}
	}()
	for i := 0; i < count; i++ {
		data, err = readData(reader, 2)
		if err != nil {
			return
		}
		size := (int(data[0]) << 8) | (int(data[1]))
		if size == 0 {
			err = errors.New("invalid sps size")
			return
		}
		data, err = readData(reader, size)
		if err != nil {
			return
		}
		lst.PushBack(data)
	}

	return
}

func readData(reader *bytes.Reader, size int) (data []byte, err error) {
	if size == 0 {
		return nil, errors.New("invalid param")
	}
	data = make([]byte, size)
	n, err := reader.Read(data)
	if err != nil {
		return nil, err
	}
	if n != size {
		return nil, errors.New("no enough data for read")
	}
	return
}

func DecodeAVC(avcData []byte) (avc *AVCDecoderConfigurationRecord, err error) {
	avc = &AVCDecoderConfigurationRecord{}
	reader := bytes.NewReader(avcData)
	var data []byte
	data, err = readData(reader, 6)
	if err != nil {
		return
	}
	if data[0] != 1 {
		return avc, errors.New("invalid avc")
	}
	avc.ConfigurationVersion = data[0]
	avc.AVCProfileIndication = data[1]
	avc.Profile_compatibility = data[2]
	avc.AVCLevelIndication = data[3]
	avc.LengthSizeMinusOne = data[4] & 3
	avc.NumOfSequenceParameterSets = int(data[5] & 0x1f)
	avc.SPS, err = readSizeDataList(reader, avc.NumOfSequenceParameterSets)
	if err != nil {
		return
	}
	data, err = readData(reader, 1)
	if err != nil {
		return
	}
	avc.NumOfPictureParameterSets = int(data[0])
	avc.PPS, err = readSizeDataList(reader, avc.NumOfPictureParameterSets)
	if err != nil {
		return
	}
	if avc.AVCProfileIndication == 100 || avc.AVCProfileIndication == 110 ||
		avc.AVCProfileIndication == 122 || avc.AVCProfileIndication == 144 {

		data, err = readData(reader, 4)
		if err != nil {
			err = nil
			return
		}
		avc.Chroma_format_idc = data[0] & 3
		avc.Bit_depth_luma_minus8 = data[1] & 7
		avc.Bit_depth_chroma_minus8 = data[2] & 7
		avc.NumOfSequenceParameterSetExt = int(data[3])
		avc.SequenceParameterSetExt, err = readSizeDataList(reader, avc.NumOfSequenceParameterSetExt)
	}

	return
}

func (this *AVCDecoderConfigurationRecord) AddSPS(sps []byte) {
	if len(sps) == 0 {
		return
	}
	if nil == this.SPS {
		this.SPS = list.New()
	}
	this.SPS.PushBack(sps)
	this.NumOfSequenceParameterSets = this.SPS.Len()
	_, _, _, this.Chroma_format_idc, this.Bit_depth_luma_minus8, this.Bit_depth_chroma_minus8 = DecodeSPS(sps)
	this.AVCProfileIndication=sps[1]
	this.Profile_compatibility=sps[2]
	this.AVCLevelIndication=sps[3]
}

func (this *AVCDecoderConfigurationRecord) AddPPS(pps []byte) {
	if len(pps) == 0 {
		return
	}
	if nil == this.PPS {
		this.PPS = list.New()
	}
	this.PPS.PushBack(pps)
	this.NumOfPictureParameterSets = this.PPS.Len()
}

func (this *AVCDecoderConfigurationRecord) AddSPSExt(spsExt []byte) {
	if len(spsExt) == 0 {
		return
	}
	if nil == this.SequenceParameterSetExt {
		this.SequenceParameterSetExt = list.New()
	}
	this.SequenceParameterSetExt.PushBack(spsExt)
	this.NumOfSequenceParameterSetExt = this.SequenceParameterSetExt.Len()
}
