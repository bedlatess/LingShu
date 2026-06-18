package upstream

// ChannelPreset describes a known upstream provider preset.
// Format is the wire protocol used by the gateway: openai or anthropic.
type ChannelPreset struct {
	Key     string `json:"key"`
	Label   string `json:"label"`
	BaseURL string `json:"base_url"`
	Format  string `json:"format"`
	Note    string `json:"note"`
}

func ChannelPresets() []ChannelPreset {
	return []ChannelPreset{
		{Key: "openai", Label: "OpenAI 官方", BaseURL: "https://api.openai.com/v1", Format: "openai", Note: "GPT 系列"},
		{Key: "deepseek", Label: "DeepSeek 深度求索", BaseURL: "https://api.deepseek.com/v1", Format: "openai", Note: "deepseek-chat / deepseek-reasoner"},
		{Key: "zhipu", Label: "智谱 GLM", BaseURL: "https://open.bigmodel.cn/api/paas/v4", Format: "openai", Note: "GLM 系列，注意是 /v4"},
		{Key: "moonshot", Label: "月之暗面 Kimi", BaseURL: "https://api.moonshot.cn/v1", Format: "openai", Note: "moonshot-v1 系列"},
		{Key: "qwen", Label: "阿里通义千问", BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1", Format: "openai", Note: "qwen 系列兼容模式"},
		{Key: "siliconflow", Label: "硅基流动 SiliconFlow", BaseURL: "https://api.siliconflow.cn/v1", Format: "openai", Note: "多模型聚合"},
		{Key: "xai", Label: "xAI Grok", BaseURL: "https://api.x.ai/v1", Format: "openai", Note: "grok 系列"},
		{Key: "openrouter", Label: "OpenRouter", BaseURL: "https://openrouter.ai/api/v1", Format: "openai", Note: "多供应商路由"},
		{Key: "anthropic", Label: "Anthropic Claude", BaseURL: "https://api.anthropic.com", Format: "anthropic", Note: "Claude 系列，不需要手动添加 /v1"},
		{Key: "custom", Label: "自定义上游", BaseURL: "", Format: "openai", Note: "自建或中转服务，填好地址后使用检测自动识别协议"},
	}
}
