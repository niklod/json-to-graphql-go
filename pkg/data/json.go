package data

import (
	"encoding/json"
	"os"
)

type JsonProvider struct {
}

func NewJSONProvider() *JsonProvider {
	return &JsonProvider{}
}

func (j *JsonProvider) GetJsonData() (map[string]interface{}, error) {
	data, err := os.ReadFile("./data.json")
	if err != nil {
		return nil, err
	}

	var res map[string]interface{}
	if err = json.Unmarshal(data, &res); err != nil {
		return nil, err
	}

	return res, nil
}

func (j *JsonProvider) GetRawJson() ([]byte, error) {
	data, err := os.ReadFile("./data.json")
	if err != nil {
		return nil, err
	}

	return data, nil
}
