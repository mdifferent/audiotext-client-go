package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

func sendBinaryData(c *websocket.Conn, data chan []byte) {
	for chunk := range data {
		//log.Printf("Read %v bytes from channel", len(chunk))
		err := c.WriteMessage(websocket.BinaryMessage, chunk)
		if err != nil {
			log.Fatal(err)
		}
	}
	c.WriteJSON(StreamingRequest{
		"command": endStream.String(),
	})
}

type StreamingRequestConfig struct {
	Language   string
	SampleRate int
	FilePath   string
	PhraseList []string
}

func handleSessionMessage(c *websocket.Conn, done chan struct{}, dataCh chan []byte, config StreamingRequestConfig) {
	defer close(done)

	request := StreamingRequest{
		"command": setLanguage.String(),
		"value":   config.Language}
	err := c.WriteJSON(request)
	if err != nil {
		log.Fatal(err)
	}

	for {
		var respJson StreamingResponse
		jsonErr := c.ReadJSON(&respJson)
		if jsonErr != nil {
			if websocket.IsUnexpectedCloseError(jsonErr) {
				break
			} else {
				log.Fatal(jsonErr)
			}
		}

		switch respJson.Type {
		case languageReady.String():
			request := StreamingRequest{"command": setSamplingRate.String(), "value": config.SampleRate}
			err := c.WriteJSON(request)
			if err != nil {
				log.Fatal(err)
			}
		case samplingRateReady.String():
			request := StreamingRequest{"command": setPhraseList.String(), "value": config.PhraseList}
			err := c.WriteJSON(request)
			if err != nil {
				log.Fatal(err)
			}
		case recognitionResult.String():
			if respJson.Status == "recognized" {
				log.Printf("%s: %s", respJson.Status, respJson.Value)
			}
		case phraseListReady.String():
			go sendBinaryData(c, dataCh)
		case languageError.String(),
			samplingRateError.String(),
			recognitionError.String(),
			phraseListError.String():
			log.Fatal("Recognization error: ", respJson.Value)
		default:
			log.Fatal("Unknown response: ", respJson)
		}
	}
}

func buildWebsocketConnection(token string) *websocket.Conn {
	addr := "translate.signans.io"

	u := url.URL{Scheme: "wss", Host: addr, Path: "/api/v1/translate/stt-streaming", RawQuery: url.PathEscape(fmt.Sprintf("token=Bearer %s", token))}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	return c
}

func readFromFile(dataCh chan []byte, filePath string) {
	defer close(dataCh)
	f, err := os.OpenFile(filePath, os.O_RDONLY, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	buffer := make([]byte, CHUNK_LENGTH)
	for {
		len, e := f.Read(buffer)
		if e != nil {
			log.Fatal(e)
		}
		//log.Printf("Write %v bytes to channel", len)
		dataCh <- buffer
		if len < CHUNK_LENGTH {
			break
		}
	}
}

func sendStreaming(token string, config StreamingRequestConfig, ch chan float64) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	c := buildWebsocketConnection(token)
	defer c.Close()

	dataCh := make(chan []byte)
	done := make(chan struct{})

	startTime := time.Now()
	go handleSessionMessage(c, done, dataCh, config)

	go readFromFile(dataCh, config.FilePath)

	for {
		select {
		case <-done:
			endTime := time.Now()
			elapsed := endTime.Sub(startTime)
			ch <- elapsed.Seconds()
			return
		case <-interrupt:
			log.Println("interrupt")
			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
