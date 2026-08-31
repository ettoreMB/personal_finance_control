import { describe, expect, it, vi, beforeEach } from "vitest";

const cookiesMock = vi.fn();
const headersMock = vi.fn();

vi.mock("next/headers", () => ({
  cookies: () => cookiesMock(),
  headers: () => headersMock(),
}));

import ProtectedLayout from "./layout";

describe("ProtectedLayout", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", vi.fn());
    cookiesMock.mockResolvedValue({ toString: () => "session=abc" });
    headersMock.mockResolvedValue(new Map([["host", "localhost:3001"]]));
  });

  it("renders children when /me succeeds", async () => {
    vi.mocked(fetch).mockResolvedValueOnce(new Response(null, { status: 200 }));

    const result = await ProtectedLayout({ children: "protected content" });

    expect(result).toBeTruthy();
    expect(fetch).toHaveBeenCalledWith(
      "http://localhost:3001/api/me",
      expect.objectContaining({ headers: { Cookie: "session=abc" } }),
    );
  });

  it("redirects to /login when /me fails", async () => {
    vi.mocked(fetch).mockResolvedValueOnce(new Response(null, { status: 401 }));

    await expect(
      ProtectedLayout({ children: "protected content" }),
    ).rejects.toThrow(/NEXT_REDIRECT/);
  });
});
