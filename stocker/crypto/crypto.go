package crypto

type Crypter interface {
	EncryptString(plaintext string) (message string, err error)
	DecryptString(message string) (plaintext string, err error)
}
