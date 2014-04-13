package chain

import (
	"testing"
)

func TestEncodeString(t *testing.T) {

	c, _ := New(GenerateKey())

	originaltext := "Test message !@#$%^&*()_1234567890{}[]."

	ciphertext, err := c.EncryptString(originaltext)
	if err != nil {
		t.Error(err)
	}
	if ciphertext == originaltext {
		t.Error("Encoding the text didn't work!")
	}

	plaintext, err := c.DecryptString(ciphertext)
	if err != nil {
		t.Error(err)
	}
	if plaintext != originaltext {
		t.Errorf("\n%X\n%X\nDecoded text did not match!", originaltext, plaintext)
	}
}

func BenchmarkEncodeString(b *testing.B) {
	b.StopTimer()

	c, err := New(GenerateKey())
	if err != nil {
		b.Error(err)
	}

	originaltext := "Test message !@#$%^&*()_1234567890{}[]."

	for i := 0; i < b.N; i++ {
		b.StartTimer()
		c.EncryptString(originaltext)
		b.StopTimer()
	}
}

func BenchmarkEncodeStringCold(b *testing.B) {
	b.StopTimer()

	originaltext := "Test message !@#$%^&*()_1234567890{}[]."

	for i := 0; i < b.N; i++ {
		c, err := New(GenerateKey())
		if err != nil {
			b.Error(err)
		}
		b.StartTimer()
		c.EncryptString(originaltext)
		b.StopTimer()
	}
}

// // func BenchmarkEncodeString(b *testing.B) {

// //  key := make([]byte, 32)
// //  io.ReadFull(rand.Reader, key)

// //  c, _ := NewCryptographer(key)
// //  plaintext := "Test message !@#$%^&*()_1234567890{}[]."

// //  for i := 0; i < b.N; i++ {
// //    c.EncodeString(plaintext)
// //  }
// // }

// // func BenchmarkDecodeString(b *testing.B) {

// //  key := make([]byte, 32)
// //  io.ReadFull(rand.Reader, key)

// //  c, _ := NewCryptographer(key)
// //  ciphertext := "4Bhd60+qVQWTdKGj4fdPe0dFzll9m1i0JwHu5swgBJRlSo5bFEfikB+OBZmpMY472OyHpuWeGoZj3iC9G2etWw=="

// //  for i := 0; i < b.N; i++ {
// //    c.DecodeString(ciphertext)
// //  }
// // }
