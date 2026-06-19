# Sprint 1 - Arquitetura Hexagonal e Modelo de Domínio

## 1. Enquadramento da Sprint

Este artefato faz parte do Documento de Arquitetura de Software (DAS) do Projeto Aplicado. Ele registra as decisões iniciais da Sprint 1, cujo foco é definir o desenho geral da arquitetura da solução para as três sprints e implementar, no backend, uma primeira fatia vertical do fluxo de importação: entrada de documento, interpretação dos dados, normalização, validação e preparação de saída para integração.

A Sprint 1 não tem como objetivo entregar uma integração real completa com o Nex, sistema da Nextar referido neste relatório como NEX, porque o mecanismo oficial de integração ainda depende de confirmação. O fluxo principal da POC considera que a loja provavelmente receberá fotos, pedidos, PDFs ou planilhas de fornecedores, e não necessariamente o XML fiscal da NF-e. Por isso, a arquitetura deve priorizar a interpretação de documentos não fiscais e manter XML de NF-e apenas como uma alternativa quando o fornecedor disponibilizar esse arquivo real. Enquanto o caminho real do NEX não for confirmado, a saída será demonstrada por JSON/CSV ou formato intermediário compatível com a POC.

Posicionamento no plano de três sprints:

- Sprint 1: arquitetura hexagonal geral, modelo de domínio, backend de entrada/interpretação/validação e saída demonstrativa.
- Sprint 2: protótipo de extração com OCR/LLM para fotos/pedidos, interface React para upload e revisão, e evolução dos adaptadores de entrada.
- Sprint 3: validação com documentos representativos, exportação para formato compatível com o NEX ou destino simulado, consolidação do DAS e retrospectiva final.

Para manter o escopo viável em sete dias, a Sprint 1 deve entregar:

- desenho geral da arquitetura hexagonal da solução completa;
- diagramas C4 de contexto e containers;
- diagrama da arquitetura hexagonal do backend;
- diagrama de sequência do fluxo de importação;
- definição do modelo de domínio limpo;
- definição do modelo de proposta de importação e proveniência dos campos;
- definição das portas de entrada e saída;
- backend em Go com primeira versão do fluxo de entrada de documento;
- interpretação inicial a partir de `sample-import-proposal.json` ou entrada manual equivalente baseada no documento fotografado;
- normalização para proposta de compra;
- validações básicas de consistência implementadas em Go;
- saída demonstrativa em JSON/CSV por meio de adaptador de saída;
- registro de riscos, limitações e decisões arquiteturais.

## 2. Critérios de Aceite da Sprint 1

| Critério | Evidência esperada | Uso no relatório |
| --- | --- | --- |
| Arquitetura hexagonal geral definida | Diagrama ou descrição de camadas, portas e adaptadores para as três sprints | Seção 2.1.1, evidência da execução |
| Visão de contexto definida | Diagrama C4 de contexto | Seção 2.1.1, evidência da solução |
| Visão de containers definida | Diagrama C4 de containers | Seção 2.1.1, evidência da solução |
| Fluxo de importação definido | Diagrama de sequência | Seção 2.1.1, evidência da solução |
| Domínio de compra modelado | Entidades, objetos de valor e regras principais | Seção 2.1.1, evidência da execução |
| Proposta de importação separada do domínio | Modelo de `ImportProposal` e `ExtractedField` | Seção 2.1.1, evidência da solução |
| Backend de entrada desenvolvido | Caso de uso capaz de receber documento ou massa simulada e iniciar importação | Seção 2.1.1, evidência da execução |
| Backend de interpretação desenvolvido | Normalização de dados extraídos para proposta de compra | Seção 2.1.1, evidência da execução |
| Validações básicas implementadas | Testes para total de item, soma dos totais, página faltante e fornecedor não identificado | Seção 2.1.1, evidência dos resultados |
| Saída demonstrativa desenvolvida | JSON/CSV ou formato intermediário gerado por adaptador de saída | Seção 2.1.1, evidência dos resultados |
| Integração com NEX analisada com cautela | Decisão arquitetural sobre API, planilha, XML real e RPA | Seção 2.1.2, retrospectiva e riscos |
| Escopo das próximas sprints delimitado | Backlog técnico por sprint | Seção 2.1.2 e planejamento das Sprints 2 e 3 |

