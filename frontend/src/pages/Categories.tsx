import { useEffect, useState } from "react";
import { Pencil, Plus, Tag } from "lucide-react";
import { categoriesApi } from "../api/services";
import type { Category } from "../types";
import { getErrorMessage } from "../api/client";
import Modal from "../components/Modal";
export default function Categories() {
  const [items, setItems] = useState<Category[]>([]);
  const [show, setShow] = useState(false);
  const [editingCategory, setEditingCategory] = useState<Category | null>(null);
  const [name, setName] = useState("");
  const [type, setType] = useState<Category["type"]>("expense");
  const [color, setColor] = useState("#8b7cff");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const load = () =>
    categoriesApi
      .list()
      .then(setItems)
      .catch((e) => setError(getErrorMessage(e)));
  useEffect(() => {
    load();
  }, []);

  function openCreateModal() {
    setEditingCategory(null);
    setName("");
    setType("expense");
    setColor("#8b7cff");
    setError("");
    setShow(true);
  }

  function openEditModal(category: Category) {
    setEditingCategory(category);
    setName(category.name);
    setType(category.type);
    setColor(category.color || "#8b7cff");
    setError("");
    setShow(true);
  }

  function closeModal() {
    setShow(false);
    setEditingCategory(null);
  }

  async function save(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError("");
    try {
      const payload = {
        name,
        type,
        color,
        icon: editingCategory?.icon || "tag",
      };
      if (editingCategory) {
        await categoriesApi.update(editingCategory.id, payload);
      } else {
        await categoriesApi.create(payload);
      }
      closeModal();
      await load();
    } catch (e) {
      setError(getErrorMessage(e));
    } finally {
      setBusy(false);
    }
  }
  return (
    <div className="content-stack">
      <div className="welcome-row">
        <div>
          <span className="eyebrow">MAKE IT YOURS</span>
          <h1>
            Categories<span className="accent-dot">.</span>
          </h1>
          <p className="muted">Give every pound a place and a purpose.</p>
        </div>
        <button className="primary-btn" onClick={openCreateModal}>
          <Plus size={17} /> New category
        </button>
      </div>
      {error && <div className="error-box">{error}</div>}
      <div className="category-columns">
        {(["expense", "income"] as const).map((t) => (
          <section className="panel category-panel" key={t}>
            <div className="panel-heading">
              <div>
                <span className="eyebrow">
                  {t === "expense" ? "MONEY OUT" : "MONEY IN"}
                </span>
                <h3>
                  {t === "expense" ? "Expense categories" : "Income categories"}
                </h3>
              </div>
              <span className="count-chip">
                {items.filter((x) => x.type === t).length}
              </span>
            </div>
            {items
              .filter((x) => x.type === t)
              .map((c) => (
                <div className="category-row" key={c.id}>
                  <span
                    className="category-color"
                    style={{ background: c.color }}
                  >
                    <Tag size={16} />
                  </span>
                  <b>{c.name}</b>
                  <span className="category-type">{t}</span>
                  <button
                    type="button"
                    className="edit-category-btn"
                    onClick={() => openEditModal(c)}
                    title="Edit category"
                    aria-label={`Edit ${c.name}`}
                  >
                    <Pencil size={15} />
                  </button>
                </div>
              ))}
            {!items.some((x) => x.type === t) && (
              <p className="muted category-empty">No categories yet.</p>
            )}
          </section>
        ))}
      </div>
      {show && (
        <Modal
          title={editingCategory ? "Edit category" : "Create category"}
          onClose={closeModal}
        >
          <form className="form-stack modal-form" onSubmit={save}>
            <label>
              Category name
              <input
                required
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g. Groceries"
              />
            </label>
            <label>
              Type
              <select
                value={type}
                onChange={(e) => setType(e.target.value as Category["type"])}
              >
                <option value="expense">Expense</option>
                <option value="income">Income</option>
              </select>
            </label>
            <label>
              Color
              <input
                type="color"
                value={color}
                onChange={(e) => setColor(e.target.value)}
              />
            </label>
            <button className="primary-btn full" disabled={busy}>
              {busy
                ? "Saving…"
                : editingCategory
                  ? "Save changes"
                  : "Create category"}
            </button>
          </form>
        </Modal>
      )}
    </div>
  );
}
