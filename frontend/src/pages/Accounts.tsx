import { useEffect, useState } from "react";
import {
  Plus,
  Wallet,
  CreditCard,
  Landmark,
  Trash2,
  Pencil,
} from "lucide-react";
import { accountsApi } from "../api/services";
import type { Account } from "../types";
import { formatCurrency, toMinorUnits } from "../lib/currency";
import { getErrorMessage } from "../api/client";
import Modal from "../components/Modal";

export default function Accounts() {
  const [items, setItems] = useState<Account[]>([]);
  const [show, setShow] = useState(false);
  const [editingAccount, setEditingAccount] = useState<Account | null>(null);
  const [name, setName] = useState("");
  const [type, setType] = useState<Account["type"]>("cash");
  const [balance, setBalance] = useState("0");
  const [currency, setCurrency] = useState("USD");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [deletingId, setDeletingId] = useState<string | null>(null);

  const load = () =>
    accountsApi
      .list()
      .then(setItems)
      .catch((e) => setError(getErrorMessage(e)));

  useEffect(() => {
    load();
  }, []);

  function openCreateModal() {
    setEditingAccount(null);
    setName("");
    setType("cash");
    setBalance("0");
    setCurrency("USD");
    setError("");
    setShow(true);
  }

  function openEditModal(account: Account) {
    setEditingAccount(account);
    setName(account.name);
    setType(account.type);
    setCurrency(account.currency || "USD");
    setError("");
    setShow(true);
  }

  function closeModal() {
    setShow(false);
    setEditingAccount(null);
  }

  async function save(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError("");

    try {
      if (editingAccount) {
        await accountsApi.update(editingAccount.id, { name, type, currency });
      } else {
        await accountsApi.create({
          name,
          type,
          balance: toMinorUnits(balance),
          currency,
        });
      }

      closeModal();
      await load();
    } catch (e) {
      setError(getErrorMessage(e));
    } finally {
      setBusy(false);
    }
  }

  async function deleteAccount(account: Account) {
    const confirmed = window.confirm(
      `Are you sure you want to delete "${account.name}"? This action cannot be undone.`,
    );

    if (!confirmed) return;

    setDeletingId(account.id);
    setError("");

    try {
      await accountsApi.delete(account.id);
      setItems((prev) => prev.filter((item) => item.id !== account.id));
    } catch (e) {
      setError(getErrorMessage(e));
    } finally {
      setDeletingId(null);
    }
  }

  const Icon = ({ kind }: { kind: Account["type"] }) =>
    kind === "credit_card" || kind === "debit_card" ? (
      <CreditCard />
    ) : kind === "savings" ? (
      <Landmark />
    ) : (
      <Wallet />
    );

  return (
    <div className="content-stack">
      <div className="welcome-row">
        <div>
          <span className="eyebrow">YOUR POCKETS</span>
          <h1>
            Accounts<span className="accent-dot">.</span>
          </h1>
          <p className="muted">See all your money, wherever you keep it.</p>
        </div>

        <button className="primary-btn" onClick={openCreateModal}>
          <Plus size={17} /> Add account
        </button>
      </div>

      {error && <div className="error-box">{error}</div>}

      <div className="accounts-grid">
        {items.map((a, i) => (
          <article className={`account-card account-color-${i % 4}`} key={a.id}>
            <div className="account-card-head">
              <span className="account-type-icon">
                <Icon kind={a.type} />
              </span>

              <div className="account-actions">
                <span className="account-type-label">{a.type}</span>

                <button
                  type="button"
                  className="edit-account-btn"
                  onClick={() => openEditModal(a)}
                  disabled={deletingId === a.id}
                  title="Edit account"
                  aria-label={`Edit ${a.name}`}
                >
                  <Pencil size={17} />
                </button>
                <button
                  type="button"
                  className="delete-account-btn"
                  onClick={() => deleteAccount(a)}
                  disabled={deletingId === a.id}
                  title="Delete account"
                  aria-label={`Delete ${a.name}`}
                >
                  <Trash2 size={17} />
                </button>
              </div>
            </div>

            <div className="account-name">{a.name}</div>

            <div className="account-balance">
              {formatCurrency(a.balance, a.currency || "USD")}
            </div>

            <div className="account-card-foot">
              <span>{a.institution || "Personal account"}</span>
              <span>•••• {a.id.slice(-4)}</span>
            </div>

            {deletingId === a.id && (
              <div className="account-deleting">Deleting...</div>
            )}
          </article>
        ))}

        {!items.length && (
          <div className="empty-state">
            <Wallet size={28} />
            <h3>Your accounts live here</h3>
            <p>
              Add cash, cards, and savings accounts to start tracking your net
              worth.
            </p>
            <button className="primary-btn" onClick={openCreateModal}>
              Create first account
            </button>
          </div>
        )}
      </div>

      {show && (
        <Modal
          title={editingAccount ? "Edit account" : "Create account"}
          onClose={closeModal}
        >
          <form className="form-stack modal-form" onSubmit={save}>
            <label>
              Account name
              <input
                required
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g. Everyday checking"
              />
            </label>

            <label>
              Account type
              <select
                value={type}
                onChange={(e) => setType(e.target.value as Account["type"])}
              >
                <option value="cash">Cash</option>
                <option value="debit_card">Debit Card</option>
                <option value="credit_card">Credit Card</option>
                <option value="savings">Savings</option>
              </select>
            </label>

            {!editingAccount && (
              <label>
                Initial balance
                <input
                  type="number"
                  min="0"
                  step="0.01"
                  required
                  value={balance}
                  onChange={(e) => setBalance(e.target.value)}
                />
              </label>
            )}

            <label>
              Currency
              <select
                value={currency}
                onChange={(e) => setCurrency(e.target.value)}
              >
                <option>USD</option>
                <option>EGP</option>
                <option>EUR</option>
                <option>GBP</option>
              </select>
            </label>

            {error && <div className="error-box">{error}</div>}

            <button className="primary-btn full" disabled={busy}>
              {busy
                ? editingAccount
                  ? "Saving…"
                  : "Creating…"
                : editingAccount
                  ? "Save changes"
                  : "Create account"}
            </button>
          </form>
        </Modal>
      )}
    </div>
  );
}
