package crypto

import (
	"fmt"
	"testing"
)

func TestEncodeString(t *testing.T) {

	c, err := NewCrypter(NewKey())
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

func TestRepeatedEncoding(t *testing.T) {

	key := NewKey()

	c1, err := NewCrypter(key)
	if err != nil {
		t.Error(err)
	}

	c2, err := NewCrypter(key)
	if err != nil {
		t.Error(err)
	}

	originaltext := "Test message !@#$%^&*()_1234567890{}[]."
	fmt.Println(originaltext)

	ciphertext1, err := c1.EncryptString(originaltext)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(ciphertext1)

	ciphertext2, err := c2.EncryptString(originaltext)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(ciphertext2)

	if ciphertext1 == ciphertext2 {
		t.Error("seperate encodings of the same string matched!")
	}
}

func TestSeperateEncoding(t *testing.T) {

	c, _ := NewCrypter(NewKey())

	originaltext := "Test message !@#$%^&*()_1234567890{}[]."
	fmt.Println(originaltext)

	ciphertext1, err := c.EncryptString(originaltext)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(ciphertext1)

	ciphertext2, err := c.EncryptString(originaltext)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(ciphertext2)

	if ciphertext1 == ciphertext2 {
		t.Error("repeated encodings of the same string matched!")
	}
}

func BenchmarkEncodeString(b *testing.B) {
	b.StopTimer()

	c, err := NewCrypter(NewKey())
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
		c, err := NewCrypter(NewKey())
		if err != nil {
			b.Error(err)
		}
		b.StartTimer()
		c.EncryptString(originaltext)
		b.StopTimer()
	}
}
