package auth

import (
	"fmt"
	"testing"
)

var ClientTestPrivateKeys = [][]byte{
	[]byte(`-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----
`),
	[]byte(`-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----
`),
	[]byte(`-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----
`),
}

var ClientTestUnauthorizedKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----
`)

func TestClient(t *testing.T) {

	server, err := newTestServer()
	if err != nil {
		t.Fatal(err)
	}

	go server.ListenAndServe(`:2022`)

	for i, privateKey := range ClientTestPrivateKeys {
		fmt.Println("key")

		// Ensure the reader connection can only read.
		rclient, err := NewClient(ReaderUser, `:2022`, privateKey)
		if err != nil {
			t.Error(err)
		} else {
			if _, err := rclient.Run(fmt.Sprintf("export A=%d", i), nil); err == nil {
				t.Error("write command allowed for reader")
			}

			if _, err := rclient.Run("unset A", nil); err == nil {
				t.Error("write command allowed for reader")
			}

			if _, err := rclient.Run("env", nil); err != nil {
				t.Error(err)
			}
		}

		// Close the reader client.
		rclient.Close()

		// Ensure the writer connection can read and write.
		wclient, err := NewClient(WriterUser, `:2022`, privateKey)
		if err != nil {
			t.Error(err)
		} else {
			if _, err := wclient.Run(fmt.Sprintf("export A=%d", i), nil); err != nil {
				t.Error("write command allowed for reader")
			}

			if out, err := wclient.Run("env", nil); err != nil {
				t.Error(err)
			} else if out != fmt.Sprintf("A=%d\n", i) {
				t.Error(out)
			}

			if _, err := wclient.Run("unset A", nil); err != nil {
				t.Error("write command allowed for reader")
			}

			if out, err := wclient.Run("env", nil); err != nil {
				t.Error(err)
			} else if out != "" {
				t.Error(out)
			}
		}

		// Close the writer client.
		wclient.Close()
	}

	if err := server.Stop(); err != nil {
		t.Fatal(err)
	}
}

func TestClientSetEnv(t *testing.T) {

	server, err := newTestServer()
	if err != nil {
		t.Fatal(err)
	}

	go server.ListenAndServe(`:2022`)

	client, err := NewClient(WriterUser, `:2022`, ClientTestPrivateKeys[0])
	if err != nil {
		t.Fatal(err)
	}

	env := make(map[string]string)

	env["A"] = "setting"

	if _, err := client.Run("export A", env); err != nil {
		t.Error("write command allowed for reader")
	}

	if out, err := client.Run("env", nil); err != nil {
		t.Error(err)
	} else if out != "A=setting\n" {
		t.Error(out)
	}

	if _, err := client.Run("unset A", nil); err != nil {
		t.Error(err)
	}

	// Close the writer client.
	client.Close()

	if err := server.Stop(); err != nil {
		t.Fatal(err)
	}
}

func TestClientUnauthorized(t *testing.T) {

	server, err := newTestServer()
	if err != nil {
		t.Fatal(err)
	}

	go server.ListenAndServe(`:2022`)

	if _, err := NewClient(WriterUser, `:2022`, ClientTestUnauthorizedKey); err == nil {
		t.Error("unauthorized client allowed to connect")
	}

	if err := server.Stop(); err != nil {
		t.Fatal(err)
	}
}
