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

type Filer struct {
	Artifacts map[string]string
}

func IndexFiles(wd string) {
	if wd != "" {
		var mu sync.Mutex
		var wg sync.WaitGroup

		filer := &Filer{
			Artifacts: map[string]string{},
		}

		numWorkers := runtime.GOMAXPROCS(0)

		jobs := make(chan string)

		worker := func(jobs <-chan string, wg *sync.WaitGroup, mu *sync.Mutex, filer *Filer) {
			defer wg.Done()

			for path := range jobs {
				hash := hex.EncodeToString(FileHash(path))

				mu.Lock()
				filer.Artifacts[path] = hash
				mu.Unlock()
			}
		}

		// Start the workers
		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go worker(jobs, &wg, &mu, filer)
		}

		// Start indexing
		err := filepath.Walk(wd, func(path string, info fs.FileInfo, err error) error {

			if !info.IsDir() && info.Mode()&fs.ModeSymlink == 0 {
				jobs <- path
			}

			return nil
		})
		if err != nil {
			log.Fatal(err)
		}

		close(jobs)

		wg.Wait()

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
