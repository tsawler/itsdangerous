# It's Dangerous

itsdangerous is a [Go](https://golang.org/) package which provides
* Methods to create and verify [MAC](https://en.wikipedia.org/wiki/Message_authentication_code) signatures of data
* Ability to add timestamps to signed tokens and use custom epoch if needed.
* BLAKE3 signatures and Base58 time encoding provides outstanding performance and security.
* A very simple to use API with good documentation and 100% test coverage.
* Various helper methods for parsing tokens

BLAKE3 support is provided by [blake3](https://github.com/lukechampine/blake3).

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