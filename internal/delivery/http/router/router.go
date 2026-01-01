package router

import (
	"time"

	"hr-management-system/internal/config"
	"hr-management-system/internal/delivery/http/handler"
	"hr-management-system/internal/delivery/http/middleware"
	"hr-management-system/internal/infrastructure/cache"
	"hr-management-system/internal/infrastructure/database"
	"hr-management-system/internal/infrastructure/email"
	"hr-management-system/internal/infrastructure/logger"
	"hr-management-system/internal/infrastructure/queue"
	"hr-management-system/internal/infrastructure/search"

	"github.com/gin-gonic/gin"
)

type Router struct {
	engine *gin.Engine
	cfg    *config.Config
	db     *database.Database
	cache  *cache.RedisCache
	queue  *queue.Queue
	es     *search.ElasticSearch
	email  *email.EmailService
	log    *logger.Logger
}

func NewRouter(
	cfg *config.Config,
	db *database.Database,
	cache *cache.RedisCache,
	queue *queue.Queue,
	es *search.ElasticSearch,
	emailSvc *email.EmailService,
	log *logger.Logger,
) *Router {
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	return &Router{
		engine: engine,
		cfg:    cfg,
		db:     db,
		cache:  cache,
		queue:  queue,
		es:     es,
		email:  emailSvc,
		log:    log,
	}
}

func (r *Router) Setup() *gin.Engine {
	// Global middlewares
	r.engine.Use(middleware.Recovery(r.log))
	r.engine.Use(middleware.RequestID())
	r.engine.Use(middleware.Logger(r.log))
	r.engine.Use(middleware.CORS(&r.cfg.Security))
	r.engine.Use(middleware.SecurityHeaders())
	r.engine.Use(middleware.Language())
	r.engine.Use(middleware.Timeout(30 * time.Second))

	// Health check
	r.engine.GET("/health", r.healthCheck)
	r.engine.GET("/ready", r.readinessCheck)

	// API v1
	v1 := r.engine.Group("/api/v1")
	{
		// Rate limiting for API
		v1.Use(middleware.RateLimiter(r.cache, &r.cfg.RateLimit))

		r.setupAuthRoutes(v1)
		r.setupEmployeeRoutes(v1)
		r.setupDepartmentRoutes(v1)
		r.setupPositionRoutes(v1)
		r.setupAttendanceRoutes(v1)
		r.setupLeaveRoutes(v1)
		r.setupOvertimeRoutes(v1)
		r.setupPayrollRoutes(v1)
		r.setupRoleRoutes(v1)
		r.setupAddressRoutes(v1)
		r.setupReportRoutes(v1)
		r.setupNotificationRoutes(v1)
	}

	return r.engine
}

func (r *Router) setupAuthRoutes(rg *gin.RouterGroup) {
	authHandler := handler.NewAuthHandler(r.db, r.cache, r.queue, r.email, r.log, r.cfg)

	auth := rg.Group("/auth")
	{
		// Public routes with stricter rate limiting
		auth.POST("/login", middleware.EndpointRateLimiter(r.cache, 5, time.Minute), authHandler.Login)
		auth.POST("/register", middleware.EndpointRateLimiter(r.cache, 3, time.Minute), authHandler.Login) // Placeholder
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.POST("/forgot-password", middleware.EndpointRateLimiter(r.cache, 3, time.Minute), authHandler.ForgotPassword)
		auth.POST("/reset-password", authHandler.ResetPassword)
		auth.POST("/verify-2fa", authHandler.Verify2FA)
		auth.POST("/send-otp", middleware.EndpointRateLimiter(r.cache, 3, 5*time.Minute), authHandler.SendOTP)
		auth.POST("/verify-otp", authHandler.VerifyOTP)

		// Protected routes
		protected := auth.Group("")
		protected.Use(middleware.JWTAuth(&r.cfg.JWT, r.cache))
		{
			protected.POST("/logout", authHandler.Logout)
			protected.POST("/change-password", authHandler.ChangePassword)
			protected.GET("/profile", authHandler.GetProfile)
		}
	}
}

func (r *Router) setupEmployeeRoutes(rg *gin.RouterGroup) {
	h := handler.NewEmployeeHandler(r.db, r.cache, r.queue, r.es, r.log, r.cfg)

	employees := rg.Group("/employees")
	employees.Use(middleware.JWTAuth(&r.cfg.JWT, r.cache))
	{
		employees.GET("", middleware.RequirePermission("employees.view"), h.List)
		employees.GET("/search", middleware.RequirePermission("employees.view"), h.Search)
		employees.GET("/:id", middleware.RequirePermission("employees.view"), h.Get)
		employees.POST("", middleware.RequirePermission("employees.create"), h.Create)
		employees.PUT("/:id", middleware.RequirePermission("employees.update"), h.Update)
		employees.DELETE("/:id", middleware.RequirePermission("employees.delete"), h.Delete)
	}
}

