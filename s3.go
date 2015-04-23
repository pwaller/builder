package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/service/s3"
)

func FetchMetadata(path string) (string, error) {
	resp, err := http.Get("http://169.254.169.254/latest/" + path)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

type InstanceInfo struct {
	InstanceID string `json:"instanceId"`
	Region     string `json:"region"`
}

func GetInstanceInfo() (*InstanceInfo, error) {

	document, err := FetchMetadata("dynamic/instance-identity/document")
	if err != nil {
		return nil, fmt.Errorf("Failed to contant metadata endpoint: %v", err)
	}

	var instanceInfo InstanceInfo

	err = json.Unmarshal([]byte(document), &instanceInfo)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse instance-identity document: %v", err)
	}

	return &instanceInfo, nil
}

func S3() {
	info, err := GetInstanceInfo()
	if err != nil {
		log.Fatal("GetInstanceInfo: %v", err)
	}

	svc := s3.New(&aws.Config{
		Region: info.Region,
	})

	_ = svc
}
