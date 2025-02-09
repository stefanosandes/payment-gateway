# Payment Gateway API

Uma API de pagamentos que atua como gateway intermediário, integrando-se a provedores de pagamento externos.
[Ver especificação](./case.pdf).

## tl;dr
- Decidi utilizar `Go` como linguagem de programação, embora não tenha experiência com a linguagem e o ecossistema, achei interessante utilizar por fazer parte da stack de desenvolvimento da empresa.
- A maior parte das práticas de desenvolvimento foram aprendidas da comunidade Go, tutoriais, documentação das bibliotecas utilizadas e outras fontes de pesquisa.
- Fiquei com dúvida específicamente no endpont de estorno de pagamento. Por não saber se um estorno pode ser feito por qualquer provedor ou se deveria ser feito pelo provedor na qual a trasação ocorreu, escolhi por restringir ao provedor utilizado.
- Outro ponto de dúvida foram os campos `originalAmount` e `currentAmount`. Se a presença destes se refere a um possível split entre os provedores ou apenas para manter o registro de transações em caso de estorno. Mantive a utilização para o caso de estorno.

## Funcionalidades

- Processamento de pagamentos com múltiplos provedores
- Fallback automático entre provedores em caso de falha
- Estorno de pagamentos
- Consulta de transações
- Circuit breaker para gerenciamento de falhas
- Política de retry para maior resiliência

## Tecnologias Utilizadas

- Go 1.23.6
- Framework web:
  - Gin
- Resiliência:
  - GoBreaker
  - Retry-Go
- Configuração:
  - Viper
- Utilitários:
  - Google UUID
  - GoFakeIt
- Testes:
  - Testify

## Estrutura do Projeto

```
.
├── api/
│   └── handlers/        # HTTP handlers e endpoints da API
├── cmd/
│   └── api/             # Entrypoint da api
├── internal/
│   ├── config/          # Gerenciamento de configuração
│   ├── domain/          # Modelos e interfaces do domínio
│   ├── providers/       # Implementação dos provedores de pagamento
│   └── service/         # Lógica de negócio e resiliência
└── mock/                # Servidores mock para simulação dos provedores
```

## Como Executar

1. Clone o repositório
2. Instale as dependências:
```bash
go mod download
```

3. Execute a aplicação:
```bash
go run cmd/api/main.go
```

A aplicação iniciará:
- API principal na porta 8080
- Mock Server 1 na porta 3001
- Mock Server 2 na porta 3002

## Resiliência

A API implementa os seguintes mecanismos de resiliência:

1. **Circuit Breaker**: Previne sobrecarga dos provedores em caso de falhas
   - Abre após 3 requisições com 60% de falha
   - Timeout de 30 segundos
   - Máximo de 3 requisições durante half-open state

2. **Retry**: Tenta novamente em caso de falhas temporárias
   - 3 tentativas por provedor
   - Delay de 1 segundo entre tentativas

3. **Fallback entre Provedores**: Tenta automaticamente o próximo provedor disponível
   - Ordem sequencial de tentativas
   - Logs detalhados do processo de fallback

## Logs

A aplicação gera logs detalhados sobre:
- Mudanças de estado do circuit breaker
- Tentativas de retry
- Fallback entre provedores
- Erros e sucessos nas operações

## Testes

Para executar os testes:
```bash
go test ./...
```

## Simulando Falhas

Os mock servers incluem um endpoint para simular falhas:

```go
mockServer1.SimulateFailure(true) // Força o servidor a retornar erro
```

Isso permite testar os mecanismos de resiliência da aplicação.

## Testes HTTP
O arquivo `test.http` na raiz do projeto contém testes HTTP que cobrem os endpoints da API. Ele pode ser executado de no VSCode com o plugin [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client). No GoLand, nativamente. Ou [Httpie](https://httpie.io/) ou [Postman](https://www.postman.com/) ou qualquer outro cliente HTTP.
