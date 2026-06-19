import React from "react";
import { BookOpen, Check, Copy, KeyRound, Terminal } from "lucide-react";
import { useTranslation } from "react-i18next";
import type { UserModelConfig } from "@lingshu/shared/user-types";

import { Badge, Button, Card, CardContent, CardHeader, CardTitle, DataTable, EmptyState, PageHeader, Tabs, toast } from "@lingshu/ui";
import { copyText } from "@/lib/clipboard";
import { trBillingMode, trType } from "@/lib/i18n";
import { useAuth } from "@/providers/auth";

const sampleKey = "sk-lingshu_xxxxxxxxxxxxxxxxxxxxxxxx";

export function DocsPage() {
  const { t } = useTranslation("docs");
  const { api } = useAuth();
  const [models, setModels] = React.useState<UserModelConfig[]>([]);
  const [tab, setTab] = React.useState("curl");
  const [loading, setLoading] = React.useState(true);
  const baseURL = React.useMemo(() => `${window.location.origin}/v1`, []);
  const chatModel = models.find((item) => item.type === "chat")?.public_name || "gpt-4o-mini";
  const embeddingModel = models.find((item) => item.type === "embedding")?.public_name || "text-embedding-3-small";

  React.useEffect(() => {
    api.userModels().then((result) => setModels(result.items)).finally(() => setLoading(false));
  }, [api]);

  const snippets = React.useMemo(() => buildSnippets(baseURL, chatModel, embeddingModel), [baseURL, chatModel, embeddingModel]);

  async function copy(value: string) {
    if (await copyText(value)) {
      toast.success(t("copied"));
    }
  }

  return (
    <div className="page-grid">
      <PageHeader
        eyebrow={t("eyebrow")}
        title={t("title")}
        description={t("description")}
        action={<Button variant="secondary" onClick={() => copy(baseURL)}><Copy className="h-4 w-4" />{t("copyBaseURL")}</Button>}
      />

      <section className="grid gap-4 lg:grid-cols-[0.9fr_1.1fr]">
        <Card>
          <CardHeader>
            <CardTitle>{t("baseUrl.title")}</CardTitle>
          </CardHeader>
          <CardContent className="grid gap-4">
            <div className="rounded-md border border-border bg-[var(--bg-subtle)] p-3">
              <p className="text-xs text-muted-foreground">{t("baseUrl.label")}</p>
              <code className="mt-2 block break-all font-mono text-sm text-foreground">{baseURL}</code>
            </div>
            <div className="grid gap-2 text-sm text-muted-foreground">
              <p>{t("baseUrl.keyHint")}</p>
              <code className="rounded-md border border-border bg-card p-2 font-mono text-xs">{sampleKey}</code>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{t("quickConfig.title")}</CardTitle>
          </CardHeader>
          <CardContent className="grid gap-3">
            <Snippet title={t("quickConfig.codex")} value={`OPENAI_BASE_URL=${baseURL}\nOPENAI_API_KEY=${sampleKey}`} onCopy={copy} />
            <Snippet title={t("quickConfig.claude")} value={`ANTHROPIC_BASE_URL=${baseURL}\nANTHROPIC_API_KEY=${sampleKey}`} onCopy={copy} />
          </CardContent>
        </Card>
      </section>

      <Card>
        <CardHeader>
          <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
            <CardTitle>{t("examples.title")}</CardTitle>
            <Tabs
              value={tab}
              onChange={setTab}
              tabs={[
                { value: "curl", label: "curl" },
                { value: "python", label: "Python" },
                { value: "node", label: "Node" }
              ]}
            />
          </div>
        </CardHeader>
        <CardContent className="grid gap-4">
          {snippets[tab as keyof typeof snippets].map((snippet) => (
            <Snippet key={snippet.title} title={t(`examples.${snippet.title}`)} value={snippet.value} onCopy={copy} />
          ))}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{t("matrix.title")}</CardTitle>
        </CardHeader>
        <CardContent>
          <DataTable
            loading={loading}
            data={models}
            rowKey={(row) => row.id}
            empty={<EmptyState title={t("matrix.emptyTitle")} description={t("matrix.emptyDescription")} icon={<BookOpen className="h-5 w-5" />} />}
            columns={[
              { key: "public_name", title: t("matrix.model"), render: (row) => <span className="font-medium">{row.public_name}</span> },
              { key: "type", title: t("matrix.type"), render: (row) => <Badge>{trType(row.type)}</Badge> },
              { key: "billing_mode", title: t("matrix.billing"), render: (row) => <Badge variant="info">{trBillingMode(row.billing_mode)}</Badge> },
              { key: "stream", title: t("matrix.stream"), render: (row) => <Capability ok={row.type === "chat"} /> },
              { key: "tools", title: t("matrix.tools"), render: (row) => <Capability ok={row.type === "chat"} /> },
              { key: "vision", title: t("matrix.vision"), render: (row) => <Capability ok={row.type === "image"} /> },
              { key: "endpoints", title: t("matrix.endpoints"), render: (row) => <EndpointBadges model={row} /> }
            ]}
          />
        </CardContent>
      </Card>
    </div>
  );
}

