package handlers

import (
	"database/sql"
	"errors"
	"financeTracker/middleware"
	"financeTracker/services"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type FinanceHandler struct {
	service *services.FinanceService
}

func NewFinanceHandler(service *services.FinanceService) *FinanceHandler {
	return &FinanceHandler{service: service}
}

// Struct لاستقبال بيانات المعاملة
// --- Accounts Handlers ---
type AccountReq struct {
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	Balance  int64   `json:"balance"` // Cents
	Currency string  `json:"currency"`
	Icon     *string `json:"icon"`
}

func (h *FinanceHandler) HandleCreateAccount(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	var req AccountReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "HND0001 Invalid request body"})
	}

	acc, err := h.service.CreateAccount(c.Context(), userID, req.Name, req.Type, req.Currency, req.Balance, req.Icon)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("[%s] %s", "HND0002", err.Error())})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": acc})
}

func (h *FinanceHandler) HandleUpdateAccount(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	var req AccountReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if req.Name == "" || req.Currency == "" || (req.Type != "cash" && req.Type != "debit_card" && req.Type != "credit_card" && req.Type != "savings") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name, currency, and a valid account type are required"})
	}

	acc, err := h.service.UpdateAccount(c.Context(), userID, c.Params("id"), req.Name, req.Type, req.Currency)
	if errors.Is(err, sql.ErrNoRows) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Account not found or unauthorized"})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": acc})
}

func (h *FinanceHandler) HandleGetAccounts(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	accounts, err := h.service.GetUserAccounts(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": accounts})
}

func (h *FinanceHandler) HandleDeleteAccount(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	accID := c.Params("id")
	err := h.service.DeleteAccount(c.Context(), userID, accID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Account deleted successfully"})
}

// --- Categories Handlers ---
type CategoryReq struct {
	Name  string  `json:"name"`
	Type  string  `json:"type"`
	Icon  *string `json:"icon"`
	Color *string `json:"color"`
}

func (h *FinanceHandler) HandleCreateCategory(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	var req CategoryReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	cat, err := h.service.CreateCategory(c.Context(), userID, req.Name, req.Type, req.Icon, req.Color)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": cat})
}

func (h *FinanceHandler) HandleUpdateCategory(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	var req CategoryReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if req.Name == "" || (req.Type != "income" && req.Type != "expense") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name and a valid category type are required"})
	}

	cat, err := h.service.UpdateCategory(c.Context(), userID, c.Params("id"), req.Name, req.Type, req.Icon, req.Color)
	if errors.Is(err, sql.ErrNoRows) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Category not found or unauthorized"})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": cat})
}

func (h *FinanceHandler) HandleGetCategories(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	catType := c.Query("type")
	var typePtr *string
	if catType != "" {
		typePtr = &catType
	}

	categories, err := h.service.GetUserCategories(c.Context(), userID, typePtr)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": categories})
}

func (h *FinanceHandler) HandleDeleteCategory(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	catID := c.Params("id")
	err := h.service.DeleteCategory(c.Context(), userID, catID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Category deleted successfully"})
}

// --- Transactions & Transfers Handlers ---
type TransactionReq struct {
	AccountID   string  `json:"accountId"`
	CategoryID  *string `json:"categoryId"`
	Type        string  `json:"type"`
	Amount      int64   `json:"amount"` // Cents
	Description string  `json:"description"`
	Date        string  `json:"date"`
}

func (h *FinanceHandler) HandleCreateTransaction(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	var req TransactionReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if req.AccountID == "" || req.Amount <= 0 || (req.Type != "income" && req.Type != "expense") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Account, positive amount, and valid transaction type are required"})
	}
	if req.CategoryID != nil && *req.CategoryID == "" {
		req.CategoryID = nil
	}

	transactionDate, err := parseFormDate(req.Date)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Date must use YYYY-MM-DD format"})
	}

	err = h.service.CreateTransaction(c.Context(), &userID, &req.AccountID, req.CategoryID, req.Type, req.Amount, req.Description, transactionDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Transaction recorded successfully"})
}

