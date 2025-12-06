package main

import (
	"github.com/fox-gonic/fox"
)

func main() {
	// Create domain engine
	de := fox.NewDomainEngine()

	// API domain: api.example.com
	de.Domain("api.example.com", func(apiRouter *fox.Engine) {
		apiRouter.GET("/", func() map[string]string {
			return map[string]string{
				"domain":  "api.example.com",
				"service": "API",
			}
		})

		apiRouter.GET("/users", func() map[string]any {
			return map[string]any{
				"users": []map[string]any{
					{"id": 1, "name": "Alice"},
					{"id": 2, "name": "Bob"},
				},
			}
		})

		apiRouter.GET("/status", func() map[string]string {
			return map[string]string{
				"status": "API service running",
			}
		})
	})

	// Admin domain: admin.example.com
	de.Domain("admin.example.com", func(adminRouter *fox.Engine) {
		adminRouter.GET("/", func() map[string]string {
			return map[string]string{
				"domain":  "admin.example.com",
				"service": "Admin Panel",
			}
		})

		adminRouter.GET("/dashboard", func() map[string]any {
			return map[string]any{
				"title": "Admin Dashboard",
				"stats": map[string]int{
					"users":  1500,
					"orders": 3200,
				},
			}
		})

		adminRouter.GET("/settings", func() map[string]string {
			return map[string]string{
				"page": "settings",
			}
		})
	})

	// Regex domain pattern: *.staging.example.com
	de.DomainRegexp(`^.*\.staging\.example\.com$`, func(stagingRouter *fox.Engine) {
		stagingRouter.GET("/", func(ctx *fox.Context) map[string]any {
			return map[string]any{
				"environment": "staging",
				"host":        ctx.Request.Host,
			}
		})

		stagingRouter.GET("/info", func(ctx *fox.Context) map[string]any {
			return map[string]any{
				"environment": "staging",
				"message":     "This is a staging environment",
				"subdomain":   ctx.Request.Host,
			}
		})
	})

	// Regex domain pattern: subdomain pattern
	de.DomainRegexp(`^[a-z0-9]+\.app\.example\.com$`, func(appRouter *fox.Engine) {
		appRouter.GET("/", func(ctx *fox.Context) map[string]any {
			return map[string]any{
				"type":   "tenant app",
				"tenant": ctx.Request.Host,
			}
		})

		appRouter.GET("/tenant-info", func(ctx *fox.Context) map[string]any {
			// Extract tenant from subdomain
			host := ctx.Request.Host
			return map[string]any{
				"tenant": host,
				"status": "active",
			}
		})
	})

	// Default domain (fallback) - www.example.com or example.com
	de.GET("/", func() map[string]string {
		return map[string]string{
			"domain":  "default",
			"service": "Main Website",
		}
	})

	de.GET("/about", func() map[string]string {
		return map[string]string{
			"page": "about",
			"info": "Main website about page",
		}
	})

	de.GET("/contact", func() map[string]string {
		return map[string]string{
			"page":  "contact",
			"email": "contact@example.com",
		}
	})

	// Start server
	if err := de.Run(":8080"); err != nil {
		panic(err)
	}
}
