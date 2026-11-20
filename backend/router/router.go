package router

import (
	"financeTracker/handlers"
	"financeTracker/middleware"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func NewRouter(app *fiber.App, financeHandler handlers.FinanceHandler) *http.ServeMux {
	mux := http.NewServeMux()

	authGroup := app.Group("/api/v1/auth")
	authGroup.Post("/register" /* handler تسجيل الدخول */)
	authGroup.Post("/login" /* handler تسجيل الدخول وإرجاع التوكن */)

	// مجموعة مسارات محمية بـ JWT Middleware
	api := app.Group("/api/v1", middleware.JWTProtected())

	// مسارات الحسابات
	api.Post("/accounts", financeHandler.HandleCreateAccount)
	api.Get("/accounts", financeHandler.HandleGetAccounts)
	api.Put("/accounts/:id", financeHandler.HandleUpdateAccount)
	api.Delete("/accounts/:id", financeHandler.HandleDeleteAccount)

	// مسارات المعاملات والتحويلات
	api.Post("/transactions", financeHandler.HandleCreateTransaction)
	api.Get("/transactions", financeHandler.HandleGetTransactions)
	api.Post("/transfers", financeHandler.HandleTransferFunds)

	// مسارات التصنيفات
	api.Post("/categories", financeHandler.HandleCreateCategory)
	api.Get("/categories", financeHandler.HandleGetCategories)
	api.Put("/categories/:id", financeHandler.HandleUpdateCategory)
	api.Delete("/categories/:id", financeHandler.HandleDeleteCategory)

	api.Get("/dashboard/summary", financeHandler.HandleGetDashboardSummary)
	return mux
}
