"use client";

import { useEffect, useState, type FormEvent } from "react";

type Category = {
  id: number;
  name: string;
};

export default function CategoriesPage() {
  const [categories, setCategories] = useState<Category[]>([]);
  const [name, setName] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [editingName, setEditingName] = useState("");

  useEffect(() => {
    async function loadCategories() {
      const response = await fetch("/api/categories");
      if (!response.ok) {
        setError("Não foi possível carregar as categorias.");
        return;
      }
      const data = (await response.json()) as Category[];
      setCategories(data);
    }
    void loadCategories();
  }, []);

  async function handleCreate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);

    const response = await fetch("/api/categories", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name }),
    });

    if (!response.ok) {
      const data = await response.json().catch(() => null);
      setError(data?.message ?? data?.error ?? "Não foi possível criar a categoria.");
      return;
    }

    const created = (await response.json()) as Category;
    setCategories((current) =>
      [...current, created].sort((a, b) => a.name.localeCompare(b.name)),
    );
    setName("");
  }

  async function handleRename(event: FormEvent<HTMLFormElement>, id: number) {
    event.preventDefault();
    setError(null);

    const response = await fetch(`/api/categories/${id}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name: editingName }),
    });

    if (!response.ok) {
      const data = await response.json().catch(() => null);
      setError(data?.message ?? data?.error ?? "Não foi possível renomear a categoria.");
      return;
    }

    const updated = (await response.json()) as Category;
    setCategories((current) =>
      current
        .map((category) => (category.id === id ? updated : category))
        .sort((a, b) => a.name.localeCompare(b.name)),
    );
    setEditingId(null);
  }

  async function handleDelete(id: number) {
    setError(null);

    const response = await fetch(`/api/categories/${id}`, {
      method: "DELETE",
    });

    if (!response.ok) {
      if (response.status === 409) {
        setError(
          "Não foi possível excluir: a categoria ainda tem lançamentos. Reclassifique-os antes.",
        );
        return;
      }
      setError("Não foi possível excluir a categoria.");
      return;
    }

    setCategories((current) => current.filter((category) => category.id !== id));
  }

  return (
    <div className="mx-auto flex w-full max-w-xl flex-col gap-6 p-6">
      <h1 className="text-2xl font-semibold">Categorias</h1>

      <form onSubmit={handleCreate} className="flex flex-col gap-2 sm:flex-row">
        <label className="flex flex-1 flex-col gap-1">
          Nome
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="rounded border px-3 py-2"
          />
        </label>
        <button
          type="submit"
          className="self-end rounded bg-foreground px-4 py-2 text-background"
        >
          Criar
        </button>
      </form>

      {error && <p className="text-red-600">{error}</p>}

      <ul className="flex flex-col gap-3">
        {categories.map((category) => (
          <li
            key={category.id}
            className="flex flex-wrap items-center gap-2 border-b pb-2"
          >
            {editingId === category.id ? (
              <form
                onSubmit={(event) => handleRename(event, category.id)}
                className="flex flex-1 flex-wrap items-center gap-2"
              >
                <label className="flex flex-1 flex-col gap-1">
                  Novo nome
                  <input
                    value={editingName}
                    onChange={(e) => setEditingName(e.target.value)}
                    className="rounded border px-3 py-2"
                  />
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
                <span className="flex-1">{category.name}</span>
                <button
                  type="button"
                  className="rounded border px-3 py-1"
                  onClick={() => {
                    setEditingId(category.id);
                    setEditingName(category.name);
                  }}
                >
                  Renomear
                </button>
                <button
                  type="button"
                  className="rounded border px-3 py-1"
                  onClick={() => handleDelete(category.id)}
                >
                  {`Excluir ${category.name}`}
                </button>
              </>
            )}
          </li>
        ))}
      </ul>
    </div>
  );
}
