import { api } from "./client";
import type {
  Account,
  AuthResponse,
  Category,
  DashboardSummary,
  Page,
  Transaction,
  TransactionQuery,
  User,
} from "../types";

export const authApi = {
  login: async (email: string, password: string) =>
    (await api.post<AuthResponse>("/auth/login", { email, password })).data,
  register: async (name: string, email: string, password: string) =>
    (await api.post<AuthResponse>("/auth/register", { name, email, password }))
      .data,
};
export const dashboardApi = {
  summary: async (month: string) =>
    (
      await api.get<{ data: DashboardSummary }>("/dashboard/summary", {
        params: { month },
      })
    ).data.data,
};
export const accountsApi = {
  list: async () => (await api.get<Account[]>("/accounts")).data.data,
  create: async (
    payload: Pick<Account, "name" | "type" | "balance" | "currency">,
  ) => (await api.post<Account>("/accounts", payload)).data,
  update: async (
    id: string,
    payload: Pick<Account, "name" | "type" | "currency">,
  ) => (await api.put<Account>(`/accounts/${id}`, payload)).data,
  delete: async (id: string) => (await api.delete(`/accounts/${id}`)).data,
};
export const categoriesApi = {
  list: async () => (await api.get<Category[]>("/categories")).data.data,
  create: async (payload: Pick<Category, "name" | "type" | "color" | "icon">) =>
    (await api.post<Category>("/categories", payload)).data,
  update: async (
    id: string,
    payload: Pick<Category, "name" | "type" | "color" | "icon">,
  ) => (await api.put<Category>(`/categories/${id}`, payload)).data,
  delete: async (id: string) => (await api.delete(`/categories/${id}`)).data,
};
export const transactionsApi = {
  list: async (params: TransactionQuery) =>
    (await api.get<Page<Transaction>>("/transactions", { params })).data,
  create: async (payload: Record<string, unknown>) =>
    (await api.post<Transaction>("/transactions", payload)).data,
  transfer: async (payload: Record<string, unknown>) =>
    (await api.post<Transaction>("/transfers", payload)).data,
};
export const profileApi = {
  me: async () => (await api.get<User>("/auth/me")).data,
};
