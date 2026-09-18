"use client";

import { useEffect, useState } from "react";

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
    <div className="flex flex-1 flex-col gap-6 p-6">
      <h1 className="text-2xl font-semibold">Painel</h1>
      {error ? <p role="alert">{error}</p> : null}
      {summary ? (
        <>
          <p>{formatPeriod(summary.period)}</p>
          <dl className="flex flex-wrap gap-8">
            <div>
              <dt>Ganhos</dt>
              <dd>{formatBRL(summary.income_cents)}</dd>
            </div>
            <div>
              <dt>Gastos</dt>
              <dd>{formatBRL(summary.expense_cents)}</dd>
            </div>
            <div>
              <dt>Saldo</dt>
              <dd>{formatBRL(summary.balance_cents)}</dd>
            </div>
          </dl>
          {summary.categories.length > 0 ? (
            <table>
              <thead>
                <tr>
                  <th>Categoria</th>
                  <th>Ganhos</th>
                  <th>Gastos</th>
                  <th>Saldo</th>
                </tr>
              </thead>
              <tbody>
                {summary.categories.map((row) => (
                  <tr key={row.category_id}>
                    <td>{row.category_name}</td>
                    <td>{formatBRL(row.income_cents)}</td>
                    <td>{formatBRL(row.expense_cents)}</td>
                    <td>{formatBRL(row.balance_cents)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          ) : null}
        </>
      ) : null}
    </div>
  );
}
