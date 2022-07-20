package wav

import (
	"bytes"
	"encoding/binary"
)

const (
	sampleRate = 48000
	bitDepth   = 16
	channel    = 1
)

func AddHeader(pcm interface{}, duration int) ([]byte, error) {
	var buf bytes.Buffer

	var pcmData bytes.Buffer
	err := binary.Write(&pcmData, binary.LittleEndian, pcm)
	if err != nil {
		return nil, err
	}

	binary.Write(&buf, binary.LittleEndian, []byte("RIFF"))
	binary.Write(&buf, binary.LittleEndian, int32(pcmData.Len()+36))
	binary.Write(&buf, binary.LittleEndian, []byte("WAVEfmt "))
	binary.Write(&buf, binary.LittleEndian, int32(16))
	binary.Write(&buf, binary.LittleEndian, int16(1))
	binary.Write(&buf, binary.LittleEndian, int16(1))
	binary.Write(&buf, binary.LittleEndian, uint32(sampleRate))
	binary.Write(&buf, binary.LittleEndian, uint32(sampleRate*channel*(bitDepth/8)))
	binary.Write(&buf, binary.LittleEndian, uint16(channel*(bitDepth/8)))
	binary.Write(&buf, binary.LittleEndian, uint16(bitDepth))
	binary.Write(&buf, binary.LittleEndian, []byte("data"))

	res := append(buf.Bytes(), pcmData.Bytes()...)

	return res, nil
}
