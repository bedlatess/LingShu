const docs = {
  eyebrow: "接入指引",
  title: "把 SDK 指向灵枢网关",
  description: "复制 base_url 与平台 API Key 后，即可用 OpenAI 兼容接口、Anthropic Messages 或 Embeddings 端点接入。示例中的 Key 是占位符。",
  copyBaseURL: "复制 base_url",
  copied: "已复制",
  baseUrl: {
    title: "网关地址",
    label: "Base URL",
    keyHint: "把下面的占位 Key 替换为你在 API 密钥页创建的真实密钥。"
  },
  quickConfig: {
    title: "Claude Code / Codex 快速配置",
    codex: "Codex / OpenAI 兼容",
    claude: "Claude / Anthropic 兼容"
  },
  examples: {
    title: "可复制示例",
    openai: "OpenAI 对话",
    anthropic: "Anthropic Messages",
    embeddings: "Embeddings 向量"
  },
  matrix: {
    title: "能力矩阵",
    model: "模型",
    type: "类型",
    billing: "计费方式",
    stream: "流式",
    tools: "工具调用",
    vision: "视觉",
    endpoints: "可用端点",
    emptyTitle: "暂无可用模型",
    emptyDescription: "管理员发布模型后，能力矩阵会在这里展示。"
  }
};

export default docs;
