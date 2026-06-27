# Sprint 2 - Fluxo de Revisao Humana e Interface

## Objetivo

A Sprint 2 transforma a fatia backend da Sprint 1 em um fluxo operavel pela dona/operadora da loja. O foco da revisao humana nao e aprovar campo a campo em um processo formal, mas permitir uma conferencia rapida em uma grade desktop, semelhante a uma planilha, antes de baixar um Excel para importacao manual.

O fluxo visual fica separado em duas telas: uma tela dedicada somente ao upload/processamento do documento e uma segunda tela dedicada a validacao, edicao da grade e download do Excel.

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
   - Divergencia de totais, pagina faltante, campos essenciais ausentes e revisao humana pendente aparecem como pontos de conferencia antes do download.
7. Frontend navega para a tela 2, dedicada a validacao.
8. Frontend exibe uma tabela editavel em formato de planilha:
   - uma linha por item importado;
   - colunas de codigo, referencia, descricao, unidade, quantidade, valor unitario e valor total;
   - cabecalho da compra em campos editaveis;
   - campos obrigatorios vazios destacados em vermelho.
9. Operadora edita diretamente as celulas incorretas ou vazias.
10. Frontend salva as alteracoes e backend reexecuta as validacoes.
11. Botao de baixar Excel fica na tela 2, junto da tabela validada.
12. Operadora baixa a planilha com os dados processados.
13. Backend gera um arquivo XLSX a partir do rascunho revisado.
14. A importacao no NEX e feita manualmente pela operadora enquanto API/modelo oficial nao estiver confirmado.

## Regras de revisao

- A revisao humana acontece em tabela editavel, nao em uma fila de decisao por campo.
- Upload e validacao ficam em telas separadas.
- Campos obrigatorios vazios devem aparecer em vermelho.
- Edicao de celula deve atualizar o rascunho da compra.
- Apos edicao, a proposta deve ser revalidada.
- O download do Excel fica na tela de validacao, na mesma tela da tabela.
- Para a POC, a saida principal e XLSX para importacao manual enquanto a integracao real com NEX nao estiver confirmada.
- Vinculo item a item com produto interno/NEX nao e obrigatorio nesta sprint e nao impede o download do Excel.

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

### `POST /api/imports/{id}/download-xlsx`

Recebe as edicoes atuais da grade, salva a revisao e retorna um arquivo XLSX para download.

Resposta: arquivo `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`.

## Configuracao de IA para imagem

Para testar extracao real de foto por IA:

1. Configurar `OPENAI_API_KEY` no ambiente do backend.
2. Opcionalmente configurar `OPENAI_VISION_MODEL`; o padrao atual do projeto e `gpt-4o`, pois ele preserva melhor tabelas longas em fotos.
3. Subir uma imagem usando `reader=imageocr`.

Ordem de resolucao do `imageocr`:

1. Usa sidecar `<imagem>.ocr.json`, quando existir.
2. Usa OpenAI Vision, quando `OPENAI_API_KEY` existir.
3. Usa JSON demonstrativo, quando nao houver chave.

Mesmo com IA real, a proposta pode entrar em revisao humana por origem OCR/imagem, totais divergentes, documento incompleto, campos obrigatorios ausentes ou campos de baixa confianca. Vinculo com produto interno/NEX nao bloqueia esta sprint.

## Escopo explicitamente fora da Sprint 2

- OCR/IA com acuracia produtiva garantida.
- Login/autenticacao.
- Integracao real com NEX sem confirmacao de API/modelo oficial.
- Persistencia definitiva em banco.
- Cadastro completo de produtos internos.
