const dashboard = {
  eyebrow: "概览",
  title: "账户运行情况",
  description: "查看余额、消费趋势和模型可用性。这里仅展示与你账户相关的数据。",
  stats: {
    balance: "余额",
    balanceHint: "可用于后续调用",
    todayCharge: "今日消费",
    todayChargeHint: "今日 {{count}} 次请求",
    monthCharge: "本月消费",
    monthChargeHint: "按自然月统计",
    models: "可用模型",
    modelsHint: "管理员已开放的模型"
  },
  trendTitle: "近 7 日消费趋势",
  trendEmptyTitle: "还没有消费趋势",
  trendEmptyDescription: "完成一次 API 调用后，日统计会出现在这里。",
  quota: {
    title: "额度使用进度",
    granted: "累计入账",
    used: "累计消费",
    remaining: "当前余额",
    description: "该进度使用账户累计入账与累计消费估算，便于快速判断额度消耗情况。"
  },
  developer: {
    title: "开发者接入",
    description: "在本地工具中使用灵枢网关，需先创建一个 API Key。",
    manageKeys: "管理 API Key",
    toggle: "查看接入配置示例"
  },
  quickConfig: {
    title: "快速接入",
    claude: "Claude Code",
    codex: "Codex / OpenAI",
    copy: "复制",
    copied: "配置已复制",
    description: "将示例中的占位 Key 替换为你创建的 API Key，即可在本地工具中使用灵枢网关。"
  }
};

export default dashboard;
