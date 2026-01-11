package utils

import (
	"hash/fnv"

	"github.com/pkg/errors"
)

// Hash - считает число-хэш от client uuid, чтобы сделать по нему лок
func Hash(value []byte) (uint32, error) {
	h := fnv.New32a()
	_, err := h.Write(value)
	if err != nil {
		return 0, errors.Wrap(err, "can't create Hash from client uuid")
	}

	return h.Sum32(), nil
}
