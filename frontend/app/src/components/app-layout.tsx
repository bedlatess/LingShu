import React from "react";
import { Activity, Bell, Command as CommandIcon, Gauge, KeyRound, LogOut, Moon, PanelTop, Settings, Sun, Ticket, WalletCards } from "lucide-react";
import { NavLink, useNavigate } from "react-router-dom";

import { Button, Command, useHotkeys, cn } from "@lingshu/ui";
import { useAuth } from "@/providers/auth";
import { useTheme } from "@/providers/theme";

const navItems = [
  { to: "/dashboard", label: "概览", icon: Gauge, hint: "g+d" },
  { to: "/api-keys", label: "API 密钥", icon: KeyRound },
  { to: "/usage", label: "用量", icon: Activity },
  { to: "/models", label: "模型", icon: PanelTop },
  { to: "/redeem", label: "兑换", icon: Ticket },
  { to: "/announcements", label: "公告", icon: Bell },
  { to: "/settings", label: "设置", icon: Settings }
];

export function AppLayout({ children }: { children: React.ReactNode }) {
  const { user, logout } = useAuth();
  const { theme, toggleTheme } = useTheme();
  const navigate = useNavigate();
  const [commandOpen, setCommandOpen] = React.useState(false);

  useHotkeys({
    "mod+k": () => setCommandOpen(true),
    "g+d": () => navigate("/dashboard"),
    "g+u": () => navigate("/usage"),
    "/": () => setCommandOpen(true),
    escape: () => setCommandOpen(false)
  });

  return (
    <div className="min-h-screen bg-background">
      <header className="sticky top-0 z-30 border-b border-border bg-card/95 backdrop-blur">
        <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6">
          <div className="flex items-center gap-3">
            <div className="grid h-9 w-9 place-items-center rounded-md border border-border bg-foreground text-sm font-black text-background">LS</div>
            <div>
              <div className="flex items-center gap-2 font-serif text-sm font-semibold text-foreground">灵枢控制台</div>
              <p className="text-xs text-muted-foreground">私有 AI 接入服务</p>
            </div>
          </div>
          <div className="flex items-center gap-2 sm:gap-3">
            <Button variant="secondary" size="sm" onClick={() => setCommandOpen(true)} className="hidden sm:inline-flex">
              <CommandIcon className="h-4 w-4" />
              快速命令
            </Button>
            <div className="hidden items-center gap-2 rounded-md border border-border bg-card px-3 py-2 text-sm text-muted-foreground sm:flex">
              <WalletCards className="h-4 w-4 text-primary" />
              {user?.username ?? "加载中"}
            </div>
            <Button variant="ghost" size="icon" onClick={toggleTheme} title={theme === "dark" ? "切换到浅色" : "切换到深色"} aria-label={theme === "dark" ? "切换到浅色" : "切换到深色"}>
              {theme === "dark" ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />}
            </Button>
            <Button variant="ghost" size="icon" onClick={logout} title="退出登录" aria-label="退出登录">
              <LogOut className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </header>

      <div className="mx-auto grid max-w-7xl gap-5 px-4 py-5 sm:px-6 lg:grid-cols-[220px_1fr]">
        <aside className="hidden rounded-lg border border-border bg-card p-2 shadow-[var(--shadow-xs)] lg:block">
          <nav className="grid gap-1">
            {navItems.map((item) => (
              <NavLink
                key={item.to}
                to={item.to}
                className={({ isActive }) =>
                  cn("flex items-center gap-3 rounded-md px-3 py-2.5 text-sm text-muted-foreground transition-colors hover:bg-[var(--bg-subtle)] hover:text-foreground", isActive && "bg-[var(--clay-soft)] text-[var(--clay-hover)]")
                }
              >
                <item.icon className="h-4 w-4" />
                <span className="flex-1">{item.label}</span>
                {item.hint ? <span className="font-mono text-[10px] text-muted-foreground">{item.hint}</span> : null}
              </NavLink>
            ))}
          </nav>
        </aside>

        <nav className="flex gap-2 overflow-x-auto lg:hidden">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              className={({ isActive }) =>
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

      <Command
        open={commandOpen}
        onClose={() => setCommandOpen(false)}
        items={navItems.map((item) => ({
          label: item.label,
          hint: item.to,
          onSelect: () => navigate(item.to)
        }))}
      />
    </div>
  );
}
