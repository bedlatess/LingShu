const docs = {
  eyebrow: "Integration Docs",
  title: "Point your SDK to the LingShu gateway",
  description: "Copy the base_url and a platform API key to use OpenAI-compatible chat, Anthropic Messages, or Embeddings endpoints. Keys in examples are placeholders.",
  copyBaseURL: "Copy base_url",
  copied: "Copied",
  baseUrl: {
    title: "Gateway endpoint",
    label: "Base URL",
    keyHint: "Replace the placeholder key below with a real key created on the API Keys page."
  },
  quickConfig: {
    title: "Claude Code / Codex quick config",
    codex: "Codex / OpenAI compatible",
    claude: "Claude / Anthropic compatible"
  },
  examples: {
    title: "Copy-ready examples",
    openai: "OpenAI Chat",
    anthropic: "Anthropic Messages",
    embeddings: "Embeddings"
  },
  matrix: {
    title: "Capability matrix",
    model: "Model",
    type: "Type",
    billing: "Billing",
    stream: "Streaming",
    tools: "Tool calls",
    vision: "Vision",
    endpoints: "Endpoints",
    emptyTitle: "No models available",
    emptyDescription: "Published models will appear in the capability matrix."
  }
};

export default docs;
