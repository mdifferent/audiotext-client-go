package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type ErrorResponse struct {
	Status string `json:"status"`
	Data   struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"data"`
}

type Account struct {
	Url           string `json:"url"`
	IsSecProtocal bool   `json:"secProtocal"`
	AccessKey     string `json:"accessKey"`
	SecretKey     string `json:"secretKey"`
}

func main() {
	// Output log to file
	logFile, err := os.OpenFile("logfile.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	//Read account info
	accountFile, err := os.OpenFile("account.json", os.O_RDONLY, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer accountFile.Close()

	byteValue, _ := ioutil.ReadAll(accountFile)
	var account Account
	err = json.Unmarshal(byteValue, &account)
	if err != nil {
		log.Fatalf("Parse account info failed: %v", err)
	}

	url := account.Url
	isSecProtocal := account.IsSecProtocal
	accessKey := account.AccessKey
	secretKey := account.SecretKey

	//Get token
	var protocal string
	if isSecProtocal {
		protocal = "https"
	} else {
		protocal = "http"
	}
	completeUrl := fmt.Sprintf("%s://%s", protocal, url)
	token, err := getToken(completeUrl, accessKey, secretKey)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(token)

	// Translate test
	requestCount := 1
	totalTime := 0.0
	totalCount := 0
	ch := make(chan float64, requestCount*2)
	/*for i := 0; i < requestCount; i++ {
		go sendTranslateRequest([]string{"更新しました。"}, "ja", "en", token, ch)
	}

	for i := 0; i < requestCount; i++ {
		totalCount++
		totalTime += <-ch
	}
	log.Printf("Averange time of text translation %v elapsed\n", totalTime/float64(totalCount))*/

	//Read streaming config
	streamingConfigFile, err := os.OpenFile("streaming-config.json", os.O_RDONLY, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer streamingConfigFile.Close()

	byteValue, _ = ioutil.ReadAll(streamingConfigFile)
	var streamingConfig StreamingRequestConfig
	err = json.Unmarshal(byteValue, &streamingConfig)
	if err != nil {
		log.Fatal(err)
	}

	//Streaming test
	totalTime = 0.0
	totalCount = 0

	if isSecProtocal {
		protocal = "wss"
	} else {
		protocal = "ws"
	}

	for i := 0; i < requestCount; i++ {
		go sendStreaming(protocal, url, token, streamingConfig, ch)
	}

	for i := 0; i < requestCount; i++ {
		totalCount++
		totalTime += <-ch
	}

	log.Printf("Averange time of streaming requests %v elapsed\n", totalTime/float64(totalCount))

}
