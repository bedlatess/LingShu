import React from "react";

import { Button } from "@/components/ui/button";

type State = { hasError: boolean };

export class ErrorBoundary extends React.Component<{ children: React.ReactNode }, State> {
  state: State = { hasError: false };

  static getDerivedStateFromError(): State {
    return { hasError: true };
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="grid min-h-screen place-items-center bg-background px-4 text-center">
          <div className="max-w-sm rounded-lg border border-border bg-card p-6">
            <h1 className="font-serif text-lg font-semibold">页面暂时不可用</h1>
            <p className="mt-2 text-sm text-muted-foreground">请刷新页面后重试，或联系管理员处理。</p>
            <Button className="mt-5" onClick={() => window.location.reload()}>刷新页面</Button>
          </div>
        </div>
      );
    }
    return this.props.children;
  }
}
