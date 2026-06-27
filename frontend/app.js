const { useEffect, useMemo, useState } = React;
const h = React.createElement;

const api = {
  async createImport({ file, reader, requiresHumanReview }) {
    const formData = new FormData();
    formData.append("file", file);
    formData.append("reader", reader);
    formData.append("requiresHumanReview", String(requiresHumanReview));
    return readResponse(await fetch("/api/imports", { method: "POST", body: formData }));
  },
  async review(id, edits) {
    return readResponse(
      await fetch(`/api/imports/${id}/review`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ edits }),
      }),
    );
  },
  async downloadXlsx(id, edits) {
    const response = await fetch(`/api/imports/${id}/download-xlsx`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ edits }),
      });
    if (!response.ok) {
      const data = await response.json().catch(() => ({}));
      throw new Error(data.error || `Erro HTTP ${response.status}`);
    }
    return response.blob();
  },
  async createDemoImport() {
    const response = await fetch("/demo-import-proposal.json", { cache: "no-store" });
    if (!response.ok) {
      throw new Error("Nao foi possivel carregar a massa de exemplo.");
    }
    const blob = await response.blob();
    const file = new File([blob], "demo-import-proposal.json", { type: "application/json" });
    return this.createImport({ file, reader: "manualjson", requiresHumanReview: false });
  },
};

async function readResponse(response) {
  const data = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error(data.error || `Erro HTTP ${response.status}`);
  }
  return data;
}

