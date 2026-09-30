import { proxyToApi } from "@/lib/proxy";

export function GET(request: Request) {
  const query = new URL(request.url).search;
  return proxyToApi(`/summary${query}`, request, "GET");
}
