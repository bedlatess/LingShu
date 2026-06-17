import { Sparkles } from "lucide-react";
import { Button } from "@/components/ui/button";

export function EmptyState({ title, description, action, onAction }: { title: string; description: string; action?: string; onAction?: () => void }) {
  return (
    <div className="grid place-items-center rounded-lg border border-dashed border-white/15 bg-white/[0.025] p-8 text-center">
      <div className="relative mb-4 h-20 w-28">
        <div className="absolute left-2 top-6 h-10 w-20 rounded-lg border border-primary/30 bg-primary/10" />
        <div className="absolute right-2 top-2 h-12 w-16 rounded-lg border border-violet-300/25 bg-violet-400/10" />
        <div className="absolute left-10 top-8 grid h-10 w-10 place-items-center rounded-full bg-background shadow-glow">
          <Sparkles className="h-5 w-5 text-primary" />
        </div>
      </div>
      <h3 className="text-sm font-semibold">{title}</h3>
      <p className="mt-2 max-w-sm text-sm leading-6 text-muted-foreground">{description}</p>
      {action && onAction ? (
        <Button className="mt-4" variant="secondary" onClick={onAction}>
          {action}
        </Button>
      ) : null}
    </div>
  );
}
