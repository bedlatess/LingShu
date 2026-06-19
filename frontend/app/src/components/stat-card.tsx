import type { LucideIcon } from "lucide-react";
import { Card, CardContent } from "@lingshu/ui";

export function StatCard({ label, value, hint, icon: Icon }: { label: string; value: string; hint: string; icon: LucideIcon }) {
  return (
    <Card className="group overflow-hidden transition-colors hover:border-[var(--border-strong)]">
      <CardContent className="p-5">
        <div className="flex items-start justify-between">
          <div>
            <p className="text-sm text-muted-foreground">{label}</p>
            <strong className="mt-2 block font-serif text-2xl font-semibold tracking-[-0.02em] text-foreground">{value}</strong>
          </div>
          <div className="grid h-10 w-10 place-items-center rounded-md border border-border bg-[var(--bg-subtle)] text-[var(--clay)]">
            <Icon className="h-5 w-5" />
          </div>
        </div>
        <p className="mt-4 text-xs text-muted-foreground">{hint}</p>
      </CardContent>
    </Card>
  );
}

