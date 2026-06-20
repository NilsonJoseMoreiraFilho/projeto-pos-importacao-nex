# Sprint 2 - Fluxo de Revisao Humana e Interface

## Objetivo

A Sprint 2 transforma a fatia backend da Sprint 1 em um fluxo operavel pela dona/operadora da loja. O foco da revisao humana nao e aprovar campo a campo em um processo formal, mas permitir uma conferencia rapida em uma grade desktop, semelhante a uma planilha, antes da confirmacao da importacao.

## Tarefas do quadro

1. Definicao do fluxo de revisao humana antes da confirmacao da importacao.
2. Desenvolvimento em frontend da interface de upload de documentos.
3. Desenvolvimento em frontend da interface para revisao humana e confirmacao.

## Fluxo funcional

1. Operadora acessa uma tela desktop de importacao.
2. Operadora seleciona um documento de fornecedor.
3. Frontend envia o arquivo para o backend.
4. Backend seleciona o reader pelo tipo do arquivo:
   - JSON controlado para massa demonstrativa;
   - XLSX para planilhas de fornecedor;
   - PDF sidecar para PDFs enquanto nao houver parser/OCR real;
   - imagem/OCR sidecar para fotos enquanto nao houver OCR real.
5. Backend cria uma `ImportProposal`.
6. Backend executa validacoes de fornecedor, itens, totais, pagina faltante e campos obrigatorios.
7. Frontend exibe uma tabela editavel em formato de planilha:
   - uma linha por item importado;
   - colunas de codigo, referencia, descricao, unidade, quantidade, valor unitario, valor total e produto interno;
   - cabecalho da compra em campos editaveis;
   - campos obrigatorios vazios destacados em vermelho.
8. Operadora edita diretamente as celulas incorretas ou vazias.
9. Frontend salva as alteracoes e backend reexecuta as validacoes.
10. Botao de aprovar fica na mesma tela.
11. Operadora aprova quando nao houver obrigatorios pendentes.
12. Backend gera `ApprovedPurchase` e aciona o adaptador de integracao/saida.
13. A integracao direta com NEX continua atras de `ManagementSystemExporterPort`; enquanto API/modelo oficial nao estiver confirmado, CSV/XLSX ou adapter simulado continuam como caminho demonstrativo.

## Regras de revisao

- A revisao humana acontece em tabela editavel, nao em uma fila de decisao por campo.
- Campos obrigatorios vazios devem aparecer em vermelho.
- Edicao de celula deve atualizar o rascunho da compra.
- Apos edicao, a proposta deve ser revalidada.
- Aprovacao fica na mesma tela da tabela.
- Aprovacao deve acionar a saida/integracao por `ManagementSystemExporterPort`.
- Para a POC, a saida pode ser CSV/XLSX ou adapter simulado enquanto a integracao real com NEX nao estiver confirmada.

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

Recebe edicoes feitas na grade:

```json
{
  "edits": [
    {
      "fieldPath": "purchase.items[0].quantity",
      "value": 24,
      "editedBy": "operadora"
    }
  ]
}
```

Resposta: `ImportProposal` revalidada.

### `POST /api/imports/{id}/approve`

Confirma a importacao revisada e aciona a saida/integracao configurada.

Resposta: `ApprovedPurchase`.

## Escopo explicitamente fora da Sprint 2

- OCR real em producao.
- Login/autenticacao.
- Integracao real com NEX sem confirmacao de API/modelo oficial.
- Persistencia definitiva em banco.
- Cadastro completo de produtos internos.
