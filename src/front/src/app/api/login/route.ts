import { apiUrl } from "@/lib/api";

export async function POST(request: Request) {
  const body = await request.text();

  const apiResponse = await fetch(apiUrl("/login"), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body,
  });

  const responseBody = await apiResponse.text();

  const response = new Response(responseBody, {
    status: apiResponse.status,
    headers: { "Content-Type": "application/json" },
  });

  for (const cookie of apiResponse.headers.getSetCookie()) {
    response.headers.append("Set-Cookie", cookie);
  }

  return response;
}
