// Package framework provides the interconnections among the components that
// make up the m-lab pipeline.
// Part or all of this package will likely become main.
//
// The components are:
//  Driver - rpc / http interface to receive file paths. Provides Reader
//   objects to downstream code.
//  Reader - unpacks tar files and provides raw test data.
//  Parsers - process individual tests
//  BigQueryWriter - inserts new test records into BigQuery
//  Status - handles status page generation, health reporting
//
// We want parallelism in the following components:
//  Reader - because file i/o might otherwise be a bottleneck.
//  Parser - because it is compute intensive.  We expect to want at least num_cpus
//   parser goroutines, and expect these to consume the bulk of the CPU.
//  BQWriter - because i/o might be a bottleneck.
//
// Alternative implementation:
//  It might work equally well, or better, to have N independent goroutines, each
//  comprising a Reader, a Parser, and a BQWriter.  We wouldn't care which
//  of these was blocking, as long as there are enough of them, (probably 2x numcpus)
//  so that CPUs are all kept busy.
//
//  One potential advantage to this is that the data would stay in the same L2 cache
//  through it's lifetime.
//
//  This scheme would be quite a bit simpler from a design point of view - basically
//  requiring only a single channel for obtaining file objects to read.
//
//  CONS:
//   This would result in many large file buffers, one for each goroutine, whereas we might
//   otherwise share a smaller number of file buffers across multiple goroutines.

package framework

import (
	"crypto/md5"
	"sync"
)

type Driver interface {
}

// Reader is a source of tests.
type Reader interface {
	// Returns test as byte slice, or nil if no more tests.
	NextTest() []byte
}

// replicate starts a goroutine to send 'count' copies of the
// data block on the data channel.
func replicate(out chan<- []byte, count int, block []byte) {
	for i := 0; i < count; i++ {
		out <- block // This copies only the slice, not the data
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
	data := make(chan []byte, 50) // data output channel, buffer 50.
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
