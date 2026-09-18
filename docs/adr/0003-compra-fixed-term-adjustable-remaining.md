# Compra tem prazo fixo, valor ajustável no restante, undo só antes do primeiro mês passado

N e data da Compra não mudam. Não existe cancelar o rabo do parcelamento (apagar parcelas futuras e deixar as pagas).

Editar o valor total não reescreve parcelas de meses calendário já passados. O novo total menos a soma do passado é rateado nas parcelas do mês corrente e futuras; o resto da divisão em centavos cai na última parcela (sempre a N-ésima, porque N não encolhe). Cada parcela ainda mutável precisa de pelo menos 1 centavo — senão 409. Compra cujas N parcelas já estão no passado não aceita editar total.

Desfazer a Compra inteira só é permitido enquanto **nenhuma** parcela cai em mês passado (data da Compra no mês corrente ou futuro, inclusive lançada hoje com âncora futura). É soft delete da Compra **e** das N parcelas: some da lista e o GET devolve 404. Hard delete não existe.

A lista de Compras mostra as não desfeitas, inclusive as já 100% no passado. Data da Compra pode ser passada, hoje ou futura — parcelas que nascem em mês já passado já nascem com valor/data imutáveis, e essa Compra não pode ser desfeita.

Isso substitui o fluxo antigo da IDEA (“cancela restantes e lança outra Compra”) por: ajustar o restante, ou desfazer o erro enquanto ainda não há mês passado.
