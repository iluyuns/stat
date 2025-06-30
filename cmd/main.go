package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iluyuns/stat"
)

func main() {
	// 初始化应用
	stat.InitApp()

	// 创建 Gin 引擎
	r := stat.InitGinEngine()

	// 添加全局错误处理中间件
	r.Use(globalErrorHandler())

	// 添加CORS中间件
	r.Use(corsMiddleware())

	// 注册路由
	registerRoutes(r)

	// 启动服务器
	port := stat.AppConfig.Server.Port
	log.Printf("服务器启动在端口 %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

// 全局错误处理中间件
func globalErrorHandler() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			log.Printf("Panic: %s", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "服务器内部错误",
			})
		} else {
			log.Printf("Panic: %v", recovered)
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "服务器内部错误",
			})
		}
		c.Abort()
	})
}

// CORS中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// 注册路由
func registerRoutes(r *gin.Engine) {
	// 不需要认证的API路由组
	api := r.Group("/api")
	{
		track := api.Group("/track")
		{
			track.POST("/pv", stat.TrackPV)
			track.POST("/event", stat.TrackEvent)
			track.POST("/duration", stat.TrackDuration)
		}
		statGroup := api.Group("/stat")
		{
			statGroup.GET("/report", stat.GetReport)
			statGroup.GET("/trend", stat.Report)
			statGroup.GET("/popular-devices", stat.GetPopularDevices)
			statGroup.GET("/popular-cities", stat.GetPopularCities)
		}
		site := api.Group("/site")
		{
			site.POST("", stat.CreateSite)
			site.GET("", stat.ListSite)
		}
		auth := api.Group("/auth")
		{
			auth.POST("/register", stat.Register)
			auth.POST("/login", stat.Login)
			auth.POST("/logout", stat.Logout)
			auth.POST("/verify-code", stat.SendVerifyCode)
			auth.POST("/reset-password", stat.ResetPassword)
		}
		script := api.Group("/script")
		{
			script.GET("/generate", stat.GenerateScript)
			script.GET("/stat.js", stat.ServeScript)
		}
	}

	// 需要认证的API路由组
	apiAuth := r.Group("/api")
	apiAuth.Use(stat.AuthMiddleware())
	{
		statAuth := apiAuth.Group("/stat")
		{
			statAuth.GET("/dashboard", stat.GetDashboardData)
			statAuth.GET("/popular-pages", stat.GetPopularPages)
			statAuth.GET("/realtime", stat.GetRealtimeData)
			statAuth.GET("/distribution", stat.GetDistributionData)
			statAuth.GET("/events", stat.ListEvents)
			statAuth.GET("/event-stats", stat.GetEventStats)
			statAuth.GET("/event-trend", stat.GetEventTrend)
			statAuth.GET("/event-distribution", stat.GetEventTypeDistribution)
			statAuth.GET("/event-category-distribution", stat.GetEventCategoryDistribution)
			statAuth.GET("/report-metrics", stat.HandleReportMetrics)
			statAuth.GET("/report-trend", stat.HandleReportTrend)
			statAuth.GET("/report-hour-distribution", stat.HandleReportHourDistribution)
			statAuth.GET("/report-daily-stats", stat.HandleReportDailyStats)
			statAuth.GET("/popular-events", stat.HandlePopularEvents)
		}
		siteAuth := apiAuth.Group("/site")
		{
			siteAuth.PUT("/:siteId", stat.UpdateSite)
			siteAuth.DELETE("/:siteId", stat.DeleteSite)
			siteAuth.GET("/:siteId/settings", stat.GetSiteSettings)
			siteAuth.PUT("/:siteId/settings", stat.UpdateSiteSettings)
			siteAuth.GET("/:siteId/export", stat.ExportSiteData)
			siteAuth.DELETE("/:siteId/data", stat.DeleteSiteData)
		}
		user := apiAuth.Group("/user")
		{
			user.GET("/profile", stat.GetCurrentUser)
		}
	}

	admin := r.Group("/admin")
	{
		admin.GET("/dashboard", renderPage("dashboard", "仪表盘"))
		admin.GET("/sites", renderPage("sites", "站点管理"))
		admin.GET("/reports", renderPage("reports", "数据报表"))
		admin.GET("/events", renderPage("events", "事件分析"))
		admin.GET("/settings", renderPage("settings", "系统设置"))

		// 站点级别的路由
		admin.GET("/site/:siteId/dashboard", renderSitePage("dashboard", "数据概览"))
		admin.GET("/site/:siteId/reports", renderSitePage("reports", "数据报表"))
		admin.GET("/site/:siteId/events", renderSitePage("events", "事件分析"))
		admin.GET("/site/:siteId/settings", renderSitePage("settings", "站点设置"))
	}
	r.Static("/static", "./static")
	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{"Title": "登录"})
	})
	r.GET("/report", func(c *gin.Context) {
		c.File("./stat.html")
	})
	r.GET("/test", func(c *gin.Context) {
		c.File("./test.html")
	})
	r.GET("/health", stat.HealthCheckAPI)
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/login")
	})
}

// 页面渲染函数
func renderPage(page, title string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 使用layout.html作为主模板，传入对应的内容块名称
		c.HTML(http.StatusOK, "layout.html", gin.H{
			"Title":           title,
			"Page":            page,
			"ContentTemplate": page + "_content", // 如 "dashboard_content"
			"User":            nil,               // 前端会通过API获取用户信息
		})
	}
}

// 站点级别页面渲染函数
func renderSitePage(page, title string) gin.HandlerFunc {
	return func(c *gin.Context) {
		siteId := c.Param("siteId")

		// 检查数据库连接是否就绪
		if stat.StatDB == nil {
			c.HTML(http.StatusServiceUnavailable, "error.html", gin.H{
				"Title": "服务暂时不可用",
				"Error": "数据库连接正在初始化中，请稍后重试",
			})
			return
		}

		// 获取站点信息
		var site stat.StatSite
		if err := stat.StatDB.Where("site_id = ?", siteId).First(&site).Error; err != nil {
			c.HTML(http.StatusNotFound, "error.html", gin.H{
				"Title": "站点不存在",
				"Error": "找不到指定的站点",
			})
			return
		}

		// 使用site_layout.html作为主模板
		c.HTML(http.StatusOK, "site_layout.html", gin.H{
			"Title":           title,
			"CurrentPage":     page,
			"ContentTemplate": "site_" + page + "_content", // 如 "site_dashboard_content"
			"SiteId":          siteId,
			"SiteName":        site.Name,
			"SiteRemark":      site.Remark,
		})
	}
}
