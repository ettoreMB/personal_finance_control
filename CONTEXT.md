# Controle financeiro pessoal

App de um único usuário para registrar ganhos e gastos em BRL, classificar por categoria e acompanhar o dinheiro em diferentes recortes de tempo.

## Language

**Lançamento**:
Um ganho ou gasto em uma data, sempre pertencente a uma categoria. Todo lançamento é avulso ou parcela.
_Avoid_: transação, movimento, entry, registro financeiro

**Lançamento avulso**:
Lançamento que não pertence a uma Compra. Edição e exclusão são livres a qualquer momento; não existe mês fechado. Gasto no cartão em 1x é isto, não Compra.
_Avoid_: lançamento simples, avulsa como tipo no banco

**Compra**:
Uma compra parcelada: descrição obrigatória, data (passada, hoje ou futura), valor total, N parcelas (2 a 24) e categoria. Gera N parcelas na criação; prazo (N e data) é fixo; valor total pode ser redistribuído só no que ainda não passou. Não existe compra à vista, Cartão nem Fatura. Desfeita some das listas; terminada (todas as parcelas no passado) continua visível.
_Avoid_: purchase, pedido, parcelamento como entidade, fatura, cartão

**Parcela**:
Gasto que pertence a uma Compra, a k-ésima de N. Não se edita nem se exclui isoladamente. Valor e data de mês calendário passado (`America/Sao_Paulo`) não mudam; categoria das ativas segue a Compra.
_Avoid_: installment como tipo separado, fatura

**Ganho**:
Lançamento que representa entrada de dinheiro. Não existe ganho parcelado.
_Avoid_: receita, income, crédito

**Gasto**:
Lançamento que representa saída de dinheiro.
_Avoid_: despesa, expense, débito

**Categoria**:
Classificação obrigatória de um lançamento e de uma Compra. O nome é único. Todas as parcelas ativas de uma Compra compartilham a categoria da Compra. Excluir uma categoria em uso é bloqueado enquanto existir lançamento ativo nela; Compra sem parcela ativa não trava. A Compra guarda o nome da categoria para a história sobreviver se a categoria for apagada.
_Avoid_: tag, label, grupo

**Período**:
Recorte de tempo calendário civil em `America/Sao_Paulo`: um mês, um semestre ou um ano civil. Não é janela móvel.
_Avoid_: rolling, últimos N meses, últimos 30 dias, trimestre

**Semestre**:
Janeiro–junho (1º) ou julho–dezembro (2º) de um ano civil em `America/Sao_Paulo`.
_Avoid_: seis meses móveis, semestre letivo

**Saldo**:
Ganhos menos gastos dos lançamentos ativos cuja data cai no Período. Pode ser negativo. Não é saldo de conta e não carrega de um Período para o outro. Nos totais por categoria, o mesmo recorte: ganhos, gastos e Saldo; categoria sem lançamento ativo no Período não entra.
_Avoid_: balanço, net, balance, saldo acumulado, saldo realizado

**Comprometido**:
Valor total de uma Compra ativa, atribuído ao mês calendário da data da Compra — não ao mês de cada parcela.
_Avoid_: fatura, limite, saldo do cartão, total do mês no cartão (ambíguo com a soma das parcelas)
