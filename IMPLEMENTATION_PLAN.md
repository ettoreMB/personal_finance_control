# Plano de implementação

Tracking das features do projeto. Cada feature é implementada em uma branch própria e mergeada via PR (ver regra de workflow no [CLAUDE.md](./CLAUDE.md)). Ao mergear e validar um PR, marque a feature correspondente como concluída aqui, com a data e o link do PR.

Status possíveis: `Não iniciado` · `Em andamento` · `Em revisão (PR aberto)` · `Concluído`

> Fases 0–3: grilling 2026-08-30. Cortes restantes: `/grill-with-docs` 2026-09-17 — norte em [IDEA.md](./IDEA.md), entrega vertical em [ADR 0006](./docs/adr/0006-remaining-work-is-vertical-slices.md). Unitários viajam no PR de cada corte; 100% de cobertura continua meta aspiracional, não gate.

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
| Registro de compra parcelada e geração das parcelas mensais | Concluído | `feat/cartao-parcelado` | [#33](https://github.com/ettoreMB/personal_finance_control/pull/33) | 2026-09-18 |
| Edição da Compra (descrição, categoria e valor restante; valor/data de parcelas de mês passado intocados) | Concluído | `feat/cartao-parcelado` | [#33](https://github.com/ettoreMB/personal_finance_control/pull/33) | 2026-09-18 |
| Desfazer a Compra no mês (soft delete da Compra e das N parcelas; bloqueado se alguma parcela já caiu em mês passado) | Concluído | `feat/cartao-parcelado` | [#33](https://github.com/ettoreMB/personal_finance_control/pull/33) | 2026-09-18 |

## Próximos cortes (verticais)

Norte do Painel (seletor de Período, semestre/ano, modo Cartão) está no [IDEA.md](./IDEA.md). Cada linha abaixo é um PR. Não reabrir o norte inteiro ao implementar — só o que o corte precisa.

| Corte | Status | Branch | PR | Data |
|---|---|---|---|---|
| Painel — mês corrente (ganhos, gastos, Saldo, breakdown por categoria) | Concluído | `feat/painel-mes-corrente` | [#39](https://github.com/ettoreMB/personal_finance_control/pull/39) | 2026-09-18 |
| Painel — seletor de Período (navegar mês/semestre/ano, mesmo shape) | Não iniciado | | | |
| Painel — modo Cartão (Comprometido pela data da Compra) | Não iniciado | | | |
| e2e API (testcontainers) e Front (Playwright) — opcional, não bloqueia o restante | Não iniciado | | | |
| WhatsApp — Evolution API, extração via LLM, lançamento + confirmação (último canal) | Não iniciado | | | |
