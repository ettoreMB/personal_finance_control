"use client";

import { useEffect, useState } from "react";
import { ChevronLeft, ChevronRight, Minus, TrendingDown, TrendingUp } from "lucide-react";

import { ExpenseCharts } from "@/components/expense-charts";
import { PageHeader } from "@/components/page-header";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
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
import { formatReaisFromCents } from "@/lib/money";

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
  const sign = cents < 0 ? "-" : "";
  return `R$ ${sign}${formatReaisFromCents(cents)}`;
}

function formatPeriod(period: SummaryPeriod): string {
  const date = new Date(period.year, period.month - 1, 1);
  return new Intl.DateTimeFormat("pt-BR", {
    month: "long",
    year: "numeric",
  }).format(date);
}

function monthInputValue(period: SummaryPeriod): string {
  return `${period.year}-${String(period.month).padStart(2, "0")}`;
}

function shiftMonth(value: string, delta: number): string {
  const [year, month] = value.split("-").map(Number);
  const date = new Date(Date.UTC(year, month - 1 + delta, 1));
  return `${date.getUTCFullYear()}-${String(date.getUTCMonth() + 1).padStart(2, "0")}`;
}

function summaryPath(monthValue: string | null): string {
  if (!monthValue) {
    return "/api/summary";
  }
  const [year, month] = monthValue.split("-");
  return `/api/summary?year=${year}&month=${Number(month)}`;
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
  const [selectedMonth, setSelectedMonth] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const controller = new AbortController();
    let ignore = false;
    async function load() {
      const response = await fetch(summaryPath(selectedMonth), {
        signal: controller.signal,
      });
      if (ignore) {
        return;
      }
      if (!response.ok) {
        setError("Não foi possível carregar o resumo.");
        return;
      }
      setError(null);
      setSummary((await response.json()) as Summary);
    }
    void load().catch((cause: unknown) => {
      if (ignore || (cause instanceof DOMException && cause.name === "AbortError")) {
        return;
      }
      setError("Não foi possível carregar o resumo.");
    });
    return () => {
      ignore = true;
      controller.abort();
    };
  }, [selectedMonth]);

  function chooseMonth(value: string) {
    setSummary(null);
    setError(null);
    setSelectedMonth(value);
  }

  const monthValue = selectedMonth ?? (summary ? monthInputValue(summary.period) : "");
  const periodLabel = summary ? formatPeriod(summary.period) : "";
  const expenseCategories =
    summary?.categories.filter((row) => row.expense_cents > 0).length ?? 0;

  return (
    <div className="@container/main flex min-w-0 flex-1 flex-col gap-4 p-4 md:gap-6 md:p-6">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <PageHeader
          title="Painel"
          description={
            summary
              ? `Resumo dos gastos em ${periodLabel}`
              : "Resumo dos gastos do mês"
          }
        />
        <div className="flex items-end gap-2">
          <Button
            type="button"
            variant="outline"
            size="icon"
            aria-label="Mês anterior"
            disabled={!monthValue}
            onClick={() => chooseMonth(shiftMonth(monthValue, -1))}
          >
            <ChevronLeft />
          </Button>
          <Label className="w-40 flex-col items-stretch gap-2">
            Mês
            <Input
              type="month"
              value={monthValue}
              onChange={(event) => {
                if (event.target.value) {
                  chooseMonth(event.target.value);
                }
              }}
            />
          </Label>
          <Button
            type="button"
            variant="outline"
            size="icon"
            aria-label="Próximo mês"
            disabled={!monthValue}
            onClick={() => chooseMonth(shiftMonth(monthValue, 1))}
          >
            <ChevronRight />
          </Button>
        </div>
      </div>
      {error ? (
        <p role="alert" className="text-sm text-destructive">
          {error}
        </p>
      ) : null}
      {summary ? (
        <>
          <div className="grid grid-cols-1 gap-4 @xl/main:grid-cols-3">
            <Card className="@container/card bg-gradient-to-t from-primary/5 to-card shadow-xs">
              <CardHeader>
                <CardDescription>Ganhos</CardDescription>
                <CardTitle
                  className={`text-2xl font-semibold tabular-nums @[250px]/card:text-3xl ${moneyClass(summary.income_cents)}`}
                >
                  {formatBRL(summary.income_cents)}
                </CardTitle>
              </CardHeader>
              <CardFooter className="text-sm text-muted-foreground">
                Entradas registradas no mês
              </CardFooter>
            </Card>
            <Card className="@container/card bg-gradient-to-t from-primary/5 to-card shadow-xs">
              <CardHeader>
                <CardDescription>Gastos</CardDescription>
                <CardTitle className="text-2xl font-semibold tabular-nums @[250px]/card:text-3xl">
                  {formatBRL(summary.expense_cents)}
                </CardTitle>
              </CardHeader>
              <CardFooter className="text-sm text-muted-foreground">
                {expenseCategories === 0
                  ? "Nenhum gasto neste mês"
                  : expenseCategories === 1
                    ? "1 categoria com gasto"
                    : `${expenseCategories} categorias com gasto`}
              </CardFooter>
            </Card>
            <Card className="@container/card bg-gradient-to-t from-primary/5 to-card shadow-xs">
              <CardHeader>
                <CardDescription>Saldo</CardDescription>
                <CardTitle
                  className={`text-2xl font-semibold tabular-nums @[250px]/card:text-3xl ${moneyClass(summary.balance_cents)}`}
                >
                  {formatBRL(summary.balance_cents)}
                </CardTitle>
              </CardHeader>
              <CardFooter className="gap-2 text-sm text-muted-foreground">
                {summary.balance_cents > 0 ? (
                  <TrendingUp className="size-4 text-emerald-700 dark:text-emerald-400" />
                ) : summary.balance_cents < 0 ? (
                  <TrendingDown className="size-4 text-destructive" />
                ) : (
                  <Minus className="size-4" />
                )}
                Ganhos menos gastos
              </CardFooter>
            </Card>
          </div>
          <ExpenseCharts
            slices={summary.categories.map((row) => ({
              category: row.category_name,
              cents: row.expense_cents,
            }))}
            periodLabel={periodLabel}
            formatBRL={formatBRL}
          />
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
