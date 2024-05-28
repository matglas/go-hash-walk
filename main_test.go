package main

import (
	"crypto/rand"
	"fmt"
	"os"
	"testing"
)

var table = []struct {
	input int
}{
	// {input: 10},
	// {input: 50},
	{input: 100},
	{input: 1000},
	{input: 10000},
}

func BenchmarkHash(b *testing.B) {

	for _, v := range table {
		workingDir := b.TempDir()

		for i := 0; i < v.input; i++ {
			file, err := os.CreateTemp(workingDir, fmt.Sprintf("file-%v.txt", i))
			if err != nil {
				b.Fatalf("Failed to create file %v for table entry %v inside %v", i, v.input, workingDir)
			}

			token := make([]byte, 8)
    		rand.Read(token)

			_, err = file.Write(token)
			if err != nil {
				b.Fatalf("Failed to write file")
			}
			file.Close()
		}

		b.Run(fmt.Sprintf("input_size_%d", v.input), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				IndexFiles(workingDir)
			}
		})
	}
}
