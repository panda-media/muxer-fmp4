package AAC

import (
	"github.com/panda-media/muxer-fmp4/utils"
	"strings"
)

const (
	AOT_NULL = iota
	// Support?                Name
	AOT_AAC_MAIN     ///< Y                       Main
	AOT_AAC_LC       ///< Y                       Low Complexity
	AOT_AAC_SSR      ///< N (code in SoC repo)    Scalable Sample Rate
	AOT_AAC_LTP      ///< Y                       Long Term Prediction
	AOT_SBR          ///< Y                       Spectral Band Replication HE-AAC
	AOT_AAC_SCALABLE ///< N                       Scalable
	AOT_TWINVQ       ///< N                       Twin Vector Quantizer
	AOT_CELP         ///< N                       Code Excited Linear Prediction
	AOT_HVXC         ///< N                       Harmonic Vector eXcitation Coding
)
const (
	AOT_TTSI      = 12 + iota ///< N                       Text-To-Speech Interface
	AOT_MAINSYNTH             ///< N                       Main Synthesis
	AOT_WAVESYNTH             ///< N                       Wavetable Synthesis
	AOT_MIDI                  ///< N                       General MIDI
	AOT_SAFX                  ///< N                       Algorithmic Synthesis and Audio Effects
	AOT_ER_AAC_LC             ///< N                       Error Resilient Low Complexity
)
const (
	AOT_ER_AAC_LTP      = 19 + iota ///< N                       Error Resilient Long Term Prediction
	AOT_ER_AAC_SCALABLE             ///< N                       Error Resilient Scalable
	AOT_ER_TWINVQ                   ///< N                       Error Resilient Twin Vector Quantizer
	AOT_ER_BSAC                     ///< N                       Error Resilient Bit-Sliced Arithmetic Coding
	AOT_ER_AAC_LD                   ///< N                       Error Resilient Low Delay
	AOT_ER_CELP                     ///< N                       Error Resilient Code Excited Linear Prediction
	AOT_ER_HVXC                     ///< N                       Error Resilient Harmonic Vector eXcitation Coding
	AOT_ER_HILN                     ///< N                       Error Resilient Harmonic and Individual Lines plus Noise
	AOT_ER_PARAM                    ///< N                       Error Resilient Parametric
	AOT_SSC                         ///< N                       SinuSoidal Coding
	AOT_PS                          ///< N                       Parametric Stereo
	AOT_SURROUND                    ///< N                       MPEG Surround
	AOT_ESCAPE                      ///< Y                       Escape Value
	AOT_L1                          ///< Y                       Layer 1
	AOT_L2                          ///< Y                       Layer 2
	AOT_L3                          ///< Y                       Layer 3
	AOT_DST                         ///< N                       Direct Stream Transfer
	AOT_ALS                         ///< Y                       Audio LosslesS
	AOT_SLS                         ///< N                       Scalable LosslesS
	AOT_SLS_NON_CORE                ///< N                       Scalable LosslesS (non core)
	AOT_ER_AAC_ELD                  ///< N                       Error Resilient Enhanced Low Delay
	AOT_SMR_SIMPLE                  ///< N                       Symbolic Music Representation Simple
	AOT_SMR_MAIN                    ///< N                       Symbolic Music Representation Main
	AOT_USAC_NOSBR                  ///< N                       Unified Speech and Audio Coding (no SBR)
	AOT_SAOC                        ///< N                       Spatial Audio Object Coding
	AOT_LD_SURROUND                 ///< N                       Low Delay MPEG Surround
	AOT_USAC                        ///< N                       Unified Speech and Audio Coding
)

const (
	SAMPLE_SIZE=1024
)

type AACAudioSpecificConfig struct {
	object_type        int
	sampling_index     int
	sample_rate        int
	chan_config        int
	sbr                int
	ext_object_type    int
	ext_sampling_index int
	ext_sample_rate    int
	ext_chan_config    int
	channels           int
	ps                 int
	frame_length_short int
}

func parseConfigALS(reader *utils.BitReader, asc *AACAudioSpecificConfig) {
	if reader.BitsLeft() < 112 {
		return
	}
	if reader.ReadBits(8) != 'A' || reader.ReadBits(8) != 'L' || reader.ReadBits(8) != 'S' || reader.ReadBits(8) != 0 {
		return
	}
	asc.sample_rate = int(reader.Read32Bits())
	reader.Read32Bits()
	asc.chan_config = 0
	asc.channels = reader.ReadBits(16) + 1
}

func getObjectType(reader *utils.BitReader) int {
	objType := reader.ReadBits(5)
	if AOT_ESCAPE == objType {
		objType = 32 + reader.ReadBits(6)
	}
	return objType
}

func getSampleRate(reader *utils.BitReader) (sampleRateIdx, sampleRate int) {
	sampleRateIdx = reader.ReadBits(4)
	if sampleRateIdx == 0xf {
		sampleRate = reader.ReadBits(24)
	} else {
		sampleRate = func(idx int) int {
			AACSampleRates := [16]int{96000, 88200, 64000, 48000, 44100, 32000,
				24000, 22050, 16000, 12000, 11025, 8000, 7350}
			return AACSampleRates[idx]
		}(sampleRateIdx)
	}
	return sampleRateIdx, sampleRate
}

