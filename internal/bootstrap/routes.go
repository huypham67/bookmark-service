package bootstrap

import (
	"github.com/newrelic/go-agent/v3/integrations/nrgin"

	"github.com/huypham67/bookmark-service/internal/api"
)

// SetupRoutes registers all API routes with the provided router and dependency container.
func SetupRoutes(router *api.Router, container *Container) {
	apiGroup := router.GroupAPI()
	apiV1Group := router.GroupV1()

	// NR traces all v1 business routes. Health check stays on apiGroup (no NR)
	// to avoid load-balancer probe noise flooding the APM transaction list.
	if container.NRApp != nil {
		apiV1Group.Use(nrgin.Middleware(container.NRApp))
	}

	// Rate limit every public endpoint. GroupV1 is a separate group object from
	// GroupAPI, so the middleware must be applied to both to cover all routes.
	apiGroup.Use(container.RateLimitMiddleware)
	apiV1Group.Use(container.RateLimitMiddleware)

	// Register all feature routes in order
	api.RegisterHealthRoutes(apiGroup, container.HealthHandler)
	api.RegisterLinkRoutes(apiV1Group, container.LinkHandler)
	api.RegisterBookmarkRoutes(apiV1Group, container.BookmarkHandler, container.JWTMiddleware)
}
