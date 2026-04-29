package handler

import (
	"math"
	"net/http"

	"github.com/dimas292/url_shortener/pkg/model"
	"github.com/dimas292/url_shortener/pkg/response"
	"github.com/dimas292/url_shortener/pkg/service"
	"github.com/gin-gonic/gin"
)

// BaseHandler provides generic CRUD HTTP handlers for any model.
// No custom code needed — just wire it up and get full CRUD endpoints.
type BaseHandler[T any, PT model.ModelPtr[T]] struct {
	Service *service.BaseService[T, PT]
}

// NewBaseHandler creates a new generic CRUD handler.
func NewBaseHandler[T any, PT model.ModelPtr[T]](svc *service.BaseService[T, PT]) *BaseHandler[T, PT] {
	return &BaseHandler[T, PT]{Service: svc}
}

// Create handles POST / — create a new record.
func (h *BaseHandler[T, PT]) Create(c *gin.Context) {
	entity := PT(new(T))
	if err := c.ShouldBindJSON(entity); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.Service.Create(entity); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Created(c, "created successfully", entity)
}

// FindAll handles GET / — list all records with pagination.
func (h *BaseHandler[T, PT]) FindAll(c *gin.Context) {
	var query response.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	entities, total, err := h.Service.FindAll(query.Page, query.PerPage)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	totalPage := int(math.Ceil(float64(total) / float64(query.PerPage)))

	response.Paginated(c, "retrieved successfully", entities, response.Meta{
		Page:      query.Page,
		PerPage:   query.PerPage,
		Total:     total,
		TotalPage: totalPage,
	})
}

// FindByID handles GET /:id — get a single record by UUID.
func (h *BaseHandler[T, PT]) FindByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return
	}

	entity, err := h.Service.FindByID(id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "not found")
		return
	}

	response.Success(c, "retrieved successfully", entity)
}

// Update handles PUT /:id — update an existing record.
func (h *BaseHandler[T, PT]) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return
	}

	entity, err := h.Service.FindByID(id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "not found")
		return
	}

	if err := c.ShouldBindJSON(entity); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.Service.Update(entity); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, "updated successfully", entity)
}

// Delete handles DELETE /:id — soft-delete a record.
func (h *BaseHandler[T, PT]) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.Service.Delete(id); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, "deleted successfully", nil)
}

// Route identifies a specific CRUD route that can be excluded.
type Route int

const (
	RouteCreate   Route = iota // POST   /
	RouteFindAll               // GET    /
	RouteFindByID              // GET    /:id
	RouteUpdate                // PUT    /:id
	RouteDelete                // DELETE /:id
)

// CRUDOptions holds configuration for RegisterCRUD.
type CRUDOptions struct {
	Exclude map[Route]bool
}

// CRUDOption is a functional option for RegisterCRUD.
type CRUDOption func(*CRUDOptions)

// WithExclude returns an option that excludes the specified routes
// from being registered, so you can provide your own custom handler.
//
// Example:
//
//	hndl.RegisterCRUD(rg, handler.WithExclude(handler.RouteFindByID))
//	rg.GET("/:id", myCustomHandler)
func WithExclude(routes ...Route) CRUDOption {
	return func(o *CRUDOptions) {
		for _, r := range routes {
			o.Exclude[r] = true
		}
	}
}

// RegisterCRUD registers CRUD routes on the given router group.
// Use WithExclude to skip specific routes you want to override.
//
//	POST   /          → Create
//	GET    /          → FindAll (paginated)
//	GET    /:id       → FindByID
//	PUT    /:id       → Update
//	DELETE /:id       → Delete
func (h *BaseHandler[T, PT]) RegisterCRUD(rg *gin.RouterGroup, opts ...CRUDOption) {
	o := &CRUDOptions{Exclude: make(map[Route]bool)}
	for _, fn := range opts {
		fn(o)
	}

	if !o.Exclude[RouteCreate] {
		rg.POST("", h.Create)
	}
	if !o.Exclude[RouteFindAll] {
		rg.GET("", h.FindAll)
	}
	if !o.Exclude[RouteFindByID] {
		rg.GET("/:id", h.FindByID)
	}
	if !o.Exclude[RouteUpdate] {
		rg.PUT("/:id", h.Update)
	}
	if !o.Exclude[RouteDelete] {
		rg.DELETE("/:id", h.Delete)
	}
}
