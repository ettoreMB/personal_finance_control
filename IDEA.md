# Ideia do produto — Personal Finance Control

> Documento vivo do produto. Core de escrita: `/grilling` (2026-08-30). Visualizações, entrega vertical e backlog restante: `/grill-with-docs` (2026-09-17).

## Visão geral

App de controle financeiro pessoal, **single-user** (uma única conta, sem conceito de multiusuário). O usuário registra ganhos e gastos (avulsos ou parcelados no cartão de crédito), categoriza os lançamentos e visualiza a saúde financeira por diferentes recortes de tempo. Um dos canais de entrada é o WhatsApp, via mensagem de texto — último canal, depois do Painel do mês corrente (não precisa esperar o norte inteiro de visualizações).

Moeda: apenas BRL (R$). Sem suporte a múltiplas moedas, sem campo de moeda no schema.

## Domínio

### Lançamentos financeiros
- Tipos: ganho ou gasto
- Data: qualquer data (passada, hoje, futura), não apenas "hoje"
- Cada lançamento pertence a uma categoria
- Todo lançamento é **avulso** ou **parcela**. Avulso: edição/exclusão livres a qualquer momento — não existe "mês fechado". Parcela não se edita nem se exclui isoladamente (mutação pela Compra)

### Compras parceladas no cartão de crédito
- Uma **Compra** (N entre 2 e 24) gera N gastos, um por mês. Compra à vista (1x) é lançamento avulso, não Compra. Sem entidade Cartão nem Fatura
- Exemplo: compra de R$150 em 3x hoje → R$50 no mês corrente + R$50 nos 2 meses seguintes (resto da divisão em centavos na última parcela)
- Cada parcela aparece na visão de "gastos do mês" normalmente, junto com os demais gastos daquele mês
- **Prazo fixo**: N e a data da Compra não mudam depois de criada. Não existe cancelar parcelas futuras e deixar as pagas
- **Editar o valor total**: parcelas de mês calendário já passado (`America/Sao_Paulo`) não são reescritas; o novo total menos a soma do passado é rateado no mês corrente e nas futuras
- **Desfazer** a Compra inteira (soft delete da Compra e das N parcelas) só vale enquanto nenhuma parcela cai em mês passado. Compra terminada continua na lista. Nunca hard delete
- Vocabulário e decisões: `CONTEXT.md`, ADRs 0001–0003 (grilling 2026-09-14/15)

### Categorias
- Categorias iniciais (casa, carro, comida, lazer) são inseridas via **seed automático em migration** — o app já nasce utilizável
- CRUD completo de categorias (criar, editar, remover)
- **Exclusão de categoria em uso é bloqueada**: se houver lançamentos vinculados, é preciso reclassificá-los antes de excluir a categoria
- Lançamentos já existentes podem ser reclassificados para outra categoria a qualquer momento

## Visualizações

Norte da sessão `/grill-with-docs` (2026-09-17). Vocabulário: `CONTEXT.md`; recorte civil, futuro-no-período e entrega vertical: ADRs 0004–0006.

Entrega em cortes: primeiro o **mês civil corrente** (ganhos, gastos, Saldo, breakdown). Seletor de Período e modo Cartão vêm depois; um PR não implementa o norte inteiro. Lista de cortes: `IMPLEMENTATION_PLAN.md`.

- Superfície: um **Painel** com dois modos (`Período` | `Cartão`). Lista de lançamentos continua em `/entries`. `/purchases` continua sendo a mutação da Compra. Sem gráfico.
- **Modo Período**: seletor mês / semestre / ano civil (mesmo shape nos três: ganhos, gastos, Saldo + breakdown). Padrão = recorte corrente em `America/Sao_Paulo`; navega passado e futuro; Período vazio mostra zeros; trocar granularidade mantém o recorte que contém o que está na tela. Sem série mês-a-mês dentro de semestre/ano.
- **Totais do Período**: lançamentos **ativos** cuja **data** cai no Período — inclusive futuros. Parcela entra no mês da parcela. Compra desfeita não entra. Sem split realizado vs previsto.
- **Por categoria**: dimensão do Período. Cada categoria com movimento mostra ganhos, gastos e Saldo; sem movimento, some. Não é tela all-time.
- **Modo Cartão**: mesmo seletor de Período, filtrando a **data da Compra**. Cada Compra ativa: descrição, nome da categoria na Compra, N, Comprometido. Agrupado pelo mês da Compra. Sem “já caído vs restante”. Exemplo: Compra de janeiro em 3x de R$50 aparece como R$150 comprometido em janeiro, mesmo que só R$50 caia como gasto em janeiro.

## Autenticação

Mesmo sendo single-user, o app tem login completo:

- **Sessão via cookie httpOnly** (não JWT) — API seta o cookie, front não manipula token diretamente
- **Criação de usuário**: existe um endpoint de registro (`/register`) e uma tela correspondente no front, usada uma única vez para criar a conta. O endpoint **não tem bloqueio automático** contra criação de uma segunda conta — a responsabilidade de não expor a API publicamente antes de criar a conta única é do usuário
- **Recuperação de senha**: rota própria na API (sem envio de e-mail — não há infraestrutura de e-mail no projeto). Protegida por uma chave secreta fixa via variável de ambiente (`RECOVERY_SECRET`)
- Todas as demais telas do front exigem login (redirecionam para login se não houver sessão válida)

## Integração com WhatsApp

Último canal. Não começa antes do Painel do mês corrente existir (senão o lançamento some num livro sem leitura). Não espera seletor de Período nem modo Cartão.

Fluxo pretendido:
1. Usuário envia mensagem de texto para o WhatsApp com o gasto/ganho (ex: "gastei 45 reais no mercado")
2. Mensagem chega via webhook através da **Evolution API** (gateway não-oficial de WhatsApp, já definido pelo usuário)
3. Texto é enviado a uma LLM para extração estruturada (`valor`, `categoria`, `data`, `tipo`) — não é um caso de uso de MCP, e sim de extração de dados estruturados via prompt/function calling
4. Dados extraídos viram uma chamada para a API criando o lançamento
5. Resposta de confirmação é enviada de volta ao usuário via WhatsApp

Detalhes de implementação (proteção do webhook, mapeamento de número de telefone, provedor de LLM) ficam para o corte de WhatsApp.

## Metas/orçamento por categoria

Fora de escopo por enquanto. Não modelado no schema — o schema normalizado + migrations (golang-migrate) tornam isso uma adição isolada no futuro, sem necessidade de antecipar.

## Testes

Cobertura de 100% (API e Front) é **meta aspiracional**, não é gate obrigatório de CI — não bloqueia merge de PR. Unitários viajam no PR de cada corte. e2e (testcontainers / Playwright) é um corte opcional de endurecimento, não uma fase que bloqueia o restante.

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
