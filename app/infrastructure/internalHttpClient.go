package infrastructure

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var expectedContentType = "application/json"

type InternalHttpClient struct {
}

func (client *InternalHttpClient) Get(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var target string
		json.NewDecoder(resp.Body).Decode(&target)

		return nil, errors.New(target)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != expectedContentType {
		return nil, errors.New(fmt.Sprintf("Wrong content type, expected '%s', received: '%s'", expectedContentType, contentType))
	}

	scanner := bufio.NewScanner(resp.Body)
	var buf bytes.Buffer
	for scanner.Scan() {
		buf.Write(scanner.Bytes())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func NewInternalHttpClient() *InternalHttpClient {
	return &InternalHttpClient{}
}
