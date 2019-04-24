package main

import (
	"testing"

	"github.com/kavehmz/prime"
)

func BenchmarkPrime(b *testing.B) {
	for n := 0; n < b.N; n++ {
		prime.Primes(250000)
	}
}
