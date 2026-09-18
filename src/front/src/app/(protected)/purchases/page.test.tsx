import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import PurchasesPage from "./page";

const categories = [
  { id: 1, name: "carro" },
  { id: 2, name: "casa" },
  { id: 3, name: "comida" },
  { id: 4, name: "lazer" },
];

const createdPurchase = {
  id: 1,
  description: "TV Samsung",
  purchase_date: "2026-01-31",
  amount_cents: 10000,
  installment_count: 3,
  category_id: 2,
  category_name: "casa",
  category: { id: 2, name: "casa" },
  installments: [
    { id: 10, installment_number: 1, amount_cents: 3333, entry_date: "2026-01-31", category_id: 2 },
    { id: 11, installment_number: 2, amount_cents: 3333, entry_date: "2026-02-28", category_id: 2 },
    { id: 12, installment_number: 3, amount_cents: 3334, entry_date: "2026-03-31", category_id: 2 },
  ],
};

describe("PurchasesPage", () => {
  beforeEach(() => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
        const url = String(input);
        const method = init?.method ?? "GET";

        if (url === "/api/categories" && method === "GET") {
          return new Response(JSON.stringify(categories), { status: 200 });
        }
        if (url === "/api/purchases" && method === "GET") {
          return new Response(JSON.stringify([]), { status: 200 });
        }
        if (url === "/api/purchases" && method === "POST") {
          return new Response(JSON.stringify(createdPurchase), { status: 201 });
        }
        return new Response(null, { status: 404 });
      }),
    );
  });

  it("shows an empty state", async () => {
    render(<PurchasesPage />);

    expect(screen.getByRole("heading", { name: /compras/i })).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText(/nenhuma compra ainda/i)).toBeInTheDocument();
    });
  });

  it("creates a purchase and shows it in the list", async () => {
    render(<PurchasesPage />);

    await waitFor(() => {
      expect(screen.getByLabelText("Descrição")).toBeInTheDocument();
    });

    fireEvent.change(screen.getByLabelText("Descrição"), {
      target: { value: "TV Samsung" },
    });
    fireEvent.change(screen.getByLabelText("Data"), {
      target: { value: "2026-01-31" },
    });
    fireEvent.change(screen.getByLabelText("Valor"), {
      target: { value: "100,00" },
    });
    fireEvent.change(screen.getByLabelText("Parcelas"), {
      target: { value: "3" },
    });
    fireEvent.change(screen.getByLabelText("Categoria"), {
      target: { value: "2" },
    });
    fireEvent.click(screen.getByRole("button", { name: /registrar compra/i }));

    await waitFor(() => {
      expect(screen.getByText("TV Samsung")).toBeInTheDocument();
    });
    expect(screen.getByText(/R\$ 100,00/)).toBeInTheDocument();
    expect(screen.queryByText(/nenhuma compra ainda/i)).not.toBeInTheDocument();
  });

  it("shows a validation error from the API", async () => {
    vi.mocked(fetch).mockImplementation(
      async (input: RequestInfo | URL, init?: RequestInit) => {
        const url = String(input);
        const method = init?.method ?? "GET";
        if (url === "/api/categories") {
          return new Response(JSON.stringify(categories), { status: 200 });
        }
        if (url === "/api/purchases" && method === "GET") {
          return new Response(JSON.stringify([]), { status: 200 });
        }
        if (url === "/api/purchases" && method === "POST") {
          return new Response(
            JSON.stringify({ error: "description is required" }),
            { status: 400 },
          );
        }
        return new Response(null, { status: 404 });
      },
    );

    render(<PurchasesPage />);

    await waitFor(() => {
      expect(screen.getByLabelText("Descrição")).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole("button", { name: /registrar compra/i }));

    await waitFor(() => {
      expect(screen.getByText(/description is required/i)).toBeInTheDocument();
    });
  });

  it("edits description, amount and category", async () => {
    const updated = {
      ...createdPurchase,
      description: "TV 55",
      amount_cents: 12000,
      category_id: 4,
      category_name: "lazer",
    };

    vi.mocked(fetch).mockImplementation(
      async (input: RequestInfo | URL, init?: RequestInit) => {
        const url = String(input);
        const method = init?.method ?? "GET";
        if (url === "/api/categories") {
          return new Response(JSON.stringify(categories), { status: 200 });
        }
        if (url === "/api/purchases" && method === "GET") {
          return new Response(JSON.stringify([createdPurchase]), { status: 200 });
        }
        if (url === "/api/purchases/1" && method === "PATCH") {
          return new Response(JSON.stringify(updated), { status: 200 });
        }
        return new Response(null, { status: 404 });
      },
    );

    render(<PurchasesPage />);

    await waitFor(() => {
      expect(screen.getByText("TV Samsung")).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole("button", { name: /editar/i }));
    fireEvent.change(screen.getAllByLabelText("Descrição")[1], {
      target: { value: "TV 55" },
    });
    fireEvent.click(screen.getByRole("button", { name: /salvar/i }));

    await waitFor(() => {
      expect(screen.getByText("TV 55")).toBeInTheDocument();
    });
  });

  it("undoes a purchase and shows 409 when blocked", async () => {
    vi.mocked(fetch).mockImplementation(
      async (input: RequestInfo | URL, init?: RequestInit) => {
        const url = String(input);
        const method = init?.method ?? "GET";
        if (url === "/api/categories") {
          return new Response(JSON.stringify(categories), { status: 200 });
        }
        if (url === "/api/purchases" && method === "GET") {
          return new Response(JSON.stringify([createdPurchase]), { status: 200 });
        }
        if (url === "/api/purchases/1" && method === "DELETE") {
          return new Response(
            JSON.stringify({ error: "cannot undo a compra with past parcelas" }),
            { status: 409 },
          );
        }
        return new Response(null, { status: 404 });
      },
    );

    render(<PurchasesPage />);

    await waitFor(() => {
      expect(screen.getByText("TV Samsung")).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole("button", { name: /desfazer/i }));

    await waitFor(() => {
      expect(screen.getByText(/cannot undo a compra with past parcelas/i)).toBeInTheDocument();
    });
  });
});
