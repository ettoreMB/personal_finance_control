# Parcelas se mutam pela Compra, não como lançamento avulso

Uma parcela é um lançamento (gasto): aparece na listagem e nos totais do mês. Mas a unidade de mutação é a Compra. PATCH/DELETE no lançamento que é parcela é rejeitado; criar, reclassificar e editar valor acontecem na Compra. Valor e data de mês calendário já passado não são reescritos; categoria de parcela ativa segue a Compra mesmo no passado. A alternativa era tratar parcela como lançamento avulso (edição livre da Fase 2) e perder a integridade do grupo.