## 3. Análise Inicial do Documento de Fornecedor

A imagem recebida representa um pedido de venda de fornecedor usado pela loja como base para abastecimento de estoque. Do ponto de vista da loja, esse documento precisa ser convertido em uma compra recebida ou entrada de mercadoria.

Campos identificados no exemplo:

- fornecedor: Comércio Ponto Aura Central Ltda.;
- tipo de documento: pedido de venda;
- número do pedido: 11441701;
- data de emissão: 01/05/2026;
- cliente: Nayara Moraes da Silva;
- tabela de preço: 2-ATACADO;
- itens com código, referência, descrição, unidade, embalagem, quantidade, valor unitário e valor total;
- total de produtos: 4.216,94;
- total geral: 4.216,94.

Riscos observados:

- a foto mostra página 1 de 2, portanto o documento está incompleto;
- o comprovante de cartão cobre parte do cabeçalho;
- o comprovante mostra valor diferente do total do pedido;
- o papel está inclinado, amassado e com perspectiva;
- marca-texto amarelo interfere no OCR;
- textos pequenos podem gerar erro em código, referência e valores;
- alguns campos de pagamento e cabeçalho não são confiáveis sem revisão humana.

Conclusão: imagens desse tipo não devem ser enviadas diretamente para o estoque. A arquitetura deve tratar a extração como uma proposta de importação, com confiança por campo, validações automáticas e aprovação humana antes da confirmação.

## 4. Escopo de OCR/LLM na Sprint 1

O fluxo alvo para imagens, recibos e papéis escaneados é:

1. Receber o arquivo original e registrar metadados.
2. Pré-processar a imagem com correção de rotação, perspectiva, contraste e nitidez.
3. Detectar regiões do documento: cabeçalho, tabela de itens, totais, rodapé e anexos sobrepostos.
4. Executar OCR com preservação de layout.
5. Usar LLM ou parser estruturado para transformar o texto extraído em JSON de proposta de compra.
6. Aplicar validações matemáticas e cadastrais.
7. Calcular confiança por campo.
8. Enviar campos incertos para revisão humana.
9. Persistir documento original, resultado extraído, versão revisada e logs de validação.
10. Liberar exportação ou integração apenas após aprovação.

Na Sprint 1, o backend deve implementar a primeira versão desse fluxo usando uma entrada controlada, como `sample-import-proposal.json` ou JSON manual equivalente baseado no documento fotografado. O objetivo é provar o encadeamento de entrada, interpretação, normalização, validação e saída a partir do cenário mais provável da loja: foto ou pedido de fornecedor. Métricas amplas de acurácia de OCR, comparação entre motores de OCR e automação completa de revisão ficam fora do escopo da Sprint 1.

Campos com revisão obrigatória:

- fornecedor;
- número do pedido;
- data;
- todos os itens;
- quantidade;
- valor unitário;
- valor total;
- total da compra;
- divergências entre comprovante e pedido;
- indicação de páginas faltantes.

## 5. Análise Inicial do NEX

Com base em pesquisa inicial nas páginas públicas da Nextar/NEX:

- O NEX é apresentado como sistema de PDV, estoque, catálogo, caixa, pedidos e relatórios.
- A página oficial informa controle de estoque integrado e entrada automática por XML de NF-e.
- A página de controle de estoque menciona entrada em lote por XML da nota eletrônica do fornecedor.
- Há indícios de uso de planilhas em cenários de migração ou carga manual, mas esse ponto ainda depende de confirmação no ambiente real da loja ou com o suporte da Nextar.
- Não foi encontrada, até este momento, documentação pública de API aberta para integração direta.

Essa evidência deve ser tratada com cautela. A entrada por XML de NF-e significa que o NEX pode processar o XML fiscal real emitido pelo fornecedor. A POC não deve tentar gerar uma NF-e a partir de uma foto ou pedido de venda, pois NF-e possui estrutura fiscal própria e pode exigir campos que o pedido fotografado não contém.

