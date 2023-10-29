// Copyright (c) 2023 BVK Chaitanya

package kvhttp

import (
	"errors"
	"os"
)

func error2string(err error) string {
	if errors.Is(err, os.ErrInvalid) {
		return "ErrInvalid"
	}
	if errors.Is(err, os.ErrNotExist) {
		return "ErrNotExist"
	}
	return err.Error()
}

func string2error(str string) error {
	if str == "ErrInvalid" {
		return os.ErrInvalid
	}
	if str == "ErrNotExist" {
		return os.ErrNotExist
	}
	return errors.New(str)
}
