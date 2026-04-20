package utils

import (
	"math/rand"
	"testing"
)

func init() {
	rand.New(rand.NewSource(12345))
}

func BenchmarkUniqueSlice(b *testing.B) {
	elems := []string{"apple", "banana", "cherry", "date", "elderberry", "fig", "grape", "honeydew", "kiwi", "lemon"}
	data := make([]string, len(elems)*10)
	for i := range elems {
		newData := make([]string, 10)
		for j := 0; j < 10; j++ {
			newData[j] = elems[i]
		}
	}
	rand.Shuffle(len(data), func(i, j int) {
		data[i], data[j] = data[j], data[i]
	})
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		UniqueSlice(&data)
	}
}
