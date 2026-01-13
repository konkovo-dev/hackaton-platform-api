package idempotency

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"google.golang.org/protobuf/proto"
)

type Helper struct {
	repo Repository
	cfg  *Config
}

func NewHelper(repo Repository, cfg *Config) *Helper {
	return &Helper{
		repo: repo,
		cfg:  cfg,
	}
}

func (h *Helper) CheckAndGet(ctx context.Context, key, scope, requestHash string, response proto.Message) (bool, error) {
	if key == "" {
		return false, nil
	}

	stored, err := h.repo.Get(ctx, key, scope)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return false, nil
		}
		return false, err
	}

	if stored.RequestHash != requestHash {
		return false, &ConflictError{Message: "idempotency key already used with different request"}
	}

	if err := proto.Unmarshal(stored.ResponseBlob, response); err != nil {
		return false, &ResponseInvalidError{Err: err}
	}

	return true, nil
}

func (h *Helper) Save(ctx context.Context, key, scope, requestHash string, response proto.Message) error {
	if key == "" {
		return nil
	}

	responseBlob, err := proto.Marshal(response)
	if err != nil {
		return err
	}

	expiresAt := time.Now().UTC().Add(h.cfg.TTL)
	return h.repo.Set(ctx, key, scope, requestHash, responseBlob, expiresAt)
}

func ComputeHash(fields ...string) string {
	h := sha256.New()
	for _, field := range fields {
		h.Write([]byte(field))
	}
	return hex.EncodeToString(h.Sum(nil))
}
