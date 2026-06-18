import { Sparkles } from "lucide-react";
import { Button } from "@/components/ui/button";

export function EmptyState({ title, description, action, onAction }: { title: string; description: string; action?: string; onAction?: () => void }) {
  return (
    <div className="grid place-items-center rounded-lg border border-dashed border-border bg-card p-8 text-center">
      <Sparkles className="mb-4 h-9 w-9 text-[var(--ink-faint)]" />
      <h3 className="font-serif text-base font-semibold text-foreground">{title}</h3>
      <p className="mt-2 max-w-sm text-sm leading-6 text-muted-foreground">{description}</p>
      {action && onAction ? (
        <Button className="mt-4" variant="secondary" onClick={onAction}>
          {action}
        </Button>
      ) : null}
    </div>
  );
}
