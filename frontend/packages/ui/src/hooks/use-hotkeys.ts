import { useEffect } from "react";

export function useHotkeys(bindings: Record<string, () => void>, enabled = true) {
  useEffect(() => {
    if (!enabled) return;
    const handler = (event: KeyboardEvent) => {
      const parts = [
        event.metaKey || event.ctrlKey ? "mod" : "",
        event.shiftKey ? "shift" : "",
        event.altKey ? "alt" : "",
        event.key.toLowerCase()
      ].filter(Boolean);
      const combo = parts.join("+");
      const action = bindings[combo] ?? bindings[event.key.toLowerCase()];
      if (action) {
        event.preventDefault();
        action();
      }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [bindings, enabled]);
}
