package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-service/internal/handler"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Router wraps the Gin engine and application server configuration.
type Router struct {
	engine *gin.Engine
}

// NewRouter creates and configures a new HTTP router with all API endpoints.
func NewRouter() *Router {
	engine := gin.Default()

	engine.GET(
		"/swagger/*any",
		ginSwagger.WrapHandler(swaggerFiles.Handler),
	)

	return &Router{
		engine: engine,
	}
}

// GroupAPI returns a router group for endpoints (no /api prefix - handled by API Gateway).
func (r *Router) GroupAPI() *gin.RouterGroup {
	return r.engine.Group("/api/bookmark_service")
}

// GroupV1 returns a router group for API version 1 endpoints.
func (r *Router) GroupV1() *gin.RouterGroup {
	return r.GroupAPI().Group("/v1")
}

// RegisterHealthRoutes registers all health-check routes.
func RegisterHealthRoutes(
	apiGroup *gin.RouterGroup,
	healthCheckHandler handler.HealthCheck,
) {
	apiGroup.GET(
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

// RegisterUserRoutes registers all user management routes.
func RegisterUserRoutes(
	routerGroup *gin.RouterGroup,
	userHandler handler.User,
) {
	routerGroup.POST(
		"/users/register",
		userHandler.Register,
	)
}

// ServeHTTP implements the http.Handler interface.
func (r *Router) ServeHTTP(
	writer http.ResponseWriter,
	request *http.Request,
) {
	r.engine.ServeHTTP(writer, request)
}

// Engine exposes underlying Gin engine
func (r *Router) Engine() *gin.Engine {
	return r.engine
}
