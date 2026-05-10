package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-management/internal/handler/health"
	"github.com/huypham67/bookmark-management/internal/handler/link"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Router wraps the Gin engine and application server configuration.
type Router struct {
	engine *gin.Engine
	port   string
}

// NewRouter creates and configures a new HTTP router with all API endpoints.
func NewRouter(
	port string,
	healthCheckHandler health.HealthCheck,
	linkHandler link.Link,
) *Router {
	engine := gin.Default()

	engine.GET(
		"/swagger/*any",
		ginSwagger.WrapHandler(swaggerFiles.Handler),
	)

	apiV1 := engine.Group("/api/v1")
	{
		apiV1.GET(
			"/health-check",
			healthCheckHandler.GetHealthCheck,
		)

		apiV1.POST("/links/shorten-url", linkHandler.ShortenURL)
	}

	return &Router{
		engine: engine,
		port:   port,
	}
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