func (r *Router) setupDepartmentRoutes(rg *gin.RouterGroup) {
	departments := rg.Group("/departments")
	departments.Use(middleware.JWTAuth(&r.cfg.JWT, r.cache))
	{
		departments.GET("", middleware.RequirePermission("departments.view"), func(c *gin.Context) {
			// Department list handler
		})
		departments.GET("/tree", middleware.RequirePermission("departments.view"), func(c *gin.Context) {
			// Department tree handler
		})
		departments.GET("/:id", middleware.RequirePermission("departments.view"), func(c *gin.Context) {})
		departments.POST("", middleware.RequirePermission("departments.create"), func(c *gin.Context) {})
		departments.PUT("/:id", middleware.RequirePermission("departments.update"), func(c *gin.Context) {})
		departments.DELETE("/:id", middleware.RequirePermission("departments.delete"), func(c *gin.Context) {})
	}
}

func (r *Router) setupPositionRoutes(rg *gin.RouterGroup) {
	positions := rg.Group("/positions")
	positions.Use(middleware.JWTAuth(&r.cfg.JWT, r.cache))
	{
		positions.GET("", middleware.RequirePermission("positions.view"), func(c *gin.Context) {})
		positions.GET("/:id", middleware.RequirePermission("positions.view"), func(c *gin.Context) {})
		positions.POST("", middleware.RequirePermission("positions.create"), func(c *gin.Context) {})
		positions.PUT("/:id", middleware.RequirePermission("positions.update"), func(c *gin.Context) {})
		positions.DELETE("/:id", middleware.RequirePermission("positions.delete"), func(c *gin.Context) {})
	}
}

func (r *Router) setupAttendanceRoutes(rg *gin.RouterGroup) {
	h := handler.NewAttendanceHandler(r.db, r.cache, r.queue, r.log, r.cfg)

	attendance := rg.Group("/attendance")
	attendance.Use(middleware.JWTAuth(&r.cfg.JWT, r.cache))
	{
		// Self-service
		attendance.POST("/check-in", h.CheckIn)
		attendance.POST("/check-out", h.CheckOut)
		attendance.GET("/my", h.GetMyAttendance)
		attendance.GET("/today", h.GetTodayStatus)

		// Management
		attendance.GET("", middleware.RequirePermission("attendance.view"), h.List)
		attendance.GET("/summary", middleware.RequirePermission("attendance.view"), h.GetSummary)
		attendance.PUT("/:id/approve", middleware.RequirePermission("attendance.approve"), func(c *gin.Context) {})
	}
}

func (r *Router) setupLeaveRoutes(rg *gin.RouterGroup) {
	leave := rg.Group("/leave")
	leave.Use(middleware.JWTAuth(&r.cfg.JWT, r.cache))
	{
		// Types
		leave.GET("/types", func(c *gin.Context) {})

		// Balance
		leave.GET("/balance", func(c *gin.Context) {})
		leave.GET("/balance/:employee_id", middleware.RequirePermission("leave.view"), func(c *gin.Context) {})

		// Requests
		leave.GET("/requests", func(c *gin.Context) {})
		leave.GET("/requests/pending", middleware.RequirePermission("leave.approve"), func(c *gin.Context) {})
		leave.GET("/requests/:id", func(c *gin.Context) {})
		leave.POST("/requests", func(c *gin.Context) {})
		leave.PUT("/requests/:id/cancel", func(c *gin.Context) {})
		leave.PUT("/requests/:id/approve", middleware.RequirePermission("leave.approve"), func(c *gin.Context) {})
	}
}

func (r *Router) setupOvertimeRoutes(rg *gin.RouterGroup) {
	overtime := rg.Group("/overtime")
	overtime.Use(middleware.JWTAuth(&r.cfg.JWT, r.cache))
	{
		overtime.GET("/requests", func(c *gin.Context) {})
		overtime.GET("/requests/pending", middleware.RequirePermission("overtime.approve"), func(c *gin.Context) {})
		overtime.GET("/requests/:id", func(c *gin.Context) {})
		overtime.POST("/requests", func(c *gin.Context) {})
		overtime.PUT("/requests/:id/cancel", func(c *gin.Context) {})
		overtime.PUT("/requests/:id/approve", middleware.RequirePermission("overtime.approve"), func(c *gin.Context) {})

		// Policy
		overtime.GET("/policy", middleware.RequirePermission("overtime.view"), func(c *gin.Context) {})
		overtime.PUT("/policy", middleware.RequirePermission("overtime.manage"), func(c *gin.Context) {})
	}
}