Decisão arquitetural provisória:

1. Foto/pedido/PDF/planilha de fornecedor: tratar como fluxo principal de entrada da POC, com interpretação assistida e revisão humana.
2. XML de NF-e: tratar como adaptador de entrada opcional quando o fornecedor fornecer o XML real da nota fiscal.
3. API do NEX: manter como preferência para saída apenas se o suporte ou o ambiente real da loja confirmar documentação, credenciais e endpoints.
4. Planilha compatível com o NEX: tratar como hipótese de saída para a POC somente se o modelo oficial for confirmado.
5. Automação de tela: manter como contingência, usando Playwright/RPA apenas se não houver API nem importação adequada por arquivo.

Implicação para a arquitetura: o domínio não deve conhecer NEX, API, XML, CSV, XLSX, OCR ou Playwright. Ele deve produzir uma compra revisada e aprovada a partir de qualquer origem suportada. A integração com o NEX pertence aos adaptadores de saída.

Fontes consultadas:

- https://www.nextar.com.br/
- https://www.nextar.com.br/recurso/controle-de-estoque
- https://ajuda.nextar.com.br/

## 6. Arquitetura Hexagonal Proposta

A arquitetura será organizada em torno do núcleo de domínio e aplicação. As portas são contratos definidos pelo núcleo; os adaptadores implementam esses contratos ou acionam os casos de uso.

Camadas:

1. Domínio: entidades, objetos de valor, regras de validação e políticas de negócio.
2. Aplicação: casos de uso que orquestram importação, revisão, aprovação e exportação.
3. Portas de entrada: contratos chamados por HTTP/API, CLI ou testes.
4. Portas de saída: contratos para leitura de documentos, persistência, catálogo de produtos e exportação.
5. Adaptadores: implementações concretas para React/HTTP, OCR, XLSX, PDF, XML real de NF-e, CSV/XLSX, banco de dados e RPA.

Fluxo conceitual:

```text
Usuário
  -> Web UI React
  -> API Backend Go
  -> Porta de entrada: ImportPurchaseProposalUseCase
  -> Porta de saída: DocumentReaderPort
  -> RawDocumentExtraction
  -> ImportProposal
  -> Validações de Aplicação e Domínio
  -> Revisão Humana
  -> ApprovedPurchase
  -> Porta de saída: ManagementSystemExporterPort
  -> Adaptador NEX / Planilha / Destino Simulado / RPA
```

Diagramas associados:

- `docs/diagrams/01-c4-contexto.mmd`: visão de contexto da solução.
- `docs/diagrams/02-c4-containers.mmd`: visão de containers e integrações.
- `docs/diagrams/03-arquitetura-hexagonal-backend.mmd`: desenho da arquitetura hexagonal do backend.
- `docs/diagrams/04-sequencia-importacao.mmd`: sequência do fluxo de importação, revisão e saída.

## 7. Separação Entre Domínio e Proposta de Importação

O domínio da compra deve permanecer limpo, sem detalhes de OCR, tela, confiança visual ou bounding box. Esses dados pertencem ao fluxo de importação e revisão.

Núcleo de domínio:

- `Supplier`;
- `Purchase`;
- `PurchaseItem`;
- `PurchaseTotals`;
- `ApprovedPurchase`;
- objetos de valor como `Money`, `Quantity`, `DocumentNumber`, `TaxDocumentKey` e `ProductIdentifier`.

Camada de aplicação/importação:

- `SourceDocument`;
- `RawDocumentExtraction`;
- `ExtractedField`;
- `ImportProposal`;
- `ReviewDecision`;
- `PurchaseImportWorkflow`.

Essa separação evita que o domínio fique dependente da tecnologia de OCR/LLM e permite trocar o leitor de documentos sem alterar regras de compra.

## 8. Modelo de Domínio Inicial

### Supplier

Representa o fornecedor.

Campos:

- id;
- legalName;
- tradeName;
- documentNumber;
- stateRegistration;
- address;
- externalCodes.

Regras:

