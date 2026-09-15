import { proxyToApi } from "@/lib/proxy";

export function GET(request: Request) {
  return proxyToApi("/entries", request, "GET");
}

export function POST(request: Request) {
  return proxyToApi("/entries", request, "POST");
}
