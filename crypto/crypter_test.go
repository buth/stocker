package crypto

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"testing"
)

func TestEncodeString(t *testing.T) {

	c, err := NewRandomCrypter()
	if err != nil {
		t.Error(err)
	}

	originaltext := "Test message !@#$%^&*()_1234567890{}[]."
	fmt.Println(originaltext)

	ciphertext, err := c.EncryptString(originaltext)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(ciphertext)

	if ciphertext == originaltext {
		t.Error("encoding the text didn't work!")
	}

	plaintext, err := c.DecryptString(ciphertext)
	if err != nil {
		t.Error(err)
	}
	if plaintext != originaltext {
		t.Errorf("\n%X\n%X\ndecoded text did not match!", originaltext, plaintext)
	}
}

func TestSeperateEncoding(t *testing.T) {

	key := make([]byte, HmacKeyLength+SymetricKeyLength)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatal(err)
	}

	c1, err := NewCrypter(bytes.NewBuffer(key))
	if err != nil {
		t.Fatal(err)
	}

	c2, err := NewCrypter(bytes.NewBuffer(key))
	if err != nil {
		t.Fatal(err)
	}

	originaltext := "Test message !@#$%^&*()_1234567890{}[]."
	fmt.Println(originaltext)

	ciphertext1, err := c1.EncryptString(originaltext)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(ciphertext1)

	ciphertext2, err := c2.EncryptString(originaltext)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(ciphertext2)

	if ciphertext1 == ciphertext2 {
		t.Error("seperate encodings of the same string matched!")
	}
}

func TestRepeatedEncoding(t *testing.T) {

	c, err := NewRandomCrypter()
	if err != nil {
		t.Fatal(err)
	}

	originaltext := "Test message !@#$%^&*()_1234567890{}[]."
	fmt.Println(originaltext)

	ciphertext1, err := c.EncryptString(originaltext)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(ciphertext1)

	ciphertext2, err := c.EncryptString(originaltext)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(ciphertext2)

	if ciphertext1 == ciphertext2 {
		t.Error("repeated encodings of the same string matched!")
	}
}

func BenchmarkEncodeString(b *testing.B) {
	b.StopTimer()

	c, err := NewRandomCrypter()
	if err != nil {
		b.Fatal(err)
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
		c, err := NewRandomCrypter()
		if err != nil {
			b.Fatal(err)
		}
		b.StartTimer()
		c.EncryptString(originaltext)
		b.StopTimer()
	}
}
