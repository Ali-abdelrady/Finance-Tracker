import { NavLink, Outlet, useNavigate } from "react-router-dom";
import {
  WalletCards,
  LayoutDashboard,
  ArrowLeftRight,
  Tags,
  LogOut,
  PlusCircle,
  Menu,
} from "lucide-react";
import { useAuth } from "../context/AuthContext";
import { useState } from "react";
const links = [
  ["/dashboard", "Overview", LayoutDashboard],
  ["/accounts", "Accounts", WalletCards],
  ["/transactions", "Transactions", ArrowLeftRight],
  ["/categories", "Categories", Tags],
] as const;
export default function Layout() {
  const { user, logout } = useAuth();
  const nav = useNavigate();
  const [open, setOpen] = useState(false);
  return (
    <div className="app-shell">
      {open && (
        <button
          type="button"
          className="sidebar-backdrop"
          aria-label="Close navigation menu"
          onClick={() => setOpen(false)}
        />
      )}
      <aside
        id="primary-sidebar"
        className={`sidebar ${open ? "sidebar-open" : ""}`}
      >
        <div className="brand">
          <span className="brand-mark">
            <WalletCards size={21} />
          </span>
          <span>
            my<span className="brand-accent">finance</span>
          </span>
        </div>
        <div className="side-label">WORKSPACE</div>
        <nav>
          {links.map(([to, label, Icon]) => (
            <NavLink
              key={to}
              to={to}
              className={({ isActive }) =>
                `nav-link ${isActive ? "active" : ""}`
              }
              onClick={() => setOpen(false)}
            >
              <Icon size={18} />
              {label}
            </NavLink>
          ))}
        </nav>
        <div className="sidebar-bottom">
          <div className="avatar">{user?.email?.slice(0, 1) || "A"}</div>
          <div className="user-meta">
            <b>{user?.email || "Your account"}</b>
            <span>{user?.email || "Personal workspace"}</span>
          </div>
          <button
            className="icon-btn"
            title="Sign out"
            onClick={() => {
              setOpen(false);
              logout();
              nav("/login");
            }}
          >
            <LogOut size={17} />
          </button>
        </div>
      </aside>
      <main className="main-area">
        <header className="topbar">
          <button
            className="icon-btn mobile-menu"
            onClick={() => setOpen(!open)}
            aria-expanded={open}
            aria-controls="primary-sidebar"
            aria-label={open ? "Close navigation menu" : "Open navigation menu"}
          >
            <Menu />
          </button>
          <div>
            <span className="eyebrow">PERSONAL FINANCE</span>
            <h2>Money overview</h2>
          </div>
          <button className="top-action" onClick={() => nav("/transactions")}>
            <PlusCircle size={17} /> Add transaction
          </button>
        </header>
        <div className="page-content">
          <Outlet />
        </div>
      </main>
    </div>
  );
}
