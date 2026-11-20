# MyFinance Frontend

A dark fintech-style React + Vite + TypeScript frontend for a Go/Fiber + PostgreSQL API.

## Run locally
1. Install Node.js 20+.
2. Copy `.env.example` to `.env` (change the URL if needed).
3. Run `npm install`.
4. Run `npm run dev`.
5. Open `http://localhost:5173`.

Build check: `npm run build`.

## Backend contract assumed by this frontend
All routes are relative to `VITE_API_BASE_URL` (default `http://localhost:8080/api/v1`).

| Method | Endpoint | Expected response / request |
|---|---|---|
| POST | `/auth/login` | `{email,password}` → `{token,user}` |
| POST | `/auth/register` | `{name,email,password}` → `{token,user}` |
| GET | `/auth/me` | `User` |
| GET | `/dashboard` | `DashboardSummary` (amounts in minor units) |
| GET/POST | `/accounts` | `Account[]` / account payload |
| GET/POST | `/categories` | `Category[]` / category payload |
| GET | `/transactions` | `Page<Transaction>`; query `type,q,page,limit` |
| POST | `/transactions` | income/expense payload |
| POST | `/transfers` | `{amount,fromAccountId,toAccountId,date,description}` |

### Important response shapes
- All money fields (`balance`, `amount`, `netWorth`, `monthlyIncome`, `monthlyExpense`, chart amounts) are integer minor units. The UI divides by 100 for display.
- `GET /transactions` returns `{data: Transaction[], page, limit, total, totalPages}`.
- `GET /dashboard` returns `{netWorth,monthlyIncome,monthlyExpense,categoryBreakdown:[{name,amount,color}],cashflow:[{month,income,expense}],accounts}`.
- Auth errors should return a useful JSON `message` or `error` field.

## Production notes / integration checklist
- Configure CORS on Fiber for the Vite origin during development and the deployed frontend origin in production.
- Use HTTPS in production; set `VITE_API_BASE_URL` at build time.
- This starter stores the JWT in `localStorage` to match the requested contract. For stronger XSS resilience, prefer secure, HttpOnly, SameSite cookies and a CSRF strategy if the backend can support them.
- Add refresh-token rotation, logout/revocation, rate limiting, server-side validation, and automated tests before production deployment.
- Confirm whether your API uses `accountId`/`fromAccountId` naming and the pagination envelope above; adjust `services.ts` and transaction payloads to match the actual Fiber handlers.
- Currency display defaults to USD. Account cards and transaction rows use the account currency when provided; dashboard summary currently defaults to USD. Adapt locale/currency policy for multi-currency accounts.
- The categories screen includes a color picker and a simple tag icon. Extend the category API/UI with an icon registry if you need selectable Lucide icon names.
