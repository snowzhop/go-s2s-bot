package tools

import (
	"io"
	"net/http"
	"os"
)

func DownloadData(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	voiceMsg, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return voiceMsg, err
}

func SaveToFile(d []byte) {
	f, _ := os.Create("output.ogg")
	defer f.Close()

	f.Write(d)
}
