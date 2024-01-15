package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

type Encrypter struct {
	block        cipher.Block
	randomSource io.Reader
}

func NewEncrypter(key []byte, options ...func(*Encrypter)) (*Encrypter, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	e := &Encrypter{
		block:        block,
		randomSource: rand.Reader,
	}
	for _, o := range options {
		o(e)
	}

	return e, nil
}

func WithRandomSource(randomSource io.Reader) func(*Encrypter) {
	return func(e *Encrypter) {
		e.randomSource = randomSource
	}
}

func (e *Encrypter) EncryptMessage(message []byte) ([]byte, error) {
	cipherText := make([]byte, aes.BlockSize+len(message))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(e.randomSource, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(e.block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], message)

	return cipherText, nil
}

func (e *Encrypter) DecryptMessage(cipherText []byte) ([]byte, error) {
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(e.block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	return cipherText, nil
}
