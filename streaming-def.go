package main

type StreamingCommand int

const (
	setLanguage StreamingCommand = iota
	setSamplingRate
	endStream
	endSession
	setPhraseList
)

func (d StreamingCommand) String() string {
	return [...]string{
		"SET_LANGUAGE",
		"SET_SAMPLING_RATE",
		"END_STREAM",
		"END_SESSION",
		"SET_PHRASE_LIST"}[d]
}

type StreamingResponseType int

const (
	languageReady StreamingResponseType = iota
	languageError
	samplingRateReady
	samplingRateError
	recognitionResult
	recognitionError
	phraseListReady
	phraseListError
)

func (d StreamingResponseType) String() string {
	return [...]string{
		"LANGUAGE_READY",
		"SET_LANGUAGE_ERROR",
		"SAMPLING_RATE_READY",
		"SET_SAMPLING_RATE_ERROR",
		"RECOGNITION_RESULT",
		"RECOGNITION_ERROR",
		"PHRASE_LIST_READY",
		"PHRASE_LIST_ERROR"}[d]
}

type StreamingRequest map[string]interface{}

type StreamingResponse struct {
	Type   string `json:"type"`
	Value  string `json:"value,omitempty"`
	Status string `json:"status,omitempty"`
}

const CHUNK_LENGTH = 8192 * 8
