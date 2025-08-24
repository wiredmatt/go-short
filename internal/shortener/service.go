package shortener

import (
	"errors"
	"log/slog"
	"math/rand"
	"os"
	"time"

	"github.com/wiredmatt/go_short/internal/model"
	"github.com/wiredmatt/go_short/internal/storage"
)

// Shortener defines the interface for URL shortening operations
type Shortener interface {
	GetBaseURL() string
	Shorten(userID, originalURL string) (string, error)
	Resolve(code string) (string, error)
	ListMappings(userID string) ([]model.URLMapping, error)
}

type ShortenerService struct {
	store           storage.Store
	baseURL         string
	shortCodeLength int
	logger          *slog.Logger
}

func NewService(store storage.Store, baseURL string, shortCodeLength int) *ShortenerService {
	return &ShortenerService{
		store:           store,
		baseURL:         baseURL,
		shortCodeLength: shortCodeLength,
		logger: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})),
	}
}

func (s *ShortenerService) GetBaseURL() string {
	return s.baseURL
}

func (s *ShortenerService) Shorten(userID, originalURL string) (string, error) {
	s.logger.Info("Shortening new url: ", slog.String("originalURL", originalURL))

	code := generateCode(s.shortCodeLength)
	mapping := model.URLMapping{
		Code:      code,
		Original:  originalURL,
		UserID:    userID,
		CreatedAt: time.Now(),
	}
	err := s.store.Save(mapping)
	if err != nil {
		s.logger.Error("Shorten failed",
			slog.Any("input", mapping),
			slog.String("error", err.Error()),
		)
		return "", err
	}
	return code, nil
}

func (s *ShortenerService) Resolve(code string) (string, error) {
	original_url, err := s.store.Get(code)
	if err != nil {
		s.logger.Error("Resolve failed",
			slog.Group("input", slog.String("code", code)),
			slog.String("error", err.Error()),
		)
		return "", err
	}

	if original_url == nil {
		return "", errors.New("code not found")
	}

	// Increment click count asynchronously to avoid blocking the redirect
	go func() {
		if err := s.store.IncrementClickCount(code); err != nil {
			s.logger.Warn("Failed to increment click count",
				slog.String("code", code),
				slog.String("error", err.Error()),
			)
		}
	}()

	return *original_url, nil
}

func (s *ShortenerService) ListMappings(userID string) ([]model.URLMapping, error) {
	mappings, err := s.store.ListByUser(userID)
	if err != nil {
		s.logger.Error("ListMappings failed",
			slog.String("userID", userID),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	return mappings, nil
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
