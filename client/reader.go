package client

import "time"

type Message []byte

type FileReaderConfig struct {
	itv  time.Duration
	file string
	lock string
}

func NewFileReaderConfig(file, lock string, itv time.Duration) FileReaderConfig {
	return FileReaderConfig{
		file: file,
		lock: lock,
		itv:  itv,
	}
}

type FileReader struct {
	cfg    FileReaderConfig
	sendch chan Message
	quitch chan struct{}
}

func NewFileReader(cfg FileReaderConfig) *FileReader {
	return &FileReader{
		cfg:    cfg,
		sendch: make(chan Message),
		quitch: make(chan struct{}),
	}
}

func (f *FileReader) Run() {

}
