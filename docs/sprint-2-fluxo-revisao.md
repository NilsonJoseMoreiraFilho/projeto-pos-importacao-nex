# Sprint 2 - Fluxo de Revisao Humana e Interface

## Objetivo

A Sprint 2 transforma a fatia backend da Sprint 1 em um fluxo operavel pela dona/operadora da loja. O foco nao e integrar definitivamente com o NEX, mas permitir que um documento seja enviado, interpretado como proposta de importacao, revisado por uma pessoa e aprovado somente quando nao houver pendencias bloqueantes.

## Tarefas do quadro

1. Definicao do fluxo de revisao humana antes da confirmacao da importacao.
2. Desenvolvimento em frontend da interface de upload de documentos.
3. Desenvolvimento em frontend da interface para revisao humana e confirmacao.

## Fluxo funcional

1. Operadora acessa a tela web.
2. Operadora seleciona um documento de fornecedor.
3. Frontend envia o arquivo para o backend.
4. Backend seleciona o reader pelo tipo do arquivo:
   - JSON controlado para massa demonstrativa;
   - XLSX para planilhas de fornecedor;
   - PDF sidecar para PDFs enquanto nao houver parser/OCR real;
   - imagem/OCR sidecar para fotos enquanto nao houver OCR real.
5. Backend cria uma `ImportProposal`.
6. Backend executa validacoes de fornecedor, itens, totais, pagina faltante e vinculo de produto.
7. Frontend exibe:
   - status da proposta;
   - validacoes bloqueantes;
   - campos extraidos com confianca;
   - rascunho da compra;
   - tabela de itens.
8. Operadora revisa campo a campo:
   - aceitar;
   - corrigir;
   - rejeitar.
9. Backend aplica as decisoes de revisao, atualiza o rascunho quando houver correcao e reexecuta validacoes.
10. Operadora so consegue confirmar/aprovar quando nao houver erro bloqueante nem campo rejeitado.
11. Backend gera `ApprovedPurchase`.
12. A exportacao real para NEX continua atras de `ManagementSystemExporterPort`; CSV/XLSX seguem como saida demonstrativa.

## Regras de revisao

- Documento vindo de foto/OCR deve exigir revisao humana.
- Campo rejeitado bloqueia aprovacao.
- Correcao de campo deve atualizar o rascunho da compra quando o campo for suportado.
- Apos revisao, a proposta deve ser revalidada.
- Aprovacao nao deve chamar diretamente NEX nesta sprint.

## Contrato HTTP proposto

### `POST /api/imports`

Recebe `multipart/form-data`.

Campos:

- `file`: documento enviado.
- `reader`: `auto`, `manualjson`, `xlsx`, `pdf` ou `imageocr`.
- `requiresHumanReview`: `true` ou `false`.

Resposta: `ImportProposal`.

### `GET /api/imports/{id}`

Retorna a proposta atual.

### `POST /api/imports/{id}/review`

Recebe lista de decisoes:

```json
{
  "decisions": [
    {
      "fieldPath": "purchase.supplier.legalName",
      "decision": "CORRECT",
      "correctedValue": "Fornecedor Corrigido",
      "reviewedBy": "operadora"
    }
  ]
}
```

Resposta: `ImportProposal` revalidada.

### `POST /api/imports/{id}/approve`

Confirma a importacao revisada.

Resposta: `ApprovedPurchase`.

## Escopo explicitamente fora da Sprint 2

- OCR real em producao.
- Login/autenticacao.
- Integracao real com NEX.
- Persistencia definitiva em banco.
- Cadastro completo de produtos internos.
