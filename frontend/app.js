const { useMemo, useState } = React;

const api = {
  async createImport({ file, reader, requiresHumanReview }) {
    const formData = new FormData();
    formData.append("file", file);
    formData.append("reader", reader);
    formData.append("requiresHumanReview", String(requiresHumanReview));
    const response = await fetch("/api/imports", { method: "POST", body: formData });
    return readResponse(response);
  },
  async review(id, decisions) {
    const response = await fetch(`/api/imports/${id}/review`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ decisions }),
    });
    return readResponse(response);
  },
  async approve(id) {
    const response = await fetch(`/api/imports/${id}/approve`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ reviewer: "operadora" }),
    });
    return readResponse(response);
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

  const blocking = useMemo(
    () => (proposal?.validationResults || []).filter((item) => item.blocking),
    [proposal],
  );
  const canApprove = proposal && blocking.length === 0 && proposal.status === "PROPOSED";

  return React.createElement(
    "div",
    { className: "app" },
    React.createElement(
      "header",
      { className: "topbar" },
      React.createElement("h1", null, "Importacao de compras"),
      React.createElement("p", null, "Sprint 2 - upload, revisao humana e confirmacao"),
    ),
    React.createElement(
      "div",
      { className: "layout" },
      React.createElement(UploadPanel, {
        busy,
        onUpload: (input) =>
          run(async () => {
            setApproved(null);
            const next = await api.createImport(input);
            setProposal(next);
          }),
      }),
      React.createElement(
        "section",
        { className: "grid" },
        error ? React.createElement("div", { className: "error" }, error) : null,
        proposal
          ? React.createElement(ProposalReview, {
              proposal,
              approved,
              busy,
              canApprove,
              onReview: (decisions) =>
                run(async () => {
                  const next = await api.review(proposal.id, decisions);
                  setProposal(next);
                  setApproved(null);
                }),
              onApprove: () =>
                run(async () => {
                  const next = await api.approve(proposal.id);
                  setApproved(next);
                }),
            })
          : React.createElement("div", { className: "panel empty" }, "Envie um documento para iniciar a revisao."),
      ),
    ),
  );
}

function UploadPanel({ busy, onUpload }) {
  const [file, setFile] = useState(null);
  const [reader, setReader] = useState("auto");
  const [requiresHumanReview, setRequiresHumanReview] = useState(true);

  return React.createElement(
    "aside",
    { className: "panel" },
    React.createElement("h2", null, "Upload do documento"),
    React.createElement(
      "div",
      { className: "field" },
      React.createElement("label", null, "Arquivo"),
      React.createElement("input", {
        type: "file",
        onChange: (event) => setFile(event.target.files[0] || null),
      }),
    ),
    React.createElement(
      "div",
      { className: "field" },
      React.createElement("label", null, "Reader"),
      React.createElement(
        "select",
        { value: reader, onChange: (event) => setReader(event.target.value) },
        ["auto", "manualjson", "xlsx", "pdf", "imageocr"].map((item) =>
          React.createElement("option", { key: item, value: item }, item),
        ),
      ),
    ),
    React.createElement(
      "label",
      { className: "field" },
      React.createElement("span", null, "Revisao humana obrigatoria"),
      React.createElement("input", {
        type: "checkbox",
        checked: requiresHumanReview,
        onChange: (event) => setRequiresHumanReview(event.target.checked),
      }),
    ),
    React.createElement(
      "button",
      {
        className: "button primary",
        disabled: busy || !file,
        onClick: () => onUpload({ file, reader, requiresHumanReview }),
      },
      busy ? "Processando..." : "Enviar para revisao",
    ),
  );
}

function ProposalReview({ proposal, approved, busy, canApprove, onReview, onApprove }) {
  const [draftDecisions, setDraftDecisions] = useState({});

  function setDecision(fieldPath, patch) {
    setDraftDecisions((current) => ({
      ...current,
      [fieldPath]: {
        fieldPath,
        decision: "ACCEPT",
        reviewedBy: "operadora",
        ...(current[fieldPath] || {}),
        ...patch,
      },
    }));
  }

  const decisions = Object.values(draftDecisions);

  return React.createElement(
    React.Fragment,
    null,
    React.createElement(
      "section",
      { className: "panel" },
      React.createElement(
        "div",
        { className: "actions" },
        React.createElement("h2", { style: { marginRight: "auto" } }, "Proposta de importacao"),
        React.createElement("span", { className: `status ${proposal.status === "PROPOSED" ? "ok" : "review"}` }, proposal.status),
      ),
      React.createElement(Summary, { proposal }),
      approved
        ? React.createElement("div", { className: "issue warning" }, `Compra aprovada por ${approved.approvedBy}.`)
        : null,
    ),
    React.createElement(ValidationPanel, { results: proposal.validationResults || [] }),
    React.createElement(FieldsPanel, { fields: proposal.extractedFields || [], draftDecisions, setDecision }),
    React.createElement(ItemsPanel, { purchase: proposal.purchaseDraft }),
    React.createElement(
      "section",
      { className: "panel actions" },
      React.createElement(
        "button",
        {
          className: "button secondary",
          disabled: busy || decisions.length === 0,
          onClick: () => onReview(decisions),
        },
        "Salvar revisao",
      ),
      React.createElement(
        "button",
        {
          className: "button primary",
          disabled: busy || !canApprove,
          onClick: onApprove,
        },
        "Confirmar importacao",
      ),
    ),
  );
}

