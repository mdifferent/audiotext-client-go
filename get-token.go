package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
)

type GetTokenRequest struct {
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
	Duration  int    `json:"duration"`
}

type GetTokenResponse struct {
	Status string `json:"status"`
	Data   struct {
		EncodedJWT string `json:"encodedJWT"`
	} `json:"data"`
}

func getToken(accessKey string, secretKey string) (string, error) {
	request := GetTokenRequest{accessKey, secretKey, 10000}
	b, err := json.Marshal(request)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := http.Post("https://translate.signans.io/api/v1/token", "application/json", bytes.NewBuffer(b))
	if err != nil {
		log.Fatal(err)
	}
	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode == 200 {
		var bodyJson GetTokenResponse
		jsonErr := json.Unmarshal(body, &bodyJson)
		if jsonErr != nil {
			log.Fatal(jsonErr)
		}
		return bodyJson.Data.EncodedJWT, nil
	} else {
		var errorJson ErrorResponse
		jsonErr := json.Unmarshal(body, &errorJson)
		if jsonErr != nil {
			log.Fatal(jsonErr)
		}
		return "", errors.New(errorJson.Data.Message)
	}
}
