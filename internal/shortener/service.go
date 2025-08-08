package shortener

import (
	"errors"
	"math/rand"
	"time"

	"github.com/wiredmatt/go-backend-template/internal/model"
	"github.com/wiredmatt/go-backend-template/internal/storage"
)

// Service defines the interface for URL shortening operations
type IShortenerService interface {
	GetBaseURL() string
	Shorten(userID, originalURL string) (string, error)
	Resolve(code string) (string, error)
}

type ShortenerService struct {
	store           storage.Store
	baseURL         string
	shortCodeLength int
}

func NewService(store storage.Store, baseURL string, shortCodeLength int) *ShortenerService {
	return &ShortenerService{
		store:           store,
		baseURL:         baseURL,
		shortCodeLength: shortCodeLength,
	}
}

func (s *ShortenerService) GetBaseURL() string {
	return s.baseURL
}

func (s *ShortenerService) Shorten(userID, originalURL string) (string, error) {
	code := generateCode(s.shortCodeLength)
	mapping := model.URLMapping{
		Code:      code,
		Original:  originalURL,
		UserID:    userID,
		CreatedAt: time.Now(),
	}
	err := s.store.Save(mapping)
	if err != nil {
		return "", err
	}
	return code, nil
}

func (s *ShortenerService) Resolve(code string) (string, error) {
	original_url, err := s.store.Get(code)
	if err != nil || original_url == nil {
		return "", errors.New("code not found")
	}
	return *original_url, nil
}

func generateCode(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
