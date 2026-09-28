"use client";

import { useEffect, useState, type FormEvent } from "react";
import { NativeSelect } from "@/components/native-select";
import { PageHeader } from "@/components/page-header";
import {
  formatReaisFromCents,
  maskReaisInput,
  parseReaisToCents,
} from "@/lib/money";
import { Badge } from "@/components/ui/badge";
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
    setEditAmount(formatReaisFromCents(purchase.amount_cents));
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
    <div className="mx-auto flex w-full max-w-3xl flex-1 flex-col gap-6 p-4 md:p-6">
      <PageHeader
        title="Compras"
        description="Compra parcelada no cartão. As parcelas entram no mês de cada uma."
      />

      <Card>
        <CardHeader>
          <CardTitle>Registrar compra</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleCreate} className="grid gap-4 sm:grid-cols-2">
            <Label className="flex-col items-stretch gap-2 sm:col-span-2">
              Descrição
              <Input
                value={description}
                onChange={(e) => setDescription(e.target.value)}
              />
            </Label>
            <Label className="flex-col items-stretch gap-2">
              Data
              <Input
                type="date"
                value={date}
                onChange={(e) => setDate(e.target.value)}
              />
            </Label>
            <Label className="flex-col items-stretch gap-2">
              Valor
              <Input
                value={amount}
                onChange={(e) => setAmount(maskReaisInput(e.target.value))}
                inputMode="numeric"
                autoComplete="off"
              />
            </Label>
            <Label className="flex-col items-stretch gap-2">
              Parcelas
              <Input
                value={installments}
                onChange={(e) => setInstallments(e.target.value)}
                inputMode="numeric"
              />
            </Label>
            <Label className="flex-col items-stretch gap-2">
              Categoria
              <NativeSelect
                value={categoryId}
                onChange={(e) => setCategoryId(e.target.value)}
              >
                <option value="">Selecione</option>
                {categories.map((category) => (
                  <option key={category.id} value={category.id}>
                    {category.name}
                  </option>
                ))}
              </NativeSelect>
            </Label>
            <Button type="submit" className="sm:col-span-2">
              Registrar compra
            </Button>
          </form>
        </CardContent>
      </Card>

      {error ? (
        <p role="alert" className="text-sm text-destructive">
          {error}
        </p>
      ) : null}

      {purchases.length === 0 && !error ? (
        <p className="text-sm text-muted-foreground">Nenhuma compra ainda.</p>
      ) : (
        <ul className="flex flex-col gap-3">
          {purchases.map((purchase) => (
            <li key={purchase.id}>
              <Card>
                <CardContent>
                  {editingId === purchase.id ? (
                    <form
                      onSubmit={(event) => handleEdit(event, purchase.id)}
                      className="grid gap-4 sm:grid-cols-2"
                    >
                      <Label className="flex-col items-stretch gap-2 sm:col-span-2">
                        Descrição
                        <Input
                          value={editDescription}
                          onChange={(e) => setEditDescription(e.target.value)}
                        />
                      </Label>
                      <Label className="flex-col items-stretch gap-2">
                        Valor
                        <Input
                          value={editAmount}
                          onChange={(e) =>
                            setEditAmount(maskReaisInput(e.target.value))
                          }
                          inputMode="numeric"
                          autoComplete="off"
                        />
                      </Label>
                      <Label className="flex-col items-stretch gap-2">
                        Categoria
                        <NativeSelect
                          value={editCategoryId}
                          onChange={(e) => setEditCategoryId(e.target.value)}
                        >
                          {categories.map((category) => (
                            <option key={category.id} value={category.id}>
                              {category.name}
                            </option>
                          ))}
                        </NativeSelect>
                      </Label>
                      <div className="flex gap-2 sm:col-span-2">
                        <Button type="submit">Salvar</Button>
                        <Button
                          type="button"
                          variant="outline"
                          onClick={() => setEditingId(null)}
                        >
                          Cancelar
                        </Button>
                      </div>
                    </form>
                  ) : (
                    <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                      <div className="flex flex-col gap-1">
                        <p className="font-medium">{purchase.description}</p>
                        <p className="text-sm text-muted-foreground">
                          {formatBRL(purchase.amount_cents)} · {purchase.purchase_date} ·{" "}
                          {purchase.category_name}
                        </p>
                      </div>
                      <div className="flex items-center gap-2">
                        <Badge variant="secondary">{purchase.installment_count}x</Badge>
                        <Button
                          type="button"
                          variant="outline"
                          size="sm"
                          onClick={() => startEdit(purchase)}
                        >
                          Editar
                        </Button>
                        <Button
                          type="button"
                          variant="destructive"
                          size="sm"
                          onClick={() => handleUndo(purchase.id)}
                        >
                          Desfazer
                        </Button>
                      </div>
                    </div>
                  )}
                </CardContent>
              </Card>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
