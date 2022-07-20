package voice

import (
	"bytes"
	"fmt"
	"io"

	"github.com/snowzhop/go-s2t-bot/internal/wav"
	"gopkg.in/hraban/opus.v2"
)

const (
	frameSizeMs = 60
	frameSize   = 1 * frameSizeMs * 48000 / 1000
)

func OpusToWav(opusData []byte, duration int) ([]byte, error) {
	voiceStream, err := opus.NewStream(bytes.NewReader(opusData))
	if err != nil {
		return nil, err
	}

	rawVoice := make([]int16, frameSize)
	var pcm []int16
	for {
		n, internalError := voiceStream.Read(rawVoice)
		if internalError != nil {
			break
		}
		pcm = append(pcm, rawVoice[:n]...)
	}

	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("stream reading error: %w", err)
	}

	wavVoice, err := wav.AddHeader(pcm, duration)
	if err != nil {
		return nil, fmt.Errorf("error adding wav header: %w", err)
	}

	return wavVoice, nil
}
