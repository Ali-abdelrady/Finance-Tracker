export type ID = string;
export type TransactionType = "income" | "expense" | "transfer";
export interface User {
  id: ID;
  name: string;
  email: string;
}
export interface Account {
  id: ID;
  name: string;
  type: "cash" | "debit_card" | "credit_card" | "savings";
  balance: number;
  currency: string;
  color?: string;
  institution?: string;
}
export interface Category {
  id: ID;
  name: string;
  type: "income" | "expense";
  color: string;
  icon?: string;
}
export interface Transaction {
  id: ID;
  type: TransactionType;
  amount: number;
  description?: string | null;
  transactionDate: string;
  accountId: ID;
  accountName?: string;
  categoryId?: ID;
  categoryName?: string;
  toAccountId?: ID;
  toAccountName?: string;
}
export interface DashboardSummary {
  netWorth: number;
  month: string;
  monthlyIncome: number;
  monthlyExpense: number;
  savingsRate?: number;
  categoryBreakdown: {
    categoryName: string;
    amount: number;
    color: string | null;
    percentage: number;
  }[];
  cashflow: { month: string; income: number; expense: number }[];
  accounts: Account[];
}
export interface Page<T> {
  data: T[];
  page: number;
  limit: number;
  total: number;
  totalPages: number;
  summary: {
    expenseTotal: number;
    incomeTotal: number;
    transferTotal: number;
  };
}
export interface AuthResponse {
  token: string;
  user: User;
}
export interface TransactionQuery {
  type?: "income" | "expense" | "transfer" | "all";
  q?: string;
  accountId?: string;
  categoryId?: string;
  dateFrom?: string;
  dateTo?: string;
  page: number;
  limit: number;
}
