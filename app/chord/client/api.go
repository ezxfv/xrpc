package client

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	DefaultURL = "http://localhost:9900/chord"
)

func NewChordClient(url string) *ChordClient {
	cc := &ChordClient{
		url: url,
		c:   &http.Client{Timeout: 3 * time.Second},
	}
	return cc
}

type Client interface {
	Set(key, value string) error
	Get(key string) (value string, err error)
	Del(key string) error
}

type ChordClient struct {
	url string
	c   *http.Client
}

func (cc *ChordClient) Set(key, value string) error {
	m := map[string]string{
		"key":   key,
		"value": value,
	}
	d, _ := json.Marshal(&m)
	r := bytes.NewReader(d)
	req, err := http.NewRequest("POST", cc.url+"/set", r)
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return err
	}
	_, err = http.DefaultClient.Do(req)
	return err
}

func (cc *ChordClient) Get(key string) (value string, err error) {
	req, err := http.NewRequest("GET", cc.url+"/get?key="+key, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	data, _ := ioutil.ReadAll(resp.Body)
	value = string(data)
	return value, err
}

func (cc *ChordClient) Del(key string) error {
	req, err := http.NewRequest("DELETE", cc.url+"/del?key="+key, nil)
	if err != nil {
		return err
	}
	_, err = http.DefaultClient.Do(req)
	return err
}
