import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import Home from "./page";

describe("Home", () => {
  it("renders the app landing with login and register links", () => {
    render(<Home />);

    expect(
      screen.getByRole("heading", { name: /controle financeiro/i }),
    ).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /entrar/i })).toHaveAttribute(
      "href",
      "/login",
    );
    expect(screen.getByRole("link", { name: /criar conta/i })).toHaveAttribute(
      "href",
      "/register",
    );
  });
});
