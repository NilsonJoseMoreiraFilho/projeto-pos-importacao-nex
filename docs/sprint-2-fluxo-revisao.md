# Sprint 2 - Fluxo de Revisao Humana e Interface

## Objetivo

A Sprint 2 transforma a fatia backend da Sprint 1 em um fluxo operavel pela dona/operadora da loja. O foco da revisao humana nao e aprovar campo a campo em um processo formal, mas permitir uma conferencia rapida em uma grade desktop, semelhante a uma planilha, antes da confirmacao da importacao.

O fluxo visual fica separado em duas telas: uma tela dedicada somente ao upload/processamento do documento e uma segunda tela dedicada a validacao, edicao da grade e aprovacao.

## Tarefas do quadro

1. Definicao do fluxo de revisao humana antes da confirmacao da importacao.
2. Desenvolvimento em frontend da interface de upload de documentos.
3. Desenvolvimento em frontend da interface para revisao humana e confirmacao.

## Fluxo funcional

1. Operadora acessa a tela 1, dedicada ao upload.
2. Operadora seleciona um documento de fornecedor.
3. Frontend envia o arquivo para o backend.
4. Backend seleciona o reader pelo tipo do arquivo:
   - JSON controlado para massa demonstrativa;
   - XLSX para planilhas de fornecedor;
   - PDF sidecar para PDFs enquanto nao houver parser/OCR real;
   - imagem via `imageocr`, usando OpenAI Vision quando `OPENAI_API_KEY` estiver configurada, sidecar `.ocr.json` quando existir, ou fallback demonstrativo sem chave.
5. Backend cria uma `ImportProposal`.
6. Backend executa validacoes de fornecedor, itens, totais, pagina faltante e campos obrigatorios.
7. Frontend navega para a tela 2, dedicada a validacao.
8. Frontend exibe uma tabela editavel em formato de planilha:
   - uma linha por item importado;
   - colunas de codigo, referencia, descricao, unidade, quantidade, valor unitario, valor total e produto interno;
   - cabecalho da compra em campos editaveis;
   - campos obrigatorios vazios destacados em vermelho.
9. Operadora edita diretamente as celulas incorretas ou vazias.
10. Frontend salva as alteracoes e backend reexecuta as validacoes.
11. Botao de aprovar fica na tela 2, junto da tabela validada.
12. Operadora aprova quando nao houver obrigatorios pendentes.
13. Backend gera `ApprovedPurchase` e aciona o adaptador de integracao/saida.
14. A integracao direta com NEX continua atras de `ManagementSystemExporterPort`; enquanto API/modelo oficial nao estiver confirmado, CSV/XLSX ou adapter simulado continuam como caminho demonstrativo.

## Regras de revisao

- A revisao humana acontece em tabela editavel, nao em uma fila de decisao por campo.
- Upload e validacao ficam em telas separadas.
- Campos obrigatorios vazios devem aparecer em vermelho.
- Edicao de celula deve atualizar o rascunho da compra.
- Apos edicao, a proposta deve ser revalidada.
- Aprovacao fica na tela de validacao, na mesma tela da tabela.
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

## Configuracao de IA para imagem

Para testar extracao real de foto por IA:

1. Configurar `OPENAI_API_KEY` no ambiente do backend.
2. Opcionalmente configurar `OPENAI_VISION_MODEL`; o padrao atual do projeto e `gpt-4o-mini`.
3. Subir uma imagem usando `reader=imageocr`.

Ordem de resolucao do `imageocr`:

1. Usa sidecar `<imagem>.ocr.json`, quando existir.
2. Usa OpenAI Vision, quando `OPENAI_API_KEY` existir.
3. Usa JSON demonstrativo, quando nao houver chave.

Mesmo com IA real, a proposta normalmente entra em revisao humana porque produto NEX, totais e campos de baixa confianca ainda precisam ser confirmados.

## Escopo explicitamente fora da Sprint 2

- OCR/IA com acuracia produtiva garantida.
- Login/autenticacao.
- Integracao real com NEX sem confirmacao de API/modelo oficial.
- Persistencia definitiva em banco.
- Cadastro completo de produtos internos.
