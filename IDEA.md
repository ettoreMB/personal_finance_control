# Ideia do produto — Personal Finance Control

> Documento vivo do produto. As decisões abaixo foram fechadas em sessão `/grilling` (2026-08-30).

## Visão geral

App de controle financeiro pessoal, **single-user** (uma única conta, sem conceito de multiusuário). O usuário registra ganhos e gastos (avulsos ou parcelados no cartão de crédito), categoriza os lançamentos e visualiza a saúde financeira por diferentes recortes de tempo. Um dos canais de entrada é o WhatsApp, via mensagem de texto — implementado por último, depois de todo o core estar pronto.

Moeda: apenas BRL (R$). Sem suporte a múltiplas moedas, sem campo de moeda no schema.

## Domínio

### Lançamentos financeiros
- Tipos: ganho ou gasto
- Data: qualquer data (passada, hoje, futura), não apenas "hoje"
- Cada lançamento pertence a uma categoria
- Edição totalmente livre (valor, data, categoria), a qualquer momento — não existe conceito de "mês fechado" que bloqueie edição

### Compras parceladas no cartão de crédito
- Uma compra parcelada gera N lançamentos, um por mês, com valor = total / N
- Exemplo: compra de R$150 em 3x hoje → R$50 no mês corrente + R$50 nos 2 meses seguintes
- Cada parcela aparece na visão de "gastos do mês" normalmente, junto com os demais gastos daquele mês
- **Edição/exclusão da compra original**: parcelas de meses já passados nunca são alteradas (histórico não é reescrito). Apenas parcelas futuras são atualizadas ou removidas
- **Alterar o valor total de uma compra parcelada não é suportado diretamente**: o fluxo é cancelar as parcelas futuras restantes e lançar uma nova compra parcelada separada
- Cancelamento de parcelas futuras é feito via **soft delete** (flag/`deleted_at`), mantendo rastreabilidade — nunca hard delete

### Categorias
- Categorias iniciais (casa, carro, comida, lazer) são inseridas via **seed automático em migration** — o app já nasce utilizável
- CRUD completo de categorias (criar, editar, remover)
- **Exclusão de categoria em uso é bloqueada**: se houver lançamentos vinculados, é preciso reclassificá-los antes de excluir a categoria
- Lançamentos já existentes podem ser reclassificados para outra categoria a qualquer momento

## Visualizações

- **Mensal**: ganhos vs. gastos do mês, incluindo parcelas de cartão que caem naquele mês
- **Por categoria**: total gasto/ganho agrupado por categoria
- **Semestral**: consolidado de 6 meses
- **Anual**: consolidado do ano
- **Cartão de crédito (visão própria)**: total comprometido por compra, agrupado pelo mês da compra original (não pelo mês da parcela). Exemplo: compra de janeiro em 3x de R$50 aparece como R$150 "comprado em janeiro no cartão", mesmo que só R$50 caia na fatura de janeiro

## Autenticação

Mesmo sendo single-user, o app tem login completo:

- **Sessão via cookie httpOnly** (não JWT) — API seta o cookie, front não manipula token diretamente
- **Criação de usuário**: existe um endpoint de registro (`/register`) e uma tela correspondente no front, usada uma única vez para criar a conta. O endpoint **não tem bloqueio automático** contra criação de uma segunda conta — a responsabilidade de não expor a API publicamente antes de criar a conta única é do usuário
- **Recuperação de senha**: rota própria na API (sem envio de e-mail — não há infraestrutura de e-mail no projeto). Protegida por uma chave secreta fixa via variável de ambiente (`RECOVERY_SECRET`)
- Todas as demais telas do front exigem login (redirecionam para login se não houver sessão válida)

## Integração com WhatsApp

Última feature a ser implementada, depois de todo o core (lançamentos, cartão, categorias, visualizações, auth) estar pronto e validado.

Fluxo pretendido:
1. Usuário envia mensagem de texto para o WhatsApp com o gasto/ganho (ex: "gastei 45 reais no mercado")
2. Mensagem chega via webhook através da **Evolution API** (gateway não-oficial de WhatsApp, já definido pelo usuário)
3. Texto é enviado a uma LLM para extração estruturada (`valor`, `categoria`, `data`, `tipo`) — não é um caso de uso de MCP, e sim de extração de dados estruturados via prompt/function calling
4. Dados extraídos viram uma chamada para a API criando o lançamento
5. Resposta de confirmação é enviada de volta ao usuário via WhatsApp

Detalhes de implementação (proteção do webhook, mapeamento de número de telefone, provedor de LLM) ficam para quando essa fase for planejada, já que é a última do roadmap.

## Metas/orçamento por categoria

Fora de escopo por enquanto. Não modelado no schema — o schema normalizado + migrations (golang-migrate) tornam isso uma adição isolada no futuro, sem necessidade de antecipar.

## Testes

Cobertura de 100% (API e Front) é **meta aspiracional**, não é gate obrigatório de CI — não bloqueia merge de PR.

## Deploy

Ambiente de produção ainda não definido. Por enquanto, execução local (via Air/Taskfile na API, `next dev` no front). Quando houver destino de produção (provável VPS pessoal), a execução será containerizada via Docker.

## Stack técnica (decidida)

| Camada | Escolha |
|---|---|
| API | Go + Fiber + GORM + golang-migrate |
| Logs | `log/slog` (nativo) |
| DB | SQLite, schema normalizado (formas normais) |
| Dev API | Air (hot reload) + Taskfile (go-task) |
| Testes API | unitários com meta de 100% de cobertura + testcontainers para e2e |
| Front | Next.js + Tailwind v4 + shadcn/ui + TanStack Table (tabelas) |
| Testes Front | unitários com meta de 100% de cobertura + Playwright para e2e |
| Arquitetura | Clean Code / boas práticas em ambas as camadas |
| WhatsApp | Evolution API |
| Workflow | 1 branch + 1 PR por feature, merge para `main` só após validação |
