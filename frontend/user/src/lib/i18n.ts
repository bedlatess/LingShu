export const statusMap: Record<string, string> = {
  active: "启用",
  enabled: "启用",
  disabled: "停用",
  banned: "封禁",
  online: "在线",
  offline: "下线",
  success: "成功",
  failed: "失败"
};

export const billingModeMap: Record<string, string> = {
  token: "按量计费",
  per_call: "按次计费"
};

export const typeMap: Record<string, string> = {
  chat: "对话",
  embedding: "向量",
  image: "图像",
  video: "视频"
};

export const ledgerTypeMap: Record<string, string> = {
  admin_grant: "充值",
  admin_deduct: "扣减",
  redeem: "兑换",
  usage_charge: "消费",
  refund: "退款",
  adjustment: "调整",
  charge: "消费",
  adjust: "调整"
};

export function zhStatus(value?: string) {
  if (!value) return "-";
  return statusMap[value] ?? value;
}

export function zhBillingMode(value?: string) {
  if (!value) return "-";
  return billingModeMap[value] ?? value;
}

export function zhType(value?: string) {
  if (!value) return "-";
  return typeMap[value] ?? value;
}

export function zhLedgerType(value?: string) {
  if (!value) return "-";
  return ledgerTypeMap[value] ?? value;
}