function Summary({ proposal }) {
  const purchase = proposal.purchaseDraft || {};
  const supplier = purchase.supplier || {};
  const totals = purchase.totals || {};
  return React.createElement(
    "div",
    { className: "summary" },
    metric("Fornecedor", supplier.legalName || "-"),
    metric("Documento", purchase.documentNumber || "-"),
    metric("Itens", String((purchase.items || []).length)),
    metric("Total", money(totals.grandTotal)),
  );
}

function metric(label, value) {
  return React.createElement("div", { className: "metric" }, React.createElement("strong", null, label), React.createElement("span", null, value));
}

function ValidationPanel({ results }) {
  return React.createElement(
    "section",
    { className: "panel" },
    React.createElement("h2", null, "Validacoes"),
    results.length === 0
      ? React.createElement("p", { className: "empty" }, "Sem validacoes bloqueantes.")
      : results.map((item) =>
          React.createElement(
            "div",
            { key: `${item.code}-${item.field}`, className: `issue ${item.severity === "WARNING" ? "warning" : ""}` },
            React.createElement("strong", null, item.code),
            React.createElement("div", null, item.message),
            React.createElement("small", null, item.field),
          ),
        ),
  );
}

function FieldsPanel({ fields, draftDecisions, setDecision }) {
  return React.createElement(
    "section",
    { className: "panel grid" },
    React.createElement("h2", null, "Campos extraidos"),
    fields.length === 0
      ? React.createElement("p", { className: "empty" }, "Nenhum campo rastreavel informado.")
      : fields.map((field) =>
          React.createElement(
            "article",
            { className: "field-card", key: field.fieldPath },
            React.createElement(
              "header",
              null,
              React.createElement("code", null, field.fieldPath),
              React.createElement("span", null, `${Math.round((field.confidence || 0) * 100)}%`),
            ),
            React.createElement("p", null, field.rawText || "-"),
            React.createElement("small", null, `Status: ${field.status}`),
            React.createElement(
              "div",
              { className: "review-row" },
              React.createElement("input", {
                placeholder: "valor corrigido",
                value: draftDecisions[field.fieldPath]?.correctedValue || "",
                onChange: (event) =>
                  setDecision(field.fieldPath, {
                    decision: "CORRECT",
                    correctedValue: event.target.value,
                  }),
              }),
              React.createElement(
                "div",
                { className: "actions" },
                React.createElement("button", { className: "button secondary", onClick: () => setDecision(field.fieldPath, { decision: "ACCEPT" }) }, "Aceitar"),
                React.createElement("button", { className: "button danger", onClick: () => setDecision(field.fieldPath, { decision: "REJECT" }) }, "Rejeitar"),
              ),
            ),
          ),
        ),
  );
}

function ItemsPanel({ purchase }) {
  const items = purchase?.items || [];
  return React.createElement(
    "section",
    { className: "panel" },
    React.createElement("h2", null, "Itens da compra"),
    React.createElement(
      "div",
      { className: "table-wrap" },
      React.createElement(
        "table",
        null,
        React.createElement(
          "thead",
          null,
          React.createElement(
            "tr",
            null,
            ["Linha", "Codigo", "Referencia", "Descricao", "Qtd", "Unitario", "Total", "Produto interno"].map((header) =>
              React.createElement("th", { key: header }, header),
            ),
          ),
        ),
        React.createElement(
          "tbody",
          null,
          items.map((item) =>
            React.createElement(
              "tr",
              { key: item.lineNumber },
              React.createElement("td", null, item.lineNumber),
              React.createElement("td", null, item.barcode || item.supplierProductCode || "-"),
              React.createElement("td", null, item.reference || "-"),
              React.createElement("td", null, item.description || "-"),
              React.createElement("td", null, item.quantity),
              React.createElement("td", null, money(item.unitCost)),
              React.createElement("td", null, money(item.totalCost)),
              React.createElement("td", null, item.matchedInternalProductId || "pendente"),
            ),
          ),
        ),
      ),
    ),
  );
}

function money(value) {
  if (!value || typeof value.cents !== "number") return "R$ 0,00";
  return (value.cents / 100).toLocaleString("pt-BR", { style: "currency", currency: "BRL" });
}

ReactDOM.createRoot(document.getElementById("root")).render(React.createElement(App));
