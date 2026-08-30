import { apiUrl } from "@/lib/api";

export async function POST(request: Request) {
  const body = await request.text();

  const apiResponse = await fetch(apiUrl("/register"), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body,
  });

  const responseBody = await apiResponse.text();

  return new Response(responseBody, {
    status: apiResponse.status,
    headers: { "Content-Type": "application/json" },
  });
}
