package goalone

import (
	"crypto/subtle"
	"fmt"
	"testing"
	"time"
)

var testSecret = "AVerySecretString"

func TestTimeCoding(t *testing.T) {
	New(nil) // needed to init the decodemap

	var now int64
	for i := 0; i < 41; i++ {
		b := make([]byte, encodeBase58Len(now))
		encodeBase58(now, b)
		id := decodeBase58(b)
		if id != now {
			fmt.Printf("Time.Now()   : %d\n", now)
			fmt.Printf("encodeBase58 : %s\n", b)
			fmt.Printf("decodeBase58 : %d\n", id)
			t.Fatal("id != now!")
		}
		now += 1 + now*2
	}
}

func TestNewNil(t *testing.T) {
	s := New(nil)
	if s == nil {
		t.Fatal("New returned a nil")
	}

	if s.dirty {
		t.Fatal("New returned a dirty hash")
	}
}

func TestNewSecret(t *testing.T) {
	secret := []byte(testSecret)
	s := New(secret)
	if s == nil {
		t.Fatal("New returned a nil")
	}

	if s.hash == nil {
		t.Fatal("New returned a Sword with a nil hash")
	}

	if s.dirty {
		t.Fatal("New returned a dirty hash")
	}
}

func TestNewSecretTooBig(t *testing.T) {
	secret := []byte(`B1nzyRateLid;flkjasdl;fjasd;lkfjkl;ljasd;fkljsda;fkljasd;klfj;asdjts`)
	s := New(secret)
	if s == nil {
		t.Fatal("New returned a nil")
	}

	if s.hash == nil {
		t.Fatal("New returned a Sword with a nil hash")
	}

	if s.dirty {
		t.Fatal("New returned a dirty hash")
	}
}

func TestUnsignTooLittle(t *testing.T) {
	secret := []byte(testSecret)
	token := []byte("9yhD.QheV.k0W")

	s := New(secret)
	got, err := s.Unsign(token)

	if got != nil {
		t.Error("Unsign returned data, but should have returned nil")
	}

	if err != ErrShortToken {
		t.Fatal("Unsign did not return the correct error")
	}
}

func TestSign(t *testing.T) {
	secret := []byte(testSecret)
	secret = padSecret(secret)
	//secret := []byte("hello")
	data := []byte(`1203981209381290.LutinRocks`)
	want := []byte(`1203981209381290.LutinRocks.p5l-D9w5MtSGs9LyTR5fWGHkP-5dW2YvBrYl1JGrA7Y`)

	s := New(secret)
	token := s.Sign(data)

	if subtle.ConstantTimeCompare(token, want) != 1 {
		t.Logf("data: \n%s\n", data)
		t.Logf("want: \n%s\n", want)
		t.Logf("got : \n%s\n", token)
		t.Fatal("token and want do not match")
	}

	// sign after a dirty hash
	if !s.dirty {
		t.Fatal("Hash is not dirty, but it should be")
	}

	token = s.Sign(data)

	if subtle.ConstantTimeCompare(token, want) != 1 {
		t.Logf("data: \n%s\n", data)
		t.Logf("want: \n%s\n", want)
		t.Logf("got : \n%s\n", token)
		t.Fatal("token and want do not match")
	}
}

func TestSignTimestamp(t *testing.T) {
	secret := []byte(testSecret)
	data := []byte(`1203981209381290.LutinRocks`)

	s := New(secret, Timestamp)
	token := s.Sign(data)

	// Make sure we got the same payload
	if subtle.ConstantTimeCompare(token[0:27], data) != 1 {
		t.Logf("token: \n%s\n", token[0:27])
		t.Logf("want: \n%s\n", data)
		t.Fatal("Payload was changed")
	}

	// Make sure we got a timestamp.... (not a great test..)
	if token[27] != '.' || token[34] != '.' {
		t.Fatal("Doesn't appear to be a timestamp.")
	}
}

func TestSignEpoch(t *testing.T) {
	secret := []byte(testSecret)
	data := []byte(`1203981209381290.LutinRocks`)

	s := New(secret, Epoch(1293840000), Timestamp)
	token := s.Sign(data)

	// Make sure we got the same payload
	if subtle.ConstantTimeCompare(token[0:27], data) != 1 {
		t.Logf("token: \n%s\n", token[0:27])
		t.Logf("want: \n%s\n", data)
		t.Fatal("Payload was changed")
	}

	// Make sure we got a timestamp.... (not a great test..)
	if token[27] != '.' || token[33] != '.' {
		t.Fatal("Doesn't appear to be a timestamp.")
	}
}

