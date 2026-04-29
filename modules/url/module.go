package url

import (
	"net/http"

	pkgauth "github.com/dimas292/url_shortener/pkg/auth"
	"github.com/dimas292/url_shortener/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type UrlModule struct {
	service *UrlService
	jwtService *pkgauth.JWTService
}

func NewUrlModule(db *gorm.DB, rdb *redis.Client, jwtService *pkgauth.JWTService) *UrlModule {
	db.AutoMigrate(&Url{})

	service := NewUrlService(db, rdb)

	return &UrlModule{service: service, jwtService: jwtService}
}

func (m *UrlModule) JWTService() *pkgauth.JWTService {
	return m.jwtService
}



func (m *UrlModule) RegisterRoutes(rg *gin.RouterGroup) {
	urls := rg.Group("/url")
	protected := urls.Group("")
	protected.Use(pkgauth.AuthMiddleware(m.JWTService()))
	{
		protected.GET("/", m.FindAll)
		protected.POST("/shorten", m.Shorten)
		protected.GET("/:shortUrl", m.Redirect)
	}
	
}

func (m *UrlModule) Shorten(c *gin.Context) {
	var url UrlRequest
	if err := c.ShouldBindJSON(&url); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := m.service.Create(url)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Created(c, "URL shortened successfully", result)
}

func (m *UrlModule) Redirect(c *gin.Context) {
	shortUrl := c.Param("shortUrl")
	result, err := m.service.Redirect(shortUrl)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.Redirect(http.StatusPermanentRedirect, result.OriginalUrl)
}

func (m *UrlModule) FindAll(c *gin.Context) {
	result, err := m.service.FindAll()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, "URLs retrieved successfully", result)
}
	
