const usage = {
  eyebrow: "用量统计",
  title: "用量与消费明细",
  description: "从每日、模型、账本和请求四个视角查看账户使用情况。",
  tabs: {
    daily: "每日统计",
    models: "按模型",
    ledger: "扣费记录",
    logs: "请求日志"
  },
  filters: {
    search: "搜索请求、模型或备注",
    allStatus: "全部状态",
    success: "成功",
    failed: "失败",
    inFlight: "进行中请求 {{count}}",
    from: "开始日期",
    to: "结束日期",
    timezone: "当前时区 {{timezone}}"
  },
  inFlight: {
    title: "进行中请求 {{count}}",
    emptyTitle: "当前没有进行中请求",
    emptyDescription: "新的请求开始后会在这里短暂出现，便于观察实时占用。",
    pollingHint: "该区块每 5 秒自动刷新一次。"
  },
  dailyTitle: "每日消费",
  dailyEmptyTitle: "还没有每日统计",
  dailyEmptyDescription: "产生调用后会按天展示请求数和消费走势。",
  modelsTitle: "模型消费分布",
  modelsEmptyTitle: "还没有模型统计",
  modelsEmptyDescription: "模型消费分布会在这里出现。",
  ledgerTable: {
    type: "类型",
    amount: "金额",
    balanceAfter: "变动后余额",
    remark: "备注",
    createdAt: "时间"
  },
  logsTable: {
    requestId: "请求",
    model: "模型",
    status: "状态",
    tokens: "Token",
    charge: "扣费",
    createdAt: "时间"
  },
  detail: {
    title: "请求详情"
  }
};

export default usage;
