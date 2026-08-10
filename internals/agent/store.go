package agent

import (
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
)

const (
	filePath = "/tmp/sgrok/"
	fileName = "sgrok.json"
)

type State struct {
	ClientID string
}

func (s State) Save() error {
	data, err := json.MarshalIndent(s, "", " ")
	if err != nil {
		return err
	}

	tmp, err := createStateFile()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}

	return nil
}

func (s *State) Load() error {
	tmp, err := createStateFile()
	if err != nil {
		return err
	}
	defer tmp.Close()

	data, err := io.ReadAll(tmp)
	if err != nil {
		return err
	}

	if data != nil {
		err = json.Unmarshal(data, &s)
		if err != nil {
			return err
		}
	}
	return nil
}

func createStateFile() (*os.File, error) {
	tmp, err := os.OpenFile(filePath+fileName, os.O_RDWR, 0o600)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if err = os.MkdirAll(filePath, 0o700); err != nil {
				return nil, err
			}
			tmp, err = os.Create(filePath + fileName)
			if err != nil {
				return nil, err
			}
			if err = tmp.Chmod(0o600); err != nil {
				tmp.Close()
				return nil, err
			}
		}
		return nil, err
	}

	return tmp, nil
}
