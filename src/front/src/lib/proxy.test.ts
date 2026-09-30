import { afterEach, describe, expect, it, vi } from "vitest";

import { proxyToApi } from "./proxy";

afterEach(() => {
  vi.unstubAllGlobals();
});

function incoming(method: string) {
  return new Request("http://front.local/api/entries/1", { method });
}

describe("proxyToApi", () => {
  it("forwards 204 with a null body", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(new Response(null, { status: 204 })),
    );

    const response = await proxyToApi("/entries/1", incoming("DELETE"), "DELETE");

    expect(response.status).toBe(204);
    expect(await response.text()).toBe("");
  });

  it("forwards a json body", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(JSON.stringify({ id: 1, description: "água" }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        }),
      ),
    );

    const response = await proxyToApi("/entries/1", incoming("PATCH"), "PATCH");

    expect(response.status).toBe(200);
    expect(response.headers.get("content-type")).toBe("application/json");
    await expect(response.json()).resolves.toEqual({
      id: 1,
      description: "água",
    });
  });
});
