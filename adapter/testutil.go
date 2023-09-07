package adapter

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func writeJson(data interface{}) string {
	bytes, _ := json.Marshal(data)
	return string(bytes)
}

func readJson(file string, obj interface{}) error {
	bytes, err := os.ReadFile(filepath.Join("testdata", file))
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &obj)
	if err != nil {
		return err
	}
	return nil
}
