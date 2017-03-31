
package framework

import (
	"crypto/md5"
	"sync"
)

// replicate starts a goroutine to send 'count' copies of the
// data block on the data channel.
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
func digester(done <-chan struct{}, data <-chan []byte) { //, c chan<- bool) {
	for block := range data {
		md5.Sum(block)
	}
}

func ManyBig(numSources int, numDigesters int, numRecords int, block []byte) {
	// closes the done channel when it returns
	done := make(chan struct{})
	defer close(done)

	var source_wg sync.WaitGroup
	source_wg.Add(numSources)
	data := make(chan []byte, 50)  // data output channel, buffer 50.
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
	//c := make(chan bool)
	var wg sync.WaitGroup
	wg.Add(numDigesters)
	for i := 0; i < numDigesters; i++ {
		go func() {
			digester(done, data) //, c)
			wg.Done()
		}()
	}
//	go func() {
		wg.Wait()
//		close(c)  // signal to all digesters to quit.
//	}()

	//	m := make(map[string]int)
	//for r := range c {
	//	m[r.path] = r.size
	//}
	//return m, nil
}

