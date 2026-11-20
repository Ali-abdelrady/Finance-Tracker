package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type FinanceService struct {
	db *sql.DB
}

func NewFinanceService(db *sql.DB) *FinanceService {
	return &FinanceService{db: db}
}

// 1. إضافة معاملة مالية (مصروف أو دخل) مع تحديث رصيد الحساب تلقائياً
type Account struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	Name      string  `json:"name"`
	Type      string  `json:"type"`    // cash, debit_card, credit_card, savings
	Balance   int64   `json:"balance"` // Minor currency units (10050 means 100.50)
	Currency  string  `json:"currency"`
	Icon      *string `json:"icon"`
	CreatedAt string  `json:"created_at"`
}

func (s *FinanceService) CreateAccount(ctx context.Context, userID, name, accType, currency string, balance int64, icon *string) (*Account, error) {
	query := `
		INSERT INTO accounts (user_id, name, type, balance, currency, icon)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, name, type, balance::bigint, currency, icon, created_at
	`
	var acc Account
	err := s.db.QueryRowContext(ctx, query, userID, name, accType, balance, currency, icon).Scan(
		&acc.ID, &acc.UserID, &acc.Name, &acc.Type, &acc.Balance, &acc.Currency, &acc.Icon, &acc.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}
	return &acc, nil
}

func (s *FinanceService) GetUserAccounts(ctx context.Context, userID string) ([]Account, error) {
	query := `SELECT id, user_id, name, type, balance::bigint, currency, icon, created_at FROM accounts WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := make([]Account, 0)
	for rows.Next() {
		var acc Account
		if err := rows.Scan(&acc.ID, &acc.UserID, &acc.Name, &acc.Type, &acc.Balance, &acc.Currency, &acc.Icon, &acc.CreatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}
	return accounts, nil
}

func (s *FinanceService) UpdateAccount(ctx context.Context, userID, accountID, name, accType, currency string) (*Account, error) {
	query := `
		UPDATE accounts
		SET name = $1, type = $2, currency = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4 AND user_id = $5
		RETURNING id, user_id, name, type, balance::bigint, currency, icon, created_at
	`
	var acc Account
	err := s.db.QueryRowContext(ctx, query, name, accType, currency, accountID, userID).Scan(
		&acc.ID, &acc.UserID, &acc.Name, &acc.Type, &acc.Balance, &acc.Currency, &acc.Icon, &acc.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &acc, nil
}

func (s *FinanceService) DeleteAccount(ctx context.Context, userID, accountID string) error {
	// منع الحذف لو في معاملات مرتبطة (لتجنب Foreign Key violation)
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM transactions WHERE account_id = $1`, accountID).Scan(&count)
	if err == nil && count > 0 {
		return errors.New("cannot delete account with active transactions. Archive it instead")
	}

	query := `DELETE FROM accounts WHERE id = $1 AND user_id = $2`
	result, err := s.db.ExecContext(ctx, query, accountID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("account not found or unauthorized")
	}
	return nil
}

// --- 2. Categories CRUD & Updates ---

type Category struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	Name      string  `json:"name"`
	Type      string  `json:"type"` // income or expense
	Icon      *string `json:"icon"`
	Color     *string `json:"color"`
	CreatedAt string  `json:"created_at"`
}

