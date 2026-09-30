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
  // 204, 205 and 304 reject any body, including the empty string from text().
  const nullBody =
    apiResponse.status === 204 ||
    apiResponse.status === 205 ||
    apiResponse.status === 304;
  const response = new Response(nullBody ? null : responseBody, {
    status: apiResponse.status,
    headers:
      !nullBody && responseBody
        ? { "Content-Type": "application/json" }
        : undefined,
  });

  for (const setCookie of apiResponse.headers.getSetCookie()) {
    response.headers.append("Set-Cookie", setCookie);
  }

  return response;
}
