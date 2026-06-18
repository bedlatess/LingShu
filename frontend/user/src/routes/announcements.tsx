import React from "react";
import { Bell, Pin } from "lucide-react";
import type { Announcement } from "@lingshu/shared/user-types";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { PageHeader } from "@/components/page-header";
import { EmptyState } from "@/components/empty-state";
import { useAuth } from "@/providers/auth";

export function AnnouncementsPage() {
  const { api } = useAuth();
  const [items, setItems] = React.useState<Announcement[]>([]);

  React.useEffect(() => {
    api.userAnnouncements().then((result) => setItems(result.items));
  }, [api]);

  return (
    <div className="page-grid">
      <PageHeader eyebrow="公告" title="公告和通知" description="服务说明、联系方式或维护通知会展示在这里。" />
      {items.length === 0 ? (
        <EmptyState title="暂无公告" description="当前没有在线公告。" />
      ) : (
        <div className="grid gap-4">
          {items.map((item) => (
            <Card key={item.id} className="glass transition-all hover:border-primary/35">
              <CardContent className="p-5">
                <div className="mb-3 flex items-center gap-2">
                  <Bell className="h-4 w-4 text-primary" />
                  <h3 className="font-semibold">{item.title}</h3>
                  {item.pinned ? <Badge><Pin className="mr-1 h-3 w-3" />置顶</Badge> : null}
                </div>
                <p className="whitespace-pre-wrap text-sm leading-7 text-muted-foreground">{item.content}</p>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