- CNPJ/CPF deve ser validado quando disponível;
- fornecedor pode ser identificado por CNPJ, nome ou combinação de dados extraídos;
- fornecedores não identificados com segurança devem gerar pendência de revisão.

### Purchase

Representa a compra normalizada antes da aprovação final.

Campos:

- id;
- supplier;
- documentNumber;
- issueDate;
- sourceDocumentType;
- priceTable;
- freightMode;
- paymentInfo;
- items;
- totals.

Regras:

- deve possuir ao menos um item;
- não pode ser aprovada com total inconsistente;
- deve aplicar estratégia de duplicidade conforme tipo documental;
- deve manter rastreabilidade para o documento de origem por meio da aplicação.

### PurchaseItem

Representa um item comprado, sem detalhes técnicos de OCR.

Campos:

- lineNumber;
- supplierProductCode;
- barcode;
- reference;
- description;
- unit;
- packageQuantity;
- quantity;
- unitCost;
- totalCost;
- matchedInternalProductId.

Regras:

- quantidade deve ser maior que zero;
- valor unitário deve ser maior ou igual a zero;
- valor total deve ser aproximadamente quantidade x valor unitário;
- item sem produto interno correspondente é bloqueante até que a pessoa revisora confirme o vínculo, cadastre o produto ou marque o item como exceção justificada;
- unidade deve pertencer a uma lista reconhecida, como UNID, CX, KG, PC ou similar.

### PurchaseTotals

Representa totais do documento.

Campos:

- productsTotal;
- discount;
- addition;
- ipi;
- taxSubstitution;
- fcpSt;
- grandTotal.

Regras:

- soma dos itens deve bater com total de produtos;
- total geral deve bater com produtos + acréscimos + impostos - descontos;
- divergência acima da tolerância deve gerar erro bloqueante.

### ApprovedPurchase

Representa uma compra revisada e aprovada para exportação.

Campos:

- purchase;
- approvedBy;
- approvedAt;
- approvalNotes.

Regras:

- só pode ser criada a partir de `Purchase` sem erros bloqueantes;
- deve ter registro de aprovação humana;
- deve preservar vínculo com a proposta de importação e documento de origem.

### ValidationResult

Representa um alerta, erro ou informação de validação.

Campos:

- severity;
- code;
- field;
- message;
- blocking.

Severidades:

- INFO;
- WARNING;
- ERROR.

## 9. Modelo de Aplicação Para Extração e Revisão

### SourceDocument

Representa o documento original recebido.

Campos:

- id;
- fileName;
- fileType;
- pageCount;
- currentPage;
- checksum;
- storageLocation;
- qualityScore;
- detectedDocumentType;
- observations.

Regras:

- documento com página atual menor que total de páginas deve ficar pendente;
- documento ilegível não pode seguir para aprovação;
- arquivo original deve ser preservado para rastreabilidade.

### RawDocumentExtraction

Representa o resultado bruto de OCR, parser de planilha, parser de PDF ou leitura manual.

Campos:

- sourceDocumentId;
- rawText;
- tables;
- detectedFields;
- extractedAt;
- extractorName;
- extractorVersion.

### ExtractedField

Representa a origem e confiabilidade de um campo extraído.

Campos:

- fieldPath;
- rawText;
- normalizedValue;
- confidence;
- page;
- boundingBox;
- originalValue;
- reviewedValue;
- reviewedBy;
- reviewedAt;
- status.

Estados sugeridos:

- EXTRACTED;
- LOW_CONFIDENCE;
- REVIEWED;
- CORRECTED;
- REJECTED.

### ImportProposal

Representa uma proposta de compra gerada a partir da extração, antes da aprovação humana.

Campos:

- id;
- sourceDocument;
- extractedFields;
- purchaseDraft;
- validationResults;
- status;
- createdAt;
- reviewedAt.

Estados sugeridos:

- RECEIVED;
- EXTRACTION_FAILED;
- PROPOSED;
- NEEDS_REVIEW;
- REVIEWED;
- APPROVED;
- CANCELLED.

### ReviewDecision

Representa uma decisão humana sobre campos extraídos.

Campos:

- proposalId;
- fieldPath;
- decision;
- correctedValue;
- reviewer;
- decidedAt;
- reason.

