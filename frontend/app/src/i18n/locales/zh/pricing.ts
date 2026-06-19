const pricing = {
  eyebrow: "价格表",
  title: "按实际用量计费",
  description: "以下价格为用户侧最终调用价格，最终消费以网关实际返回的 usage 结算。",
  enterConsole: "进入控制台",
  perCall: "单次价格",
  input: "输入价格",
  output: "输出价格",
  emptyTitle: "暂未公开模型",
  emptyDescription: "管理员开放模型后会在这里展示价格。",
  filters: {
    search: "搜索模型、类型或分组",
    allBilling: "全部计费方式",
    token: "Token 计费",
    perCall: "按次计费",
    allGroups: "全部分组",
    defaultGroup: "默认分组"
  },
  detail: {
    priceTitle: "价格明细",
    capabilityTitle: "调用能力",
    modelId: "模型标识",
    context: "上下文长度",
    contextUnavailable: "以模型实际响应为准",
    endpoints: "支持端点"
  },
  landing: {
    badge: "AI API 聚合网关",
    heroTitle: "把多个上游模型整理成一个清爽的 API 入口",
    heroDescription: "灵枢面向私有运营场景：统一 OpenAI 兼容调用入口、平台 API Key、余额预扣、真实 usage 结算和渠道故障转移。少一点复杂，多一点可控。",
    primaryCta: "开始使用",
    secondaryCta: "查看价格",
    previewEyebrow: "Gateway Flow",
    previewTitle: "一次调用的清晰路径",
    stats: {
      openaiCompatible: "OpenAI 兼容接口，现有 SDK 改 base_url 即可接入。",
      realtimeBilling: "按上游 usage 回灌结算，余额和账本同步落库。",
      operatorManaged: "私有运营，仅管理员和普通用户两种角色。"
    },
    flow: {
      client: "客户端发起请求",
      clientHint: "使用平台 API Key 调用 /v1/chat/completions 等接口。",
      gateway: "灵枢完成调度与结算",
      gatewayHint: "预扣余额、选择健康渠道、转发请求并记录日志。",
      upstream: "上游返回真实 usage",
      upstreamHint: "网关按最终用量结算，用户端只看到最终消费。"
    },
    features: {
      multiModel: {
        title: "多模型聚合",
        description: "统一管理 OpenAI 兼容、Anthropic 等上游渠道，把可用模型整理给用户。"
      },
      balance: {
        title: "余额预扣",
        description: "调用前检查余额，调用后按真实 usage 结算，减少欠费和对账风险。"
      },
      failover: {
        title: "渠道故障转移",
        description: "健康检查、冷却窗口和重试排除让异常渠道短期避让。"
      },
      privateOps: {
        title: "私有运营友好",
        description: "兑换码和管理员手动充值即可运营，不引入支付、分销或复杂权限。"
      }
    },
    pricingEyebrow: "公开价格",
    pricingTitle: "只展示对客最终价格",
    pricingDescription: "用户端价格页只呈现最终单价，不展示上游成本、倍率或毛利。管理端仍保留完整运营报表。",
    viewAllPrices: "查看完整价格",
    ctaTitle: "准备把调用入口收束起来？",
    ctaDescription: "登录后即可创建平台 API Key、查看用量和兑换额度。管理员可在后台维护渠道、模型、报表和公告。"
  }
};

export default pricing;
