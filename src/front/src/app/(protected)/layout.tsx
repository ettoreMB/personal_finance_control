import { cookies, headers } from "next/headers";
import { redirect } from "next/navigation";
import type { ReactNode } from "react";

export default async function ProtectedLayout({
  children,
}: {
  children: ReactNode;
}) {
  const [cookieStore, headersList] = await Promise.all([cookies(), headers()]);
  const host = headersList.get("host");
  const protocol = host?.startsWith("localhost") ? "http" : "https";

  const res = await fetch(`${protocol}://${host}/api/me`, {
    headers: { Cookie: cookieStore.toString() },
    cache: "no-store",
  });

  if (!res.ok) {
    redirect("/login");
  }

  return <>{children}</>;
}
