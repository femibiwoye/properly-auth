package test

import (
	"properlyauth/routes"
)

var (
	router, chatServer = routes.Router()
	tokens             = []string{}
	propertyID         = []string{}
	documents          = []string{}
	images             = []string{}
	inspectionID       = []string{}
	complaitsID        = []string{}
)

type mockReadCloser struct {
	data []byte
}

func (mrc mockReadCloser) Read(data []byte) (int, error) {
	return copy(data, mrc.data), nil
}

func (mrc mockReadCloser) Close() error {
	return nil
}
