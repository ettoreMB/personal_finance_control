# Personal Finance Control

App de controle financeiro pessoal. Ver [IDEA.md](./IDEA.md) e [IMPLEMENTATION_PLAN.md](./IMPLEMENTATION_PLAN.md) para a visão de produto e o status de implementação.

## Pré-requisitos

- [Go](https://go.dev/) (versão de `api/go.mod`)
- [Node.js](https://nodejs.org/) 24+
- [go-task](https://taskfile.dev/) (`task`)
- [Air](https://github.com/air-verse/air) — hot reload da API (necessário para `task dev:api`)
- [golang-migrate](https://github.com/golang-migrate/migrate) — migrations do banco (necessário para `task migrate:up`/`migrate:down`)

## Comandos (Taskfile)

| Comando | Descrição |
|---|---|
| `task dev:api` | Sobe a API localmente com hot reload |
| `task dev:front` | Sobe o Front localmente (`next dev`) |
| `task test:api` | Roda os testes da API |
| `task test:front` | Roda os testes do Front |
| `task build:api` | Compila a API |
| `task build:front` | Build de produção do Front |
| `task migrate:up` | Aplica as migrations do banco |
| `task migrate:down` | Reverte a última migration |

## Estrutura

- `api/` — API em Go (Fiber, GORM, golang-migrate)
- `front/` — Front em Next.js (Tailwind v4, shadcn/ui)