function App() {
  const [proposal, setProposal] = useState(null);
  const [downloaded, setDownloaded] = useState(false);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const onUploadScreen = !proposal;

  async function run(action) {
    setError("");
    setBusy(true);
    try {
      await action();
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  }

  function resetFlow() {
    setProposal(null);
    setDownloaded(false);
    setError("");
  }

  return h(
    "div",
    { className: "app" },
    h(
      "header",
      { className: "topbar" },
      h("div", null, h("h1", null, "Importacao de compras NEX"), h("p", null, onUploadScreen ? "Tela 1 - upload do documento" : "Tela 2 - validacao, edicao e download")),
      h(Stepper, { active: onUploadScreen ? "upload" : "review" }),
    ),
    error ? h("div", { className: "global-error" }, error) : null,
    onUploadScreen
      ? h(UploadScreen, {
          busy,
          onUpload: (input) =>
            run(async () => {
              setDownloaded(false);
              setProposal(await api.createImport(input));
            }),
          onDemo: () =>
            run(async () => {
              setDownloaded(false);
              setProposal(await api.createDemoImport());
            }),
        })
      : h(ReviewScreen, {
          proposal,
          downloaded,
          busy,
          onNewUpload: resetFlow,
          onReview: (edits) =>
            run(async () => {
              setProposal(await api.review(proposal.id, edits));
              setDownloaded(false);
            }),
          onDownload: (edits) =>
            run(async () => {
              const blob = await api.downloadXlsx(proposal.id, edits);
              downloadBlob(blob, `nex-import-${proposal.purchaseDraft?.documentNumber || proposal.id}.xlsx`);
              setDownloaded(true);
            }),
        }),
  );
}

function Stepper({ active }) {
  return h(
    "nav",
    { className: "stepper", "aria-label": "Etapas da importacao" },
    h("span", { className: active === "upload" ? "step active" : "step done" }, "1 Upload"),
    h("span", { className: active === "review" ? "step active" : "step" }, "2 Validacao"),
  );
}

function countLabel(count, singular, plural) {
  return `${count} ${count === 1 ? singular : plural}`;
}

function UploadScreen({ busy, onUpload, onDemo }) {
  const [file, setFile] = useState(null);
  const [reader, setReader] = useState("auto");
  const [requiresHumanReview, setRequiresHumanReview] = useState(true);

  return h(
    "main",
    { className: "screen upload-screen" },
    h(
      "section",
      { className: "panel upload-card" },
      h("div", { className: "screen-kicker" }, "Tela 1"),
      h("h2", null, "Upload do documento"),
      h("p", { className: "muted" }, "Selecione o arquivo recebido do fornecedor. Depois do processamento, o sistema abre a tela de validacao em tabela."),
      h(
        "div",
        { className: "upload-dropzone" },
        h("strong", null, file ? file.name : "Escolha um arquivo para processar"),
        h("span", null, "JSON controlado, XLSX, PDF ou imagem com IA/OCR"),
        h("input", { type: "file", onChange: (event) => setFile(event.target.files[0] || null) }),
      ),
      h(
        "div",
        { className: "upload-options" },
        h(
          "div",
          { className: "field" },
          h("label", null, "Reader"),
          h(
            "select",
            { value: reader, onChange: (event) => setReader(event.target.value) },
            ["auto", "manualjson", "xlsx", "pdf", "imageocr"].map((item) => h("option", { key: item, value: item }, item)),
          ),
        ),
        h(
          "label",
          { className: "checkbox" },
          h("input", { type: "checkbox", checked: requiresHumanReview, onChange: (event) => setRequiresHumanReview(event.target.checked) }),
          h("span", null, "Exigir conferencia humana"),
        ),
      ),
      h(
        "div",
        { className: "upload-actions" },
        h(
          "button",
          { className: "button primary upload-action", disabled: busy || !file, onClick: () => onUpload({ file, reader, requiresHumanReview }) },
          busy ? "Processando..." : "Processar e abrir validacao",
        ),
        h(
          "button",
          { className: "button secondary upload-action", disabled: busy, onClick: onDemo },
          busy ? "Carregando..." : "Usar dados de exemplo",
        ),
      ),
    ),
  );
}

function ReviewScreen({ proposal, downloaded, busy, onNewUpload, onReview, onDownload }) {
  const [draft, setDraft] = useState(clonePurchase(proposal.purchaseDraft));
  const blocking = useMemo(() => (proposal.validationResults || []).filter((item) => item.blocking), [proposal]);
  const warnings = useMemo(() => (proposal.validationResults || []).filter((item) => !item.blocking), [proposal]);

  useEffect(() => {
    setDraft(clonePurchase(proposal.purchaseDraft));
  }, [proposal.id, proposal.purchaseDraft]);

  return h(
    "main",
    { className: "screen review-screen" },
    h(
      "section",
      { className: "panel review-shell" },
      h(
        "div",
        { className: "review-header" },
        h("div", null, h("div", { className: "screen-kicker" }, "Tela 2"), h("h2", null, "Validacao da importacao"), h("p", { className: "muted" }, "Confira a tabela processada. Campos obrigatorios pendentes aparecem em vermelho.")),
        h(
          "div",
          { className: "review-status" },
          h("span", { className: `status ${blocking.length === 0 ? "ok" : "review"}` }, blocking.length === 0 ? "Pronto para baixar" : countLabel(blocking.length, "ponto para conferir", "pontos para conferir")),
          warnings.length > 0 ? h("span", { className: "status warning" }, countLabel(warnings.length, "aviso", "avisos")) : null,
          h("button", { className: "button ghost", disabled: busy, onClick: onNewUpload }, "Novo upload"),
        ),
      ),
      h(SourcePanel, { proposal }),
      h(HeaderGrid, { draft, setDraft, results: proposal.validationResults || [] }),
      h(ItemsGrid, { draft, setDraft, results: proposal.validationResults || [] }),
      h(ValidationPanel, { results: proposal.validationResults || [] }),
      h(
        "div",
        { className: "review-actions" },
        h("button", { className: "button secondary", disabled: busy, onClick: () => onReview(buildEdits(draft)) }, "Salvar alteracoes"),
        h("button", { className: "button primary", disabled: busy, onClick: () => onDownload(buildEdits(draft)) }, downloaded ? "Baixar Excel novamente" : "Baixar Excel"),
      ),
    ),
    downloaded ? h(DownloadPanel) : null,
  );
}

function SourcePanel({ proposal }) {
  const observations = proposal.sourceDocument?.observations || [];
  const isOpenAI = observations.some((item) => String(item).toLowerCase().includes("openai"));
  const isFallback = observations.some((item) => String(item).toLowerCase().includes("fallback"));
  const label = isOpenAI ? "Processado por OpenAI Vision" : isFallback ? "Fallback demonstrativo" : "Origem estruturada";
  const tone = isOpenAI ? "ai" : isFallback ? "demo" : "structured";

  return h(
    "section",
    { className: `source-panel ${tone}` },
    h("div", null, h("strong", null, label), h("span", null, proposal.sourceDocument?.fileName || "documento enviado")),
    h(
      "small",
      null,
      isOpenAI
        ? "Dados extraidos da imagem por IA. Revise com atencao: a IA pode errar ou inventar linhas quando a foto estiver ruim."
        : isFallback
          ? "Dados ficticios carregados para demonstracao do fluxo."
          : "Dados carregados de JSON/XLSX/sidecar estruturado.",
    ),
  );
}

function HeaderGrid({ draft, setDraft, results }) {
  return h(
    "div",
    { className: "sheet header-sheet" },
    h("div", { className: "sheet-title" }, "Cabecalho da compra"),
    inputCell("Fornecedor", draft.supplier?.legalName || "", (value) => setDraft(update(draft, ["supplier", "legalName"], value)), invalidSupplier(draft, results)),
    inputCell("CNPJ/CPF", draft.supplier?.documentNumber || "", (value) => setDraft(update(draft, ["supplier", "documentNumber"], value)), invalidSupplier(draft, results)),
    inputCell("Documento", draft.documentNumber || "", (value) => setDraft(update(draft, ["documentNumber"], value)), isInvalid(results, "purchase.documentNumber")),
    inputCell("Tabela preco", draft.priceTable || "", (value) => setDraft(update(draft, ["priceTable"], value)), false),
    inputCell("Frete", draft.freightMode || "", (value) => setDraft(update(draft, ["freightMode"], value)), false),
    moneyCell("Total produtos", draft.totals?.productsTotal, (value) => setDraft(update(draft, ["totals", "productsTotal"], value)), isInvalid(results, "purchase.totals.productsTotal")),
    moneyCell("Total geral", draft.totals?.grandTotal, (value) => setDraft(update(draft, ["totals", "grandTotal"], value)), isInvalid(results, "purchase.totals.grandTotal")),
  );
}

function ItemsGrid({ draft, setDraft, results }) {
  const items = draft.items || [];
  return h(
    "div",
    { className: "table-wrap sheet-table" },
    h(
      "table",
      null,
      h(
        "thead",
        null,
        h("tr", null, ["Linha", "Codigo", "Referencia", "Descricao", "Unid.", "Qtd.", "Vlr. Unit.", "Vlr. Total"].map((header) => h("th", { key: header }, header))),
      ),
      h(
        "tbody",
        null,
        items.map((item, index) =>
          h(
            "tr",
            { key: item.lineNumber || index },
            h("td", null, item.lineNumber || index + 1),
            h("td", null, cellInput(item.supplierProductCode || item.barcode || "", (value) => setDraft(updateItem(draft, index, "supplierProductCode", value)), isInvalid(results, `purchase.items[${index}].supplierProductCode`))),
            h("td", null, cellInput(item.reference || "", (value) => setDraft(updateItem(draft, index, "reference", value)), false)),
            h("td", null, cellInput(item.description || "", (value) => setDraft(updateItem(draft, index, "description", value)), isInvalid(results, `purchase.items[${index}].description`))),
            h("td", null, cellInput(item.unit || "", (value) => setDraft(updateItem(draft, index, "unit", value)), false)),
            h("td", null, numberInput(item.quantity, (value) => setDraft(updateItem(draft, index, "quantity", value)), isInvalid(results, `purchase.items[${index}].quantity`))),
            h("td", null, moneyInput(item.unitCost, (value) => setDraft(updateItem(draft, index, "unitCost", value)), isInvalid(results, `purchase.items[${index}].unitCost`))),
            h("td", null, moneyInput(item.totalCost, (value) => setDraft(updateItem(draft, index, "totalCost", value)), isInvalid(results, `purchase.items[${index}].totalCost`))),
          ),
        ),
      ),
    ),
  );
}

function ValidationPanel({ results }) {
  const issues = results.map(readableValidation).sort((a, b) => Number(b.blocking) - Number(a.blocking));
  return h(
    "section",
    { className: "validation-panel" },
    h("h3", null, "Pontos para conferir"),
    issues.length === 0
      ? h("p", { className: "success" }, "Tudo certo para baixar o Excel da compra.")
      : issues.map((item) =>
          h(
            "div",
            { key: `${item.code}-${item.field}`, className: `issue ${item.blocking ? "blocking" : "warning"}` },
            h("div", { className: "issue-header" }, h("strong", null, item.title), h("span", null, item.blocking ? "Precisa resolver" : "Apenas conferir")),
            h("span", null, item.description),
            item.detail ? h("code", null, item.detail) : null,
          ),
        ),
  );
}

function readableValidation(result) {
  const field = readableField(result.field);
  const messages = {
    DOCUMENT_INCOMPLETE: {
      title: "O documento parece ter mais páginas",
      description: "A imagem indica que existe outra página. Confira se os itens desta compra estão todos na tabela antes de baixar o Excel.",
    },
    HUMAN_REVIEW_REQUIRED: {
      title: "Conferência humana necessária",
      description: "Revise os dados extraídos da imagem. O Excel pode ser baixado depois da conferência.",
    },
    ITEM_TOTAL_MISMATCH: {
      title: "Total de item para conferir",
      description: `${field || "Um item"} não bate exatamente com quantidade x valor unitário. Ajuste se estiver errado ou siga se o valor da nota estiver correto.`,
    },
    PRODUCTS_TOTAL_MISMATCH: {
      title: "Total de produtos para conferir",
      description: "A soma dos itens extraídos não bate com o total de produtos informado no documento. Isso pode acontecer quando a foto não permite ler todas as linhas.",
    },
    GRAND_TOTAL_MISMATCH: {
      title: "Total geral para conferir",
      description: "O total geral calculado ficou diferente dos totais informados. Confira desconto, acréscimo, impostos e total da compra.",
    },
    SUPPLIER_NOT_IDENTIFIED: {
      title: "Fornecedor não identificado",
      description: "Informe o nome ou CNPJ/CPF do fornecedor antes de baixar o Excel.",
    },
    PURCHASE_WITHOUT_ITEMS: {
      title: "Nenhum item encontrado",
      description: "A importação não encontrou itens de compra. Envie outro arquivo ou preencha os itens antes de baixar o Excel.",
    },
    INVALID_QUANTITY: {
      title: "Quantidade inválida",
      description: `${field || "Um item"} está com quantidade vazia ou menor que zero.`,
    },
    INVALID_UNIT_COST: {
      title: "Valor unitário inválido",
      description: `${field || "Um item"} está com valor unitário negativo.`,
    },
  };
  const fallback = {
    title: "Ponto para conferir",
    description: result.message || "Confira este dado antes de baixar o Excel.",
  };
  const message = messages[result.code] || fallback;
  return {
    ...result,
    title: message.title,
    description: message.description,
    detail: field,
  };
}

function readableField(path) {
  const normalized = normalizePath(path);
  const itemMatch = normalized.match(/^items\[(\d+)\]\.(.+)$/);
  if (itemMatch) {
    return `Linha ${Number(itemMatch[1]) + 1}, ${readableFieldName(itemMatch[2])}`;
  }
  if (normalized === "sourceDocument.pageCount") return "Páginas do documento";
  return readableFieldName(normalized);
}

function readableFieldName(path) {
  const names = {
    supplier: "Fornecedor",
    "supplier.legalName": "Nome do fornecedor",
    "supplier.documentNumber": "CNPJ/CPF do fornecedor",
    documentNumber: "Número do documento",
    "totals.productsTotal": "Total de produtos",
    "totals.grandTotal": "Total geral",
    quantity: "Quantidade",
    unitCost: "Valor unitário",
    totalCost: "Valor total",
  };
  return names[path] || "";
}

function DownloadPanel() {
  return h(
    "section",
    { className: "panel approval-panel" },
    h("h2", null, "Excel gerado"),
    h("p", null, "A planilha foi baixada com os dados processados para importacao manual no NEX."),
  );
}

function downloadBlob(blob, fileName) {
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = fileName;
  document.body.appendChild(link);
  link.click();
  link.remove();
  URL.revokeObjectURL(url);
}

function inputCell(label, value, onChange, invalid) {
  return h("label", { className: "sheet-cell" }, h("span", null, label), cellInput(value, onChange, invalid));
}

function moneyCell(label, value, onChange, invalid) {
  return h("label", { className: "sheet-cell" }, h("span", null, label), moneyInput(value, onChange, invalid));
}

function cellInput(value, onChange, invalid) {
  return h("input", { className: `cell-input ${invalid ? "invalid" : ""}`, value, onChange: (event) => onChange(event.target.value) });
}

function numberInput(value, onChange, invalid) {
  return h("input", { className: `cell-input number ${invalid ? "invalid" : ""}`, type: "number", step: "0.01", value: value ?? 0, onChange: (event) => onChange(Number(event.target.value)) });
}

function moneyInput(value, onChange, invalid) {
  return h("input", {
    className: `cell-input money ${invalid ? "invalid" : ""}`,
    type: "number",
    step: "0.01",
    value: moneyNumber(value),
    onChange: (event) => onChange({ cents: Math.round(Number(event.target.value || 0) * 100) }),
  });
}

function buildEdits(draft) {
  const edits = [
    edit("purchase.supplier.legalName", draft.supplier?.legalName || ""),
    edit("purchase.supplier.documentNumber", draft.supplier?.documentNumber || ""),
    edit("purchase.documentNumber", draft.documentNumber || ""),
    edit("purchase.priceTable", draft.priceTable || ""),
    edit("purchase.freightMode", draft.freightMode || ""),
    edit("purchase.totals.productsTotal", draft.totals?.productsTotal || { cents: 0 }),
    edit("purchase.totals.grandTotal", draft.totals?.grandTotal || { cents: 0 }),
  ];
  (draft.items || []).forEach((item, index) => {
    edits.push(
      edit(`purchase.items[${index}].supplierProductCode`, item.supplierProductCode || ""),
      edit(`purchase.items[${index}].reference`, item.reference || ""),
      edit(`purchase.items[${index}].description`, item.description || ""),
      edit(`purchase.items[${index}].unit`, item.unit || ""),
      edit(`purchase.items[${index}].quantity`, item.quantity || 0),
      edit(`purchase.items[${index}].unitCost`, item.unitCost || { cents: 0 }),
      edit(`purchase.items[${index}].totalCost`, item.totalCost || { cents: 0 }),
    );
  });
  return edits;
}

function edit(fieldPath, value) {
  return { fieldPath, value, editedBy: "operadora" };
}

function clonePurchase(purchase) {
  return JSON.parse(JSON.stringify(purchase || { supplier: {}, totals: {}, items: [] }));
}

function update(source, path, value) {
  const next = clonePurchase(source);
  let target = next;
  for (let index = 0; index < path.length - 1; index += 1) {
    const key = path[index];
    target[key] = target[key] || {};
    target = target[key];
  }
  target[path[path.length - 1]] = value;
  return next;
}

function updateItem(draft, index, field, value) {
  const next = clonePurchase(draft);
  next.items[index] = { ...(next.items[index] || {}), [field]: value };
  return next;
}

function invalidSupplier(draft, results) {
  const supplier = draft.supplier || {};
  return (!supplier.legalName && !supplier.documentNumber) || results.some((item) => normalizePath(item.field) === "supplier");
}

function isInvalid(results, path) {
  const expected = normalizePath(path);
  return results.some((item) => item.blocking && normalizePath(item.field) === expected);
}

function normalizePath(path) {
  return String(path || "").replace(/^purchase\./, "").replace(/^purchaseDraft\./, "");
}

function moneyNumber(value) {
  if (!value || typeof value.cents !== "number") return "0.00";
  return (value.cents / 100).toFixed(2);
}

ReactDOM.createRoot(document.getElementById("root")).render(h(App));
