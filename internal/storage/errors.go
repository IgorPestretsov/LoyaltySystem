package storage

import "errors"

var ErrLoginExist = errors.New("login already exist")
var ErrAlreadyLoadedByThisUser = errors.New("order has been loaded already by this user")
var ErrAlreadyLoadedByDifferentUser = errors.New("order has been loaded already by different user")
var ErrDBInteraction = errors.New("database interaction error")
