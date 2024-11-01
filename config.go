package prue

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
)

type config struct {
	structs interface{}
}

var (
	Config     *config
	configOnce sync.Once
)

func NewConfig(filePath string, structs interface{}) (*config, error) {
	if reflect.TypeOf(structs).Kind() != reflect.Ptr {
		return nil, fmt.Errorf("structs must be a pointer to a struct")
	}

	configOnce.Do(func() {
		Config = &config{structs: structs}
	})

	if err := Config.jsonLoader(filePath, structs); err != nil {
		return nil, err
	}

	return Config, nil
}

func (c *config) InitStructs() error {
	ref := reflect.ValueOf(c.structs)
	if ref.Kind() == reflect.Ptr && ref.Elem().Kind() == reflect.Struct {
		init := ref.MethodByName("Init")
		if init.IsValid() && init.Type().NumIn() == 0 && init.Type().NumOut() == 1 {
			result := init.Call(nil)
			if err, ok := result[0].Interface().(error); ok && err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *config) Get(key string) interface{} {
	ref := reflect.ValueOf(c.structs).Elem()
	field := ref.FieldByName(key)
	if !field.IsValid() {
		return nil
	}
	return field.Interface()
}

func (c *config) jsonLoader(path string, structure interface{}) error {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("json: %v", err)
	}
	defer file.Close()

	if err = json.NewDecoder(file).Decode(structure); err != nil {
		return fmt.Errorf("json: %v", err)
	}
	return nil
}