Decisões:

- ACCEPT;
- CORRECT;
- REJECT;
- REQUEST_NEW_DOCUMENT.

## 10. Regras de Validação

Validações bloqueantes:

- documento incompleto, por exemplo página 1 de 2 sem a página 2;
- quantidade menor ou igual a zero;
- valor unitário negativo;
- total do item incompatível com quantidade x valor unitário acima da tolerância;
- soma dos itens incompatível com total de produtos acima da tolerância;
- total geral incompatível com total de produtos, impostos, acréscimos e descontos;
- fornecedor não identificado;
- tentativa de aprovação sem revisão humana quando a origem for OCR/imagem.

Validações de alerta:

- baixa confiança em descrição de produto;
- produto encontrado por correspondência aproximada, exigindo confirmação humana;
- divergência entre comprovante de pagamento e total do pedido;
- campos de pagamento ausentes;
- documento com marcações, sombras ou áreas obstruídas;
- fornecedor sem CNPJ legível.

Tolerância monetária inicial:

- diferenças de até R$ 0,01 por item podem ser tratadas como arredondamento;
- diferenças maiores devem gerar erro ou revisão obrigatória;
- diferenças no total geral devem ser bloqueantes até correção ou justificativa.

Estratégia de duplicidade:

- para XML de NF-e real: usar chave de acesso da NF-e;
- para pedido fotografado: usar fornecedor + número do pedido + data + tipo documental;
- para documento sem número confiável: usar checksum do arquivo, fornecedor, data e total;
- em caso de baixa confiança, bloquear aprovação automática e exigir revisão.

Contrato lógico de duplicidade:

```go
type DuplicateKey struct {
    SourceDocumentType string
    SupplierDocument   string
    DocumentNumber     string
    IssueDate          time.Time
    TaxDocumentKey     string
    FileChecksum       string
    GrandTotal         Money
}
```

Nem todos os campos serão preenchidos em todos os tipos de documento. A estratégia de comparação deve considerar o tipo documental: XML de NF-e prioriza `TaxDocumentKey`; pedido fotografado prioriza fornecedor, número, data e tipo; documentos sem número confiável usam checksum, fornecedor, data e total como fallback.

## 11. Portas Principais

### Portas de Entrada

```go
type ImportPurchaseProposalUseCase interface {
    Import(ctx context.Context, input ImportPurchaseInput) (ImportProposal, error)
}

type ReviewImportProposalUseCase interface {
    Review(ctx context.Context, proposalID string, decisions []ReviewDecision) (ImportProposal, error)
}

type ApprovePurchaseUseCase interface {
    Approve(ctx context.Context, proposalID string, reviewer string) (ApprovedPurchase, error)
}
```

Na Sprint 1, `ImportPurchaseProposalUseCase` deve ser implementado no backend e exercitado por testes ou por uma entrada manual em JSON. `ReviewImportProposalUseCase` e `ApprovePurchaseUseCase` devem ficar definidos como contratos e regras de transição, sem exigir interface funcional completa. A implementação HTTP completa pode ficar para a Sprint 2.

### DocumentReaderPort

Responsável por extrair dados brutos de um documento.

```go
type DocumentReaderPort interface {
    Read(ctx context.Context, doc SourceDocument) (RawDocumentExtraction, error)
}
```

Implementações previstas:

- `ManualJsonDocumentReader`, para testes e massa simulada da Sprint 1;
- `XlsxDocumentReader`, para planilhas;
- `PdfDocumentReader`, para PDFs;
- `ImageOcrDocumentReader`, para imagem/OCR;
- `NfeXmlReaderAdapter`, para XML real de NF-e quando fornecido pelo fornecedor, sem tratar esse XML como premissa do fluxo principal.

### PurchaseNormalizerPort

Responsável por transformar dados brutos em proposta de compra.

```go
type PurchaseNormalizerPort interface {
    Normalize(ctx context.Context, raw RawDocumentExtraction) (ImportProposal, error)
}
```

### PurchaseRepositoryPort

Responsável por persistir importações, revisões, estados e resultados.

