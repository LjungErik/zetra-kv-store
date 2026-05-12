package models

import "errors"

type KeyStoreInsertRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type KeyStoreDeleteRequest struct {
	Key string `json:"key"`
}

func (r KeyStoreInsertRequest) Validate() error {
	if r.Key == "" {
		return errors.New("key not set")
	}

	return nil
}

func (r KeyStoreDeleteRequest) Validate() error {
	if r.Key == "" {
		return errors.New("key not set")
	}

	return nil
}
