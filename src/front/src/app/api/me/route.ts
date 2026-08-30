import { apiUrl } from "@/lib/api";

export async function GET(request: Request) {
  const cookie = request.headers.get("cookie");

  const apiResponse = await fetch(apiUrl("/me"), {
    headers: cookie ? { Cookie: cookie } : undefined,
    cache: "no-store",
  });

  const responseBody = await apiResponse.text();

  return new Response(responseBody, {
    status: apiResponse.status,
    headers: { "Content-Type": "application/json" },
  });
}
