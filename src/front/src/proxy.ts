import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

// Cookie-presence check only, for UX. The real check is always the
// server-side /me call made by the (protected) route group's layout.
export function proxy(request: NextRequest) {
  if (!request.cookies.has("session")) {
    return NextResponse.redirect(new URL("/login", request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/dashboard/:path*"],
};
