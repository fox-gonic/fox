package main

import (
	"github.com/fox-gonic/fox"
)

func main() {
	// Create domain engine
	de := fox.NewDomainEngine()

	// API domain: api.example.com
	de.Domain("api.example.com", func(apiRouter *fox.Engine) {
		apiRouter.GET("/", func(ctx *fox.Context) map[string]string {
			return map[string]string{
				"domain":  "api.example.com",
				"service": "API",
			}
		})

		apiRouter.GET("/users", func(ctx *fox.Context) map[string]interface{} {
			return map[string]interface{}{
				"users": []map[string]interface{}{
					{"id": 1, "name": "Alice"},
					{"id": 2, "name": "Bob"},
				},
			}
		})

		apiRouter.GET("/status", func(ctx *fox.Context) map[string]string {
			return map[string]string{
				"status": "API service running",
			}
		})
	})

	// Admin domain: admin.example.com
	de.Domain("admin.example.com", func(adminRouter *fox.Engine) {
		adminRouter.GET("/", func(ctx *fox.Context) map[string]string {
			return map[string]string{
				"domain":  "admin.example.com",
				"service": "Admin Panel",
			}
		})

		adminRouter.GET("/dashboard", func(ctx *fox.Context) map[string]interface{} {
			return map[string]interface{}{
				"title": "Admin Dashboard",
				"stats": map[string]int{
					"users":  1500,
					"orders": 3200,
				},
			}
		})

		adminRouter.GET("/settings", func(ctx *fox.Context) map[string]string {
			return map[string]string{
				"page": "settings",
			}
		})
	})

	// Regex domain pattern: *.staging.example.com
	de.DomainRegexp(`^.*\.staging\.example\.com$`, func(stagingRouter *fox.Engine) {
		stagingRouter.GET("/", func(ctx *fox.Context) map[string]interface{} {
			return map[string]interface{}{
				"environment": "staging",
				"host":        ctx.Request.Host,
			}
		})

		stagingRouter.GET("/info", func(ctx *fox.Context) map[string]interface{} {
			return map[string]interface{}{
				"environment": "staging",
				"message":     "This is a staging environment",
				"subdomain":   ctx.Request.Host,
			}
		})
	})

	// Regex domain pattern: subdomain pattern
	de.DomainRegexp(`^[a-z0-9]+\.app\.example\.com$`, func(appRouter *fox.Engine) {
		appRouter.GET("/", func(ctx *fox.Context) map[string]interface{} {
			return map[string]interface{}{
				"type":   "tenant app",
				"tenant": ctx.Request.Host,
			}
		})

		appRouter.GET("/tenant-info", func(ctx *fox.Context) map[string]interface{} {
			// Extract tenant from subdomain
			host := ctx.Request.Host
			return map[string]interface{}{
				"tenant": host,
				"status": "active",
			}
		})
	})

	// Default domain (fallback) - www.example.com or example.com
	de.GET("/", func(ctx *fox.Context) map[string]string {
		return map[string]string{
			"domain":  "default",
			"service": "Main Website",
		}
	})

	de.GET("/about", func(ctx *fox.Context) map[string]string {
		return map[string]string{
			"page": "about",
			"info": "Main website about page",
		}
	})

	de.GET("/contact", func(ctx *fox.Context) map[string]string {
		return map[string]string{
			"page":  "contact",
			"email": "contact@example.com",
		}
	})

	// Start server
	de.Run(":8080")
}
