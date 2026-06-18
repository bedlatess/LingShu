import { Link } from "react-router-dom";
import { ArrowRight } from "lucide-react";

import { Button } from "@/components/ui/button";
import { useAuth } from "@/providers/auth";

export function SiteNav({ siteName = "LingShu" }: { siteName?: string }) {
  const { token, user } = useAuth();
  return (
    <header className="sticky top-0 z-40 border-b border-white/10 bg-background/75 backdrop-blur-xl">
      <div className="mx-auto flex h-16 w-full max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
        <Link to="/" className="flex items-center gap-3">
          <span className="grid h-9 w-9 place-items-center rounded-lg bg-primary text-xs font-black text-primary-foreground shadow-glow">LS</span>
          <span className="text-sm font-semibold text-foreground">{siteName}</span>
        </Link>
        <nav className="hidden items-center gap-6 text-sm text-muted-foreground md:flex">
          <Link className="transition hover:text-foreground" to="/pricing">
            价格
          </Link>
          <a className="transition hover:text-foreground" href="/api.md">
            文档
          </a>
        </nav>
        <div className="flex items-center gap-2">
          {token ? (
            <>
              <span className="hidden text-sm text-muted-foreground sm:inline">{user?.username ?? "已登录"}</span>
              <Button asChild size="sm">
                <Link to="/dashboard">
                  控制台 <ArrowRight className="h-4 w-4" />
                </Link>
              </Button>
            </>
          ) : (
            <Button asChild size="sm" variant="outline">
              <Link to="/login">登录</Link>
            </Button>
          )}
        </div>
      </div>
    </header>
  );
}
