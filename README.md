[![Version](https://img.shields.io/badge/goversion-1.20.x-blue.svg)](https://go.dev)
<a href="https://go.dev"><img src="https://img.shields.io/badge/powered_by-Go-3362c2.svg?style=flat-square" alt="Built with GoLang"></a>
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/tsawler/itsdangerous/master/LICENSE)
<a href="https://pkg.go.dev/github.com/tsawler/itsdangerous"><img src="https://img.shields.io/badge/godoc-reference-%23007d9c.svg"></a>
[![Go Report Card](https://goreportcard.com/badge/github.com/tsawler/itsdangerous)](https://goreportcard.com/report/github.com/tsawler/itsdangerous)
![Tests](https://github.com/tsawler/itsdangerous/actions/workflows/tests.yml/badge.svg)

# It's Dangerous

itsdangerous is a [Go](https://golang.org/) package which provides
* Methods to create and verify [MAC](https://en.wikipedia.org/wiki/Message_authentication_code) signatures of data
* Ability to add timestamps to signed tokens and use custom epoch if needed.
* [BLAKE3](https://github.com/BLAKE3-team/BLAKE3) signatures and Base58 time encoding provides outstanding performance and security.
* A very simple to use API with good documentation and 100% test coverage.
* Various helper methods for parsing tokens

BLAKE3 support is provided by [blake3](https://github.com/lukechampine/blake3).

## Installation
Go get the package in the usual way:

~~~
go get -u github.com/tsawler/itsdangerous
~~~

## Usage
A simple example:

~~~go
package main

import (
	"fmt"
	"github.com/tsawler/itsdangerous"
	"log"
	"time"
)

func main() {
	// Create a secret; typically, make this 32 characters long.
	var secret = []byte("some-very-good-secret")

	// The data to be signed.
	var data = []byte("https://example.com?id=10")

	// Create a new instance of itsdangerous.
	s := itsdangerous.New(secret)

	// Sign the data; we get back a byte array with the signature appended.
	token := s.Sign(data)

	// Unsign the token to verify it. If successful the data portion of the
	// token is returned.  If unsuccessful then d will be nil, and an error
	// is returned.
	d, err := s.Unsign(token)
	if err != nil {
		// The signature is not valid. The token was tampered with, forged, or maybe it's
		// not even a token at all.It's not safe to use it.
		log.Println("Invalid signature error:", err)
	} else {
		// The signature is valid, so it is safe to use the data.
		fmt.Println("Unsigned:", string(d))
		fmt.Println("Signed:", string(token))
		fmt.Println("Valid signature!")
	}

	// Create a new signer, this time with a timestamp.
	s2 := itsdangerous.New(secret, itsdangerous.Timestamp)
	token = s2.Sign(data)

	// print out new signed data
	fmt.Println("Signed with timestamp:", string(token))

	// Parse the token.
	ts := s2.Parse(token)

	// The token should be not expired at this point.
	if time.Since(ts.Timestamp) < time.Second {
		fmt.Println("Token with timestamp has not expired!")
	}

	// Wait two seconds to expire the token.
	time.Sleep(2 * time.Second)

	// The token should be expired at this point.
	if time.Since(ts.Timestamp) > time.Second {
		fmt.Println("Token with timestamp is expired!")
	}
}

~~~

The output of this program is something like:

~~~
tcs@Grendel example-itsdangerous % go run .
Signed: https://example.com?id=10.eoaZ-lDxxuZ-BK5PFmbZlVps0Htqi6NILs0HR47Dvs0
Unsigned: https://example.com?id=10
Valid signature!
Token with timestamp has not expired!
Token with timestamp is expired!
~~~

## Tests
To run all tests, execute this command: 

~~~
go test -v .
~~~

## Benchmarks
To run benchmarks, execute this command:

~~~
go test -v -bench=.
~~~

## Credit
This code is an updated version of [go-alone](https://github.com/bwmarrin/go-alone). Major changes include updating to use go modules, and 
moving from BLAKE2b to BLAKE3.