package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

func main() {
	startTime := time.Now()
	fmt.Printf("%v\n", startTime.Format(time.RFC3339))

	if len(os.Args) > 1 {
		wd := os.Args[1]
		IndexFiles(wd)
	}

	fmt.Printf("%v\n", time.Since(startTime).Seconds())
}

func IndexFiles(wd string) {
	if wd != "" {
		artifacts := make(map[string]string, 0)

		err := filepath.Walk(wd, func(path string, info fs.FileInfo, err error) error {

			if !info.IsDir() && info.Mode()&fs.ModeSymlink == 0 {
				artifacts[path] = path
			}

			return nil
		})
		if err != nil {
			log.Fatal(err)
		}

		numWorkers := runtime.GOMAXPROCS(0)
		jobs := make(chan string, len(artifacts))
		results := make(chan int, len(artifacts))

		var mu sync.Mutex

		worker := func(jobs <-chan string, results chan<- int) {

			for path := range jobs {
				hash := hex.EncodeToString(FileHash(path))

				mu.Lock()
				artifacts[path] = hash
				results <- 1
				mu.Unlock()
			}
		}

		// Start the workers
		for i := 0; i < numWorkers; i++ {
			go worker(jobs, results)
		}

		for key := range artifacts {
			mu.Lock()
			jobs <- key
			mu.Unlock()
		}

		// Close the jobs channel to signal the workers to exit
		close(jobs)

		for a := 1; a <= len(artifacts); a++ {
			<-results
		}

		fmt.Printf("%v\n", len(artifacts))

		// No concurrency:
		//
		// Indexing
		// 1902748 files + dirs, ~141 seconds (135s, 141s)
		// 1269867 files + dirs, ~77 seconds
		//
		// Hashing
		// 1271473 files, ~486 seconds / 8 min
		// 278130 files, ~261 seconds / 4.3 min

		// Concurrency:
		//
		// 849971 files ~164 seconds / 2.7 min

		// fmt.Printf("%v\n", time.Since(startTime).Seconds())
		// fmt.Printf("Total files: %v", len(artifacts))
	}
}

func FileHash(file string) []byte {
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}

	return h.Sum(nil)
}
