package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	DefaultMemory      = 64 * 1024
	DefaultIterations  = 3
	DefaultParallelism = 2
	DefaultSaltLength  = 16
	DefaultKeyLength   = 32
)

type Service struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

func NewService() *Service {
	return &Service{
		memory:      DefaultMemory,
		iterations:  DefaultIterations,
		parallelism: DefaultParallelism,
		saltLength:  DefaultSaltLength,
		keyLength:   DefaultKeyLength,
	}
}

func (s *Service) Hash(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	salt := make([]byte, s.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		s.iterations,
		s.memory,
		s.parallelism,
		s.keyLength,
	)

	saltEncoded := base64.RawStdEncoding.EncodeToString(salt)
	hashEncoded := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		s.memory,
		s.iterations,
		s.parallelism,
		saltEncoded,
		hashEncoded,
	)

	return encoded, nil
}

func (s *Service) Verify(password, encodedHash string) error {
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	if encodedHash == "" {
		return fmt.Errorf("hash cannot be empty")
	}

	salt, hash, params, err := s.decodeHash(encodedHash)
	if err != nil {
		return fmt.Errorf("failed to decode hash: %w", err)
	}

	computedHash := argon2.IDKey(
		[]byte(password),
		salt,
		params.iterations,
		params.memory,
		params.parallelism,
		uint32(len(hash)),
	)

	if subtle.ConstantTimeCompare(hash, computedHash) != 1 {
		return fmt.Errorf("invalid password")
	}

	return nil
}

type hashParams struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
}

func (s *Service) decodeHash(encodedHash string) (salt, hash []byte, params hashParams, err error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return nil, nil, hashParams{}, fmt.Errorf("invalid hash format")
	}

	if parts[1] != "argon2id" {
		return nil, nil, hashParams{}, fmt.Errorf("unsupported algorithm: %s", parts[1])
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return nil, nil, hashParams{}, fmt.Errorf("invalid version: %w", err)
	}

	if version != argon2.Version {
		return nil, nil, hashParams{}, fmt.Errorf("incompatible version: %d", version)
	}

	_, err = fmt.Sscanf(
		parts[3],
		"m=%d,t=%d,p=%d",
		&params.memory,
		&params.iterations,
		&params.parallelism,
	)
	if err != nil {
		return nil, nil, hashParams{}, fmt.Errorf("invalid params: %w", err)
	}

	salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, hashParams{}, fmt.Errorf("invalid salt encoding: %w", err)
	}

	hash, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, hashParams{}, fmt.Errorf("invalid hash encoding: %w", err)
	}

	return salt, hash, params, nil
}
