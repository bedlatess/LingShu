const usage = {
  eyebrow: "Usage",
  title: "Usage and spending details",
  description: "Inspect your account from four views: daily stats, models, ledger, and request logs.",
  tabs: {
    daily: "Daily",
    models: "By Model",
    ledger: "Charges",
    logs: "Requests"
  },
  filters: {
    search: "Search request, model, or remark",
    allStatus: "All statuses",
    success: "Success",
    failed: "Failed",
    inFlight: "In-flight requests {{count}}",
    from: "Start date",
    to: "End date",
    timezone: "Current timezone {{timezone}}"
  },
  inFlight: {
    title: "In-flight requests {{count}}",
    emptyTitle: "No in-flight requests",
    emptyDescription: "New requests will briefly appear here so you can watch active usage.",
    pollingHint: "This section refreshes every 5 seconds."
  },
  dailyTitle: "Daily spend",
  dailyEmptyTitle: "No daily stats yet",
  dailyEmptyDescription: "After requests happen, counts and spend will be grouped by day here.",
  modelsTitle: "Spend by model",
  modelsEmptyTitle: "No model stats yet",
  modelsEmptyDescription: "Model spending distribution will appear here.",
  ledgerTable: {
    type: "Type",
    amount: "Amount",
    balanceAfter: "Balance After",
    remark: "Remark",
    createdAt: "Time"
  },
  logsTable: {
    requestId: "Request",
    model: "Model",
    status: "Status",
    tokens: "Tokens",
    charge: "Charge",
    createdAt: "Time"
  },
  detail: {
    title: "Request detail"
  }
};

export default usage;
