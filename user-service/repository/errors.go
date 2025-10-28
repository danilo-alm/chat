package repository

import "errors"

var ErrDuplicateKey = errors.New("repository: duplicate key constraint violation")
var ErrEntityNotFound = errors.New("repository: entity was not found")
