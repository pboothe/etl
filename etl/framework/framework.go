
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

// digester reads data blocks and sends digests on c
func digester(data <-chan []byte, c chan<- [md5.Size]byte) {
	for block := range data {
		c <- md5.Sum(block)
	}
}

func ManyBig(numSources int, numDigesters int, numRecords int, block []byte) {
	var source_wg sync.WaitGroup
	source_wg.Add(numSources)
	data := make(chan []byte, 50)  // data output channel, buffer 50.
	for s := 0; s < numSources; s++ {
		go func() {
			replicate(data, numRecords/numSources, block)
			source_wg.Done()
		}()
	}

	// Start a fixed number of goroutines to read and digest files.
	c := make(chan [md5.Size]byte)
	var digest_wg sync.WaitGroup
	digest_wg.Add(numDigesters)
	for i := 0; i < numDigesters; i++ {
		go func() {
			digester(data, c)
			digest_wg.Done()	
		}()
	}

	go func() {
		source_wg.Wait()
		close(data)
	}()

	go func() {
		digest_wg.Wait()
		close(c)
	}()

	// Consume the md5 values.
	count := 0
	for _ = range c {
		count += 1
	}
}

