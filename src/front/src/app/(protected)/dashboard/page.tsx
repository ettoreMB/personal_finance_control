"use client";

import { useEffect, useState } from "react";
import { PageHeader } from "@/components/page-header";
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

type SummaryPeriod = {
  kind: string;
  year: number;
  month: number;
};

type SummaryCategory = {
  category_id: number;
  category_name: string;
  income_cents: number;
  expense_cents: number;
  balance_cents: number;
};

type Summary = {
  period: SummaryPeriod;
  income_cents: number;
  expense_cents: number;
  balance_cents: number;
  categories: SummaryCategory[];
};

function formatBRL(cents: number): string {
  return `R$ ${(cents / 100).toFixed(2).replace(".", ",")}`;
}

function formatPeriod(period: SummaryPeriod): string {
  const date = new Date(period.year, period.month - 1, 1);
  return new Intl.DateTimeFormat("pt-BR", {
    month: "long",
    year: "numeric",
  }).format(date);
}

function moneyClass(cents: number): string {
  if (cents < 0) {
    return "text-destructive";
  }
  if (cents > 0) {
    return "text-emerald-700 dark:text-emerald-400";
  }
  return "";
}

export default function DashboardPage() {
  const [summary, setSummary] = useState<Summary | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function load() {
      const response = await fetch("/api/summary");
      if (!response.ok) {
        setError("Não foi possível carregar o resumo.");
        return;
      }
      setSummary((await response.json()) as Summary);
    }
    void load();
  }, []);

  return (
    <div className="@container/main flex min-w-0 flex-1 flex-col gap-4 p-4 md:gap-6 md:p-6">
      <PageHeader
        title="Painel"
        description={summary ? formatPeriod(summary.period) : undefined}
      />
      {error ? (
        <p role="alert" className="text-sm text-destructive">
          {error}
        </p>
      ) : null}
      {summary ? (
        <>
          <div className="grid grid-cols-1 gap-4 @xl/main:grid-cols-3">
            <Card className="bg-gradient-to-t from-primary/5 to-card shadow-xs">
              <CardHeader>
                <CardDescription>Ganhos</CardDescription>
                <CardTitle
                  className={`text-2xl tabular-nums ${moneyClass(summary.income_cents)}`}
                >
                  {formatBRL(summary.income_cents)}
                </CardTitle>
              </CardHeader>
            </Card>
            <Card className="bg-gradient-to-t from-primary/5 to-card shadow-xs">
              <CardHeader>
                <CardDescription>Gastos</CardDescription>
                <CardTitle className="text-2xl tabular-nums">
                  {formatBRL(summary.expense_cents)}
                </CardTitle>
              </CardHeader>
            </Card>
            <Card className="bg-gradient-to-t from-primary/5 to-card shadow-xs">
              <CardHeader>
                <CardDescription>Saldo</CardDescription>
                <CardTitle
                  className={`text-2xl tabular-nums ${moneyClass(summary.balance_cents)}`}
                >
                  {formatBRL(summary.balance_cents)}
                </CardTitle>
              </CardHeader>
            </Card>
          </div>
          {summary.categories.length > 0 ? (
            <Card className="py-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Categoria</TableHead>
                    <TableHead className="text-right">Ganhos</TableHead>
                    <TableHead className="text-right">Gastos</TableHead>
                    <TableHead className="text-right">Saldo</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {summary.categories.map((row) => (
                    <TableRow key={row.category_id}>
                      <TableCell className="font-medium">
                        {row.category_name}
                      </TableCell>
                      <TableCell
                        className={`text-right tabular-nums ${moneyClass(row.income_cents)}`}
                      >
                        {formatBRL(row.income_cents)}
                      </TableCell>
                      <TableCell className="text-right tabular-nums">
                        {formatBRL(row.expense_cents)}
                      </TableCell>
                      <TableCell
                        className={`text-right tabular-nums ${moneyClass(row.balance_cents)}`}
                      >
                        {formatBRL(row.balance_cents)}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </Card>
          ) : null}
        </>
      ) : null}
    </div>
  );
}
