"use client";

import { useEffect, useState, type FormEvent } from "react";

type Category = {
  id: number;
  name: string;
};

type PurchaseInstallment = {
  id: number;
  installment_number: number;
  amount_cents: number;
  entry_date: string;
  category_id: number;
};

type Purchase = {
  id: number;
  description: string;
  purchase_date: string;
  amount_cents: number;
  installment_count: number;
  category_id: number;
  category_name: string;
  category?: Category;
  installments: PurchaseInstallment[];
};

function formatBRL(cents: number): string {
  return `R$ ${(cents / 100).toFixed(2).replace(".", ",")}`;
}

function parseReaisToCents(raw: string): number {
  const normalized = raw.trim().replace(",", ".");
  const value = Number(normalized);
  if (!Number.isFinite(value)) {
    return 0;
  }
  return Math.round(value * 100);
}

export default function PurchasesPage() {
  const [purchases, setPurchases] = useState<Purchase[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [description, setDescription] = useState("");
  const [date, setDate] = useState("");
  const [amount, setAmount] = useState("");
  const [installments, setInstallments] = useState("2");
  const [categoryId, setCategoryId] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [editDescription, setEditDescription] = useState("");
  const [editAmount, setEditAmount] = useState("");
  const [editCategoryId, setEditCategoryId] = useState("");

  useEffect(() => {
    async function load() {
      const [purchasesResponse, categoriesResponse] = await Promise.all([
        fetch("/api/purchases"),
        fetch("/api/categories"),
      ]);
      if (purchasesResponse.ok) {
        setPurchases((await purchasesResponse.json()) as Purchase[]);
      } else {
        setError("Não foi possível carregar as compras.");
      }
      if (categoriesResponse.ok) {
        setCategories((await categoriesResponse.json()) as Category[]);
      }
    }
    void load();
  }, []);

  async function handleCreate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);

    const response = await fetch("/api/purchases", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        description,
        purchase_date: date,
        amount_cents: parseReaisToCents(amount),
        installment_count: Number(installments),
        category_id: Number(categoryId),
      }),
    });

    if (!response.ok) {
      const data = await response.json().catch(() => null);
      setError(data?.message ?? data?.error ?? "Não foi possível criar a compra.");
      return;
    }

    const created = (await response.json()) as Purchase;
    setPurchases((current) =>
      [...current, created].sort((a, b) => {
        if (a.purchase_date === b.purchase_date) {
          return b.id - a.id;
        }
        return a.purchase_date < b.purchase_date ? 1 : -1;
      }),
    );
    setDescription("");
    setDate("");
    setAmount("");
    setInstallments("2");
    setCategoryId("");
  }

  function startEdit(purchase: Purchase) {
    setEditingId(purchase.id);
    setEditDescription(purchase.description);
    setEditAmount((purchase.amount_cents / 100).toFixed(2).replace(".", ","));
    setEditCategoryId(String(purchase.category_id));
    setError(null);
  }

  async function handleEdit(event: FormEvent<HTMLFormElement>, id: number) {
    event.preventDefault();
    setError(null);

    const response = await fetch(`/api/purchases/${id}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        description: editDescription,
        amount_cents: parseReaisToCents(editAmount),
        category_id: Number(editCategoryId),
      }),
    });

    if (!response.ok) {
      const data = await response.json().catch(() => null);
      setError(data?.message ?? data?.error ?? "Não foi possível salvar a compra.");
      return;
    }

    const updated = (await response.json()) as Purchase;
    setPurchases((current) =>
      current.map((purchase) => (purchase.id === id ? updated : purchase)),
    );
    setEditingId(null);
  }

  async function handleUndo(id: number) {
    setError(null);
    const response = await fetch(`/api/purchases/${id}`, { method: "DELETE" });
    if (!response.ok) {
      const data = await response.json().catch(() => null);
      if (response.status === 409) {
        setError(
          data?.message ??
            data?.error ??
            "Não é possível desfazer: já há parcela em mês passado.",
        );
        return;
      }
      setError("Não foi possível desfazer a compra.");
      return;
    }
    setPurchases((current) => current.filter((purchase) => purchase.id !== id));
    if (editingId === id) {
      setEditingId(null);
    }
  }

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-col gap-6 p-6">
      <h1 className="text-2xl font-semibold">Compras</h1>

      <form onSubmit={handleCreate} className="grid gap-3 sm:grid-cols-2">
        <label className="flex flex-col gap-1 sm:col-span-2">
          Descrição
          <input
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            className="rounded border px-3 py-2"
          />
        </label>
        <label className="flex flex-col gap-1">
          Data
          <input
            type="date"
            value={date}
            onChange={(e) => setDate(e.target.value)}
            className="rounded border px-3 py-2"
          />
        </label>
        <label className="flex flex-col gap-1">
          Valor
          <input
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            className="rounded border px-3 py-2"
            inputMode="decimal"
          />
        </label>
        <label className="flex flex-col gap-1">
          Parcelas
          <input
            value={installments}
            onChange={(e) => setInstallments(e.target.value)}
            className="rounded border px-3 py-2"
            inputMode="numeric"
          />
        </label>
        <label className="flex flex-col gap-1">
          Categoria
          <select
            value={categoryId}
            onChange={(e) => setCategoryId(e.target.value)}
            className="rounded border px-3 py-2"
          >
            <option value="">Selecione</option>
            {categories.map((category) => (
              <option key={category.id} value={category.id}>
                {category.name}
              </option>
            ))}
          </select>
        </label>
        <button
          type="submit"
          className="rounded bg-foreground px-4 py-2 text-background sm:col-span-2"
        >
          Registrar compra
        </button>
      </form>

      {error && <p className="text-red-600">{error}</p>}

      {purchases.length === 0 && !error ? (
        <p>Nenhuma compra ainda.</p>
      ) : (
        <ul className="flex flex-col gap-4">
          {purchases.map((purchase) => (
            <li key={purchase.id} className="border-b pb-3">
              {editingId === purchase.id ? (
                <form
                  onSubmit={(event) => handleEdit(event, purchase.id)}
                  className="grid gap-3 sm:grid-cols-2"
                >
                  <label className="flex flex-col gap-1 sm:col-span-2">
                    Descrição
                    <input
                      value={editDescription}
                      onChange={(e) => setEditDescription(e.target.value)}
                      className="rounded border px-3 py-2"
                    />
                  </label>
                  <label className="flex flex-col gap-1">
                    Valor
                    <input
                      value={editAmount}
                      onChange={(e) => setEditAmount(e.target.value)}
                      className="rounded border px-3 py-2"
                      inputMode="decimal"
                    />
                  </label>
                  <label className="flex flex-col gap-1">
                    Categoria
                    <select
                      value={editCategoryId}
                      onChange={(e) => setEditCategoryId(e.target.value)}
                      className="rounded border px-3 py-2"
                    >
                      {categories.map((category) => (
                        <option key={category.id} value={category.id}>
                          {category.name}
                        </option>
                      ))}
                    </select>
                  </label>
                  <button
                    type="submit"
                    className="rounded bg-foreground px-3 py-2 text-background"
                  >
                    Salvar
                  </button>
                  <button
                    type="button"
                    className="rounded border px-3 py-2"
                    onClick={() => setEditingId(null)}
                  >
                    Cancelar
                  </button>
                </form>
              ) : (
                <>
                  <p className="font-medium">{purchase.description}</p>
                  <p>
                    {formatBRL(purchase.amount_cents)} · {purchase.installment_count}x ·{" "}
                    {purchase.purchase_date} · {purchase.category_name}
                  </p>
                  <div className="mt-2 flex gap-2">
                    <button
                      type="button"
                      className="rounded border px-3 py-1"
                      onClick={() => startEdit(purchase)}
                    >
                      Editar
                    </button>
                    <button
                      type="button"
                      className="rounded border px-3 py-1"
                      onClick={() => handleUndo(purchase.id)}
                    >
                      Desfazer
                    </button>
                  </div>
                </>
              )}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
