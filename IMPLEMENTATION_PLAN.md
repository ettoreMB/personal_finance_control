# Plano de implementação

Tracking das features do projeto. Cada feature é implementada em uma branch própria e mergeada via PR (ver regra de workflow no [CLAUDE.md](./CLAUDE.md)). Ao mergear e validar um PR, marque a feature correspondente como concluída aqui, com a data e o link do PR.

Status possíveis: `Não iniciado` · `Em andamento` · `Em revisão (PR aberto)` · `Concluído`

> Fases e escopo refinados via sessão `/grilling` (2026-08-30) — ver decisões em [IDEA.md](./IDEA.md).

## Fase 0 — Fundação

| Feature | Status | Branch | PR | Data |
|---|---|---|---|---|
| Setup do repositório (estrutura API + Front, Taskfile, CI base) | Concluído | `feat/setup-fundacao` | [#5](https://github.com/ettoreMB/personal_finance_control/pull/5) | 2026-08-30 |
| Setup da API (Fiber, GORM, golang-migrate, slog, Air) | Concluído | `feat/fase0-restante` | [#13](https://github.com/ettoreMB/personal_finance_control/pull/13) | 2026-08-30 |
| Setup do Front (Next.js, Tailwind v4, shadcn/ui) | Concluído | `feat/fase0-restante` | [#13](https://github.com/ettoreMB/personal_finance_control/pull/13) | 2026-08-30 |
| Design do schema do banco de dados (SQLite, normalizado) | Concluído | `feat/fase0-restante` | [#13](https://github.com/ettoreMB/personal_finance_control/pull/13) | 2026-08-30 |
| Criação dos arquivos de docker compose | Concluído | `feat/fase0-restante` | [#13](https://github.com/ettoreMB/personal_finance_control/pull/13) | 2026-08-30 |

## Fase 1 — Autenticação

| Feature | Status | Branch | PR | Data |
|---|---|---|---|---|
| Endpoint + tela de criação de usuário (registro único) | Concluído | `feat/fase1-autenticacao` | [#21](https://github.com/ettoreMB/personal_finance_control/pull/21) | 2026-08-31 |
| Login com sessão via cookie httpOnly | Concluído | `feat/fase1-autenticacao` | [#21](https://github.com/ettoreMB/personal_finance_control/pull/21) | 2026-08-31 |
| Middleware de proteção de rotas (API) e redirecionamento de telas não autenticadas (Front) | Concluído | `feat/fase1-autenticacao` | [#21](https://github.com/ettoreMB/personal_finance_control/pull/21) | 2026-08-31 |
| Rota de recuperação de senha protegida por `RECOVERY_SECRET` | Concluído | `feat/fase1-autenticacao` | [#21](https://github.com/ettoreMB/personal_finance_control/pull/21) | 2026-08-31 |

## Fase 2 — Core: lançamentos e categorias

| Feature | Status | Branch | PR | Data |
|---|---|---|---|---|
| Seed automático das categorias iniciais (casa, carro, comida, lazer) via migration | Concluído | `feat/fase0-restante` | [#13](https://github.com/ettoreMB/personal_finance_control/pull/13) | 2026-08-30 |
| CRUD de categorias (exclusão bloqueada se houver lançamentos vinculados) | Concluído | `feat/lancamentos-e-categorias` | [#27](https://github.com/ettoreMB/personal_finance_control/pull/27) | 2026-09-14 |
| CRUD de lançamentos avulsos (ganho/gasto), edição livre sem restrição de "mês fechado" | Concluído | `feat/lancamentos-e-categorias` | [#27](https://github.com/ettoreMB/personal_finance_control/pull/27) | 2026-09-14 |
| Reclassificação de categoria em lançamentos existentes | Concluído | `feat/lancamentos-e-categorias` | [#27](https://github.com/ettoreMB/personal_finance_control/pull/27) | 2026-09-14 |

## Fase 3 — Cartão de crédito parcelado

| Feature | Status | Branch | PR | Data |
|---|---|---|---|---|
| Registro de compra parcelada e geração das parcelas mensais | Não iniciado | | | |
| Edição/exclusão da compra original (parcelas passadas intocadas, só futuras mudam) | Não iniciado | | | |
| Cancelamento de parcelas futuras via soft delete (para "trocar" o valor total de uma compra) | Não iniciado | | | |

## Fase 4 — Visualizações

| Feature | Status | Branch | PR | Data |
|---|---|---|---|---|
| Visão mensal (ganhos vs. gastos) | Não iniciado | | | |
| Visão por categoria | Não iniciado | | | |
| Visão semestral | Não iniciado | | | |
| Visão anual | Não iniciado | | | |
| Visão de cartão de crédito (total por compra, agrupado por mês da compra) | Não iniciado | | | |

## Fase 5 — Qualidade e testes

| Feature | Status | Branch | PR | Data |
|---|---|---|---|---|
| Cobertura de testes unitários API (meta aspiracional 100%, não bloqueia CI) | Não iniciado | | | |
| Testes e2e API com testcontainers | Não iniciado | | | |
| Cobertura de testes unitários Front (meta aspiracional 100%, não bloqueia CI) | Não iniciado | | | |
| Testes e2e Front com Playwright | Não iniciado | | | |

## Fase 6 — Integração WhatsApp (última fase)

| Feature | Status | Branch | PR | Data |
|---|---|---|---|---|
| Setup Evolution API + webhook de recebimento | Não iniciado | | | |
| Extração estruturada de lançamento via LLM | Não iniciado | | | |
| Criação de lançamento a partir da mensagem + resposta de confirmação | Não iniciado | | | |
