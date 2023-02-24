package utils

import (
	"encoding/json"
	"io"
)

func JsonEncode(payload any) []byte {
	bytes, err := json.Marshal(payload)
	if err != nil {
		panic("Error serializing JSON")
	}
	return bytes
}

func JsonDecode[T any](body io.ReadCloser) T {
	var value T
	json.NewDecoder(body).Decode(&value)
	return value
}

func JsonDecodeByteStream[T any](data []byte) (*T, error) {
	var value T
	err := json.Unmarshal(data, &value)
	if err != nil {
		return nil, err
	}
	return &value, nil
}
