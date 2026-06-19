import React from "react";
import { Alert, Button } from "@lingshu/ui";

export class ErrorBoundary extends React.Component<{ children: React.ReactNode }, { error: Error | null }> {
  state: { error: Error | null } = { error: null };

  static getDerivedStateFromError(error: Error) {
    return { error };
  }

  render() {
    if (!this.state.error) return this.props.children;
    return (
      <main className="grid min-h-screen place-items-center bg-background p-6">
        <div className="w-full max-w-lg space-y-4">
          <Alert variant="danger" title="页面暂时无法显示">
            {this.state.error.message || "发生了未知错误，请刷新页面后重试。"}
          </Alert>
          <Button onClick={() => window.location.reload()}>刷新页面</Button>
        </div>
      </main>
    );
  }
}
