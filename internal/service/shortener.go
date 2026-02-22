package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	"github.com/kumakikensuke/url-shortener/internal/repository"
)

const (
	alphabet   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	codeLength = 8
	maxRetries = 3
)

type Repo interface {
	Put(ctx context.Context, code, originalURL string) error
	Get(ctx context.Context, code string) (*repository.URLItem, error)
}

type Shortener struct {
	repo    Repo
	baseURL string
}

func NewShortener(repo Repo, baseURL string) *Shortener {
	return &Shortener{repo: repo, baseURL: baseURL}
}

func (s *Shortener) Shorten(ctx context.Context, originalURL string) (code, shortURL string, err error) {
	for i := 0; i < maxRetries; i++ {
		code, err = generateCode()
		if err != nil {
			return "", "", fmt.Errorf("generate code: %w", err)
		}

		err = s.repo.Put(ctx, code, originalURL)
		if err == nil {
			return code, s.baseURL + "/" + code, nil
		}
		if !errors.Is(err, repository.ErrConflict) {
			return "", "", fmt.Errorf("put item: %w", err)
		}
	}
	return "", "", fmt.Errorf("failed to generate unique code after %d retries", maxRetries)
}

func (s *Shortener) Resolve(ctx context.Context, code string) (string, error) {
	item, err := s.repo.Get(ctx, code)
	if err != nil {
		return "", err
	}
	return item.OriginalURL, nil
}

func generateCode() (string, error) {
	b := make([]byte, codeLength)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		b[i] = alphabet[n.Int64()]
	}
	return string(b), nil
}
