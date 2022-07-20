package voskws

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/snowzhop/go-s2t-bot/internal/vosk"
)

type WebSocketAdapter struct {
	lastID  uint64
	wsConn  *websocket.Conn
	results chan *vosk.Answer

	Url string
}

func NewAdapter() vosk.Adapter {
	return &WebSocketAdapter{
		results: make(chan *vosk.Answer, 20),
	}
}

func NewVoskConn(url string) (*WebSocketAdapter, error) {
	w, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	return &WebSocketAdapter{wsConn: w}, nil
}

func (v *WebSocketAdapter) SetUrl(u url.URL) {
	v.Url = u.String()
}

func (v *WebSocketAdapter) ResultsChan() <-chan *vosk.Answer {
	return v.results
}

func (v *WebSocketAdapter) RecognizeDebug(wavVoice []byte) (string, error) {
	log.Printf("debug: len(pcm) = %d", len(wavVoice))

	err := v.wsConn.WriteMessage(websocket.BinaryMessage, wavVoice)
	if err != nil {
		return "", fmt.Errorf("ws writing message error: %w", err)
	}

	err = v.wsConn.WriteMessage(websocket.TextMessage, []byte("{\"eof\" : 1}"))
	if err != nil {
		return "", fmt.Errorf("ws writing end message error: %w", err)
	}

	// Возможно стоит удалить
	_, _, err = v.wsConn.ReadMessage()
	if err != nil {
		return "", fmt.Errorf("ws reading error: %w", err)
	}
	_, msg, err := v.wsConn.ReadMessage()
	if err != nil {
		return "", fmt.Errorf("ws reading (2) error: %w", err)
	}

	var m vosk.Message
	err = json.Unmarshal(msg, &m)
	if err != nil {
		return "", fmt.Errorf("umarhaling error: %w", err)
	}

	log.Printf("debug result: %v", m)

	return m.Text, nil
}

func (v *WebSocketAdapter) Recognize(voice []byte) uint64 {
	v.lastID++

	go v.recognize(voice, v.lastID)

	return v.lastID
}

func (v *WebSocketAdapter) recognize(voice []byte, id uint64) {
	text, err := recognizeHelper(v.Url, voice)

	ans := &vosk.Answer{ID: id}
	if err != nil {
		ans.Error = err
	} else {
		ans.Text = text
	}

	v.results <- ans
}

func recognizeHelper(url string, voice []byte) (string, error) {
	w, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return "", fmt.Errorf("dialing error: %w", err)
	}

	err = w.WriteMessage(websocket.BinaryMessage, voice)
	if err != nil {
		return "", fmt.Errorf("error writing voice message: %w", err)
	}

	err = w.WriteMessage(websocket.TextMessage, []byte("{\"eof\" : 1}"))
	if err != nil {
		return "", fmt.Errorf("error writing EOF message: %w", err)
	}

	_, _, err = w.ReadMessage()
	if err != nil {
		return "", fmt.Errorf("reading error: %w", err)
	}
	_, msg, err := w.ReadMessage()
	if err != nil {
		return "", fmt.Errorf("reading (2) error: %w", err)
	}

	var m vosk.Message
	err = json.Unmarshal(msg, &m)
	if err != nil {
		return "", fmt.Errorf("umarhaling error: %w", err)
	}

	return m.Text, nil
}
