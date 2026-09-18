import { proxyToApi } from "@/lib/proxy";

export function GET(request: Request) {
  return proxyToApi("/summary", request, "GET");
}
