
package framework_test

import (
	"etl"
	"testing"
)

func benchmarkMD5(numDigesters int, b *testing.B) {
        for n := 0; n < b.N; n++ {
		framework.ManyBig(4, numDigesters, 10000, "small")
        }
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
