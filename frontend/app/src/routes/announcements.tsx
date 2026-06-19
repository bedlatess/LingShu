import React from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { useTranslation } from "react-i18next";
import type { Announcement } from "@lingshu/shared/user-types";

import { Badge, Card, CardContent, EmptyState, PageHeader } from "@lingshu/ui";
import { useAuth } from "@/providers/auth";

export function AnnouncementsPage() {
  const { t, i18n } = useTranslation("announcements");
  const { api } = useAuth();
  const [items, setItems] = React.useState<Announcement[]>([]);

  React.useEffect(() => {
    api.userAnnouncements().then((result) => setItems(result.items));
  }, [api]);

  return (
    <div className="page-grid">
      <PageHeader eyebrow={t("eyebrow")} title={t("title")} description={t("description")} />
      {items.length ? (
        <section className="grid gap-4">
          {items.map((item) => (
            <Card key={item.id}>
              <CardContent className="p-5">
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div>
                    <h2 className="font-serif text-xl font-semibold">{item.title}</h2>
                    <p className="mt-1 text-xs text-muted-foreground">{new Date(item.created_at).toLocaleString(i18n.resolvedLanguage === "zh" ? "zh-CN" : "en-US", { hour12: false })}</p>
                  </div>
                  {item.pinned ? <Badge variant="warning">{t("pinned")}</Badge> : null}
                </div>
                <div className="prose mt-4 max-w-none">
                  <ReactMarkdown remarkPlugins={[remarkGfm]}>{item.content}</ReactMarkdown>
                </div>
              </CardContent>
            </Card>
          ))}
        </section>
      ) : (
        <EmptyState title={t("emptyTitle")} description={t("emptyDescription")} />
      )}
    </div>
  );
}
