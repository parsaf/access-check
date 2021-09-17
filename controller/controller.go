package controller

import "github.com/parsaf/access-check/storage"

// Controller provides methods for the API's main functionality.
type Controller struct {
	store storage.AccessIntervalStorage
}

func New(
	store storage.AccessIntervalStorage,
) (*Controller){
	return &Controller{
		store: store,
	}
}