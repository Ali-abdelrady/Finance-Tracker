package main

import (
	"database/sql"
	"financeTracker/handlers"
	"financeTracker/middleware"
	"financeTracker/services"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	_ "github.com/lib/pq" // PostgreSQL driver
)

func main() {
	// قراءة اتصال قاعدة البيانات من متغيرات البيئة أو استخدام رابط افتراضي
	dbConnStr := os.Getenv("DATABASE_URL")
	if dbConnStr == "" {
		dbConnStr = "postgres://postgres:admin@localhost:5432/finance_db?sslmode=disable"
	}

	// الاتصال بقاعدة بيانات PostgreSQL
	db, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Database connection is dead: %v", err)
	}
	log.Println("Connected to PostgreSQL successfully!")

	// حقن التبعيات (Dependency Injection)
	financeService := services.NewFinanceService(db)
	financeHandler := handlers.NewFinanceHandler(financeService)

	// إعداد تطبيق Fiber
	app := fiber.New(fiber.Config{
		AppName: "Personal Finance API v1.0",
	})

	// إعداد الـ Middlewares العامة
	app.Use(logger.New()) // لتسجيل الطلبات في الـ Terminal
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE",
	}))

	// مسار تجريبي للتأكد أن السيرفر يعمل
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Welcome to Personal Finance API - Running Smoothly!"})
	})

	// تجميع مسارات الـ API
	apiV1 := app.Group("/api/v1")

	// --- 1. مسارات عامة (بدون حماية) مثل التسجيل وتسجيل الدخول ---
	authGroup := apiV1.Group("/auth")
	authGroup.Post("/register", financeHandler.HandleRegister)
	authGroup.Post("/login", financeHandler.HandleLogin)

	// --- 2. مسارات محمية بـ JWT Authentication Middleware ---
	protected := apiV1.Group("", middleware.JWTProtected())

	// أ. إدارة الحسابات (Accounts)
	// أ. إدارة الحسابات (Accounts CRUD)
	protected.Post("/accounts", financeHandler.HandleCreateAccount)
	protected.Get("/accounts", financeHandler.HandleGetAccounts)
	protected.Put("/accounts/:id", financeHandler.HandleUpdateAccount)
	protected.Delete("/accounts/:id", financeHandler.HandleDeleteAccount)

	// ب. إدارة التصنيفات (Categories CRUD)
	protected.Post("/categories", financeHandler.HandleCreateCategory)
	protected.Get("/categories", financeHandler.HandleGetCategories)
	protected.Put("/categories/:id", financeHandler.HandleUpdateCategory)
	protected.Delete("/categories/:id", financeHandler.HandleDeleteCategory)

	// ج. المعاملات المالية والتحويلات (Transactions & Transfers)
	protected.Post("/transactions", financeHandler.HandleCreateTransaction)
	protected.Get("/transactions", financeHandler.HandleGetTransactions) // تدعم ?type=&q=&page=&limit=
	protected.Post("/transfers", financeHandler.HandleTransferFunds)

	// د. لوحة التحليلات والملخص (Dashboard Summary)
	protected.Get("/dashboard/summary", financeHandler.HandleGetDashboardSummary)

	// تشغيل السيرفر على البورت 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server is running on port %s...", port)
	log.Fatal(app.Listen(":" + port))
}
