"use client";

import { useEffect, useState, type FormEvent } from "react";
import {
  createColumnHelper,
  tableFeatures,
  useTable,
} from "@tanstack/react-table";
import {
  formatReaisFromCents,
  maskReaisInput,
  parseReaisToCents,
} from "@/lib/money";

type Category = {
  id: number;
  name: string;
};

type Entry = {
  id: number;
  type: "income" | "expense";
  amount_cents: number;
  entry_date: string;
  category_id: number;
  category?: Category;
  purchase_id?: number;
  installment_number?: number;
  purchase?: {
    id: number;
    description: string;
    installment_count: number;
  };
};

function isParcela(entry: Entry): boolean {
  return entry.purchase_id != null;
}

function parcelaLabel(entry: Entry): string {
  if (!entry.purchase || entry.installment_number == null) {
    return "";
  }
  return `${entry.installment_number}/${entry.purchase.installment_count} · ${entry.purchase.description}`;
}

const features = tableFeatures({});
const helper = createColumnHelper<typeof features, Entry>();
const columns = helper.columns([
  helper.accessor("type", {
    header: "Tipo",
    cell: ({ getValue }) => (getValue() === "income" ? "Ganho" : "Gasto"),
  }),
  helper.accessor("amount_cents", {
    header: "Valor",
    cell: ({ getValue }) => formatBRL(getValue()),
  }),
  helper.accessor("entry_date", { header: "Data" }),
  helper.accessor((row) => row.category?.name ?? "", { id: "category", header: "Categoria" }),
  helper.accessor((row) => parcelaLabel(row), { id: "parcela", header: "Parcela" }),
]);

const EMPTY_ENTRIES: Entry[] = [];

function formatBRL(cents: number): string {
  return `R$ ${(cents / 100).toFixed(2).replace(".", ",")}`;
}

export default function EntriesPage() {
  const [entries, setEntries] = useState<Entry[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [type, setType] = useState<"income" | "expense">("expense");
  const [amount, setAmount] = useState("");
  const [date, setDate] = useState("");
  const [categoryId, setCategoryId] = useState("");
  const [editingId, setEditingId] = useState<number | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function load() {
      const [entriesResponse, categoriesResponse] = await Promise.all([
        fetch("/api/entries"),
        fetch("/api/categories"),
      ]);
      if (entriesResponse.ok) {
        setEntries((await entriesResponse.json()) as Entry[]);
      } else {
        setError("Não foi possível carregar os lançamentos.");
      }
      if (categoriesResponse.ok) {
        setCategories((await categoriesResponse.json()) as Category[]);
      }
    }
    void load();
  }, []);

  const table = useTable({
    features,
    columns,
    data: entries.length === 0 ? EMPTY_ENTRIES : entries,
  });

  function sortEntries(list: Entry[]): Entry[] {
    return [...list].sort((a, b) => {
      if (a.entry_date === b.entry_date) {
        return b.id - a.id;
      }
      return a.entry_date < b.entry_date ? 1 : -1;
    });
  }

  function resetForm() {
    setType("expense");
    setAmount("");
    setDate("");
    setCategoryId("");
    setEditingId(null);
  }

  function startEdit(entry: Entry) {
    setEditingId(entry.id);
    setType(entry.type);
    setAmount(formatReaisFromCents(entry.amount_cents));
    setDate(entry.entry_date);
    setCategoryId(String(entry.category_id));
    setError(null);
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);

    const payload = {
      type,
      amount_cents: parseReaisToCents(amount),
      entry_date: date,
      category_id: Number(categoryId),
    };

    const response = await fetch(
      editingId === null ? "/api/entries" : `/api/entries/${editingId}`,
      {
        method: editingId === null ? "POST" : "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      },
    );

    if (!response.ok) {
      const data = await response.json().catch(() => null);
      setError(
        data?.message ??
          data?.error ??
          "Não foi possível salvar o lançamento.",
      );
      return;
    }

    const saved = (await response.json()) as Entry;
    setEntries((current) => {
      const without = current.filter((entry) => entry.id !== saved.id);
      return sortEntries([...without, saved]);
    });
    resetForm();
  }

  async function handleDelete(id: number) {
    setError(null);
    const response = await fetch(`/api/entries/${id}`, { method: "DELETE" });
    if (!response.ok) {
      setError("Não foi possível excluir o lançamento.");
      return;
    }
    setEntries((current) => current.filter((entry) => entry.id !== id));
    if (editingId === id) {
      resetForm();
    }
  }

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-col gap-6 p-6">
      <h1 className="text-2xl font-semibold">Lançamentos</h1>

      <form onSubmit={handleSubmit} className="grid gap-3 sm:grid-cols-2">
        <label className="flex flex-col gap-1">
          Tipo
          <select
            value={type}
            onChange={(e) => setType(e.target.value as "income" | "expense")}
            className="rounded border px-3 py-2"
          >
            <option value="expense">Gasto</option>
            <option value="income">Ganho</option>
          </select>
        </label>
        <label className="flex flex-col gap-1">
          Valor
          <input
            value={amount}
            onChange={(e) => setAmount(maskReaisInput(e.target.value))}
            className="rounded border px-3 py-2"
            inputMode="numeric"
            autoComplete="off"
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
          {editingId === null ? "Lançar" : "Salvar"}
        </button>
      </form>

      {error && <p className="text-red-600">{error}</p>}

      {entries.length === 0 && !error ? (
        <p>Nenhum lançamento ainda.</p>
      ) : entries.length === 0 ? null : (
        <table className="w-full border-collapse text-left">
          <thead>
            {table.getHeaderGroups().map((group) => (
              <tr key={group.id} className="border-b">
                {group.headers.map((header) => (
                  <th key={header.id} className="py-2">
                    {header.isPlaceholder ? null : (
                      <table.FlexRender header={header} />
                    )}
                  </th>
                ))}
              </tr>
            ))}
          </thead>
          <tbody>
            {table.getRowModel().rows.map((row) => (
              <tr key={row.id} className="border-b">
                {row.getAllCells().map((cell) => (
                  <td key={cell.id} className="py-2">
                    <table.FlexRender cell={cell} />
                  </td>
                ))}
                <td className="flex gap-2 py-2">
                  {isParcela(row.original) ? null : (
                    <>
                      <button
                        type="button"
                        className="rounded border px-2 py-1"
                        onClick={() => startEdit(row.original)}
                      >
                        Editar
                      </button>
                      <button
                        type="button"
                        className="rounded border px-2 py-1"
                        onClick={() => handleDelete(row.original.id)}
                      >
                        Excluir
                      </button>
                    </>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
