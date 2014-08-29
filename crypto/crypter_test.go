package crypto

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"testing"
)

func TestEncode(t *testing.T) {

	c, err := NewRandomCrypter()
	if err != nil {
		t.Error(err)
	}

	originalbytes := []byte("Test message !@#$%^&*()_1234567890{}[].")
	fmt.Println(originalbytes)

	cipherbytes, err := c.Encrypt(originalbytes)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(cipherbytes)

	if bytes.Equal(cipherbytes, originalbytes) {
		t.Error("encoding the bytes didn't work!")
	}

	plainbytes, err := c.Decrypt(cipherbytes)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(plainbytes, originalbytes) {
		t.Errorf("\n%X\n%X\ndecoded bytes did not match!", originalbytes, plainbytes)
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

	originalbytes := []byte("Test message !@#$%^&*()_1234567890{}[].")
	fmt.Println(originalbytes)

	cipherbytes1, err := c1.Encrypt(originalbytes)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(cipherbytes1)

	cipherbytes2, err := c2.Encrypt(originalbytes)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(cipherbytes2)

	if bytes.Equal(cipherbytes1, cipherbytes2) {
		t.Error("seperate encodings of the same string matched!")
	}
}

func TestRepeatedEncoding(t *testing.T) {

	c, err := NewRandomCrypter()
	if err != nil {
		t.Fatal(err)
	}

	originalbytes := []byte("Test message !@#$%^&*()_1234567890{}[].")
	fmt.Println(originalbytes)

	cipherbytes1, err := c.Encrypt(originalbytes)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(cipherbytes1)

	cipherbytes2, err := c.Encrypt(originalbytes)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(cipherbytes2)

	if bytes.Equal(cipherbytes1, cipherbytes2) {
		t.Error("repeated encodings of the same string matched!")
	}
}

func BenchmarkEncode(b *testing.B) {
	b.StopTimer()

	c, err := NewRandomCrypter()
	if err != nil {
		b.Fatal(err)
	}

	originalbytes := []byte("Test message !@#$%^&*()_1234567890{}[].")

	for i := 0; i < b.N; i++ {
		b.StartTimer()
		c.Encrypt(originalbytes)
		b.StopTimer()
	}
}

func BenchmarkEncodeCold(b *testing.B) {
	b.StopTimer()

	originalbytes := []byte("Test message !@#$%^&*()_1234567890{}[].")

	for i := 0; i < b.N; i++ {
		c, err := NewRandomCrypter()
		if err != nil {
			b.Fatal(err)
		}
		b.StartTimer()
		c.Encrypt(originalbytes)
		b.StopTimer()
	}
}
