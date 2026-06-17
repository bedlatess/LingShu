import { Activity, Bell, Gauge, KeyRound, LogOut, MessageSquareText, PanelTop, Settings, Sparkles, Ticket, WalletCards } from "lucide-react";
import { NavLink } from "react-router-dom";

import { Button } from "@/components/ui/button";
import { useAuth } from "@/providers/auth";
import { cn } from "@/lib/utils";

const navItems = [
  { to: "/dashboard", label: "Dashboard", icon: Gauge },
  { to: "/api-keys", label: "API Keys", icon: KeyRound },
  { to: "/usage", label: "Usage", icon: Activity },
  { to: "/models", label: "Models", icon: PanelTop },
  { to: "/redeem", label: "Redeem", icon: Ticket },
  { to: "/announcements", label: "News", icon: Bell },
  { to: "/settings", label: "Settings", icon: Settings }
];

export function AppLayout({ children }: { children: React.ReactNode }) {
  const { user, logout } = useAuth();

  return (
    <div className="min-h-screen soft-grid">
      <header className="sticky top-0 z-30 border-b border-white/10 bg-background/70 backdrop-blur-2xl">
        <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6">
          <div className="flex items-center gap-3">
            <div className="grid h-9 w-9 place-items-center rounded-lg bg-primary text-sm font-black text-primary-foreground shadow-glow">LS</div>
            <div>
              <div className="flex items-center gap-2 text-sm font-semibold">
                LingShu Console <Sparkles className="h-3.5 w-3.5 text-primary" />
              </div>
              <p className="text-xs text-muted-foreground">Private AI API gateway</p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <div className="hidden items-center gap-2 rounded-lg border border-white/10 bg-white/[0.04] px-3 py-2 text-sm text-muted-foreground sm:flex">
              <WalletCards className="h-4 w-4 text-primary" />
              {user?.username ?? "Loading"}
            </div>
            <Button variant="ghost" size="icon" onClick={logout} title="退出登录">
              <LogOut className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </header>

      <div className="mx-auto grid max-w-7xl gap-5 px-4 py-5 sm:px-6 lg:grid-cols-[220px_1fr]">
        <aside className="glass hidden rounded-lg p-2 lg:block">
          <nav className="grid gap-1">
            {navItems.map((item) => (
              <NavLink
                key={item.to}
                to={item.to}
                className={({ isActive }: { isActive: boolean }) =>
                  cn("flex items-center gap-3 rounded-md px-3 py-2.5 text-sm text-muted-foreground transition-all hover:bg-white/[0.06] hover:text-foreground", isActive && "bg-primary/12 text-primary")
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
                cn("flex shrink-0 items-center gap-2 rounded-lg border border-white/10 bg-white/[0.04] px-3 py-2 text-sm text-muted-foreground", isActive && "border-primary/40 text-primary")
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