func (s *FinanceService) CreateCategory(ctx context.Context, userID, name, catType string, icon, color *string) (*Category, error) {
	query := `
		INSERT INTO categories (user_id, name, type, icon, color)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, name, type, icon, color, created_at
	`
	var cat Category
	err := s.db.QueryRowContext(ctx, query, userID, name, catType, icon, color).Scan(
		&cat.ID, &cat.UserID, &cat.Name, &cat.Type, &cat.Icon, &cat.Color, &cat.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &cat, nil
}

func (s *FinanceService) UpdateCategory(ctx context.Context, userID, categoryID, name, catType string, icon, color *string) (*Category, error) {
	query := `
		UPDATE categories
		SET name = $1, type = $2, icon = $3, color = $4
		WHERE id = $5 AND user_id = $6
		RETURNING id, user_id, name, type, icon, color, created_at
	`
	var cat Category
	err := s.db.QueryRowContext(ctx, query, name, catType, icon, color, categoryID, userID).Scan(
		&cat.ID, &cat.UserID, &cat.Name, &cat.Type, &cat.Icon, &cat.Color, &cat.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &cat, nil
}

func (s *FinanceService) GetUserCategories(ctx context.Context, userID string, catType *string) ([]Category, error) {
	query := `SELECT id, user_id, name, type, icon, color, created_at FROM categories WHERE user_id = $1`
	args := []interface{}{userID}

	if catType != nil && (*catType == "income" || *catType == "expense") {
		query += ` AND type = $2`
		args = append(args, *catType)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]Category, 0)
	for rows.Next() {
		var cat Category
		if err := rows.Scan(&cat.ID, &cat.UserID, &cat.Name, &cat.Type, &cat.Icon, &cat.Color, &cat.CreatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}
	return categories, nil
}

func (s *FinanceService) DeleteCategory(ctx context.Context, userID, categoryID string) error {
	query := `DELETE FROM categories WHERE id = $1 AND user_id = $2`
	result, err := s.db.ExecContext(ctx, query, categoryID, userID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("category not found or unauthorized")
	}
	return nil
}

// --- 3. Transactions CRUD (With Filtering & Pagination) ---

type Transaction struct {
	ID              string    `json:"id"`
	UserID          string    `json:"userId"`
	AccountID       string    `json:"accountId"`
	AccountName     string    `json:"accountName"`
	ToAccountID     *string   `json:"toAccountId,omitempty"`
	CategoryID      *string   `json:"categoryId"`
	CategoryName    *string   `json:"categoryName,omitempty"`
	ToAccountName   *string   `json:"toAccountName,omitempty"`
	Type            string    `json:"type"`   // income, expense
	Amount          int64     `json:"amount"` // Cents
	Description     *string   `json:"description"`
	TransactionDate time.Time `json:"transactionDate"`
	CreatedAt       string    `json:"created_at"`
}

type TransactionPage struct {
	Data       []Transaction      `json:"data"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	Total      int                `json:"total"`
	TotalPages int                `json:"totalPages"`
	Summary    TransactionSummary `json:"summary"`
}

type TransactionSummary struct {
	ExpenseTotal  int64 `json:"expenseTotal"`
	IncomeTotal   int64 `json:"incomeTotal"`
	TransferTotal int64 `json:"transferTotal"`
}

type TransactionFilters struct {
	Type       *string
	Search     *string
	AccountID  *string
	CategoryID *string
	DateFrom   *time.Time
	DateTo     *time.Time
}

func (s *FinanceService) CreateTransaction(ctx context.Context, userID, accountID, categoryID *string, txType string, amount int64, description string, transactionDate time.Time) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queryTx := `
		INSERT INTO transactions (user_id, account_id, category_id, type, amount, description, transaction_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = tx.ExecContext(ctx, queryTx, userID, accountID, categoryID, txType, amount, description, transactionDate)
	if err != nil {
		return fmt.Errorf("SVCTX0001 %v", err)
	}

	balanceChange := "balance - $1"
	if txType == "income" {
		balanceChange = "balance + $1"
	}

	queryUpdateAcc := fmt.Sprintf("UPDATE accounts SET balance = %s, updated_at = CURRENT_TIMESTAMP WHERE id = $2 AND user_id = $3", balanceChange)
	_, err = tx.ExecContext(ctx, queryUpdateAcc, amount, accountID, userID)
	if err != nil {
		return fmt.Errorf("SVCTX0002 %v", err)
	}

	return tx.Commit()
}

// جلب المعاملات والتحويلات مع دعم الـ Pagination والفلترة بالبحث والنوع والحساب والتصنيف والتاريخ
func (s *FinanceService) GetTransactions(ctx context.Context, userID string, filters TransactionFilters, page, limit int) (*TransactionPage, error) {
	offset := (page - 1) * limit
	const activityQuery = `
		SELECT t.id, t.user_id, t.account_id, a.name AS account_name,
			NULL::uuid AS to_account_id, t.category_id, c.name AS category_name, NULL::text AS to_account_name,
			t.type::text AS type, t.amount::bigint, t.description,
			t.transaction_date AS transaction_date, t.created_at
		FROM transactions t
		LEFT JOIN categories c ON t.category_id = c.id
		LEFT JOIN accounts a ON t.account_id = a.id
		UNION ALL
		SELECT tr.id, tr.user_id, tr.from_account_id, source.name,
			tr.to_account_id, NULL::uuid, NULL::text, destination.name, 'transfer'::text,
			tr.amount::bigint, tr.description, tr.transfer_date, tr.created_at
		FROM transfers tr
		LEFT JOIN accounts source ON tr.from_account_id = source.id
		LEFT JOIN accounts destination ON tr.to_account_id = destination.id
	`
	filter := ` WHERE user_id = $1`
	args := []interface{}{userID}
	argIdx := 2

	if filters.Type != nil && *filters.Type != "" && *filters.Type != "all" {
		filter += fmt.Sprintf(" AND type = $%d", argIdx)
		args = append(args, *filters.Type)
		argIdx++
	}

	if filters.Search != nil && *filters.Search != "" {
		filter += fmt.Sprintf(` AND (
			COALESCE(description, '') ILIKE $%d OR
			COALESCE(account_name, '') ILIKE $%d OR
			COALESCE(category_name, '') ILIKE $%d OR
			COALESCE(to_account_name, '') ILIKE $%d
		)`, argIdx, argIdx, argIdx, argIdx)
		args = append(args, "%"+*filters.Search+"%")
		argIdx++
	}

	if filters.AccountID != nil && *filters.AccountID != "" {
		filter += fmt.Sprintf(" AND (account_id::text = $%d OR to_account_id::text = $%d)", argIdx, argIdx)
		args = append(args, *filters.AccountID)
		argIdx++
	}
	if filters.CategoryID != nil && *filters.CategoryID != "" {
		filter += fmt.Sprintf(" AND category_id::text = $%d", argIdx)
		args = append(args, *filters.CategoryID)
		argIdx++
	}
	if filters.DateFrom != nil {
		filter += fmt.Sprintf(" AND transaction_date >= $%d", argIdx)
		args = append(args, *filters.DateFrom)
		argIdx++
	}
	if filters.DateTo != nil {
		filter += fmt.Sprintf(" AND transaction_date < $%d", argIdx)
		args = append(args, *filters.DateTo)
		argIdx++
	}

	var total int
	summaryQuery := `SELECT COUNT(*),
		COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0)::bigint,
		COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0)::bigint,
		COALESCE(SUM(CASE WHEN type = 'transfer' THEN amount ELSE 0 END), 0)::bigint
		FROM (` + activityQuery + `) activity` + filter
	var summary TransactionSummary
	if err := s.db.QueryRowContext(ctx, summaryQuery, args...).Scan(
		&total, &summary.ExpenseTotal, &summary.IncomeTotal, &summary.TransferTotal,
	); err != nil {
		return nil, fmt.Errorf("SVCTX0100 %v", err)
	}

	query := `SELECT id, user_id, account_id, account_name, to_account_id, category_id, category_name,
		to_account_name, type, amount, description, transaction_date, created_at
		FROM (` + activityQuery + `) activity` + filter +
		fmt.Sprintf(" ORDER BY transaction_date DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	queryArgs := append(append([]interface{}{}, args...), limit, offset)

	rows, err := s.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("SVCTX0101 %v", err)
	}
	defer rows.Close()

	transactions := make([]Transaction, 0)
	for rows.Next() {
		var t Transaction
		if err := rows.Scan(&t.ID, &t.UserID, &t.AccountID, &t.AccountName, &t.ToAccountID, &t.CategoryID, &t.CategoryName, &t.ToAccountName, &t.Type, &t.Amount, &t.Description, &t.TransactionDate, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("SVCTX0102 %v", err)
		}
		transactions = append(transactions, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("SVCTX0103 %v", err)
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}
	return &TransactionPage{
		Data:       transactions,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		Summary:    summary,
	}, nil
}

// --- 4. Transfers Logic ---

func (s *FinanceService) TransferFunds(ctx context.Context, userID, fromAccountID, toAccountID string, amount int64, description string, transferDate time.Time) error {
	if fromAccountID == toAccountID {
		return errors.New("cannot transfer to the same account")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queryTransfer := `
		INSERT INTO transfers (user_id, from_account_id, to_account_id, amount, description, transfer_date)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = tx.ExecContext(ctx, queryTransfer, userID, fromAccountID, toAccountID, amount, description, transferDate)
	if err != nil {
		return err
	}

	// خصم من الحساب المرسل
	_, err = tx.ExecContext(ctx, `UPDATE accounts SET balance = balance - $1 WHERE id = $2 AND user_id = $3`, amount, fromAccountID, userID)
	if err != nil {
		return err
	}

	// إضافة للحساب المستقبل
	_, err = tx.ExecContext(ctx, `UPDATE accounts SET balance = balance + $1 WHERE id = $2 AND user_id = $3`, amount, toAccountID, userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

type DashboardSummary struct {
	NetWorth          float64           `json:"netWorth"`
	Month             string            `json:"month"`
	MonthlyIncome     float64           `json:"monthlyIncome"`  // renamed from TotalIncome
	MonthlyExpense    float64           `json:"monthlyExpense"` // renamed from TotalExpense
	CategoryBreakdown []CategoryExpense `json:"categoryBreakdown"`
	Cashflow          []CashflowEntry   `json:"cashflow"` // added
	Accounts          []Account         `json:"accounts"` // added
}

type CategoryExpense struct {
	CategoryName string  `json:"categoryName"` // maps to 'name' in TS
	Color        *string `json:"color"`
	TotalAmount  float64 `json:"amount"` // JSON tag must be 'amount' to match TS dataKey
	Percentage   float64 `json:"percentage"`
}

type CashflowEntry struct {
	Month   string  `json:"month"`
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
}

func (s *FinanceService) GetDashboardSummary(ctx context.Context, userID string, month time.Time) (*DashboardSummary, error) {
	// 1. Calculate Net Worth
	var netWorth float64
	queryNetWorth := `SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE user_id = $1`
	if err := s.db.QueryRowContext(ctx, queryNetWorth, userID).Scan(&netWorth); err != nil {
		return nil, fmt.Errorf("failed to calculate net worth: %w", err)
	}

	// 2. Fetch User Accounts for the strip
	accountsQuery := `SELECT id, user_id, name, type, balance::bigint, currency, icon, created_at FROM accounts WHERE user_id = $1 ORDER BY created_at DESC`
	rowsAcc, err := s.db.QueryContext(ctx, accountsQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch accounts: %w", err)
	}
	defer rowsAcc.Close()

	accounts := make([]Account, 0)
	for rowsAcc.Next() {
		var acc Account
		if err := rowsAcc.Scan(&acc.ID, &acc.UserID, &acc.Name, &acc.Type, &acc.Balance, &acc.Currency, &acc.Icon, &acc.CreatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}

	// 3. Current month totals
	startOfMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	var totalIncome, totalExpense float64
	queryTotals := `
        SELECT 
            COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0),
            COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0)
        FROM transactions
        WHERE user_id = $1 AND transaction_date >= $2 AND transaction_date < $3
    `
	err = s.db.QueryRowContext(ctx, queryTotals, userID, startOfMonth, endOfMonth).Scan(&totalIncome, &totalExpense)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate monthly totals: %w", err)
	}

	// 4. Category Breakdown
	queryBreakdown := `
        SELECT 
            c.name as category_name,
            c.color,
            SUM(t.amount) as total_amount
        FROM transactions t
        JOIN categories c ON t.category_id = c.id
        WHERE t.user_id = $1 AND t.type = 'expense' AND t.transaction_date >= $2 AND t.transaction_date < $3
        GROUP BY c.name, c.color
        ORDER BY total_amount DESC
    `
	rowsBreakdown, err := s.db.QueryContext(ctx, queryBreakdown, userID, startOfMonth, endOfMonth)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch category breakdown: %w", err)
	}
	defer rowsBreakdown.Close()

	breakdown := make([]CategoryExpense, 0)
	for rowsBreakdown.Next() {
		var item CategoryExpense
		if err := rowsBreakdown.Scan(&item.CategoryName, &item.Color, &item.TotalAmount); err != nil {
			return nil, err
		}
		if totalExpense > 0 {
			item.Percentage = (item.TotalAmount / totalExpense) * 100
		}
		breakdown = append(breakdown, item)
	}
	// 5. Generate cashflow for the selected month and five preceding months.
	queryCashflow := `
        WITH months AS (
            SELECT generate_series(
                date_trunc('month', $2::date) - INTERVAL '5 months',
                date_trunc('month', $2::date),
                '1 month'::interval
            ) AS month_start
        )
        SELECT 
            TO_CHAR(m.month_start, 'Mon YYYY') AS month_name,
            COALESCE(SUM(CASE WHEN t.type = 'income' THEN t.amount ELSE 0 END), 0) AS income,
            COALESCE(SUM(CASE WHEN t.type = 'expense' THEN t.amount ELSE 0 END), 0) AS expense
        FROM months m
        LEFT JOIN transactions t ON t.user_id = $1 
            AND t.transaction_date >= m.month_start
            AND t.transaction_date < m.month_start + INTERVAL '1 month'
        GROUP BY m.month_start
        ORDER BY m.month_start ASC;
    `

	rowsCashflow, err := s.db.QueryContext(ctx, queryCashflow, userID, startOfMonth)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch cashflow trend: %w", err)
	}
	defer rowsCashflow.Close()

	cashflow := make([]CashflowEntry, 0)
	for rowsCashflow.Next() {
		var entry CashflowEntry
		var rawIncome, rawExpense float64

		if err := rowsCashflow.Scan(&entry.Month, &rawIncome, &rawExpense); err != nil {
			return nil, err
		}

		// If storing as cents in DB, convert back to float/dollars if your frontend expects it,
		// or keep as-is depending on how formatCurrency handles it.
		entry.Income = rawIncome
		entry.Expense = rawExpense
		cashflow = append(cashflow, entry)
	}

	return &DashboardSummary{
		NetWorth:          netWorth,
		Month:             startOfMonth.Format("2006-01"),
		MonthlyIncome:     totalIncome,
		MonthlyExpense:    totalExpense,
		CategoryBreakdown: breakdown,
		Cashflow:          cashflow,
		Accounts:          accounts,
	}, nil
}

type User struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"` // منع إرجاع الـ Hash في الـ JSON responses
	Currency     string `json:"currency"`
	CreatedAt    string `json:"created_at"`
}

// 1. تسجيل مستخدم جديد
func (s *FinanceService) RegisterUser(ctx context.Context, email, password, currency string) (*User, error) {
	// تشفير كلمة المرور
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	if currency == "" {
		currency = "EGP"
	}

	query := `
		INSERT INTO users (email, password_hash, currency)
		VALUES ($1, $2, $3)
		RETURNING id, email, currency, created_at
	`
	var user User
	err = s.db.QueryRowContext(ctx, query, email, string(hashedPassword), currency).Scan(
		&user.ID, &user.Email, &user.Currency, &user.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("email might already be registered: %w", err)
	}

	return &user, nil
}

// 2. التحقق من بيانات تسجيل الدخول
func (s *FinanceService) AuthenticateUser(ctx context.Context, email, password string) (*User, error) {
	query := `SELECT id, email, password_hash, currency, created_at FROM users WHERE email = $1`

	var user User
	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Currency, &user.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("invalid email or password")
	}
	if err != nil {
		return nil, err
	}

	// مقارنة كلمة المرور المدخلة بالمشفرة
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	return &user, nil
}