func TestUnsign(t *testing.T) {
	secret := []byte(testSecret)
	secret = padSecret(secret)
	token := []byte(`1203981209381290.LutinRocks.p5l-D9w5MtSGs9LyTR5fWGHkP-5dW2YvBrYl1JGrA7Y`)
	want := []byte(`1203981209381290.LutinRocks`)

	s := New(secret)
	got, err := s.Unsign(token)
	if err != nil {
		t.Fatal("Unsign returned an err,", err)
	}
	if subtle.ConstantTimeCompare(got, want) != 1 {
		t.Logf("token: \n%s\n", token)
		t.Logf("want: \n%s\n", want)
		t.Logf("got : \n%s\n", got)
		t.Fatal("data of token does not match original data")
	}

	if !s.dirty {
		t.Fatal("Hash is not dirty, but it should be")
	}

	token = []byte(`1203981209381290.LutinRocks.ZGRsRXvTb08ld7xmJImL1ykGr8D1JmrSPGc134nBNR0`)
	got, err = s.Unsign(token)
	if err != ErrInvalidSignature {
		t.Fatal("Unsign returned incorrect error")
	}
	if got != nil {
		t.Logf("token: \n%s\n", token)
		t.Logf("want: \n%s\n", want)
		t.Logf("got : \n%s\n", got)
		t.Fatal("got should be nil")
	}
}

func BenchmarkNew(b *testing.B) {
	secret := []byte(testSecret)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		New(secret)
	}
}

func BenchmarkSignLittle(b *testing.B) {
	secret := []byte(testSecret)
	data := []byte(`1203981209381290.LutinRocks`)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		New(secret).Sign(data)
	}
}

func BenchmarkReuseSignLittle(b *testing.B) {
	s := New([]byte(testSecret))
	data := []byte(`1203981209381290.LutinRocks`)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		s.Sign(data)
	}
}

func BenchmarkReuseSignTimestampLittle(b *testing.B) {
	s := New([]byte(testSecret), Timestamp)
	data := []byte(`1203981209381290.LutinRocks`)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		s.Sign(data)
	}
}

func BenchmarkSignBig(b *testing.B) {
	secret := []byte(testSecret)
	data := []byte(`1203981209381290.7h90g7h089234g75908347gh09384h7v0897fg08947f5097423058974h908fg702f9j75028fg5704239hg7053498dj7249038jd57j097g5v029dh79hc47f507v9082h7f509234j7dc02d750j24935h7f924`)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		New(secret).Sign(data)
	}
}

func BenchmarkSignBigReuse(b *testing.B) {
	s := New([]byte(testSecret))
	data := []byte(`1203981209381290.7h90g7h089234g75908347gh09384h7v0897fg08947f5097423058974h908fg702f9j75028fg5704239hg7053498dj7249038jd57j097g5v029dh79hc47f507v9082h7f509234j7dc02d750j24935h7f924`)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		s.Sign(data)
	}
}

func BenchmarkUnsignLittle(b *testing.B) {
	secret := []byte(testSecret)
	data := []byte(`1203981209381290.LutinRocks`)

	s := New([]byte(testSecret))
	t := s.Sign(data)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		New(secret).Unsign(t)
	}
}

func BenchmarkUnsignLittleReuse(b *testing.B) {
	s := New([]byte(testSecret))
	data := []byte(`1203981209381290.LutinRocks`)
	t := s.Sign(data)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		s.Unsign(t)
	}
}

func BenchmarkUnsignBig(b *testing.B) {
	secret := []byte(testSecret)
	data := []byte(`1203981209381290.7h90g7h089234g75908347gh09384h7v0897fg08947f5097423058974h908fg702f9j75028fg5704239hg7053498dj7249038jd57j097g5v029dh79hc47f507v9082h7f509234j7dc02d750j24935h7f924`)
	s := New([]byte(`B1nzyRateLimits`))
	t := s.Sign(data)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		New(secret).Unsign(t)
	}
}

func BenchmarkUnsignReuseBig(b *testing.B) {
	s := New([]byte(testSecret))
	data := []byte(`1203981209381290.7h90g7h089234g75908347gh09384h7v0897fg08947f5097423058974h908fg702f9j75028fg5704239hg7053498dj7249038jd57j097g5v029dh79hc47f507v9082h7f509234j7dc02d750j24935h7f924`)
	t := s.Sign(data)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		s.Unsign(t)
	}
}

func BenchmarkEncodeBase58(b *testing.B) {
	now := time.Now().Unix()
	by := make([]byte, 6)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		encodeBase58(now, by)
	}
}

func BenchmarkDecodeBase58(b *testing.B) {
	now := time.Now().Unix()
	by := make([]byte, 6)
	encodeBase58(now, by)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		decodeBase58([]byte(by))
	}
}
