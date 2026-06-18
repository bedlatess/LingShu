import { Activity, Bell, Gauge, KeyRound, LogOut, PanelTop, Settings, Ticket, WalletCards } from "lucide-react";
import { NavLink } from "react-router-dom";

import { Button } from "@/components/ui/button";
import { useAuth } from "@/providers/auth";
import { cn } from "@/lib/utils";

const navItems = [
  { to: "/dashboard", label: "概览", icon: Gauge },
  { to: "/api-keys", label: "API 密钥", icon: KeyRound },
  { to: "/usage", label: "用量", icon: Activity },
  { to: "/models", label: "模型", icon: PanelTop },
  { to: "/redeem", label: "充值兑换", icon: Ticket },
  { to: "/announcements", label: "公告", icon: Bell },
  { to: "/settings", label: "设置", icon: Settings }
];

export function AppLayout({ children }: { children: React.ReactNode }) {
  const { user, logout } = useAuth();

  return (
    <div className="min-h-screen bg-background">
      <header className="sticky top-0 z-30 border-b border-border bg-surface">
        <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6">
          <div className="flex items-center gap-3">
            <div className="grid h-9 w-9 place-items-center rounded-md border border-border bg-foreground text-sm font-black text-background">LS</div>
            <div>
              <div className="flex items-center gap-2 font-serif text-sm font-semibold text-foreground">灵枢控制台</div>
              <p className="text-xs text-muted-foreground">私有 AI 接入服务</p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <div className="hidden items-center gap-2 rounded-md border border-border bg-card px-3 py-2 text-sm text-muted-foreground sm:flex">
              <WalletCards className="h-4 w-4 text-primary" />
              {user?.username ?? "加载中"}
            </div>
            <Button variant="ghost" size="icon" onClick={logout} title="退出登录">
              <LogOut className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </header>

      <div className="mx-auto grid max-w-7xl gap-5 px-4 py-5 sm:px-6 lg:grid-cols-[220px_1fr]">
        <aside className="hidden rounded-lg border border-border bg-surface p-2 lg:block">
          <nav className="grid gap-1">
            {navItems.map((item) => (
              <NavLink
                key={item.to}
                to={item.to}
                className={({ isActive }: { isActive: boolean }) =>
                  cn("flex items-center gap-3 rounded-md px-3 py-2.5 text-sm text-muted-foreground transition-colors hover:bg-[var(--bg-subtle)] hover:text-foreground", isActive && "bg-[var(--clay-soft)] text-[var(--clay-hover)]")
                }
              >
                <item.icon className="h-4 w-4" />
                {item.label}
              </NavLink>
            ))}
          </nav>
        </aside>

        <nav className="flex gap-2 overflow-x-auto lg:hidden">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              className={({ isActive }: { isActive: boolean }) =>
                cn("flex shrink-0 items-center gap-2 rounded-md border border-border bg-card px-3 py-2 text-sm text-muted-foreground", isActive && "border-[var(--clay-soft)] bg-[var(--clay-soft)] text-[var(--clay-hover)]")
              }
            >
              <item.icon className="h-4 w-4" />
              {item.label}
            </NavLink>
          ))}
        </nav>

        <main className="min-w-0">{children}</main>
      </div>
    </div>
  );
}
