# Pesquisa de Integração com NEX/Nextar

## Objetivo

Avaliar caminhos de integração entre a solução de importação de compras e o NEX, considerando:

- existência de API pública;
- importação por arquivo;
- entrada de compras/estoque;
- possibilidade de automação de tela;
- impacto na arquitetura hexagonal.

Premissa revisada: na operação real da loja, não se deve assumir que a usuária terá o XML de NF-e. O fluxo principal da POC deve partir de fotos, pedidos, PDFs ou planilhas recebidas de fornecedores. XML de NF-e deve ser tratado como caminho alternativo quando existir, não como premissa do projeto.

## Fontes oficiais verificadas

- Site oficial Nextar: https://www.nextar.com.br/
- Recurso de controle de estoque: https://www.nextar.com.br/recurso/controle-de-estoque
- Central de ajuda Nextar: https://ajuda.nextar.com.br/
- Categoria Estoque: https://ajuda.nextar.com.br/categoria/estoque
- Tutorial oficial: https://ajuda.nextar.com.br/tutorial/entrada-de-estoque-via-xml-nfe
- Categoria Integrações: https://ajuda.nextar.com.br/categoria/integracoes

## Achados

### 1. Entrada de estoque via XML de NF-e existe

A Central de Ajuda da Nextar possui tutorial oficial chamado **Como fazer uma entrada de estoque via XML**.

O tutorial informa que:

- o recurso permite dar entrada no estoque por meio do XML da nota de compra emitida pelo fornecedor;
- está disponível apenas nos planos Premium e Fiscal;
- é necessário possuir o arquivo XML de uma nota fiscal emitida pelo fornecedor;
- o Nex reconhece dados da nota, como preço de custo e quantidade comprada;
- o fluxo ocorre na tela de Estoque > Transações > Compra > Ler XML;
- o usuário precisa vincular fornecedor;
- cada produto da nota precisa ser vinculado a um produto do NEX;
- quando o produto ainda não existe, é possível cadastrá-lo durante o processo;
- o NEX pode sugerir produto já cadastrado;
- pode ser necessário converter unidades;
- ao final, o estoque é atualizado.

Conclusão: XML de NF-e é o caminho oficial mais forte encontrado para entrada de compra/estoque no NEX, mas depende de a loja receber o XML fiscal real do fornecedor e de possuir plano compatível. Portanto, é uma oportunidade de integração, não o fluxo principal da POC.

### 2. XML de NF-e é entrada, não saída artificial da POC

O XML aceito pelo NEX é o XML fiscal real da nota emitida pelo fornecedor. A solução não deve gerar uma NF-e sintética a partir de foto ou pedido de venda, porque NF-e contém dados fiscais próprios.

Implicação: se a loja receber XML real do fornecedor, a solução pode tratar esse XML como adaptador de entrada. Se a loja receber apenas foto, pedido, PDF ou planilha, a solução deve gerar uma proposta de compra revisada e uma saída intermediária, mas não uma NF-e.

### 3. API pública não foi encontrada

Na busca em páginas oficiais e Central de Ajuda, não foi encontrada documentação pública de API para cadastro de compras, produtos ou movimentações de estoque.

A categoria **Integrações** da Central de Ajuda trata principalmente de:

- modo plataforma;
- uso em nuvem/web/app/computador;
- meios de pagamento;
- migração para modo plataforma.

Não apareceu documentação pública de endpoints, autenticação, tokens, SDK ou API REST para integração externa.

Conclusão: API deve permanecer como hipótese dependente de confirmação com suporte Nextar ou acesso ao ambiente real.

### 4. Importação por planilha ainda não está confirmada para este caso

Há sinais de migração/carga de dados na Central de Ajuda, mas não foi encontrado, nas páginas verificadas, um modelo oficial público de planilha para importar compras/entrada de estoque no cenário da POC.

Conclusão: planilha CSV/XLSX deve ser tratada como hipótese de saída intermediária ou demonstrativa até confirmação de modelo oficial pelo suporte ou dentro do próprio NEX.

### 5. Automação de tela é contingência viável, mas mais frágil

Como o fluxo oficial de XML acontece por telas do NEX, e como não há API pública confirmada, automação de tela com Playwright/RPA pode ser contingência.

Riscos:

- login/autenticação;
- mudanças de layout;
- campos dependentes de estado;
- necessidade de vincular produtos manualmente;
- baixa robustez para produção.

Conclusão: RPA deve ser última opção, não caminho principal.

## Decisão arquitetural recomendada

Ordem de preferência:

1. **Foto, pedido, PDF ou planilha do fornecedor como entrada principal da POC**, com interpretação assistida e revisão humana.
2. **XML real de NF-e como entrada alternativa**, quando o fornecedor fornecer esse arquivo.
3. **API NEX**, somente se o suporte/ambiente real confirmar documentação oficial.
4. **Planilha CSV/XLSX**, somente se houver modelo oficial aceito pelo NEX para compra/estoque.
5. **Automação de tela**, como contingência.
6. **JSON/CSV demonstrativo**, como saída da Sprint 1 enquanto a integração real não estiver confirmada.

## Impacto na arquitetura hexagonal

A aplicação deve ter uma porta de saída:

```go
type ManagementSystemExporterPort interface {
    ExportApprovedPurchase(ctx context.Context, purchase ApprovedPurchase) (ExportResult, error)
}
```

Adaptadores possíveis:

- `NexApiExporter`, se API for confirmada;
- `NexSpreadsheetExporter`, se modelo de planilha for confirmado;
- `NexRpaExporter`, se for necessário preencher tela;
- `CsvExportAdapter`, como evidência demonstrativa da POC.

Para entrada:

- `ImageOcrDocumentReader`, como caminho principal para fotos e pedidos digitalizados;
- `PdfDocumentReader`, quando fornecedores enviarem PDFs;
- `XlsxDocumentReader`, quando fornecedores enviarem planilhas;
- `NfeXmlReaderAdapter`, quando existir XML real da NF-e;

## Conclusão

O caminho técnico mais defensável hoje é projetar o domínio e os casos de uso independentes do NEX, mantendo o NEX como adaptador externo. A Sprint 1 deve demonstrar a geração de uma compra normalizada e revisada a partir de foto/pedido ou massa equivalente, com saída intermediária JSON/CSV. A integração real deve ser decidida após confirmar se a loja recebe XML de NF-e, se possui plano Premium/Fiscal, e se o suporte Nextar fornece API ou modelo de planilha. Sem essas confirmações, o fluxo principal continua sendo interpretação de documento não fiscal + revisão humana + saída intermediária/adaptador substituível.
