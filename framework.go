
package framework

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"sort"
	"sync"
)

// replicate starts a goroutine to send 'count' copies of the
// data block on the data channel.  On completion, it sends OK on the error channel.
// If done is closed, it abandons its work.
func replicate(out chan<- []byte, count int, block []byte) {
	for i := 0; i < count; i++ {
		out <- block  // This copies only the slice, not the data
	}
}

// A result is the product of reading and summing a file using MD5.
type result struct {
	path string
	sum  [md5.Size]byte
	size int
}

// digester reads data blocks and sends digests
// on c until either data or done is closed.
func digester(done <-chan struct{}, data <-chan []byte, c chan<- result) {
	for block := range data {
		select {
		case c <- result{"foobar", md5.Sum(block), len(block)}:
		case <-done:
			return
		}
	}
}

func ManyBig(numSources int, numDigesters int, numRecords int, fname string) (map[string]int, error) {
	// closes the done channel when it returns
	done := make(chan struct{})
	defer close(done)

	block, err := ioutil.ReadFile(fname)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var source_wg sync.WaitGroup
	source_wg.Add(numSources)
	data := make(chan []byte, 50)  // data output channel, buffer 10.
	for s := 0; s < numSources; s++ {
		go func() {
			replicate(data, numRecords/numSources, block)
			source_wg.Done()
		}()
	}
	go func() {
		source_wg.Wait()
		close(data)
	}()

	// Start a fixed number of goroutines to read and digest files.
	c := make(chan result)
	var wg sync.WaitGroup
	wg.Add(numDigesters)
	for i := 0; i < numDigesters; i++ {
		go func() {
			digester(done, data, c)
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(c)
	}()

	m := make(map[string]int)
	for r := range c {
		m[r.path] = r.size
	}
	return m, nil
}

func main() {
	fmt.Println("max procs", runtime.GOMAXPROCS(0))
	fmt.Println("cpus ", runtime.NumCPU())
	m, err := ManyBig(10, 20, 1000, os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	var paths []string
	for path := range m {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		fmt.Printf("%d  %s\n", m[path], path)
	}
}

