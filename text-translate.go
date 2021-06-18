package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type TextTranslateRequest struct {
	Text       []string `json:"text"`
	SourceLang string   `json:"sourceLang"`
	TargetLang string   `json:"targetLang"`
}

type TranslationResponse struct {
	Status string `json:"status"`
	Data   struct {
		TaskId            string `json:"taskId"`
		TranslationResult []struct {
			SourceText     string `json:"sourceText"`
			TranslatedText string `json:"translatedText"`
		} `json:"translationResult"`
		WordCount int `json:"wordCount"`
	} `json:"data"`
}

func sendTranslateRequest(text []string, sourceLang string, targetLang string, token string, c chan float64) {
	requestBody := TextTranslateRequest{text, sourceLang, targetLang}
	b, err := json.Marshal(requestBody)
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{}
	req, _ := http.NewRequest("POST", "https://translate.signans.io/api/v1/translate", bytes.NewBuffer(b))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", token))
	req.Header.Add("Content-Type", "application/json")

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	end := time.Now()
	elapsed := end.Sub(start)
	log.Printf("%v elapsed", elapsed.Seconds())
	c <- elapsed.Seconds()

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode == 200 {
		var bodyJson TranslationResponse
		jsonErr := json.Unmarshal(body, &bodyJson)
		if jsonErr != nil {
			log.Fatal(jsonErr)
		}
		translatedList := make([]string, len(bodyJson.Data.TranslationResult))
		for i, v := range bodyJson.Data.TranslationResult {
			translatedList[i] = v.TranslatedText
		}
	} else {
		var errorJson ErrorResponse
		jsonErr := json.Unmarshal(body, &errorJson)
		if jsonErr != nil {
			log.Fatal(jsonErr)
		}
	}
}