```go
type PurchaseRepositoryPort interface {
    SaveProposal(ctx context.Context, proposal ImportProposal) error
    FindProposalByID(ctx context.Context, id string) (ImportProposal, error)
    SaveReview(ctx context.Context, proposalID string, decisions []ReviewDecision) error
    SaveApprovedPurchase(ctx context.Context, purchase ApprovedPurchase) error
    SaveExportResult(ctx context.Context, proposalID string, result ExportResult) error
    ExistsDuplicate(ctx context.Context, key DuplicateKey) (bool, error)
}
```

### ProductCatalogPort

Responsável por localizar produtos internos.

```go
type ProductCatalogPort interface {
    MatchProduct(ctx context.Context, item PurchaseItem) (ProductMatchResult, error)
}
```

Critérios de correspondência:

- código de barras/EAN;
- código do fornecedor;
- referência;
- descrição aproximada;
- revisão manual quando houver baixa confiança.

### ManagementSystemExporterPort

Responsável por enviar ou preparar dados para o sistema de gestão.

```go
type ManagementSystemExporterPort interface {
    ExportApprovedPurchase(ctx context.Context, purchase ApprovedPurchase) (ExportResult, error)
}
```

Implementações previstas:

- `NexApiExporter`, se houver API confirmada;
- `NexSpreadsheetExporter`, se houver modelo oficial de planilha compatível;
- `CsvExportAdapter`, como destino simulado e evidência da POC;
- `NexRpaExporter`, se for necessário preencher tela.

## 12. Papel da Interface React

Na Sprint 1, a interface React não precisa estar funcional. O DAS deve, porém, deixar claro o que a tela de revisão precisará exibir nas próximas sprints:

- imagem ou documento original;
- campos extraídos com valor original, valor normalizado e confiança;
- alertas e erros de validação;
- tabela de itens com quantidade, valor unitário e total;
- indicação de campos obrigatórios pendentes;
- comparação com produtos internos encontrados;
- ações de aceitar, corrigir, rejeitar ou solicitar novo documento;
- botão de aprovação disponível apenas quando não houver erros bloqueantes.

Evidência suficiente para a Sprint 1:

- wireframe simples da tela de revisão; ou
- contrato JSON esperado entre backend e frontend; ou
- descrição dos estados de tela no DAS.

## 13. Casos de Uso Planejados

### Sprint 1

Casos de uso entregáveis no backend:

- importar proposta a partir de JSON manual ou documento de entrada controlado;
- interpretar dados extraídos;
- normalizar compra;
- validar itens e totais;
- detectar página faltante;
- detectar divergência de total;
- detectar item sem produto interno;
- gerar saída demonstrativa em JSON/CSV por adaptador de saída.

### Sprints seguintes

Casos de uso planejados para evolução:

- importar documento por imagem/OCR, priorizando foto ou pedido de fornecedor;
- revisar proposta na interface React;
- aprovar compra revisada;
- exportar compra aprovada para CSV/XLSX, API ou RPA conforme integração viável.

## 14. Estrutura Inicial Sugerida Para Backend Go

```text
backend/
  cmd/
    api/
      main.go
  internal/
    domain/
      supplier.go
      purchase.go
      purchase_item.go
      purchase_totals.go
      approved_purchase.go
      validation.go
      money.go
      quantity.go
    application/
      import_proposal.go
      review_proposal.go
      approve_purchase.go
      export_purchase.go
    ports/
      inbound/
        import_purchase_proposal.go
        review_import_proposal.go
        approve_purchase.go
      outbound/
        document_reader.go
        purchase_repository.go
        product_catalog.go
        management_system_exporter.go
    adapters/
      http/
      persistence/
      documentreader/
        manualjson/
        xlsx/
        pdf/
        imageocr/
        nfexml/
      export/
        csv/
        nexspreadsheet/
        nexapi/
        rpa/
    tests/
```

## 15. Evidências Esperadas Para Sprint 1

