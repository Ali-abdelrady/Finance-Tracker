import { useCallback, useEffect, useState } from "react";
import {
  Search,
  Plus,
  ChevronLeft,
  ChevronRight,
  ArrowDownLeft,
  ArrowUpRight,
  ArrowLeftRight,
  Filter,
  LayoutGrid,
  List,
} from "lucide-react";
import { accountsApi, categoriesApi, transactionsApi } from "../api/services";
import type {
  Account,
  Category,
  Page,
  Transaction,
  TransactionQuery,
} from "../types";
import { formatCurrency, toMinorUnits } from "../lib/currency";
import { getErrorMessage } from "../api/client";
import Modal from "../components/Modal";
const empty: Page<Transaction> = {
  data: [],
  page: 1,
  limit: 10,
  total: 0,
  totalPages: 0,
  summary: { expenseTotal: 0, incomeTotal: 0, transferTotal: 0 },
};
const localDate = () => {
  const now = new Date();
  const year = now.getFullYear();
  const month = String(now.getMonth() + 1).padStart(2, "0");
  const day = String(now.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
};
export default function Transactions() {
  const [result, setResult] = useState(empty);
  const [query, setQuery] = useState("");
  const [type, setType] = useState<TransactionQuery["type"]>("all");
  const [accountFilter, setAccountFilter] = useState("");
  const [categoryFilter, setCategoryFilter] = useState("");
  const [dateFrom, setDateFrom] = useState("");
  const [dateTo, setDateTo] = useState("");
  const [view, setView] = useState<"table" | "cards">("table");
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(10);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [modal, setModal] = useState(false);
  const [tab, setTab] = useState<"expense" | "income" | "transfer">("expense");
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [amount, setAmount] = useState("");
  const [description, setDescription] = useState("");
  const [accountId, setAccountId] = useState("");
  const [toAccountId, setToAccountId] = useState("");
  const [categoryId, setCategoryId] = useState("");
  const [date, setDate] = useState(localDate);
  const [saving, setSaving] = useState(false);
  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const res = await transactionsApi.list({
        type,
        q: query.trim() || undefined,
        accountId: accountFilter || undefined,
        categoryId: categoryFilter || undefined,
        dateFrom: dateFrom || undefined,
        dateTo: dateTo || undefined,
        page,
        limit,
      });
      setResult(res);
    } catch (e) {
      setError(getErrorMessage(e));
    } finally {
      setLoading(false);
    }
  }, [type, query, accountFilter, categoryFilter, dateFrom, dateTo, page, limit]);

  useEffect(() => {
    const t = setTimeout(() => {
      load();
    }, 250);
    return () => clearTimeout(t);
  }, [load]);

  useEffect(() => {
    Promise.all([accountsApi.list(), categoriesApi.list()])
      .then(([loadedAccounts, loadedCategories]) => {
        setAccounts(loadedAccounts);
        setCategories(loadedCategories);
      })
      .catch((e) => setError(getErrorMessage(e)));
  }, []);

  async function openModal() {
    setError("");
    setDate(localDate());
    try {
      const [a, c] = await Promise.all([
        accountsApi.list(),
        categoriesApi.list(),
      ]);
      setAccounts(a);
      setCategories(c);
      setAccountId(a[0]?.id || "");
      setToAccountId(a[1]?.id || "");
      setCategoryId(c.find((x: Category) => x.type === tab)?.id || "");
      setModal(true);
    } catch (e) {
      setError(getErrorMessage(e));
    }
  }

  async function save(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    setError("");
    try {
      const cents = toMinorUnits(amount);
      if (cents <= 0) throw new Error("Amount must be greater than zero.");
      if (tab === "transfer") {
        if (accountId === toAccountId)
          throw new Error("Choose two different accounts.");
        await transactionsApi.transfer({
          amount: cents,
          from_account_id: accountId,
          to_account_id: toAccountId,
          date,
          description,
        });
      } else {
        await transactionsApi.create({
          type: tab,
          amount: cents,
          accountId,
          categoryId,
          date,
          description,
        });
      }
      setModal(false);
      setAmount("");
      setDescription("");
      await load();
    } catch (e) {
      setError(getErrorMessage(e));
    } finally {
      setSaving(false);
    }
  }

  const rows = result.data || [];
  const currentPage = result.page || page;
  const resetToFirstPage = () => setPage(1);
  const transactionIcon = (txType: Transaction["type"]) =>
    txType === "income" ? (
      <ArrowDownLeft size={17} />
    ) : txType === "transfer" ? (
      <ArrowLeftRight size={17} />
    ) : (
      <ArrowUpRight size={17} />
    );
  const transactionTitle = (tx: Transaction) =>
    tx.description || tx.categoryName || "Transaction";
  const displayAmount = (tx: Transaction) =>
    `${tx.type === "income" ? "+" : tx.type === "expense" ? "−" : ""}${formatCurrency(tx.amount)}`;

  return (
    <div className="content-stack">
      <div className="welcome-row">
        <div>
          <span className="eyebrow">ACTIVITY</span>
          <h1>
            Transactions<span className="accent-dot">.</span>
          </h1>
          <p className="muted">Every movement, all in one place.</p>
        </div>
        <button className="primary-btn" onClick={openModal}>
          <Plus size={17} /> Add transaction
        </button>
      </div>
      <section className="panel transactions-panel">
        <div className="table-toolbar">
          <div className="search-box">
            <Search size={17} />
            <input
              value={query}
              onChange={(e) => {
                setQuery(e.target.value);
                setPage(1);
              }}
              placeholder="Search transactions…"
            />
          </div>
          <div className="filter-wrap">
            <Filter size={16} />
            <select
              value={type}
              onChange={(e) => {
                setType(e.target.value as TransactionQuery["type"]);
                resetToFirstPage();
              }}
            >
              <option value="all">All activity</option>
              <option value="expense">Expenses</option>
              <option value="income">Income</option>
              <option value="transfer">Transfers</option>
            </select>
          </div>
          <div className="view-toggle" aria-label="Transaction display mode">
            <button
              type="button"
              className={view === "table" ? "selected" : ""}
              onClick={() => setView("table")}
              aria-label="Show transactions as a table"
              title="Table view"
            >
              <List size={17} />
            </button>
            <button
              type="button"
              className={view === "cards" ? "selected" : ""}
              onClick={() => setView("cards")}
              aria-label="Show transactions as cards"
              title="Card view"
            >
              <LayoutGrid size={17} />
            </button>
          </div>
        </div>
        <div className="transaction-filters">
          <label>
            <span>Account</span>
            <select
              value={accountFilter}
              onChange={(e) => {
                setAccountFilter(e.target.value);
                resetToFirstPage();
              }}
            >
              <option value="">All accounts</option>
              {accounts.map((account) => (
                <option key={account.id} value={account.id}>
                  {account.name}
                </option>
              ))}
            </select>
          </label>
          <label>
            <span>Category</span>
            <select
              value={categoryFilter}
              onChange={(e) => {
                setCategoryFilter(e.target.value);
                resetToFirstPage();
              }}
            >
              <option value="">All categories</option>
              {categories.map((category) => (
                <option key={category.id} value={category.id}>
                  {category.name}
                </option>
              ))}
            </select>
          </label>
          <label>
            <span>From</span>
            <input
              type="date"
              value={dateFrom}
              max={dateTo || undefined}
              onChange={(e) => {
                setDateFrom(e.target.value);
                resetToFirstPage();
              }}
            />
          </label>
          <label>
            <span>To</span>
            <input
              type="date"
              value={dateTo}
              min={dateFrom || undefined}
              onChange={(e) => {
                setDateTo(e.target.value);
                resetToFirstPage();
              }}
            />
          </label>
          {(accountFilter || categoryFilter || dateFrom || dateTo) && (
            <button
              type="button"
              className="clear-filters-btn"
              onClick={() => {
                setAccountFilter("");
                setCategoryFilter("");
                setDateFrom("");
                setDateTo("");
                resetToFirstPage();
              }}
            >
              Clear filters
            </button>
          )}
        </div>
        <div className="transaction-summary">
          <article className="transaction-summary-card expense">
            <span>Spent in filtered results</span>
            <b>{formatCurrency(result.summary.expenseTotal)}</b>
          </article>
          <article className="transaction-summary-card income">
            <span>Income in filtered results</span>
            <b>{formatCurrency(result.summary.incomeTotal)}</b>
          </article>
          <article className="transaction-summary-card transfer">
            <span>Transfers in filtered results</span>
            <b>{formatCurrency(result.summary.transferTotal)}</b>
          </article>
        </div>
        {error && <div className="error-box">{error}</div>}
        {view === "table" ? (
          <div className="table-scroll">
            <table className="data-table">
              <thead>
                <tr>
                  <th>TRANSACTION</th>
                  <th>CATEGORY</th>
                  <th>ACCOUNT</th>
                  <th>DATE</th>
                  <th className="align-right">AMOUNT</th>
                </tr>
              </thead>
              <tbody>
                {loading ? (
                  <tr>
                    <td colSpan={5} className="table-empty">
                      Loading transactions…
                    </td>
                  </tr>
                ) : rows.length ? (
                  rows.map((tx) => (
                    <tr key={tx.id}>
                      <td>
                        <div className="tx-title">
                          <span className={`tx-icon ${tx.type}`}>
                            {transactionIcon(tx.type)}
                          </span>
                          <div>
                            <b>{transactionTitle(tx)}</b>
                            <small>{tx.type}</small>
                          </div>
                        </div>
                      </td>
                      <td>{tx.categoryName || "—"}</td>
                      <td>
                        {tx.accountName || "—"}
                        {tx.toAccountName ? ` → ${tx.toAccountName}` : ""}
                      </td>
                      <td>{new Date(tx.transactionDate).toLocaleDateString()}</td>
                      <td className={`align-right amount-${tx.type}`}>
                        {displayAmount(tx)}
                      </td>
                    </tr>
                  ))
                ) : (
                  <tr>
                    <td colSpan={5} className="table-empty">
                      No transactions found. Try changing your filters or add one.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        ) : loading ? (
          <div className="table-empty">Loading transactions…</div>
        ) : rows.length ? (
          <div className="transaction-card-grid">
            {rows.map((tx) => (
              <article className="transaction-card" key={tx.id}>
                <div className="transaction-card-top">
                  <div className="tx-title">
                    <span className={`tx-icon ${tx.type}`}>
                      {transactionIcon(tx.type)}
                    </span>
                    <div>
                      <b>{transactionTitle(tx)}</b>
                      <small>{tx.type}</small>
                    </div>
                  </div>
                  <b className={`amount-${tx.type}`}>{displayAmount(tx)}</b>
                </div>
                <div className="transaction-card-details">
                  <span>{tx.categoryName || "Uncategorized"}</span>
                  <span>
                    {tx.accountName || "—"}
                    {tx.toAccountName ? ` → ${tx.toAccountName}` : ""}
                  </span>
                  <time>{new Date(tx.transactionDate).toLocaleDateString()}</time>
                </div>
              </article>
            ))}
          </div>
        ) : (
          <div className="table-empty">
            No transactions found. Try changing your filters or add one.
          </div>
        )}
        <div className="pagination">
          <span>
            Showing{" "}
            <b>
              {rows.length ? (currentPage - 1) * result.limit + 1 : 0}–
              {(currentPage - 1) * result.limit + rows.length}
            </b>{" "}
            of <b>{result.total}</b>
          </span>
          <div>
            <label className="page-size-label">
              Rows
              <select
                value={limit}
                onChange={(e) => {
                  setLimit(Number(e.target.value));
                  setPage(1);
                }}
                aria-label="Transactions per page"
              >
                <option value={10}>10</option>
                <option value={25}>25</option>
                <option value={50}>50</option>
              </select>
            </label>
            <button
              className="page-btn"
              disabled={currentPage <= 1 || loading}
              onClick={() => setPage((p) => p - 1)}
            >
              <ChevronLeft size={17} />
            </button>
            <span>
              Page {currentPage} of {Math.max(result.totalPages, 1)}
            </span>
            <button
              className="page-btn"
              disabled={currentPage >= result.totalPages || loading}
              onClick={() => setPage((p) => p + 1)}
            >
              <ChevronRight size={17} />
            </button>
          </div>
        </div>
      </section>
      {modal && (
        <Modal title="New transaction" onClose={() => setModal(false)}>
          <div className="tab-row">
            {(["expense", "income", "transfer"] as const).map((t) => (
              <button
                key={t}
                className={tab === t ? "selected" : ""}
                onClick={() => {
                  setTab(t);
                  setCategoryId(categories.find((c) => c.type === t)?.id || "");
                }}
              >
                {t[0].toUpperCase() + t.slice(1)}
              </button>
            ))}
          </div>
          <form className="form-stack modal-form" onSubmit={save}>
            <label>
              Amount
              <input
                required
                type="number"
                min="0.01"
                step="0.01"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                placeholder="0.00"
              />
            </label>
            <label>
              Description
              <input
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="What was this for?"
              />
            </label>
            <label>
              {tab === "transfer" ? "From account" : "Account"}
              <select
                required
                value={accountId}
                onChange={(e) => setAccountId(e.target.value)}
              >
                {accounts.map((a) => (
                  <option key={a.id} value={a.id}>
                    {a.name}
                  </option>
                ))}
              </select>
            </label>
            {tab === "transfer" ? (
              <label>
                To account
                <select
                  required
                  value={toAccountId}
                  onChange={(e) => setToAccountId(e.target.value)}
                >
                  {accounts.map((a) => (
                    <option key={a.id} value={a.id}>
                      {a.name}
                    </option>
                  ))}
                </select>
              </label>
            ) : (
              <label>
                Category
                <select
                  value={categoryId}
                  onChange={(e) => setCategoryId(e.target.value)}
                >
                  <option value="">Uncategorized</option>
                  {categories
                    .filter((c) => c.type === tab)
                    .map((c) => (
                      <option key={c.id} value={c.id}>
                        {c.name}
                      </option>
                    ))}
                </select>
              </label>
            )}
            <label>
              Date
              <input
                type="date"
                required
                value={date}
                onChange={(e) => setDate(e.target.value)}
              />
            </label>
            {error && <div className="error-box">{error}</div>}
            <button
              className="primary-btn full"
              disabled={saving || !accountId}
            >
              {saving ? "Saving…" : "Save transaction"}
            </button>
            <p className="form-hint">
              Amounts are converted to integer minor units before sending.
            </p>
          </form>
        </Modal>
      )}
    </div>
  );
}
