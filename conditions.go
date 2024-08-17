package main

import (
	"errors"
	"os"
)

type FileExistsCondition struct {
	FileName string
}

func (self FileExistsCondition) Check() (bool, error) {
	_, err := os.Stat(self.FileName)
	if err != nil {
		return false, err
	}

	return true, nil
}

type FileNotExistsCondition struct {
	FileName string
}

func (self FileNotExistsCondition) Check() (bool, error) {
	_, err := os.Stat(self.FileName)
	if errors.Is(err, os.ErrNotExist) {
		return true, nil
	}

	if err != nil {
		return false, err
	}

	return false, nil
}
