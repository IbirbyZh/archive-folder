package crypto

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"
)

const (
	key            = "this is key for cipher0123456789"
	random         = "BPyAZgnkTsc3OqUv"
	testMessage    = "some text"
	testEncMessage = "\x42\x50\x79\x41\x5a\x67\x6e\x6b\x54\x73\x63\x33\x4f\x71\x55\x76\x23\xf8\x19\x6e\x24\x6b\x63\x63\xdf"
)

func TestEncryptMessage(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString(random)
	e, err := NewEncrypter([]byte(key), WithRandomSource(&buf))
	if err != nil {
		t.Fatal(err)
	}

	encrypted, err := e.EncryptMessage([]byte(testMessage))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encrypted, []byte(testEncMessage)) {
		t.Fatalf("Wrong result: %x\nExpected: %x", encrypted, testEncMessage)
	}
}

func TestDecryptMessage(t *testing.T) {
	e, err := NewEncrypter([]byte(key))
	if err != nil {
		t.Fatal(err)
	}
	message, err := e.DecryptMessage([]byte(testEncMessage))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(message, []byte(testMessage)) {
		t.Fatalf("Wrong result: %x\nExpected: %x", message, testMessage)
	}
}

func getRandomBytes(length int64) []byte {
	var buf bytes.Buffer
	io.CopyN(&buf, rand.Reader, length)
	return buf.Bytes()
}

func TestFull(t *testing.T) {
	message := getRandomBytes(10000)

	e, err := NewEncrypter(getRandomBytes(32))
	if err != nil {
		t.Fatal(err)
	}

	eMessage, err := e.EncryptMessage(message)
	if err != nil {
		t.Fatal(err)
	}

	dMessage, err := e.DecryptMessage(eMessage)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(message, dMessage) {
		t.Fatalf("Wrong result: %x\nExpected: %x", dMessage, message)
	}
}
