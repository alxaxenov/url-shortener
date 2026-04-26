package service

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

// Hasher структура генератора хешей для коротких URL
type Hasher struct {
	base62Chars string
}

// NewHasher конструктор Hasher
func NewHasher() *Hasher {
	return &Hasher{"0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"}
}

// // GetShort генерация случайного хэша для короткого URL.
func (s *Hasher) GetShort() (string, error) {
	b := make([]byte, 6)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("getShort rand error: %w", err)
	}
	num := binary.BigEndian.Uint64(append([]byte{0, 0}, b...))
	var res = make([]byte, 0, 8)
	for num > 0 {
		res = append(res, s.base62Chars[num%62])
		num /= 62
	}
	return string(res), nil
}
