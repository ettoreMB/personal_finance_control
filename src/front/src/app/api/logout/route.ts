import { apiUrl } from "@/lib/api";

export async function POST(request: Request) {
  const cookie = request.headers.get("cookie");

  const apiResponse = await fetch(apiUrl("/logout"), {
    method: "POST",
    headers: cookie ? { Cookie: cookie } : undefined,
  });

  const response = new Response(null, { status: apiResponse.status });

  for (const cookie of apiResponse.headers.getSetCookie()) {
    response.headers.append("Set-Cookie", cookie);
  }

  return response;
}