function Snippet({ title, value, onCopy }: { title: string; value: string; onCopy: (value: string) => void }) {
  return (
    <section className="overflow-hidden rounded-lg border border-border bg-card">
      <div className="flex items-center justify-between gap-3 border-b border-border bg-[var(--bg-subtle)] px-4 py-3">
        <div className="flex items-center gap-2 text-sm font-medium text-foreground">
          <Terminal className="h-4 w-4 text-[var(--clay)]" />
          {title}
        </div>
        <Button type="button" variant="secondary" size="sm" onClick={() => onCopy(value)}><Copy className="h-4 w-4" />Copy</Button>
      </div>
      <pre className="overflow-x-auto p-4 text-xs leading-6"><code>{value}</code></pre>
    </section>
  );
}

function Capability({ ok }: { ok: boolean }) {
  return ok ? <span className="inline-flex items-center gap-1 text-sm text-[var(--success)]"><Check className="h-4 w-4" />Yes</span> : <span className="text-sm text-muted-foreground">-</span>;
}

function EndpointBadges({ model }: { model: UserModelConfig }) {
  return (
    <div className="flex flex-wrap gap-1">
      {supportedEndpoints(model).map((endpoint) => <Badge key={endpoint} variant="muted">{endpoint}</Badge>)}
    </div>
  );
}

function supportedEndpoints(model: UserModelConfig) {
  if (model.type === "embedding") return ["/v1/embeddings"];
  if (model.type === "image") return ["/v1/images/generations"];
  return ["/v1/chat/completions", "/v1/messages"];
}

function buildSnippets(baseURL: string, chatModel: string, embeddingModel: string) {
  return {
    curl: [
      {
        title: "openai",
        value: `curl ${baseURL}/chat/completions \\\n  -H "Authorization: Bearer ${sampleKey}" \\\n  -H "Content-Type: application/json" \\\n  -d '{"model":"${chatModel}","messages":[{"role":"user","content":"Hello"}],"stream":false}'`
      },
      {
        title: "anthropic",
        value: `curl ${baseURL}/messages \\\n  -H "x-api-key: ${sampleKey}" \\\n  -H "anthropic-version: 2023-06-01" \\\n  -H "Content-Type: application/json" \\\n  -d '{"model":"${chatModel}","max_tokens":256,"messages":[{"role":"user","content":"Hello"}]}'`
      },
      {
        title: "embeddings",
        value: `curl ${baseURL}/embeddings \\\n  -H "Authorization: Bearer ${sampleKey}" \\\n  -H "Content-Type: application/json" \\\n  -d '{"model":"${embeddingModel}","input":"LingShu gateway"}'`
      }
    ],
    python: [
      {
        title: "openai",
        value: `from openai import OpenAI\n\nclient = OpenAI(base_url="${baseURL}", api_key="${sampleKey}")\nresp = client.chat.completions.create(\n    model="${chatModel}",\n    messages=[{"role": "user", "content": "Hello"}],\n)\nprint(resp.choices[0].message.content)`
      },
      {
        title: "anthropic",
        value: `import anthropic\n\nclient = anthropic.Anthropic(base_url="${baseURL}", api_key="${sampleKey}")\nmsg = client.messages.create(\n    model="${chatModel}",\n    max_tokens=256,\n    messages=[{"role": "user", "content": "Hello"}],\n)\nprint(msg.content[0].text)`
      },
      {
        title: "embeddings",
        value: `from openai import OpenAI\n\nclient = OpenAI(base_url="${baseURL}", api_key="${sampleKey}")\nresp = client.embeddings.create(model="${embeddingModel}", input="LingShu gateway")\nprint(resp.data[0].embedding[:5])`
      }
    ],
    node: [
      {
        title: "openai",
        value: `import OpenAI from "openai";\n\nconst client = new OpenAI({ baseURL: "${baseURL}", apiKey: "${sampleKey}" });\nconst resp = await client.chat.completions.create({\n  model: "${chatModel}",\n  messages: [{ role: "user", content: "Hello" }],\n});\nconsole.log(resp.choices[0].message.content);`
      },
      {
        title: "anthropic",
        value: `import Anthropic from "@anthropic-ai/sdk";\n\nconst client = new Anthropic({ baseURL: "${baseURL}", apiKey: "${sampleKey}" });\nconst msg = await client.messages.create({\n  model: "${chatModel}",\n  max_tokens: 256,\n  messages: [{ role: "user", content: "Hello" }],\n});\nconsole.log(msg.content[0].text);`
      },
      {
        title: "embeddings",
        value: `import OpenAI from "openai";\n\nconst client = new OpenAI({ baseURL: "${baseURL}", apiKey: "${sampleKey}" });\nconst resp = await client.embeddings.create({ model: "${embeddingModel}", input: "LingShu gateway" });\nconsole.log(resp.data[0].embedding.slice(0, 5));`
      }
    ]
  };
}
