const keys = {
  eyebrow: "API 密钥",
  title: "管理访问密钥",
  description: "创建、停用或删除你的平台 API Key。新密钥只会在创建时完整展示一次。",
  createTitle: "创建新密钥",
  namePlaceholder: "密钥名称，例如：生产环境",
  defaultName: "默认密钥",
  createAction: "创建",
  createSuccess: "API 密钥已创建",
  copyBaseURLSuccess: "base_url 已复制",
  updateSuccess: "API 密钥已更新",
  deleteSuccess: "API 密钥已删除",
  deleteConfirm: "确认删除这个 API 密钥？删除后使用它的请求将无法继续访问。",
  emptyTitle: "还没有 API 密钥",
  emptyDescription: "创建一个密钥后，就可以在 OpenAI SDK 中把 base_url 指向灵枢网关。",
  emptyAction: "创建密钥",
  table: {
    name: "名称",
    mask: "密钥",
    endpoints: "允许端点",
    status: "状态",
    createdAt: "创建时间",
    actions: "操作",
    copyBaseURL: "复制 base_url",
    docs: "示例",
    edit: "编辑",
    delete: "删除"
  },
  endpoints: {
    title: "端点权限",
    all: "允许全部端点",
    hint: "留空表示该密钥可访问全部网关端点；也可以只勾选需要的接口。",
    chat: "OpenAI 对话",
    messages: "Anthropic Messages",
    embeddings: "向量 Embeddings",
    models: "模型列表"
  },
  edit: {
    title: "编辑密钥",
    save: "保存",
    cancel: "取消"
  },
  dialog: {
    title: "请立即复制新密钥",
    description: "完整密钥只展示一次，关闭后无法再次查看。",
    saved: "我已保存",
    copied: "已复制",
    ariaCopy: "复制密钥"
  }
};

export default keys;
