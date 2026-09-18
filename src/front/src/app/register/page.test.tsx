import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import RegisterPage from "./page";

const pushMock = vi.fn();

vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: pushMock }),
}));

describe("RegisterPage", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", vi.fn());
    pushMock.mockClear();
  });

  it("renders the registration form", () => {
    render(<RegisterPage />);

    expect(
      screen.getByRole("heading", { name: /criar conta/i }),
    ).toBeInTheDocument();
  });

  it("shows an error message when the API call fails", async () => {
    vi.mocked(fetch).mockResolvedValueOnce(
      new Response(JSON.stringify({ message: "Senha inválida." }), {
        status: 400,
      }),
    );

    render(<RegisterPage />);

    fireEvent.change(screen.getByLabelText("Email"), {
      target: { value: "user@example.com" },
    });
    fireEvent.change(screen.getByLabelText("CPF"), {
      target: { value: "529.982.247-25" },
    });
    fireEvent.change(screen.getByLabelText("Senha"), {
      target: { value: "abcdefg1!" },
    });
    fireEvent.change(screen.getByLabelText("Confirmar senha"), {
      target: { value: "abcdefg1!" },
    });

    fireEvent.click(screen.getByRole("button", { name: /cadastrar/i }));

    await waitFor(() => {
      expect(screen.getByText("Senha inválida.")).toBeInTheDocument();
    });

    expect(fetch).toHaveBeenCalledWith(
      "/api/register",
      expect.objectContaining({ method: "POST" }),
    );
    expect(pushMock).not.toHaveBeenCalled();
  });

  it("redirects to login when the API call succeeds", async () => {
    vi.mocked(fetch).mockResolvedValueOnce(
      new Response(JSON.stringify({ id: 1 }), { status: 201 }),
    );

    render(<RegisterPage />);

    fireEvent.change(screen.getByLabelText("Email"), {
      target: { value: "user@example.com" },
    });
    fireEvent.change(screen.getByLabelText("CPF"), {
      target: { value: "529.982.247-25" },
    });
    fireEvent.change(screen.getByLabelText("Senha"), {
      target: { value: "abcdefg1!" },
    });
    fireEvent.change(screen.getByLabelText("Confirmar senha"), {
      target: { value: "abcdefg1!" },
    });

    fireEvent.click(screen.getByRole("button", { name: /cadastrar/i }));

    await waitFor(() => {
      expect(pushMock).toHaveBeenCalledWith("/login");
    });
  });
});
