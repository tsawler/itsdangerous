package itsdangerous

import (
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"hash"
	"lukechampine.com/blake3"
	"sync"
	"time"
)

// Sword is the main struct used to sign and unsign data using this package.
type Sword struct {
	sync.Mutex
	hash      hash.Hash
	dirty     bool
	timestamp bool
	epoch     int64
}

// ErrInvalidSignature is returned by Unsign when the provided token's
// signature is not valid.
var ErrInvalidSignature = errors.New("invalid signature")

// ErrShortToken is returned by Unsign when the provided token's length
// is too short to be a valid token.
var ErrShortToken = errors.New("token is too small to be valid")

// New takes a secret key and returns a new Sword.  If no Options are provided
// then minimal defaults will be used. NOTE: The key must be 32 bytes.
// If a key < 32 bytes is provided, it will be padded to 32 bytes.
// If a larger key is provided it will be truncated to 64 bytes.
// func New(key []byte, o *Options) *Sword {
func New(key []byte, options ...func(*Sword)) *Sword {
	// Make the key at least 32 bytes long.
	key = padSecret(key)

	// Create a map for decoding Base58.  This speeds up the process a great deal.
	for i := 0; i < len(encodeBase58Map); i++ {
		decodeBase58Map[encodeBase58Map[i]] = byte(i)
	}

	// Create an empty variable of type Sword.
	s := &Sword{}

	// Add options, if any.
	for _, opt := range options {
		opt(s)
	}

	// Add the hash to the Sword.
	s.hash = blake3.New(32, key)

	return s
}

// padSecret will make the key at least 32 bytes long.
func padSecret(key []byte) []byte {
	// Make sure key is at least 32 characters long.
	if len(key) < 32 {
		str := string(key)
		for i := 0; i < (32 - len(key)); i++ {
			str = str + " "
		}
		key = []byte(str)
	}
	return key
}

// Epoch is a functional option that can be passed to New() to set the Epoch
// to be used.
func Epoch(e int64) func(*Sword) {
	return func(s *Sword) {
		s.epoch = e
	}
}

// Timestamp is a functional option that can be passed to New() to add a
// timestamp to signatures.
func Timestamp(s *Sword) {
	s.timestamp = true
}

// Sign signs data and returns []byte in the format `data.signature`. Optionally
// add a timestamp and return in the format `data.timestamp.signature`
func (s *Sword) Sign(data []byte) []byte {
	// Build the payload
	el := base64.RawURLEncoding.EncodedLen(s.hash.Size())
	var t []byte

	if s.timestamp {
		ts := time.Now().Unix() - s.epoch
		etl := encodeBase58Len(ts)
		t = make([]byte, 0, len(data)+etl+el+2) // +2 for "." chars
		t = append(t, data...)
		t = append(t, '.')
		t = t[0 : len(t)+etl] // expand for timestamp
		encodeBase58(ts, t)
	} else {
		t = make([]byte, 0, len(data)+el+1)
		t = append(t, data...)
	}

	// Append and encode signature to token
	t = append(t, '.')
	tl := len(t)
	t = t[0 : tl+el]

	// Add the signature to the token
	s.sign(t[tl:], t[0:tl-1])

	// Return the token to the caller
	return t
}

// Unsign validates a signature and if successful returns the data portion of
// the []byte. If unsuccessful it will return an error and nil for the data.
func (s *Sword) Unsign(token []byte) ([]byte, error) {
	tl := len(token)
	el := base64.RawURLEncoding.EncodedLen(s.hash.Size())

	// A token must be at least el+2 bytes long to be valid.
	if tl < el+2 {
		return nil, ErrShortToken
	}

	// Get the signature of the payload
	dst := make([]byte, el)
	s.sign(dst, token[0:tl-(el+1)])

	if subtle.ConstantTimeCompare(token[tl-el:], dst) != 1 {
		return nil, ErrInvalidSignature
	}

	return token[0 : tl-(el+1)], nil
}

// encodeBase58Map is the map of characters used during base58 encoding.
const encodeBase58Map = "123456789abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"

// decodeBase58Map is used to create a decode map, so we can decode base58 fairly fast.
var decodeBase58Map [256]byte

// sign creates the encoded signature of payload and writes to dst.
func (s *Sword) sign(dst, payload []byte) {
	s.Lock()
	if s.dirty {
		s.hash.Reset()
	}
	s.dirty = true
	s.hash.Write(payload)
	h := s.hash.Sum(nil)
	s.Unlock()

	base64.RawURLEncoding.Encode(dst, h)
}

// encodeBase58Len returns the len of base58 encoded i.
func encodeBase58Len(i int64) int {
	var l = 1
	for i >= 58 {
		l++
		i /= 58
	}
	return l
}

// encodeBase58 encodes time int64 into b []byte.
func encodeBase58(i int64, b []byte) {
	p := len(b) - 1
	for i >= 58 {
		b[p] = encodeBase58Map[i%58]
		p--
		i /= 58
	}
	b[p] = encodeBase58Map[i]
}

// decodeBase58 parses a base58 []byte into a int64.
func decodeBase58(b []byte) int64 {
	var id int64
	for p := range b {
		id = id*58 + int64(decodeBase58Map[b[p]])
	}
	return id
}
