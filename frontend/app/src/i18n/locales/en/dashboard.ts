const dashboard = {
  eyebrow: "Overview",
  title: "Account activity",
  description: "Check your balance, spending trend, and model availability. Only data related to your account appears here.",
  stats: {
    balance: "Balance",
    balanceHint: "Available for future requests",
    todayCharge: "Today's Spend",
    todayChargeHint: "{{count}} requests today",
    monthCharge: "This Month",
    monthChargeHint: "Natural month total",
    models: "Available Models",
    modelsHint: "Models enabled by admin"
  },
  trendTitle: "Last 7 days",
  trendEmptyTitle: "No trend yet",
  trendEmptyDescription: "After you complete an API call, daily stats will show up here.",
  quota: {
    title: "Quota progress",
    granted: "Total granted",
    used: "Total spent",
    remaining: "Current balance",
    description: "This progress uses total granted balance and total spend to help you understand quota consumption."
  },
  developer: {
    title: "Developer access",
    description: "To use LingShu as a gateway in local tools, create an API key first.",
    manageKeys: "Manage API Keys",
    toggle: "View setup examples"
  },
  quickConfig: {
    title: "Quick setup",
    claude: "Claude Code",
    codex: "Codex / OpenAI",
    copy: "Copy",
    copied: "Configuration copied",
    description: "Replace the placeholder key with an API key you created, then use LingShu as the gateway for local tools."
  }
};

export default dashboard;
