import { describe, expect, it, vi, beforeEach } from "vitest";

const cookiesMock = vi.fn();

vi.mock("next/headers", () => ({
  cookies: () => cookiesMock(),
}));

import ProtectedLayout from "./layout";

describe("ProtectedLayout", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", vi.fn());
    cookiesMock.mockResolvedValue({ toString: () => "session=abc" });
  });

  it("renders children when /me succeeds", async () => {
    vi.mocked(fetch).mockResolvedValueOnce(new Response(null, { status: 200 }));

    const result = await ProtectedLayout({ children: "protected content" });

    expect(result).toBeTruthy();
    expect(fetch).toHaveBeenCalledWith(
      "http://localhost:3000/me",
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
