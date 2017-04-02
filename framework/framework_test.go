
package framework_test

import (
	"github.com/m-lab/etl/framework"
	"testing"
)


func benchmarkMD5(numDigesters int, b *testing.B) {
	var data [11000]byte
	framework.ManyBig(4, numDigesters, b.N, data[:])
}


func Benchmark1(b *testing.B) { benchmarkMD5(1, b) }
func Benchmark4(b *testing.B) { benchmarkMD5(4, b) }
func Benchmark12(b *testing.B) { benchmarkMD5(12, b) }

/*
* With md5, Looks like 100 * 74k (7.4 MB), takes around 5msec when all is good.
* This is about 1.4 GB/sec ???
* Without md5, about 2 msec, so about 3.5 GB/sec.
*
*
*
*/
