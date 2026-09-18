import { proxyToApi } from "@/lib/proxy";

export function GET(request: Request) {
  return proxyToApi("/purchases", request, "GET");
}

export function POST(request: Request) {
  return proxyToApi("/purchases", request, "POST");
}