| Evidência | Critério avaliado | Onde entra no relatório |
| --- | --- | --- |
| Diagrama da arquitetura hexagonal | Separação entre domínio, aplicação, portas e adaptadores da solução completa | 2.1.1 Solução |
| Diagrama C4 de contexto | Relação entre loja, fornecedores, documentos, solução e NEX | 2.1.1 Solução |
| Diagrama C4 de containers | Distribuição entre React, backend Go, armazenamento e adaptadores | 2.1.1 Solução |
| Diagrama de sequência | Fluxo de upload, interpretação, validação, revisão e saída | 2.1.1 Solução |
| Modelo de domínio | Entidades e regras de compra | 2.1.1 Solução |
| Modelo de `ImportProposal` e `ExtractedField` | Rastreabilidade e revisão humana | 2.1.1 Solução |
| Backend de importação em Go | Entrada, interpretação, normalização e validação da proposta | 2.1.1 Evidência da execução |
| `sample-import-proposal.json` baseado na foto | Massa de entrada representativa para desenvolvimento e testes | 2.1.1 Evidência dos resultados |
| Validações executáveis em Go | Critérios para bloquear ou liberar aprovação | 2.1.1 Evidência dos resultados |
| Saída JSON/CSV demonstrativa | Preparação para integração com NEX por adaptador substituível | 2.1.1 Evidência dos resultados |
| Testes unitários implementados | Total de item, soma dos totais, página faltante e fornecedor não identificado | 2.1.1 Evidência da execução |
| Decisão sobre NEX | Risco de integração e estratégia de adaptadores | 2.1.2 Retrospectiva |
| Wireframe ou contrato JSON da revisão | Preparação para React na Sprint 2 | 2.1.1 Solução |

## 16. Decisões Arquiteturais Iniciais

### ADR-001 - Usar arquitetura hexagonal

Decisão: separar domínio, aplicação, portas e adaptadores.

Justificativa: o projeto precisa lidar com múltiplos formatos de entrada e possíveis mecanismos diferentes de saída para o NEX. O domínio deve permanecer estável mesmo se o formato do documento, OCR ou sistema de destino mudar.

### ADR-002 - Revisão humana obrigatória

Decisão: nenhuma compra extraída de OCR/imagem será exportada automaticamente sem revisão.

Justificativa: o exemplo real possui página faltante, cabeçalho coberto, valor divergente de comprovante e baixa legibilidade em alguns campos.

### ADR-003 - Integração com NEX por adaptador substituível

Decisão: a saída para o NEX será representada por `ManagementSystemExporterPort`.

Justificativa: a documentação pública indica XML de NF-e como entrada, mas a operação real da loja provavelmente trabalha com fotos e pedidos de fornecedor. Além disso, a documentação pública não confirma API nem modelo oficial de planilha para o cenário da POC. A arquitetura deve aceitar API, arquivo ou RPA sem alterar o domínio.

### ADR-004 - XML de NF-e é entrada, não saída gerada pela POC

Decisão: XML de NF-e só será usado quando fornecido como documento fiscal real pelo fornecedor. O fluxo principal da POC não depende desse arquivo.

Justificativa: gerar NF-e a partir de foto ou pedido de venda não é adequado para a POC, pois NF-e contém dados fiscais específicos e não deve ser sintetizada como simples formato de importação.

### ADR-005 - Dados extraídos devem guardar confiança e origem

Decisão: campos extraídos devem carregar confiança, origem, página, região e status de revisão.

Justificativa: em documentos fotografados, a rastreabilidade de cada campo é essencial para revisão, auditoria e explicação de erros.

## 17. Próximas Ações

1. Confirmar com a loja qual versão/plano do NEX é usada.
2. Verificar no NEX se existe menu de importação por planilha para produtos, fornecedores ou estoque.
3. Solicitar ao suporte NEX modelo oficial de planilha ou documentação de importação.
4. Verificar se a loja recebe XML de NF-e dos fornecedores; caso contrário, manter foto/pedido como fluxo principal.
5. Solicitar a segunda página do pedido fotografado.
6. Criar `sample-import-proposal.json` com base no pedido recebido.
7. Implementar validações de domínio da Sprint 1 em Go.
8. Gerar saída JSON/CSV como adaptador demonstrativo, sem prometer integração real com o NEX nesta sprint.
