package local

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	vosk_go "github.com/alphacep/vosk-api/go"
	"github.com/snowzhop/go-s2t-bot/internal/vosk"
)

const (
	voskModelPath   = "VOSK_MODEL_PATH"
	sampleRate      = 48000
	bufferSizeLimit = 1331599 // ~ 10 sec
)

type Local struct {
	idMutex sync.Mutex
	lastID  uint64

	recognizerMutex sync.Mutex
	recognizer      *vosk_go.VoskRecognizer

	results chan *vosk.Answer
}

func NewAdapter() (vosk.Adapter, error) {
	modelPath := os.Getenv(voskModelPath)
	if modelPath == "" {
		return nil, fmt.Errorf("empty '%s'", voskModelPath)
	}

	model, err := vosk_go.NewModel(modelPath)
	if err != nil {
		return nil, fmt.Errorf("unable to create new model: %w", err)
	}

	recognizer, err := vosk_go.NewRecognizer(model, sampleRate)
	if err != nil {
		return nil, fmt.Errorf("unabke to create new recognizer: %w", err)
	}
	recognizer.SetWords(1)

	return &Local{
		results:    make(chan *vosk.Answer, 20),
		recognizer: recognizer,
	}, nil
}

func (l *Local) ResultsChan() <-chan *vosk.Answer {
	return l.results
}

func (l *Local) Recognize(voice []byte) uint64 {
	l.idMutex.Lock()
	l.lastID++
	currentID := l.lastID
	l.idMutex.Unlock()

	go l.recognize(voice, currentID)

	return currentID
}

func (l *Local) recognize(voice []byte, id uint64) {
	ans := &vosk.Answer{ID: id}
	defer func() {
		l.results <- ans
	}()

	var text []string
	var err error

	reader := bytes.NewReader(voice)
	bufSize := bufferSizeLimit
	if len(voice) < bufSize {
		bufSize = len(voice)
	}

	buf := make([]byte, bufSize)

	l.recognizerMutex.Lock()
	defer l.recognizerMutex.Unlock()
	for {
		read, err := reader.Read(buf)
		if err != nil {
			break
		}
		fmt.Printf("\tDEBUG: read = %d\n", read)

		if l.recognizer.AcceptWaveform(buf) != 0 {
			var m vosk.Message
			err = json.Unmarshal(l.recognizer.Result(), &m)
			if err != nil {
				break
			}
			text = append(text, m.Text)
		}
	}
	if err != nil && err != io.EOF {
		ans.Error = err
		return
	}

	var lastResult vosk.Message
	err = json.Unmarshal(l.recognizer.FinalResult(), &lastResult)
	if err != nil {
		ans.Error = err
		return
	}
	text = append(text, lastResult.Text)
	ans.Text = strings.Join(text, " ")
}
