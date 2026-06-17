import type { LucideIcon } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { cn } from "@/lib/utils";

export function StatCard({ label, value, hint, icon: Icon, tone = "teal" }: { label: string; value: string; hint: string; icon: LucideIcon; tone?: "teal" | "violet" | "blue" | "amber" }) {
  const tones = {
    teal: "text-primary bg-primary/10",
    violet: "text-violet-300 bg-violet-400/10",
    blue: "text-sky-300 bg-sky-400/10",
    amber: "text-amber-300 bg-amber-400/10"
  };
  return (
    <Card className="glass group overflow-hidden transition-all duration-300 hover:-translate-y-1 hover:border-primary/35">
      <CardContent className="p-5">
        <div className="flex items-start justify-between">
          <div>
            <p className="text-sm text-muted-foreground">{label}</p>
            <strong className="mt-2 block text-2xl font-semibold tracking-[-0.02em]">{value}</strong>
          </div>
          <div className={cn("grid h-10 w-10 place-items-center rounded-lg", tones[tone])}>
            <Icon className="h-5 w-5" />
          </div>
        </div>
        <p className="mt-4 text-xs text-muted-foreground">{hint}</p>
      </CardContent>
    </Card>
  );
}
