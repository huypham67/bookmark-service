package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-management/internal/handler"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Router struct {
	engine *gin.Engine
	port   string
}

func NewRouter(
	port string,
	healthCheckHandler handler.HealthCheck,
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
	}

	return &Router{
		engine: engine,
		port:   port,
	}
}

func (r *Router) ServeHTTP(
	writer http.ResponseWriter,
	request *http.Request,
) {
	r.engine.ServeHTTP(writer, request)
}

func (r *Router) Run() error {
	return r.engine.Run(":" + r.port)
}
