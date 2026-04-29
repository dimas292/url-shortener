package url

import "github.com/dimas292/url_shortener/pkg/model"


type Url struct {
	model.BaseModel
	ShortUrl    string `json:"short_url" gorm:"type:varchar(10);not null" binding:"required"`
	OriginalUrl string `json:"original_url" gorm:"type:varchar(255);not null" binding:"required"`
}

type UrlRequest struct {
	OriginalUrl string `json:"original_url" gorm:"type:varchar(255);not null" binding:"required"`
}

type UrlResponse struct {
	ShortUrl    string `json:"short_url" gorm:"type:varchar(10);not null" binding:"required"`
	OriginalUrl string `json:"original_url" gorm:"type:varchar(255);not null" binding:"required"`
}

func (u *Url) ToResponse() UrlResponse {
	return UrlResponse{
		ShortUrl:    u.ShortUrl,
		OriginalUrl: u.OriginalUrl,
	}
}

func (u *Url) TableName() string {
	return "t_urls"
}
