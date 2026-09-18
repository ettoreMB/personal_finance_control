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

  return (
    <div className="flex min-h-full flex-col">
      <nav className="flex gap-4 border-b px-6 py-3">
        <a href="/dashboard" className="underline">
          Painel
        </a>
        <a href="/categories" className="underline">
          Categorias
        </a>
        <a href="/entries" className="underline">
          Lançamentos
        </a>
        <a href="/purchases" className="underline">
          Compras
        </a>
      </nav>
      {children}
    </div>
  );
}
