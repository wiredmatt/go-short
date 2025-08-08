package shortener

import (
	"errors"
	"math/rand"
	"time"

	"github.com/wiredmatt/go-backend-template/internal/model"
	"github.com/wiredmatt/go-backend-template/internal/storage"
)

type Service struct {
	store   storage.Store
	baseURL string
}

func NewService(store storage.Store, baseURL string) *Service {
	return &Service{store: store, baseURL: baseURL}
}

func (s *Service) GetBaseURL() string {
	return s.baseURL
}

func (s *Service) Shorten(userID, originalURL string) (string, error) {
	code := generateCode(6)
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

func (s *Service) Resolve(code string) (string, error) {
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
