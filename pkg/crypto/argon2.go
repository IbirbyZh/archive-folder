package crypto

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/crypto/argon2"
)

type Salt [256]byte

func GenerateSalt() (s Salt, err error) {
	_, err = io.ReadFull(rand.Reader, s[:])
	return
}

type Argon2Settings struct {
	Memory  uint32 `json:"memory"`
	Threads uint8  `json:"parallelism"`
	Time    uint32 `json:"iterations"`
}

func DefaultArgon2Settings() Argon2Settings {
	return Argon2Settings{
		Memory:  32768, // 32 MB
		Threads: 2,
		Time:    2,
	}
}

func LoadArgon2Settings(filePath string) (a2s Argon2Settings, err error) {
	f, err := os.Open(filePath)
	if err != nil {
		return
	}

	content, err := io.ReadAll(f)
	if err != nil {
		return
	}

	err = json.Unmarshal(content, &a2s)
	return
}

func GenerateArgonKey(password []byte, salt Salt, a2s Argon2Settings, verbose bool) []byte {
	now := time.Now()
	defer func() {
		if verbose {
			fmt.Printf("Argon2 key time: %v\n", time.Since(now))
		}
	}()
	return argon2.IDKey(
		password,
		salt[:],
		a2s.Time,
		a2s.Memory,
		a2s.Threads,
		32, // AES-256
	)
}
