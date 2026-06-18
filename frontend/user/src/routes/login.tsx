import React from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { ArrowRight, KeyRound, ShieldCheck, Sparkles } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { SiteNav } from "@/components/site-nav";
import { useAuth } from "@/providers/auth";

export function LoginPage() {
  const { token, login } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [loginName, setLoginName] = React.useState("");
  const [password, setPassword] = React.useState("");
  const [error, setError] = React.useState("");
  const [loading, setLoading] = React.useState(false);
  const from = (location.state as { from?: { pathname?: string } } | null)?.from?.pathname ?? "/dashboard";

  React.useEffect(() => {
    if (token) navigate(from, { replace: true });
  }, [token, from, navigate]);

  async function onSubmit(event: React.FormEvent) {
    event.preventDefault();
    setError("");
    setLoading(true);
    try {
      await login(loginName, password);
      toast.success("登录成功");
      navigate(from, { replace: true });
    } catch (err) {
      const message = err instanceof Error ? err.message : "登录失败";
      setError(message);
      toast.error(message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="min-h-screen overflow-hidden bg-background">
      <SiteNav />
      <div className="mx-auto grid w-full max-w-5xl gap-6 px-4 py-10 lg:grid-cols-[1fr_420px]">
        <section className="flex flex-col justify-between rounded-lg border border-border bg-card p-8">
          <div>
            <div className="mb-8 grid h-11 w-11 place-items-center rounded-md border border-border bg-foreground text-sm font-black text-background">LS</div>
            <p className="mb-3 text-xs font-semibold text-primary">灵枢 API 控制台</p>
            <h1 className="max-w-2xl font-serif text-5xl font-semibold text-foreground md:text-7xl">私有 AI 网关，安全可控。</h1>
            <p className="mt-5 max-w-xl text-sm leading-6 text-muted-foreground">
              查看余额、创建平台 API 密钥、了解调用和消费情况。登录后所有数据来自你的灵枢账户。
            </p>
          </div>
          <div className="mt-12 grid gap-3 text-sm text-muted-foreground sm:grid-cols-3">
            <div className="rounded-lg border border-border bg-[var(--bg-subtle)] p-4"><ShieldCheck className="mb-3 h-5 w-5 text-primary" />余额实时消费</div>
            <div className="rounded-lg border border-border bg-[var(--bg-subtle)] p-4"><KeyRound className="mb-3 h-5 w-5 text-primary" />密钥安全管理</div>
            <div className="rounded-lg border border-border bg-[var(--bg-subtle)] p-4"><Sparkles className="mb-3 h-5 w-5 text-primary" />统一模型接入</div>
          </div>
        </section>

        <Card className="self-center">
          <CardContent className="p-6">
            <div className="mb-6">
              <h2 className="font-serif text-xl font-semibold">登录用户端</h2>
              <p className="mt-2 text-sm text-muted-foreground">使用用户名或邮箱登录。</p>
            </div>
            <form className="grid gap-4" onSubmit={onSubmit}>
              <label className="grid gap-2 text-sm">
                账号
                <Input value={loginName} onChange={(event) => setLoginName(event.target.value)} autoComplete="username" placeholder="用户名或邮箱" required />
              </label>
              <label className="grid gap-2 text-sm">
                密码
                <Input value={password} onChange={(event) => setPassword(event.target.value)} type="password" autoComplete="current-password" placeholder="请输入密码" required />
              </label>
              {error ? <div className="rounded-md border border-destructive/30 bg-[var(--danger-soft)] px-3 py-2 text-sm text-destructive">{error}</div> : null}
              <Button type="submit" disabled={loading}>
                {loading ? "登录中" : "进入控制台"} <ArrowRight className="h-4 w-4" />
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
    </main>
  );
}
