package crypter

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"testing"
)

func TestEncodeString(t *testing.T) {

	keybytes := make([]byte, 288)
	io.ReadFull(rand.Reader, keybytes)

	key := base64.StdEncoding.EncodeToString(keybytes)

	c := &AES{}

	c.Load(key)

	originaltext := "Test message !@#$%^&*()_1234567890{}[]."

	ciphertext, _ := c.EncodeString(originaltext)
	if ciphertext == originaltext {
		t.Errorf("Encoding the text didn't work!")
	}

	plaintext, _ := c.DecodeString(ciphertext)
	if plaintext != originaltext {
		t.Errorf("\n%X\n%X\nDecoded text did not match!", originaltext, plaintext)
	}
}

func BenchmarkEncodeString(b *testing.B) {
	b.StopTimer()

	originaltext := "Test message !@#$%^&*()_1234567890{}[]."

	keybytes := make([]byte, 288)
	io.ReadFull(rand.Reader, keybytes)
	key := base64.StdEncoding.EncodeToString(keybytes)
	c := &AES{}
	c.Load(key)

	for i := 0; i < b.N; i++ {

		b.StartTimer()
		c.EncodeString(originaltext)
		b.StopTimer()
	}
}

func BenchmarkEncodeStringCold(b *testing.B) {
	b.StopTimer()

	originaltext := "Test message !@#$%^&*()_1234567890{}[]."

	for i := 0; i < b.N; i++ {
		keybytes := make([]byte, 288)
		io.ReadFull(rand.Reader, keybytes)
		key := base64.StdEncoding.EncodeToString(keybytes)
		c := &AES{}
		c.Load(key)
		b.StartTimer()
		c.EncodeString(originaltext)
		b.StopTimer()
	}
}

// func BenchmarkEncodeString(b *testing.B) {

// 	key := make([]byte, 32)
// 	io.ReadFull(rand.Reader, key)

// 	c, _ := NewCryptographer(key)
// 	plaintext := "Test message !@#$%^&*()_1234567890{}[]."

// 	for i := 0; i < b.N; i++ {
// 		c.EncodeString(plaintext)
// 	}
// }

// func BenchmarkDecodeString(b *testing.B) {

// 	key := make([]byte, 32)
// 	io.ReadFull(rand.Reader, key)

// 	c, _ := NewCryptographer(key)
// 	ciphertext := "4Bhd60+qVQWTdKGj4fdPe0dFzll9m1i0JwHu5swgBJRlSo5bFEfikB+OBZmpMY472OyHpuWeGoZj3iC9G2etWw=="

// 	for i := 0; i < b.N; i++ {
// 		c.DecodeString(ciphertext)
// 	}
// }
