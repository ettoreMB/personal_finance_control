import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import DashboardPage from "./page";

const emptySummary = {
  period: { kind: "month", year: 2026, month: 9 },
  income_cents: 0,
  expense_cents: 0,
  balance_cents: 0,
  categories: [],
};

const filledSummary = {
  period: { kind: "month", year: 2026, month: 9 },
  income_cents: 10000,
  expense_cents: 15000,
  balance_cents: -5000,
  categories: [
    {
      category_id: 2,
      category_name: "casa",
      income_cents: 0,
      expense_cents: 10000,
      balance_cents: -10000,
    },
    {
      category_id: 3,
      category_name: "comida",
      income_cents: 10000,
      expense_cents: 5000,
      balance_cents: 5000,
    },
  ],
};

describe("DashboardPage", () => {
  beforeEach(() => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: RequestInfo | URL) => {
        const url = String(input);
        if (url === "/api/summary") {
          return new Response(JSON.stringify(emptySummary), { status: 200 });
        }
        return new Response(null, { status: 404 });
      }),
    );
  });

  it("shows the period and zeros for an empty month", async () => {
    render(<DashboardPage />);

    expect(screen.getByRole("heading", { name: /painel/i })).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText(/setembro de 2026/i)).toBeInTheDocument();
    });
    expect(screen.getByText(/ganhos/i)).toBeInTheDocument();
    expect(screen.getByText(/gastos/i)).toBeInTheDocument();
    expect(screen.getAllByText("R$ 0,00").length).toBeGreaterThanOrEqual(3);
    expect(screen.queryByText(/casa/i)).not.toBeInTheDocument();
  });

  it("shows a load error instead of zeros", async () => {
    vi.mocked(fetch).mockImplementation(async () => {
      return new Response(null, { status: 500 });
    });

    render(<DashboardPage />);

    await waitFor(() => {
      expect(
        screen.getByText(/não foi possível carregar o resumo/i),
      ).toBeInTheDocument();
    });
    expect(screen.queryByText("R$ 0,00")).not.toBeInTheDocument();
  });

  it("shows totals, negative saldo, and category breakdown in BRL", async () => {
    vi.mocked(fetch).mockImplementation(async (input: RequestInfo | URL) => {
      if (String(input) === "/api/summary") {
        return new Response(JSON.stringify(filledSummary), { status: 200 });
      }
      return new Response(null, { status: 404 });
    });

    render(<DashboardPage />);

    await waitFor(() => {
      expect(screen.getByText("casa")).toBeInTheDocument();
    });
    expect(screen.getByText("comida")).toBeInTheDocument();
    expect(screen.getByText("R$ 150,00")).toBeInTheDocument();
    expect(screen.getByText("R$ -50,00")).toBeInTheDocument();
    expect(screen.getAllByText("R$ 100,00").length).toBeGreaterThanOrEqual(1);
  });
});