func (r *Router) setupPayrollRoutes(rg *gin.RouterGroup) {
	payroll := rg.Group("/payroll")
	payroll.Use(middleware.JWTAuth(&r.cfg.JWT, r.cache))
	{
		// Periods
		payroll.GET("/periods", middleware.RequirePermission("payroll.view"), func(c *gin.Context) {})
		payroll.GET("/periods/:id", middleware.RequirePermission("payroll.view"), func(c *gin.Context) {})
		payroll.POST("/periods", middleware.RequirePermission("payroll.create"), func(c *gin.Context) {})
		payroll.POST("/periods/:id/calculate", middleware.RequirePermission("payroll.calculate"), func(c *gin.Context) {})
		payroll.PUT("/periods/:id/approve", middleware.RequirePermission("payroll.approve"), func(c *gin.Context) {})
		payroll.PUT("/periods/:id/pay", middleware.RequirePermission("payroll.pay"), func(c *gin.Context) {})

		// Payslips
		payroll.GET("/payslips", func(c *gin.Context) {})
		payroll.GET("/payslips/my", func(c *gin.Context) {})
		payroll.GET("/payslips/:id", func(c *gin.Context) {})
		payroll.GET("/payslips/:id/pdf", func(c *gin.Context) {})
		payroll.POST("/payslips/:id/send-email", middleware.RequirePermission("payroll.view"), func(c *gin.Context) {})

		// Allowances
		payroll.GET("/allowances", middleware.RequirePermission("payroll.view"), func(c *gin.Context) {})
		payroll.POST("/allowances", middleware.RequirePermission("payroll.manage"), func(c *gin.Context) {})

		// Deductions
		payroll.GET("/deductions", middleware.RequirePermission("payroll.view"), func(c *gin.Context) {})
		payroll.POST("/deductions", middleware.RequirePermission("payroll.manage"), func(c *gin.Context) {})
	}
}

func (r *Router) setupRoleRoutes(rg *gin.RouterGroup) {
	roles := rg.Group("/roles")
	roles.Use(middleware.JWTAuth(&r.cfg.JWT, r.cache))
	roles.Use(middleware.RequirePermission("roles.view"))
	{
		roles.GET("", func(c *gin.Context) {})
		roles.GET("/:id", func(c *gin.Context) {})
		roles.POST("", middleware.RequirePermission("roles.create"), func(c *gin.Context) {})
		roles.PUT("/:id", middleware.RequirePermission("roles.update"), func(c *gin.Context) {})
		roles.DELETE("/:id", middleware.RequirePermission("roles.delete"), func(c *gin.Context) {})
	}

	permissions := rg.Group("/permissions")
	permissions.Use(middleware.JWTAuth(&r.cfg.JWT, r.cache))
	{
		permissions.GET("", middleware.RequirePermission("permissions.view"), func(c *gin.Context) {})
	}
}

func (r *Router) setupAddressRoutes(rg *gin.RouterGroup) {
	address := rg.Group("/address")
	{
		address.GET("/provinces", func(c *gin.Context) {
			// Return provinces
		})
		address.GET("/provinces/:id/districts", func(c *gin.Context) {
			// Return districts by province
		})
		address.GET("/districts/:id/wards", func(c *gin.Context) {
			// Return wards by district
		})
	}
}

func (r *Router) setupReportRoutes(rg *gin.RouterGroup) {
	reports := rg.Group("/reports")
	reports.Use(middleware.JWTAuth(&r.cfg.JWT, r.cache))
	reports.Use(middleware.RequirePermission("reports.view"))
	{
		reports.POST("/generate", func(c *gin.Context) {})
		reports.GET("/:id/status", func(c *gin.Context) {})
		reports.GET("/:id/download", func(c *gin.Context) {})

		// Pre-built reports
		reports.GET("/attendance", func(c *gin.Context) {})
		reports.GET("/payroll", func(c *gin.Context) {})
		reports.GET("/leave", func(c *gin.Context) {})
		reports.GET("/overtime", func(c *gin.Context) {})
		reports.GET("/employees", func(c *gin.Context) {})
	}
}

func (r *Router) setupNotificationRoutes(rg *gin.RouterGroup) {
	notifications := rg.Group("/notifications")
	notifications.Use(middleware.JWTAuth(&r.cfg.JWT, r.cache))
	{
		notifications.GET("", func(c *gin.Context) {})
		notifications.GET("/unread-count", func(c *gin.Context) {})
		notifications.PUT("/:id/read", func(c *gin.Context) {})
		notifications.PUT("/read-all", func(c *gin.Context) {})
	}
}

func (r *Router) healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "healthy",
		"service": "hr-management-api",
		"version": "1.0.0",
	})
}

func (r *Router) readinessCheck(c *gin.Context) {
	// Check database
	if err := r.db.HealthCheck(c.Request.Context()); err != nil {
		c.JSON(503, gin.H{"status": "unhealthy", "database": "down"})
		return
	}

	// Check redis
	if err := r.cache.HealthCheck(c.Request.Context()); err != nil {
		c.JSON(503, gin.H{"status": "unhealthy", "redis": "down"})
		return
	}

	c.JSON(200, gin.H{
		"status":   "ready",
		"database": "up",
		"redis":    "up",
	})
}
