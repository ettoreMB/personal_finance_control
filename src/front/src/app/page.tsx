export default function Home() {
  return (
    <div className="flex flex-1 items-center justify-center">
      <main className="flex w-full max-w-sm flex-col gap-6">
        <h1 className="text-2xl font-semibold">Controle financeiro</h1>
        <p>
          Registre ganhos e gastos, classifique por categoria e acompanhe o
          saldo do mês.
        </p>
        <div className="flex flex-col gap-3">
          <a
            href="/login"
            className="rounded bg-foreground px-4 py-2 text-center text-background"
          >
            Entrar
          </a>
          <a href="/register" className="rounded border px-4 py-2 text-center">
            Criar conta
          </a>
        </div>
      </main>
    </div>
  );
}
