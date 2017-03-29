
package framework

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"sync"
)

type content struct {
  buf [12000]byte
}

// replicate starts a goroutine to send 'count' copies of the
// data block on the data channel.  On completion, it sends OK on the error channel.
// If done is closed, it abandons its work.
func replicate(count int, done <-chan struct{}, block []byte) (<-chan content) {
	out := make(chan content, 20)  // data output channel, buffer 10.
	var c content
	copy(c.buf[:len(block)], block)
	go func() {
		// Close the data channel on completion.
		defer close(out)
		for i := 0; i < count; i++ {
			//	var b []byte
			//	b = make([]byte, len(block))
			//	copy(b[:], block)
			out <- c
		}
	}()
	return out
}

// A result is the product of reading and summing a file using MD5.
type result struct {
	path string
//	sum  [md5.Size]byte
	size int
}

// digester reads data blocks and sends digests
// on c until either data or done is closed.
func digester(done <-chan struct{}, data <-chan content, c chan<- result) {
	for block := range data {
		select {
		//case c <- result{"foobar", md5.Sum(block), len(block)}:
		case c <- result{"foobar", len(block.buf)}:
		case <-done:
			return
		}
	}
}

func ManyBig(numDigesters int, numRecords int, fname string) (map[string]int, error) {
	// closes the done channel when it returns
	done := make(chan struct{})
	defer close(done)

	block, err := ioutil.ReadFile(fname)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	data := replicate(numRecords, done, block)
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
	m, err := ManyBig(20, 100, os.Args[1])
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

