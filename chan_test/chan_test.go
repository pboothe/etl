package chan_test

import (
	"testing"
)

func BenchmarkStructChan(b *testing.B) {
	ch := make(chan struct{}, 100)
	go func() {
		for {
			<-ch
		}
	}()
	go func() {
		for {
			<-ch
		}
	}()
	go func() {
		for {
			<-ch
		}
	}()
	for i := 0; i < b.N; i++ {
		ch <- struct{}{}
	}
}

func BenchmarkBoolChan(b *testing.B) {
	ch := make(chan bool, 100)
	go func() {
		for {
			<-ch
		}
	}()
	go func() {
		for {
			<-ch
		}
	}()
	go func() {
		for {
			<-ch
		}
	}()
	for i := 0; i < b.N; i++ {
		ch <- true
	}
}

func BenchmarkIntChan(b *testing.B) {
	ch := make(chan int, 100)
	go func() {
		for {
			<-ch
		}
	}()
	go func() {
		for {
			<-ch
		}
	}()
	go func() {
		for {
			<-ch
		}
	}()
	for i := 0; i < b.N; i++ {
		ch <- 1
	}
}
