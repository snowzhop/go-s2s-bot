package vosk

type Message struct {
	Result []struct {
		Conf  float64
		End   float64
		Start float64
		Word  string
	}
	Text string
}

type Answer struct {
	ID    uint64
	Text  string
	Error error
}

type Adapter interface {
	ResultsChan() <-chan *Answer
	Recognize(voice []byte) uint64
}
