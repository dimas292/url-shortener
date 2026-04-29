package url

import (
	"net/http"

	pkgauth "github.com/dimas292/url_shortener/pkg/auth"
	"github.com/dimas292/url_shortener/pkg/handler"
	"github.com/dimas292/url_shortener/pkg/repository"
	"github.com/dimas292/url_shortener/pkg/response"
	"github.com/dimas292/url_shortener/pkg/service"
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
	urls.Use(pkgauth.AuthMiddleware(m.JWTService()))
	urls.GET("/:shortUrl", m.Redirect)
	urls.POST("/shorten", m.Shorten)
	
	// Generic CRUD — skip GET /:id so we can use custom Redirect handler
	repo := repository.NewBaseRepository[Url, *Url](m.service.db)
	svc := service.NewBaseService[Url, *Url](repo)
	hndl := handler.NewBaseHandler(svc)
	hndl.RegisterCRUD(urls, handler.WithExclude(handler.RouteFindByID))

	// Custom GET /:id → Redirect by short URL
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
	
