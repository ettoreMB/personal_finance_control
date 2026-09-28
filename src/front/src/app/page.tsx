import { Wallet } from "lucide-react";
import Link from "next/link";
import { buttonVariants } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

export default function Home() {
  return (
    <div className="flex flex-1 items-center justify-center bg-muted p-6">
      <div className="w-full max-w-sm">
        <Card>
          <CardHeader>
            <div className="mb-2 flex size-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
              <Wallet className="size-4" />
            </div>
            <CardTitle>
              <h1 className="text-xl">Controle financeiro</h1>
            </CardTitle>
            <CardDescription>
              Registre ganhos e gastos, classifique por categoria e acompanhe o
              saldo do mês.
            </CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col gap-3">
            <Link href="/login" className={buttonVariants({ size: "lg" })}>
              Entrar
            </Link>
            <Link
              href="/register"
              className={buttonVariants({ variant: "outline", size: "lg" })}
            >
              Criar conta
            </Link>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