// GET /api/v1/transactions?type=expense&q=coffee&page=1&limit=10
func (h *FinanceHandler) HandleGetTransactions(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	txType := c.Query("type")
	if txType != "" && txType != "all" && txType != "income" && txType != "expense" && txType != "transfer" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid transaction type"})
	}
	var typePtr *string
	if txType != "" {
		typePtr = &txType
	}

	search := c.Query("q")
	var searchPtr *string
	if search != "" {
		searchPtr = &search
	}
	accountID := c.Query("accountId")
	var accountIDPtr *string
	if accountID != "" {
		accountIDPtr = &accountID
	}
	categoryID := c.Query("categoryId")
	var categoryIDPtr *string
	if categoryID != "" {
		categoryIDPtr = &categoryID
	}

	var dateFrom, dateTo *time.Time
	if from := c.Query("dateFrom"); from != "" {
		parsed, err := time.Parse("2006-01-02", from)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "dateFrom must use YYYY-MM-DD format"})
		}
		dateFrom = &parsed
	}
	if to := c.Query("dateTo"); to != "" {
		parsed, err := time.Parse("2006-01-02", to)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "dateTo must use YYYY-MM-DD format"})
		}
		endExclusive := parsed.AddDate(0, 0, 1)
		dateTo = &endExclusive
	}
	if dateFrom != nil && dateTo != nil && !dateFrom.Before(*dateTo) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "dateFrom must be on or before dateTo"})
	}

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Page must be a positive integer"})
	}
	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil || limit < 1 || limit > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Limit must be between 1 and 100"})
	}

	filters := services.TransactionFilters{
		Type:       typePtr,
		Search:     searchPtr,
		AccountID:  accountIDPtr,
		CategoryID: categoryIDPtr,
		DateFrom:   dateFrom,
		DateTo:     dateTo,
	}
	transactions, err := h.service.GetTransactions(c.Context(), userID, filters, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(transactions)
}

type TransferReq struct {
	FromAccountID string `json:"from_account_id"`
	ToAccountID   string `json:"to_account_id"`
	Amount        int64  `json:"amount"` // Cents
	Description   string `json:"description"`
	Date          string `json:"date"`
}

func parseFormDate(date string) (time.Time, error) {
	if date == "" {
		return time.Now().UTC(), nil
	}
	return time.Parse("2006-01-02", date)
}

func (h *FinanceHandler) HandleTransferFunds(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	var req TransferReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if req.FromAccountID == "" || req.ToAccountID == "" || req.Amount <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Source account, destination account, and positive amount are required"})
	}

	transferDate, err := parseFormDate(req.Date)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Date must use YYYY-MM-DD format"})
	}

	err = h.service.TransferFunds(c.Context(), userID, req.FromAccountID, req.ToAccountID, req.Amount, req.Description, transferDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Transfer completed successfully"})
}

func (h *FinanceHandler) HandleGetDashboardSummary(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string) // جلب الـ userID من الـ Auth Middleware

	monthParam := c.Query("month")
	month := time.Now().UTC()
	if monthParam != "" {
		parsedMonth, err := time.Parse("2006-01", monthParam)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Month must use YYYY-MM format",
			})
		}
		month = parsedMonth
	}

	summary, err := h.service.GetDashboardSummary(c.Context(), userID, month)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": summary,
	})
}

//0---------------

type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Currency string `json:"currency,omitempty"` // اختياري وقت التسجيل
}

// 1. معالجة التسجيل (Register)
func (h *FinanceHandler) HandleRegister(c *fiber.Ctx) error {
	var req AuthRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Email and password are required"})
	}

	user, err := h.service.RegisterUser(c.Context(), req.Email, req.Password, req.Currency)
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	}

	// توليد توكن JWT تلقائياً بعد التسجيل الناجح
	token, err := middleware.GenerateToken(user.ID, user.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User registered successfully",
		"token":   token,
		"data":    user,
	})
}

// 2. معالجة تسجيل الدخول (Login)
func (h *FinanceHandler) HandleLogin(c *fiber.Ctx) error {
	var req AuthRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	user, err := h.service.AuthenticateUser(c.Context(), req.Email, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	// توليد JWT Token
	token, err := middleware.GenerateToken(user.ID, user.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Login successful",
		"token":   token,
		"data":    user,
	})
}
