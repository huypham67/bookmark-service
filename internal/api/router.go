package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-management/internal/handler"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Router wraps the Gin engine and application server configuration.
type Router struct {
	engine *gin.Engine
	port   string
}

// NewRouter creates and configures a new HTTP router with all API endpoints.
func NewRouter(port string) *Router {
	engine := gin.Default()

	engine.GET(
		"/swagger/*any",
		ginSwagger.WrapHandler(swaggerFiles.Handler),
	)

	return &Router{
		engine: engine,
		port:   port,
	}
}

// GroupV1 returns a router group for API version 1 endpoints.
func (r *Router) GroupV1() *gin.RouterGroup {
	return r.engine.Group("/api/v1")
}

// RegisterHealthRoutes registers all health-check routes.
func RegisterHealthRoutes(
	routerGroup *gin.RouterGroup,
	healthCheckHandler handler.HealthCheck,
) {
	routerGroup.GET(
		"/health-check",
		healthCheckHandler.GetHealthCheck,
	)
}

// RegisterLinkRoutes registers all link management routes.
func RegisterLinkRoutes(
	routerGroup *gin.RouterGroup,
	linkHandler handler.Link,
) {
	routerGroup.POST(
		"/links/shorten",
		linkHandler.ShortenURL,
	)

	routerGroup.GET(
		"/links/redirect/:code",
		linkHandler.RedirectToURL,
	)
}

// ServeHTTP implements the http.Handler interface.
func (r *Router) ServeHTTP(
	writer http.ResponseWriter,
	request *http.Request,
) {
	r.engine.ServeHTTP(writer, request)
}

// Run starts the HTTP server on the configured port.
func (r *Router) Run() error {
	return r.engine.Run(":" + r.port)
}
