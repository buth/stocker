package auth

import (
	"bytes"
	"code.google.com/p/go.crypto/ssh"
	"container/list"
	"encoding/binary"
)

func UnpackMessage(message []byte) ([]string, error) {

	// Create a new buffer based on the message byte slice and initialize an
	// empty list.
	buf := bytes.NewBuffer(message)
	l := list.New()

	// We need 4 bytes to define a number of bytes to read. If there are only
	// 4 bytes left we can assume that the number is zero an move on.
	for buf.Len() > 4 {

		// Create a 4-byte unsigned integer to contain the read value from the
		// message and read from the buffer.
		var n uint32
		if err := binary.Read(buf, binary.BigEndian, &n); err != nil {
			return nil, err
		}

		// Convert the unsigned 4-byte integer to an int and read out the
		// specified number of bytes. Store the resulting byte slice in the
		// list.
		l.PushBack(buf.Next(int(n)))
	}

	rval := make([]string, l.Len())
	i := 0
	for e := l.Front(); e != nil; e = e.Next() {
		rval[i] = string(e.Value.([]byte))
		i++
	}

	return rval, nil
}

func SerializeKey(key ssh.PublicKey) string {
	return string(key.Marshal())
}

// NotAuthority is intended to be used in conjunction with an SSH CertChecker
// in order to indicate that we are not accepting any certificate as an
// authority.
func NotAnAuthority(auth ssh.PublicKey) bool { return false }
