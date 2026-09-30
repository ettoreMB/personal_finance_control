import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import EntriesPage from "./page";

const categories = [
  { id: 1, name: "carro" },
  { id: 2, name: "casa" },
  { id: 3, name: "comida" },
  { id: 4, name: "lazer" },
];

const createdEntry = {
  id: 1,
  type: "expense",
  amount_cents: 4500,
  entry_date: "2026-01-15",
  description: "água",
  category_id: 2,
  category: { id: 2, name: "casa" },
};

describe("EntriesPage", () => {
  beforeEach(() => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
        const url = String(input);
        const method = init?.method ?? "GET";

        if (url === "/api/categories" && method === "GET") {
          return new Response(JSON.stringify(categories), { status: 200 });
        }
        if (url === "/api/entries" && method === "GET") {
          return new Response(JSON.stringify([]), { status: 200 });
        }
        if (url === "/api/entries" && method === "POST") {
          return new Response(JSON.stringify(createdEntry), { status: 201 });
        }
        return new Response(null, { status: 404 });
      }),
    );
  });

  it("shows a load error instead of the empty state", async () => {
    vi.mocked(fetch).mockImplementation(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url === "/api/categories") {
        return new Response(JSON.stringify(categories), { status: 200 });
      }
      if (url === "/api/entries") {
        return new Response(null, { status: 500 });
      }
      return new Response(null, { status: 404 });
    });

    render(<EntriesPage />);

    await waitFor(() => {
      expect(
        screen.getByText(/não foi possível carregar os lançamentos/i),
      ).toBeInTheDocument();
    });
    expect(screen.queryByText(/nenhum lançamento/i)).not.toBeInTheDocument();
  });

  it("shows an empty state", async () => {
    render(<EntriesPage />);

    expect(
      screen.getByRole("heading", { name: /lançamentos/i }),
    ).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText(/nenhum lançamento/i)).toBeInTheDocument();
    });
    expect(screen.queryByLabelText("Valor")).not.toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /novo lançamento/i }),
    ).toBeInTheDocument();
  });

  it("masks the amount as the user types", async () => {
    render(<EntriesPage />);

    fireEvent.click(screen.getByRole("button", { name: /novo lançamento/i }));
    const dialog = await screen.findByRole("dialog");

    fireEvent.change(within(dialog).getByLabelText("Valor"), {
      target: { value: "11050" },
    });

    expect(within(dialog).getByLabelText("Valor")).toHaveValue("110,50");
  });

  it("creates an expense and shows it in the list", async () => {
    render(<EntriesPage />);

    fireEvent.click(screen.getByRole("button", { name: /novo lançamento/i }));
    const dialog = await screen.findByRole("dialog");

    fireEvent.change(within(dialog).getByLabelText("Tipo"), {
      target: { value: "expense" },
    });
    fireEvent.change(within(dialog).getByLabelText("Valor"), {
      target: { value: "45,00" },
    });
    fireEvent.change(within(dialog).getByLabelText("Data"), {
      target: { value: "2026-01-15" },
    });
    fireEvent.change(within(dialog).getByLabelText("Descrição"), {
      target: { value: "água" },
    });
    fireEvent.change(within(dialog).getByLabelText("Categoria"), {
      target: { value: "2" },
    });
    fireEvent.click(within(dialog).getByRole("button", { name: /lançar/i }));

    await waitFor(() => {
      expect(screen.getByText("R$ 45,00")).toBeInTheDocument();
    });
    expect(screen.getByRole("cell", { name: "água" })).toBeInTheDocument();

    const postCall = vi.mocked(fetch).mock.calls.find((call) => {
      const [input, init] = call;
      return String(input) === "/api/entries" && init?.method === "POST";
    });
    expect(JSON.parse(String(postCall?.[1]?.body))).toMatchObject({
      description: "água",
      amount_cents: 4500,
      category_id: 2,
    });
    expect(screen.getByText("Gasto")).toBeInTheDocument();
    expect(screen.getByRole("cell", { name: "casa" })).toBeInTheDocument();
    expect(screen.queryByText(/nenhum lançamento/i)).not.toBeInTheDocument();
  });

  it("shows an error when the amount is invalid", async () => {
    vi.mocked(fetch).mockImplementation(
      async (input: RequestInfo | URL, init?: RequestInit) => {
        const url = String(input);
        const method = init?.method ?? "GET";
        if (url === "/api/categories") {
          return new Response(JSON.stringify(categories), { status: 200 });
        }
        if (url === "/api/entries" && method === "GET") {
          return new Response(JSON.stringify([]), { status: 200 });
        }
        if (url === "/api/entries" && method === "POST") {
          return new Response(
            JSON.stringify({ error: "amount_cents must be greater than 0" }),
            { status: 400 },
          );
        }
        return new Response(null, { status: 404 });
      },
    );

    render(<EntriesPage />);

    fireEvent.click(screen.getByRole("button", { name: /novo lançamento/i }));
    const dialog = await screen.findByRole("dialog");

    fireEvent.change(within(dialog).getByLabelText("Valor"), {
      target: { value: "0" },
    });
    fireEvent.change(within(dialog).getByLabelText("Data"), {
      target: { value: "2026-01-15" },
    });
    fireEvent.change(within(dialog).getByLabelText("Categoria"), {
      target: { value: "2" },
    });
    fireEvent.click(within(dialog).getByRole("button", { name: /lançar/i }));

    await waitFor(() => {
      expect(
        screen.getByText(/amount_cents must be greater than 0/i),
      ).toBeInTheDocument();
    });
  });

  it("edits an existing entry", async () => {
    const updated = {
      ...createdEntry,
      amount_cents: 9900,
      category_id: 3,
      category: { id: 3, name: "comida" },
    };

    vi.mocked(fetch).mockImplementation(
      async (input: RequestInfo | URL, init?: RequestInit) => {
        const url = String(input);
        const method = init?.method ?? "GET";
        if (url === "/api/categories") {
          return new Response(JSON.stringify(categories), { status: 200 });
        }
        if (url === "/api/entries" && method === "GET") {
          return new Response(JSON.stringify([createdEntry]), { status: 200 });
        }
        if (url === "/api/entries/1" && method === "PATCH") {
          return new Response(JSON.stringify(updated), { status: 200 });
        }
        return new Response(null, { status: 404 });
      },
    );

    render(<EntriesPage />);

    await waitFor(() => {
      expect(screen.getByText("R$ 45,00")).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole("button", { name: /editar/i }));
    const dialog = await screen.findByRole("dialog");
    fireEvent.change(within(dialog).getByLabelText("Valor"), {
      target: { value: "99,00" },
    });
    fireEvent.change(within(dialog).getByLabelText("Categoria"), {
      target: { value: "3" },
    });
    fireEvent.click(within(dialog).getByRole("button", { name: /salvar/i }));

    await waitFor(() => {
      expect(screen.getByText("R$ 99,00")).toBeInTheDocument();
    });
    expect(screen.getByRole("cell", { name: "comida" })).toBeInTheDocument();
  });

  it("shows parcela label and hides edit/delete for installment rows", async () => {
    const parcela = {
      ...createdEntry,
      purchase_id: 9,
      installment_number: 2,
      purchase: { id: 9, description: "TV Samsung", installment_count: 6 },
    };

    vi.mocked(fetch).mockImplementation(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url === "/api/categories") {
        return new Response(JSON.stringify(categories), { status: 200 });
      }
      if (url === "/api/entries") {
        return new Response(JSON.stringify([parcela]), { status: 200 });
      }
      return new Response(null, { status: 404 });
    });

    render(<EntriesPage />);

    await waitFor(() => {
      expect(screen.getByText("2/6 · TV Samsung")).toBeInTheDocument();
    });
    expect(screen.queryByRole("button", { name: /editar/i })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /excluir/i })).not.toBeInTheDocument();
  });

  it("deletes an existing entry", async () => {
    vi.mocked(fetch).mockImplementation(
      async (input: RequestInfo | URL, init?: RequestInit) => {
        const url = String(input);
        const method = init?.method ?? "GET";
        if (url === "/api/categories") {
          return new Response(JSON.stringify(categories), { status: 200 });
        }
        if (url === "/api/entries" && method === "GET") {
          return new Response(JSON.stringify([createdEntry]), { status: 200 });
        }
        if (url === "/api/entries/1" && method === "DELETE") {
          return new Response(null, { status: 204 });
        }
        return new Response(null, { status: 404 });
      },
    );

    render(<EntriesPage />);

    await waitFor(() => {
      expect(screen.getByText("R$ 45,00")).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole("button", { name: /excluir/i }));

    await waitFor(() => {
      expect(screen.getByText(/nenhum lançamento/i)).toBeInTheDocument();
    });
  });

  it("filters the table by month and category", async () => {
    const february = {
      ...createdEntry,
      id: 2,
      entry_date: "2026-02-02",
      description: "pizza",
      amount_cents: 2000,
      category_id: 3,
      category: { id: 3, name: "comida" },
    };

    vi.mocked(fetch).mockImplementation(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url === "/api/categories") {
        return new Response(JSON.stringify(categories), { status: 200 });
      }
      if (url === "/api/entries") {
        return new Response(JSON.stringify([createdEntry, february]), {
          status: 200,
        });
      }
      return new Response(null, { status: 404 });
    });

    render(<EntriesPage />);

    await waitFor(() => {
      expect(screen.getByText("pizza")).toBeInTheDocument();
    });
    expect(screen.getByText("água")).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText("Mês"), {
      target: { value: "2026-01" },
    });
    expect(screen.queryByText("pizza")).not.toBeInTheDocument();
    expect(screen.getByText("água")).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText("Mês"), { target: { value: "" } });
    fireEvent.change(screen.getByLabelText("Categoria"), {
      target: { value: "3" },
    });
    expect(screen.queryByText("água")).not.toBeInTheDocument();
    expect(screen.getByText("pizza")).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText("Mês"), {
      target: { value: "2026-01" },
    });
    expect(screen.getByText(/nenhum lançamento neste filtro/i)).toBeInTheDocument();
  });

  it("filters the table by gastos and recebimentos", async () => {
    const income = {
      ...createdEntry,
      id: 2,
      type: "income" as const,
      description: "salário",
      amount_cents: 100000,
    };

    vi.mocked(fetch).mockImplementation(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url === "/api/categories") {
        return new Response(JSON.stringify(categories), { status: 200 });
      }
      if (url === "/api/entries") {
        return new Response(JSON.stringify([createdEntry, income]), {
          status: 200,
        });
      }
      return new Response(null, { status: 404 });
    });

    render(<EntriesPage />);

    await waitFor(() => {
      expect(screen.getByText("salário")).toBeInTheDocument();
    });
    expect(screen.getByText("água")).toBeInTheDocument();
    expect(screen.getByText("R$ 45,00")).toHaveClass("text-destructive");
    expect(screen.getByText("R$ 1000,00")).toHaveClass("text-emerald-700");

    fireEvent.change(screen.getByLabelText("Tipo"), {
      target: { value: "expense" },
    });
    expect(screen.getByText("água")).toBeInTheDocument();
    expect(screen.queryByText("salário")).not.toBeInTheDocument();

    fireEvent.change(screen.getByLabelText("Tipo"), {
      target: { value: "income" },
    });
    expect(screen.queryByText("água")).not.toBeInTheDocument();
    expect(screen.getByText("salário")).toBeInTheDocument();
  });
});
