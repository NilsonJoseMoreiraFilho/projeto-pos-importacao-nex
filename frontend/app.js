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
  async approve(id) {
    return readResponse(
      await fetch(`/api/imports/${id}/approve`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ reviewer: "operadora" }),
      }),
    );
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
  const [approved, setApproved] = useState(null);
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
    setApproved(null);
    setError("");
  }

  return h(
    "div",
    { className: "app" },
    h(
      "header",
      { className: "topbar" },
      h("div", null, h("h1", null, "Importacao de compras NEX"), h("p", null, onUploadScreen ? "Tela 1 - upload do documento" : "Tela 2 - validacao, edicao e aprovacao")),
      h(Stepper, { active: onUploadScreen ? "upload" : "review" }),
    ),
    error ? h("div", { className: "global-error" }, error) : null,
    onUploadScreen
      ? h(UploadScreen, {
          busy,
          onUpload: (input) =>
            run(async () => {
              setApproved(null);
              setProposal(await api.createImport(input));
            }),
        })
      : h(ReviewScreen, {
          proposal,
          approved,
          busy,
          onNewUpload: resetFlow,
          onReview: (edits) =>
            run(async () => {
              setProposal(await api.review(proposal.id, edits));
              setApproved(null);
            }),
          onApprove: () =>
            run(async () => {
              setApproved(await api.approve(proposal.id));
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

function UploadScreen({ busy, onUpload }) {
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
        h("span", null, "JSON controlado, XLSX, PDF ou imagem com sidecar OCR"),
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
        "button",
        { className: "button primary upload-action", disabled: busy || !file, onClick: () => onUpload({ file, reader, requiresHumanReview }) },
        busy ? "Processando..." : "Processar e abrir validacao",
      ),
    ),
  );
}

function ReviewScreen({ proposal, approved, busy, onNewUpload, onReview, onApprove }) {
  const [draft, setDraft] = useState(clonePurchase(proposal.purchaseDraft));
  const blocking = useMemo(() => (proposal.validationResults || []).filter((item) => item.blocking), [proposal]);
  const canApprove = blocking.length === 0 && proposal.status === "PROPOSED" && !approved;

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
          h("span", { className: `status ${canApprove ? "ok" : "review"}` }, canApprove ? "Pronto para aprovar" : `${blocking.length} pendencia(s)`),
          h("button", { className: "button ghost", disabled: busy, onClick: onNewUpload }, "Novo upload"),
        ),
      ),
      h(HeaderGrid, { draft, setDraft, results: proposal.validationResults || [] }),
      h(ItemsGrid, { draft, setDraft, results: proposal.validationResults || [] }),
      h(ValidationPanel, { results: proposal.validationResults || [] }),
      h(
        "div",
        { className: "review-actions" },
        h("button", { className: "button secondary", disabled: busy, onClick: () => onReview(buildEdits(draft)) }, "Salvar alteracoes"),
        h("button", { className: "button primary", disabled: busy || !canApprove, onClick: onApprove }, "Aprovar e integrar NEX"),
      ),
    ),
    approved ? h(ApprovalPanel, { approved }) : null,
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
        h("tr", null, ["Linha", "Codigo", "Referencia", "Descricao", "Unid.", "Qtd.", "Vlr. Unit.", "Vlr. Total", "Produto NEX"].map((header) => h("th", { key: header }, header))),
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
            h("td", null, cellInput(item.matchedInternalProductId || "", (value) => setDraft(updateItem(draft, index, "matchedInternalProductId", value)), isInvalid(results, `purchase.items[${index}].matchedInternalProductId`))),
          ),
        ),
      ),
    ),
  );
}

function ValidationPanel({ results }) {
  const blocking = results.filter((item) => item.blocking);
  return h(
    "section",
    { className: "validation-panel" },
    h("h3", null, "Pendencias de validacao"),
    blocking.length === 0
      ? h("p", { className: "success" }, "Sem campos obrigatorios pendentes. A compra pode ser aprovada.")
      : blocking.map((item) =>
          h(
            "div",
            { key: `${item.code}-${item.field}`, className: "issue" },
            h("strong", null, item.code),
            h("span", null, item.message),
            h("code", null, item.field),
          ),
        ),
  );
}

function ApprovalPanel({ approved }) {
  const purchase = approved.approvedPurchase || approved;
  const exportResult = approved.exportResult;
  return h(
    "section",
    { className: "panel approval-panel" },
    h("h2", null, "Compra aprovada"),
    h("p", null, `Aprovada por ${purchase.approvedBy || "operadora"}.`),
    exportResult ? h("p", null, `Saida gerada pelo adapter: ${exportResult.destination} - ${exportResult.reference}`) : null,
  );
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
      edit(`purchase.items[${index}].matchedInternalProductId`, item.matchedInternalProductId || ""),
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