func AACGetConfig(data []byte) (asc *AACAudioSpecificConfig) {
	reader := &utils.BitReader{}
	reader.Init(data)
	asc = &AACAudioSpecificConfig{}
	asc.object_type = getObjectType(reader)
	asc.sampling_index, asc.sample_rate = getSampleRate(reader)

	asc.chan_config = reader.ReadBits(4)
	if asc.chan_config < 8 {
		asc.channels = func(idx int) int {
			arr := []int{0, 1, 2, 3, 4, 5, 6, 8}
			return arr[idx]
		}(asc.chan_config)
	}

	asc.sbr = -1
	asc.ps = -1

	if AOT_SBR == asc.object_type ||
		(AOT_PS == asc.object_type && 0 == (reader.CopyBits(3)&0x03) && 0 == (reader.CopyBits(9)&0x3f)) {
		if AOT_PS == asc.object_type {
			asc.ps = 1
		}
		asc.ext_object_type = AOT_SBR
		asc.sbr = 1
		asc.ext_sampling_index, asc.ext_sample_rate = getSampleRate(reader)
		asc.object_type = getObjectType(reader)
		if asc.object_type == AOT_ER_BSAC {
			asc.ext_chan_config = reader.ReadBits(4)
		}
	} else {
		asc.ext_object_type = AOT_NULL
		asc.ext_sample_rate = 0
	}

	if AOT_ALS == asc.object_type {
		reader.ReadBits(5)
		als := reader.CopyBits(24)
		if ((als>>16)&0xff) != 'A' || ((als>>8)&0xff) != 'L' || ((als)&0xff) != 'S' {
			reader.ReadBits(24)
		}
		parseConfigALS(reader, asc)
	}

	if asc.ext_object_type != AOT_SBR {
		for reader.BitsLeft() > 15 {
			if 0x2b7 == reader.CopyBits(11) {
				reader.ReadBits(11)
				asc.ext_object_type = getObjectType(reader)
				if asc.ext_object_type == AOT_SBR {
					asc.sbr = reader.ReadBit()
					if asc.sbr == 1 {
						asc.ext_sampling_index, asc.ext_sample_rate = getSampleRate(reader)
						if asc.ext_sample_rate == asc.sample_rate {
							asc.sbr = -1
						}
					}
					if reader.BitsLeft() > 11 && reader.ReadBits(11) == 0x548 {
						asc.ps = reader.ReadBit()
					}
					break
				}
			} else {
				reader.ReadBit()
			}
		}
	}

	if asc.sbr == 0 {
		asc.ps = 0
	}
	if (asc.ps == -1 && asc.object_type == AOT_AAC_LC) || (asc.channels&^0x01) != 0 {
		asc.ps = 0
	}

	return
}

func (this *AACAudioSpecificConfig) ObjectType() int {
	return this.object_type
}

func (this *AACAudioSpecificConfig) SampleRate() int {
	if this.ext_sample_rate > 0 {
		return this.ext_sample_rate
	}
	return this.sample_rate
}

func (this *AACAudioSpecificConfig) Channel() int {
	return this.channels
}

func ASCForMP4(ascData []byte, useragent string) (cfg []byte) {
	asc := AACGetConfig(ascData)
	if len(useragent) > 0 {
		useragent = strings.ToLower(useragent)
	}
	switch useragent {
	case "firefox":
		if asc.sampling_index >= AOT_AAC_SCALABLE {
			asc.object_type = AOT_SBR
			asc.ext_sampling_index = asc.sampling_index - 3
			cfg = make([]byte, 4)
		} else {
			asc.object_type = AOT_AAC_LC
			asc.ext_sampling_index = asc.sampling_index
			cfg = make([]byte, 2)
		}
	case "android":
		asc.object_type = AOT_AAC_LC
		asc.ext_sampling_index = asc.sampling_index
		cfg = make([]byte, 2)
	default:
		asc.object_type = AOT_SBR
		asc.ext_sampling_index = asc.sampling_index
		cfg = make([]byte, 4)
		if asc.sampling_index >= AOT_AAC_SCALABLE {
			asc.ext_sampling_index = asc.sampling_index - 3
		} else if asc.chan_config == 1 {
			asc.object_type = AOT_AAC_LC
			asc.ext_sampling_index = asc.sampling_index
			cfg = make([]byte, 2)
		}
	}
	cfg[0] = byte(asc.object_type << 3)
	cfg[0] |= byte((asc.sampling_index & 0xf) >> 1)
	cfg[1] = byte((asc.sampling_index & 0xf) << 7)
	cfg[1] |= byte((asc.chan_config & 0xf) << 3)
	if asc.object_type == AOT_SBR {
		cfg[1] |= byte((asc.ext_sampling_index & 0xf) >> 1)
		cfg[2] = byte((asc.ext_sampling_index & 1) << 7)
		cfg[2] |= (2 << 2)
		cfg[3] = 0
	}
	return
}
