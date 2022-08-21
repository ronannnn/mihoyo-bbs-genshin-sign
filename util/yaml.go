package util

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

func WriteYaml(filename string, data interface{}) (err error) {
	var file *os.File
	if file, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600); err != nil {
		return fmt.Errorf("error opening/creating file: %v", err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	enc := yaml.NewEncoder(file)
	err = enc.Encode(data)
	return
}

// ReadYaml TODO
func ReadYaml(filename string) (data interface{}, err error) {
	return
}
