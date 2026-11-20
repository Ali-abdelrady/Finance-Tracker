import { useEffect, useState } from "react";
import {
  ArrowUpRight,
  ArrowDownRight,
  Wallet,
  TrendingUp,
  TrendingDown,
  Plus,
  ArrowRight,
} from "lucide-react";
import {
  Area,
  AreaChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
  PieChart,
  Pie,
  Cell,
} from "recharts";
import { dashboardApi } from "../api/services";
import type { DashboardSummary } from "../types";
import { formatCurrency } from "../lib/currency";
import { getErrorMessage } from "../api/client";
import { useNavigate } from "react-router-dom";
const currentMonth = () => {
  const now = new Date();
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}`;
};
const fallback: DashboardSummary = {
  netWorth: 0,
  month: currentMonth(),
  monthlyIncome: 0,
  monthlyExpense: 0,
  categoryBreakdown: [],
  cashflow: [],
  accounts: [],
};
export default function Dashboard() {
  const [data, setData] = useState(fallback);
  const [selectedMonth, setSelectedMonth] = useState(currentMonth);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const nav = useNavigate();
  useEffect(() => {
    setLoading(true);
    setError("");
    dashboardApi
      .summary(selectedMonth)
      .then(setData)
      .catch((e) => setError(getErrorMessage(e)))
      .finally(() => setLoading(false));
  }, [selectedMonth]);

  const selectedMonthLabel = new Date(`${selectedMonth}-01T00:00:00`).toLocaleDateString(
    undefined,
    { month: "long", year: "numeric" },
  );
  const rate =
    data.monthlyIncome > 0
      ? Math.max(
          0,
          Math.min(
            100,
            ((data.monthlyIncome - data.monthlyExpense) / data.monthlyIncome) *
              100,
          ),
        )
      : 0;

  const metrics = [
    {
      label: "Net worth",
      value: data.netWorth,
      Icon: Wallet,
      note: "Current balance across all accounts",
      tone: "violet",
    },
    {
      label: "Monthly income",
      value: data.monthlyIncome,
      Icon: TrendingUp,
      note: `Money in during ${selectedMonthLabel}`,
      tone: "green",
    },
    {
      label: "Monthly expenses",
      value: data.monthlyExpense,
      Icon: TrendingDown,
      note: `Money out during ${selectedMonthLabel}`,
      tone: "orange",
    },
  ];

  return (
    <div className="content-stack">
      <div className="welcome-row">
        <div>
          <span className="eyebrow">YOUR FINANCIAL PICTURE</span>
          <h1>
            Good afternoon<span className="accent-dot">.</span>
          </h1>
          <p className="muted">
            Summary and spending breakdown for {selectedMonthLabel}.
          </p>
        </div>
        <div className="dashboard-header-actions">
          <label className="month-picker">
            <span>Summary month</span>
            <input
              type="month"
              value={selectedMonth}
              max={currentMonth()}
              onChange={(e) => setSelectedMonth(e.target.value)}
              aria-label="Select summary month"
            />
          </label>
          <button className="primary-btn" onClick={() => nav("/transactions")}>
            <Plus size={17} /> New transaction
          </button>
        </div>
      </div>
      {error && <div className="error-box">{error}</div>}
      <section className="metric-grid">
        {metrics.map(({ label, value, Icon, note, tone }) => (
          <article className="metric-card" key={label}>
            <div className="metric-top">
              <span>{label}</span>
              <span className={`metric-icon ${tone}`}>
                <Icon size={18} />
              </span>
            </div>
            <div className="metric-value">
              {loading ? "—" : formatCurrency(value)}
            </div>
            <div className="metric-note">{note}</div>
          </article>
        ))}
      </section>
      <section className="savings-card">
        <div>
          <span className="eyebrow">SAVINGS RATE · {selectedMonthLabel}</span>
          <h3>
            {rate.toFixed(1)}% <span>of your income saved</span>
          </h3>
          <p className="muted">
            A little progress adds up. Keep your momentum going.
          </p>
        </div>
        <div className="savings-progress">
          <div className="progress-track">
            <span style={{ width: `${rate}%` }} />
          </div>
          <div className="progress-labels">
            <span>0%</span>
            <b>{rate.toFixed(0)}%</b>
            <span>100%</span>
          </div>
        </div>
      </section>
      <div className="dashboard-grid">
        <section className="panel chart-panel">
          <div className="panel-heading">
            <div>
              <span className="eyebrow">CASHFLOW</span>
              <h3>Income vs. expenses</h3>
            </div>
            <span className="period-chip">Ending {selectedMonthLabel}</span>
          </div>
          <div className="chart-wrap">
            {data?.cashflow?.length ? (
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={data.cashflow}>
                  <defs>
                    <linearGradient id="incomeFill" x1="0" y1="0" x2="0" y2="1">
                      <stop
                        offset="0%"
                        stopColor="#4adeb5"
                        stopOpacity={0.25}
                      />
                      <stop offset="100%" stopColor="#4adeb5" stopOpacity={0} />
                    </linearGradient>
                    <linearGradient
                      id="expenseFill"
                      x1="0"
                      y1="0"
                      x2="0"
                      y2="1"
                    >
                      <stop offset="0%" stopColor="#fb9b72" stopOpacity={0.2} />
                      <stop offset="100%" stopColor="#fb9b72" stopOpacity={0} />
                    </linearGradient>
                  </defs>
                  <CartesianGrid
                    stroke="#242d40"
                    strokeDasharray="3 5"
                    vertical={false}
                  />
                  <XAxis
                    dataKey="month"
                    stroke="#737f96"
                    tickLine={false}
                    axisLine={false}
                  />
                  <YAxis
                    stroke="#737f96"
                    tickLine={false}
                    axisLine={false}
                    tickFormatter={(v) => `${(v / 100).toFixed(0)}`}
                  />
                  <Tooltip
                    contentStyle={{
                      background: "#151d2d",
                      border: "1px solid #29344a",
                      borderRadius: 12,
                    }}
                    formatter={(v: number) => formatCurrency(v)}
                  />
                  <Area
                    type="monotone"
                    dataKey="income"
                    stroke="#4adeb5"
                    strokeWidth={2.5}
                    fill="url(#incomeFill)"
                  />
                  <Area
                    type="monotone"
                    dataKey="expense"
                    stroke="#fb9b72"
                    strokeWidth={2.5}
                    fill="url(#expenseFill)"
                  />
                </AreaChart>
              </ResponsiveContainer>
            ) : (
              <div className="empty-chart">
                Your cashflow trend will appear here once you add transactions.
              </div>
            )}
          </div>
          <div className="chart-legend">
            <span>
              <i className="legend-dot green" /> Income
            </span>
            <span>
              <i className="legend-dot orange" /> Expenses
            </span>
          </div>
        </section>
        <section className="panel chart-panel">
          <div className="panel-heading">
            <div>
              <span className="eyebrow">SPENDING MIX</span>
              <h3>By category</h3>
            </div>
            <button className="text-btn" onClick={() => nav("/transactions")}>
              Details <ArrowUpRight size={15} />
            </button>
          </div>
          <div className="donut-wrap">
            {data?.categoryBreakdown?.length ? (
              <>
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie
                      data={data.categoryBreakdown}
                      dataKey="amount"
                      nameKey="categoryName"
                      innerRadius={65}
                      outerRadius={91}
                      paddingAngle={4}
                      stroke="none"
                    >
                      {data.categoryBreakdown?.map((c, i) => (
                        <Cell key={i} fill={c.color || "#8b7cff"} />
                      ))}
                    </Pie>
                    <Tooltip formatter={(v: number) => formatCurrency(v)} />
                  </PieChart>
                </ResponsiveContainer>
                <div className="donut-center">
                  <b>{formatCurrency(data.monthlyExpense)}</b>
                  <span>Total spent</span>
                </div>
              </>
            ) : (
              <div className="empty-chart">No spending data yet.</div>
            )}
          </div>
          <div className="category-legend">
            {data.categoryBreakdown?.slice(0, 5)?.map((c) => (
              <div key={c.categoryName}>
                <span>
                  <i style={{ background: c.color || "#8b7cff" }} />
                  {c.categoryName}
                </span>
                <b>{formatCurrency(c.amount)}</b>
              </div>
            ))}
          </div>
        </section>
      </div>
      <section className="panel accounts-panel">
        <div className="panel-heading">
          <div>
            <span className="eyebrow">YOUR MONEY</span>
            <h3>Accounts</h3>
          </div>
          <button className="text-btn" onClick={() => nav("/accounts")}>
            All accounts <ArrowRight size={15} />
          </button>
        </div>
        <div className="account-strip">
          {data.accounts?.map((a) => (
            <article className="account-mini" key={a.id}>
              <div className="account-mini-icon">
                <Wallet size={18} />
              </div>
              <div>
                <span>{a.name}</span>
                <b>{formatCurrency(a.balance, a.currency || "EGP")}</b>
              </div>
              <ArrowUpRight size={16} />
            </article>
          ))}
          {!data?.accounts?.length && (
            <p className="muted">
              Connect the dots. Add your first account to get started.
            </p>
          )}
        </div>
      </section>
    </div>
  );
}
