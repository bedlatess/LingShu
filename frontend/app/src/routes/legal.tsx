import React from "react";
import { useParams } from "react-router-dom";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { useTranslation } from "react-i18next";
import { createAPI } from "@lingshu/shared";
import { Alert, Card, CardContent, PageHeader, Skeleton } from "@lingshu/ui";

import { PublicFooter } from "@/components/public-footer";
import { SiteNav } from "@/components/site-nav";

export function LegalPage() {
  const { t } = useTranslation("auth");
  const { slug = "tos" } = useParams();
  const safeSlug = slug === "privacy" ? "privacy" : "tos";
  const [markdown, setMarkdown] = React.useState("");
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState("");

  React.useEffect(() => {
    setLoading(true);
    setError("");
    createAPI()
      .legal(safeSlug)
      .then((result) => setMarkdown(result.markdown || t(`legal.${safeSlug}.fallback`)))
      .catch((err) => setError(err instanceof Error ? err.message : t("legal.loadFailed")))
      .finally(() => setLoading(false));
  }, [safeSlug, t]);

  return (
    <main className="min-h-screen bg-background">
      <SiteNav />
      <section className="mx-auto max-w-4xl px-4 py-12 sm:px-6">
        <PageHeader eyebrow={t("legal.eyebrow")} title={t(`legal.${safeSlug}.title`)} description={t(`legal.${safeSlug}.description`)} />
        {loading ? (
          <div className="grid gap-3">
            <Skeleton className="h-8" />
            <Skeleton className="h-64" />
          </div>
        ) : error ? (
          <Alert variant="danger">{error}</Alert>
        ) : (
          <Card>
            <CardContent className="prose max-w-none p-6">
              <ReactMarkdown remarkPlugins={[remarkGfm]}>{markdown}</ReactMarkdown>
            </CardContent>
          </Card>
        )}
      </section>
      <PublicFooter />
    </main>
  );
}
