"use client";

import { useEffect, useState, type FormEvent } from "react";
import {
  createColumnHelper,
  tableFeatures,
  useTable,
} from "@tanstack/react-table";
import { NativeSelect } from "@/components/native-select";
import { PageHeader } from "@/components/page-header";
import {
  formatReaisFromCents,
  maskReaisInput,
  parseReaisToCents,
} from "@/lib/money";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

type Category = {
  id: number;
  name: string;
};

type Entry = {
  id: number;
  type: "income" | "expense";
  amount_cents: number;
  entry_date: string;
  description: string;
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
    cell: ({ getValue }) =>
      getValue() === "income" ? (
        <Badge variant="secondary">Ganho</Badge>
      ) : (
        <Badge variant="outline">Gasto</Badge>
      ),
  }),
  helper.accessor("amount_cents", {
    header: "Valor",
    cell: ({ getValue }) => formatBRL(getValue()),
  }),
  helper.accessor("entry_date", { header: "Data" }),
  helper.accessor((row) => row.category?.name ?? "", {
    id: "category",
    header: "Categoria",
  }),
  helper.accessor("description", { header: "Descrição" }),
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
  const [description, setDescription] = useState("");
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
    setDescription("");
    setCategoryId("");
    setEditingId(null);
  }

  function startEdit(entry: Entry) {
    setEditingId(entry.id);
    setType(entry.type);
    setAmount(formatReaisFromCents(entry.amount_cents));
    setDate(entry.entry_date);
    setDescription(entry.description);
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
      description,
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
    <div className="mx-auto flex w-full min-w-0 max-w-4xl flex-1 flex-col gap-6 p-4 md:p-6">
      <PageHeader
        title="Lançamentos"
        description="Ganhos e gastos avulsos. Parcela de compra aparece aqui, mas se edita na compra."
      />

      <Card>
        <CardHeader>
          <CardTitle>{editingId === null ? "Novo lançamento" : "Editar lançamento"}</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="grid gap-4 sm:grid-cols-2">
            <Label className="flex-col items-stretch gap-2">
              Tipo
              <NativeSelect
                value={type}
                onChange={(e) => setType(e.target.value as "income" | "expense")}
              >
                <option value="expense">Gasto</option>
                <option value="income">Ganho</option>
              </NativeSelect>
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
              Data
              <Input
                type="date"
                value={date}
                onChange={(e) => setDate(e.target.value)}
              />
            </Label>
            <Label className="flex-col items-stretch gap-2 sm:col-span-2">
              Descrição
              <Input
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Ex.: água, luz, internet"
                autoComplete="off"
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
            <div className="flex gap-2 sm:col-span-2">
              <Button type="submit">
                {editingId === null ? "Lançar" : "Salvar"}
              </Button>
              {editingId !== null ? (
                <Button type="button" variant="outline" onClick={resetForm}>
                  Cancelar
                </Button>
              ) : null}
            </div>
          </form>
        </CardContent>
      </Card>

      {error ? (
        <p role="alert" className="text-sm text-destructive">
          {error}
        </p>
      ) : null}

      {entries.length === 0 && !error ? (
        <p className="text-sm text-muted-foreground">Nenhum lançamento ainda.</p>
      ) : entries.length === 0 ? null : (
        <Card className="py-0">
          <Table>
            <TableHeader>
              {table.getHeaderGroups().map((group) => (
                <TableRow key={group.id}>
                  {group.headers.map((header) => (
                    <TableHead key={header.id}>
                      {header.isPlaceholder ? null : (
                        <table.FlexRender header={header} />
                      )}
                    </TableHead>
                  ))}
                  <TableHead className="text-right">Ações</TableHead>
                </TableRow>
              ))}
            </TableHeader>
            <TableBody>
              {table.getRowModel().rows.map((row) => (
                <TableRow key={row.id}>
                  {row.getAllCells().map((cell) => (
                    <TableCell key={cell.id} className="tabular-nums">
                      <table.FlexRender cell={cell} />
                    </TableCell>
                  ))}
                  <TableCell>
                    {isParcela(row.original) ? null : (
                      <div className="flex justify-end gap-2">
                        <Button
                          type="button"
                          variant="outline"
                          size="sm"
                          onClick={() => startEdit(row.original)}
                        >
                          Editar
                        </Button>
                        <Button
                          type="button"
                          variant="destructive"
                          size="sm"
                          onClick={() => handleDelete(row.original.id)}
                        >
                          Excluir
                        </Button>
                      </div>
                    )}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </Card>
      )}
    </div>
  );
}
