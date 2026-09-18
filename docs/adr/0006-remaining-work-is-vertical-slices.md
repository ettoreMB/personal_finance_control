# Trabalho restante é corte vertical, não fase-camada

Depois da Fase 3 o mapa não é mais “todas as visões, depois todos os testes, depois WhatsApp”. Cada corte é um PR fino (API + front + unitários) que dá para usar sozinho. O norte do Painel vive no `IDEA.md`; o PR só implementa o corte da vez. e2e (testcontainers / Playwright) é endurecimento opcional, não portão. WhatsApp continua o último canal, mas não espera semestre nem modo Cartão — espera pelo menos o Painel do mês corrente, senão grava no escuro.

A alternativa (1 PR = Fase 4 inteira, Fase 5 de qualidade) forçava decidir navegação de semestre antes de existir um Saldo na tela.
