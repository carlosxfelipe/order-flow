# O Guia Definitivo do Padrão REPR (Request-Endpoint-Response)

Bem-vindo ao guia detalhado sobre o padrão arquitetural **REPR**. Este documento foi criado para ajudar a entender profundamente o que é, por que usá-lo e como ele se diferencia de abordagens tradicionais, usando as implementações deste projeto (`OrderFlow`) como estudo de caso.

---

## 1. O que é o padrão REPR?

O termo **REPR** (lê-se *Reaper*) significa **Request-Endpoint-Response**. Ele define que as rotas de uma API web devem ser desenhadas ao redor de três componentes fundamentais perfeitamente alinhados:

1. **Request:** Os dados exatos e específicos necessários para a execução daquela requisição.
2. **Endpoint:** O controlador da lógica. Um manipulador que recebe a requisição, processa a regra de negócio (sozinho ou delegando) e retorna uma resposta.
3. **Response:** Os dados de saída desenhados *exclusivamente* para o consumidor daquele endpoint.

A premissa básica é o **Isolamento Absoluto**. Cada caso de uso da aplicação é autocontido.

---

## 2. O Problema do MVC e dos Controllers "Deuses"

Na arquitetura clássica baseada em MVC (Model-View-Controller), costuma-se agrupar endpoints por *Recurso* (Entidade). Por exemplo, um `OrderController`.

Com o tempo, o `OrderController` passa a ter:
- `CreateOrder`
- `GetOrder`
- `UpdateOrder`
- `CancelOrder`
- `PayOrder`
- `ShipOrder`

**O resultado?**
- Injeção de dezenas de dependências no construtor do Controller, mesmo que o `GetOrder` só precise de uma, enquanto o `PayOrder` precisa de cinco integrações de pagamento diferentes.
- Arquivos gigantescos, difíceis de ler e com alta chance de gerar conflitos de *merge* quando vários desenvolvedores trabalham na mesma entidade (Pedidos).

---

## 3. A Solução: Arquitetura Vertical (Vertical Slicing)

O padrão REPR é a fundação da **Arquitetura Vertical** para a camada de entrada (Web/API). Em vez de dividir o código por camadas horizontais (Controllers de um lado, DTOs de outro, Services de outro), o sistema é fatiado por **Funcionalidades (Features)**.

Quando os arquivos são agrupados em uma mesma pasta por "Feature", o código que muda junto passa a viver junto. É exatamente o que está estruturado neste projeto `OrderFlow`.

> **Coesão Extrema**
> Ao abrir a pasta de um caso de uso, o desenvolvedor encontra em um só lugar tudo (Request, Response e Endpoint) que compõe a interação com a API, sem precisar saltar por várias pastas diferentes no projeto.

---

## 4. O REPR na Prática (Analisando o `OrderFlow`)

Analisando a estrutura deste projeto, na pasta `Features/Orders`, os "Vertical Slices" estão bem definidos:
- `CancelOrder/`
- `CreateOrder/`
- `PayOrder/`
- etc...

Ao abrir o fluxo de `CreateOrder`, vemos o padrão REPR sendo aplicado em três arquivos limpos e sucintos, utilizando o pacote `FastEndpoints`:

### O (R)equest
Os dados de entrada para a criação de um pedido, estritamente o necessário.

```csharp
// Features/Orders/CreateOrder/Request.cs
namespace OrderFlow.Features.Orders.CreateOrder;

public class Request
{
    public string CustomerName { get; set; } = string.Empty;
}
```

### O (R)esponse
O que a API devolverá ao cliente. Não é uma "Entidade Order" genérica cheia de valores nulos, é apenas o que o cliente precisa após um POST bem-sucedido.

```csharp
// Features/Orders/CreateOrder/Response.cs
namespace OrderFlow.Features.Orders.CreateOrder;

public class Response
{
    public Guid OrderId { get; set; }
    public string Message { get; set; } = string.Empty;
}
```

### O (E)ndpoint
A cola entre a entrada, o banco de dados e a saída. Repare que não há injeção de dependências não utilizadas; este endpoint recebe exclusivamente o `DbContext`.

```csharp
// Features/Orders/CreateOrder/Endpoint.cs
using FastEndpoints;
using OrderFlow.Data;
using OrderFlow.Domain;

namespace OrderFlow.Features.Orders.CreateOrder;

public class CreateOrderEndpoint : Endpoint<Request, Response>
{
    public AppDbContext Db { get; set; } = null!;

    public override void Configure()
    {
        Post("/api/orders");
        AllowAnonymous();
    }

    public override async Task HandleAsync(Request req, CancellationToken ct)
    {
        var order = new OrderFlow.Domain.Order
        {
            Id = Guid.NewGuid(),
            CustomerName = req.CustomerName,
            Status = OrderStatus.Created
        };

        await Db.Orders.AddAsync(order, ct);
        await Db.SaveChangesAsync(ct);

        await Send.OkAsync(new Response
        {
            OrderId = order.Id,
            Message = "Pedido criado com sucesso!"
        }, cancellation: ct);
    }
}
```

> Este endpoint implementa o padrão REPR de forma objetiva. Ele define a rota em `Configure()`, processa o payload tipado através do `HandleAsync` e retorna uma resposta modelada para esta exata operação.

---

## 5. O Mito da Violação do DRY

É comum o seguinte questionamento: *"usando o REPR pattern eu tenho problemas de DRY (Don't Repeat Yourself)?"*

Quando olhamos para a arquitetura acima, muitas vezes surge a dúvida: *"E se for necessário criar um `UpdateOrder/Request.cs` que também tem a propriedade `CustomerName`? Isso não é duplicação de código?"*

> **A Falácia da Reutilização Prematura**
> Compartilhar o mesmo DTO entre rotas diferentes acopla regras de negócio distintas à mesma estrutura. Quando uma regra muda, todos os outros endpoints são impactados.

**No REPR, prioriza-se o Isolamento em vez do DRY superficial:**
1. Os DTOs de Request e Response são considerandos **"Mensagens"**. Eles pertencem ao Endpoint, não ao modelo de domínio.
2. Não há problema em ter DTOs estruturalmente semelhantes em pastas separadas.
3. Isso garante que ao adicionar um campo obrigatório para "Cancelar um Pedido", a validação do formulário de "Criar um Pedido" não será quebrada por acidente.

*Obs: Apenas garanta que o princípio DRY seja mantido em regras de domínio, consultas complexas no banco de dados e validações genéricas.*

---

## 6. Conclusão

Ao utilizar o padrão **REPR** (e adotando frameworks baseados neste padrão, como o `FastEndpoints`), ganha-se:
- **Testabilidade:** Endpoints são extremamente fáceis de mockar e testar.
- **Leitura:** É possível entender o fluxo inteiro de um "caso de uso" lendo um único arquivo e seus tipos de entrada/saída.
- **Escalabilidade:** Novas features são adicionadas criando novas pastas, sem o perigo de inchar controladores legados.
- **Desacoplamento:** Mudanças em uma feature não introduzem bugs (efeitos colaterais) em outras features aparentemente similares.

Este projeto adota uma das arquiteturas mais modernas e com alta manutenibilidade do ecossistema .NET, focada na evolução independente de cada funcionalidade.
