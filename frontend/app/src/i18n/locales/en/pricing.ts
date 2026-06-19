const pricing = {
  eyebrow: "Pricing",
  title: "Pay for actual usage",
  description: "The prices below are customer-facing prices. Final charges are settled against the gateway usage response.",
  enterConsole: "Enter Console",
  perCall: "Per call",
  input: "Input",
  output: "Output",
  emptyTitle: "No public models yet",
  emptyDescription: "Prices will appear here once admins publish models.",
  filters: {
    search: "Search model, type, or group",
    allBilling: "All billing modes",
    token: "Token billing",
    perCall: "Per-call billing",
    allGroups: "All groups",
    defaultGroup: "Default group"
  },
  detail: {
    priceTitle: "Price details",
    capabilityTitle: "Capabilities",
    modelId: "Model ID",
    context: "Context length",
    contextUnavailable: "Depends on the model response",
    endpoints: "Supported endpoints"
  },
  landing: {
    badge: "AI API aggregation gateway",
    heroTitle: "Bring multiple upstream models into one clean API entrance",
    heroDescription: "LingShu is built for private operations: OpenAI-compatible access, platform API keys, balance reservation, real usage settlement, and channel failover. Less sprawl, more control.",
    primaryCta: "Start using",
    secondaryCta: "View pricing",
    previewEyebrow: "Gateway Flow",
    previewTitle: "A clear path for every request",
    stats: {
      openaiCompatible: "OpenAI-compatible endpoints, so existing SDKs can switch by changing base_url.",
      realtimeBilling: "Settlement follows upstream usage, with balance and ledger recorded together.",
      operatorManaged: "Private operation with only administrator and regular user roles."
    },
    flow: {
      client: "Client sends a request",
      clientHint: "Use a platform API key to call /v1/chat/completions and related endpoints.",
      gateway: "LingShu schedules and settles",
      gatewayHint: "Reserve balance, choose a healthy channel, forward the request, and write logs.",
      upstream: "Upstream returns real usage",
      upstreamHint: "The gateway settles by final usage; users only see final charges."
    },
    features: {
      multiModel: {
        title: "Multi-model aggregation",
        description: "Manage OpenAI-compatible, Anthropic, and other upstream channels, then publish available models to users."
      },
      balance: {
        title: "Balance reservation",
        description: "Check balance before calls and settle by real usage afterwards to reduce debt and reconciliation risk."
      },
      failover: {
        title: "Channel failover",
        description: "Health checks, cooldown windows, and retry exclusions keep troubled channels out of the hot path."
      },
      privateOps: {
        title: "Private-ops friendly",
        description: "Operate with redeem codes and admin grants, without payment gateways, distribution, or complex permissions."
      }
    },
    pricingEyebrow: "Public pricing",
    pricingTitle: "Customer-facing final prices only",
    pricingDescription: "The user-side pricing page shows final unit prices only. Upstream cost, multipliers, and gross profit stay in admin reports.",
    viewAllPrices: "View all prices",
    ctaTitle: "Ready to consolidate your API access?",
    ctaDescription: "After signing in, users can create platform API keys, review usage, and redeem balance. Admins can maintain channels, models, reports, and announcements."
  }
};

export default pricing;
