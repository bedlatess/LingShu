import type { LucideIcon } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { cn } from "@/lib/utils";

export function StatCard({ label, value, hint, icon: Icon, tone = "teal" }: { label: string; value: string; hint: string; icon: LucideIcon; tone?: "teal" | "violet" | "blue" | "amber" }) {
  const tones = {
    teal: "text-[var(--clay)] bg-[var(--bg-subtle)]",
    violet: "text-[var(--clay)] bg-[var(--bg-subtle)]",
    blue: "text-[var(--clay)] bg-[var(--bg-subtle)]",
    amber: "text-[var(--clay)] bg-[var(--bg-subtle)]"
  };
  return (
    <Card className="group overflow-hidden transition-colors hover:border-[var(--border-strong)]">
      <CardContent className="p-5">
        <div className="flex items-start justify-between">
          <div>
            <p className="text-sm text-muted-foreground">{label}</p>
            <strong className="mt-2 block font-serif text-2xl font-semibold tracking-[-0.02em] text-foreground">{value}</strong>
          </div>
          <div className={cn("grid h-10 w-10 place-items-center rounded-md border border-border", tones[tone])}>
            <Icon className="h-5 w-5" />
          </div>
        </div>
        <p className="mt-4 text-xs text-muted-foreground">{hint}</p>
      </CardContent>
    </Card>
  );
}
