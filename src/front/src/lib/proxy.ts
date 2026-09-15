import { apiUrl } from "@/lib/api";

export async function proxyToApi(
  path: string,
  request: Request,
  method?: string,
): Promise<Response> {
  const cookie = request.headers.get("cookie");
  const verb = method ?? request.method;
  const hasBody = verb !== "GET" && verb !== "HEAD" && verb !== "DELETE";

  const headers: Record<string, string> = {};
  if (cookie) {
    headers.Cookie = cookie;
  }
  if (hasBody) {
    headers["Content-Type"] = "application/json";
  }

  const apiResponse = await fetch(apiUrl(path), {
    method: verb,
    headers,
    body: hasBody ? await request.text() : undefined,
    cache: "no-store",
  });

  const responseBody = await apiResponse.text();
  const response = new Response(responseBody, {
    status: apiResponse.status,
    headers: responseBody
      ? { "Content-Type": "application/json" }
      : undefined,
  });

  for (const setCookie of apiResponse.headers.getSetCookie()) {
    response.headers.append("Set-Cookie", setCookie);
  }

  return response;
}
