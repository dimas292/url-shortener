package url

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type UrlService struct {
	db *gorm.DB
	rdb *redis.Client
}

func NewUrlService(db *gorm.DB, rdb *redis.Client) *UrlService {
	return &UrlService{db: db, rdb: rdb}
}

func (s *UrlService) Create(req UrlRequest) (*UrlResponse, error) {
	
	var url Url
	url.OriginalUrl = req.OriginalUrl
	url.ShortUrl = s.GenerateShortUrl()
	if err := s.db.Create(&url).Error; err != nil {
		return nil, err
	}
	return &UrlResponse{
		ShortUrl: url.ShortUrl,
		OriginalUrl: url.OriginalUrl,
	}, nil
}

func (s *UrlService) Redirect(shortUrl string) (*UrlResponse, error) {
	ctx := context.Background()
	chachedData, err := s.rdb.Get(ctx, shortUrl).Result()
	if err == nil && len(chachedData) > 0 {
		return &UrlResponse{
			ShortUrl: shortUrl,
			OriginalUrl: chachedData,
		}, nil
	}

	if err != nil && err != redis.Nil {
        log.Printf("warning: redis get url error: %v", err)
    }

	var url Url
	if err := s.db.Where("short_url = ?", shortUrl).First(&url).Error; err != nil {
		return nil, err
	}

	if err := s.rdb.Set(ctx, url.ShortUrl, url.OriginalUrl, 30*time.Minute).Err(); err != nil {
        log.Printf("warning: redis set url error: %v", err)
    }

	return &UrlResponse{
		ShortUrl: url.ShortUrl,
		OriginalUrl: url.OriginalUrl,
	}, nil
}

func(s *UrlService) FindAll() ([]UrlResponse, error) {
	var urls []Url
	if err := s.db.Find(&urls).Error; err != nil {
		return nil, err
	}
	var responses []UrlResponse
	for _, url := range urls {
		responses = append(responses, UrlResponse{
			ShortUrl: url.ShortUrl,
			OriginalUrl: url.OriginalUrl,
		})
	}
	return responses, nil
}

func (s *UrlService) GenerateShortUrl() string {
	characters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var shortUrl string
	for i := 0; i < 8; i++ {
		shortUrl += string(characters[rand.Intn(len(characters))])
	}
	return shortUrl
}




