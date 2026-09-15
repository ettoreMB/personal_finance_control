import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import CategoriesPage from "./page";

const seed = [
  { id: 1, name: "carro" },
  { id: 2, name: "casa" },
  { id: 3, name: "comida" },
  { id: 4, name: "lazer" },
];

describe("CategoriesPage", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", vi.fn());
  });

  it("renders seeded categories", async () => {
    vi.mocked(fetch).mockResolvedValueOnce(
      new Response(JSON.stringify(seed), { status: 200 }),
    );

    render(<CategoriesPage />);

    expect(
      screen.getByRole("heading", { name: /categorias/i }),
    ).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText("casa")).toBeInTheDocument();
    });
    expect(screen.getByText("comida")).toBeInTheDocument();
  });

  it("shows an error when creating a category fails", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(
        new Response(JSON.stringify(seed), { status: 200 }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ message: "name is required" }), {
          status: 400,
        }),
      );

    render(<CategoriesPage />);

    await waitFor(() => {
      expect(screen.getByText("casa")).toBeInTheDocument();
    });

    fireEvent.change(screen.getByLabelText("Nome"), {
      target: { value: "   " },
    });
    fireEvent.click(screen.getByRole("button", { name: /criar/i }));

    await waitFor(() => {
      expect(screen.getByText("name is required")).toBeInTheDocument();
    });

    expect(fetch).toHaveBeenCalledWith(
      "/api/categories",
      expect.objectContaining({ method: "POST" }),
    );
  });

  it("adds a created category to the list", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(
        new Response(JSON.stringify(seed), { status: 200 }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ id: 5, name: "saúde" }), {
          status: 201,
        }),
      );

    render(<CategoriesPage />);

    await waitFor(() => {
      expect(screen.getByText("casa")).toBeInTheDocument();
    });

    fireEvent.change(screen.getByLabelText("Nome"), {
      target: { value: "saúde" },
    });
    fireEvent.click(screen.getByRole("button", { name: /criar/i }));

    await waitFor(() => {
      expect(screen.getByText("saúde")).toBeInTheDocument();
    });
  });

  it("renames a category", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(
        new Response(JSON.stringify(seed), { status: 200 }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ id: 2, name: "moradia" }), {
          status: 200,
        }),
      );

    render(<CategoriesPage />);

    await waitFor(() => {
      expect(screen.getByText("casa")).toBeInTheDocument();
    });

    fireEvent.click(screen.getAllByRole("button", { name: /renomear/i })[1]);
    fireEvent.change(screen.getByLabelText("Novo nome"), {
      target: { value: "moradia" },
    });
    fireEvent.click(screen.getByRole("button", { name: /salvar/i }));

    await waitFor(() => {
      expect(screen.getByText("moradia")).toBeInTheDocument();
    });
    expect(screen.queryByText("casa")).not.toBeInTheDocument();
  });

  it("removes a category after a successful delete", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(
        new Response(JSON.stringify(seed), { status: 200 }),
      )
      .mockResolvedValueOnce(new Response(null, { status: 204 }));

    render(<CategoriesPage />);

    await waitFor(() => {
      expect(screen.getByText("casa")).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole("button", { name: /excluir casa/i }));

    await waitFor(() => {
      expect(screen.queryByText("casa")).not.toBeInTheDocument();
    });
  });

  it("shows an error when deleting a category fails", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(
        new Response(JSON.stringify(seed), { status: 200 }),
      )
      .mockResolvedValueOnce(new Response(null, { status: 409 }));

    render(<CategoriesPage />);

    await waitFor(() => {
      expect(screen.getByText("casa")).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole("button", { name: /excluir casa/i }));

    await waitFor(() => {
      expect(
        screen.getByText(/ainda tem lançamentos/i),
      ).toBeInTheDocument();
    });
  });
});
