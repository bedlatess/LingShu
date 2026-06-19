import { useTranslation } from "react-i18next";
import type { ModelConfig } from "@lingshu/shared";
import { Input, Select, Switch } from "@lingshu/ui";

export function ModelForm({ value, onChange }: { value: Omit<ModelConfig, "id">; onChange: (value: Omit<ModelConfig, "id">) => void }) {
  const { t } = useTranslation("admin");
  return (
    <div className="grid gap-3">
      <Input value={value.public_name} onChange={(event) => onChange({ ...value, public_name: event.target.value })} placeholder={t("models.publicName")} />
      <Select value={value.type} onChange={(event) => onChange({ ...value, type: event.target.value })}><option value="chat">{t("models.types.chat")}</option><option value="embedding">{t("models.types.embedding")}</option><option value="image">{t("models.types.image")}</option></Select>
      <Select value={value.billing_mode} onChange={(event) => onChange({ ...value, billing_mode: event.target.value })}><option value="token">{t("models.billing.token")}</option><option value="per_call">{t("models.billing.per_call")}</option></Select>
      <Input value={value.input_price_per_1k} onChange={(event) => onChange({ ...value, input_price_per_1k: event.target.value })} placeholder={t("models.inputCost")} />
      <Input value={value.output_price_per_1k} onChange={(event) => onChange({ ...value, output_price_per_1k: event.target.value })} placeholder={t("models.outputCost")} />
      <Input value={value.cache_creation_price_per_1k} onChange={(event) => onChange({ ...value, cache_creation_price_per_1k: event.target.value })} placeholder={t("models.cacheCreateCost")} />
      <Input value={value.cache_read_price_per_1k} onChange={(event) => onChange({ ...value, cache_read_price_per_1k: event.target.value })} placeholder={t("models.cacheReadCost")} />
      <Input value={value.rate_multiplier} onChange={(event) => onChange({ ...value, rate_multiplier: event.target.value })} placeholder={t("common.multiplier")} />
      <fieldset className="grid gap-3 rounded-md border border-border bg-[var(--bg-subtle)] p-3">
        <legend className="px-1 text-sm font-medium text-foreground">{t("models.capabilities.title")}</legend>
        <CapabilitySwitch label={t("models.capabilities.stream")} checked={value.supports_stream} onChange={(checked) => onChange({ ...value, supports_stream: checked })} />
        <CapabilitySwitch label={t("models.capabilities.tools")} checked={value.supports_tools} onChange={(checked) => onChange({ ...value, supports_tools: checked })} />
        <CapabilitySwitch label={t("models.capabilities.vision")} checked={value.supports_vision} onChange={(checked) => onChange({ ...value, supports_vision: checked })} />
      </fieldset>
    </div>
  );
}

function CapabilitySwitch({ label, checked, onChange }: { label: string; checked: boolean; onChange: (checked: boolean) => void }) {
  return (
    <label className="flex min-h-9 cursor-pointer items-center justify-between gap-3 text-sm text-foreground">
      <span>{label}</span>
      <Switch checked={checked} onCheckedChange={onChange} />
    </label>
  );
}
