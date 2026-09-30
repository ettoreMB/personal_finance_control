"use client";

import { useEffect, useState, type FormEvent } from "react";
import { PageHeader } from "@/components/page-header";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

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
    <div className="mx-auto flex w-full max-w-2xl flex-1 flex-col gap-6 p-4 md:p-6">
      <PageHeader
        title="Categorias"
        description="Casa, carro, comida e o que mais você quiser classificar."
      />

      <Card>
        <CardHeader>
          <CardTitle>Nova categoria</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleCreate} className="flex flex-col gap-3 sm:flex-row sm:items-end">
            <Label className="flex-1 flex-col items-stretch gap-2">
              Nome
              <Input value={name} onChange={(e) => setName(e.target.value)} />
            </Label>
            <Button type="submit">Criar</Button>
          </form>
        </CardContent>
      </Card>

      {error ? (
        <p role="alert" className="text-sm text-destructive">
          {error}
        </p>
      ) : null}

      <Card className="py-0">
        <ul className="flex flex-col">
          {categories.map((category) => (
            <li
              key={category.id}
              className="flex flex-wrap items-center gap-2 border-b px-4 py-3 last:border-b-0"
            >
              {editingId === category.id ? (
                <form
                  onSubmit={(event) => handleRename(event, category.id)}
                  className="flex flex-1 flex-wrap items-end gap-2"
                >
                  <Label className="min-w-40 flex-1 flex-col items-stretch gap-2">
                    Novo nome
                    <Input
                      value={editingName}
                      onChange={(e) => setEditingName(e.target.value)}
                    />
                  </Label>
                  <Button type="submit">Salvar</Button>
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => setEditingId(null)}
                  >
                    Cancelar
                  </Button>
                </form>
              ) : (
                <>
                  <span className="flex-1 font-medium">{category.name}</span>
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => {
                      setEditingId(category.id);
                      setEditingName(category.name);
                    }}
                  >
                    Renomear
                  </Button>
                  <Button
                    type="button"
                    variant="destructive"
                    onClick={() => handleDelete(category.id)}
                  >
                    {`Excluir ${category.name}`}
                  </Button>
                </>
              )}
            </li>
          ))}
        </ul>
      </Card>
    </div>
  );
}
