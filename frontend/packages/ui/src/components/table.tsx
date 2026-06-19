import * as React from "react";
import { ChevronLeft, ChevronRight } from "lucide-react";

import { cn } from "../lib/cn";
import { Button } from "./button";
import { EmptyState, Skeleton } from "./feedback";

export type DataColumn<T> = {
  key: string;
  title: React.ReactNode;
  render?: (row: T) => React.ReactNode;
  className?: string;
  sortable?: boolean;
};

export function Table({ children, className }: { children: React.ReactNode; className?: string }) {
  return <table className={cn("w-full border-collapse text-sm", className)}>{children}</table>;
}

export function DataTable<T>({
  columns,
  data,
  rowKey,
  loading,
  empty,
  emptyTitle = "No data",
  emptyDescription = "New records will appear here.",
  className,
  onRowClick
}: {
  columns: DataColumn<T>[];
  data: T[];
  rowKey: (row: T, index: number) => string;
  loading?: boolean;
  empty?: React.ReactNode;
  emptyTitle?: string;
  emptyDescription?: string;
  className?: string;
  onRowClick?: (row: T) => void;
}) {
  if (loading) {
    return (
      <div className="rounded-lg border border-border bg-card p-4">
        <Skeleton className="h-8 w-1/3" />
        <div className="mt-4 grid gap-3">
          {Array.from({ length: 5 }).map((_, index) => <Skeleton key={index} className="h-10 w-full" />)}
        </div>
      </div>
    );
  }
  if (data.length === 0) {
    return <>{empty ?? <EmptyState title={emptyTitle} description={emptyDescription} />}</>;
  }
  return (
    <div className={cn("overflow-hidden rounded-lg border border-border bg-card", className)}>
      <div className="overflow-x-auto">
        <Table>
          <thead className="bg-[var(--bg-subtle)]">
            <tr>
              {columns.map((column) => (
                <th key={column.key} className={cn("whitespace-nowrap px-4 py-3 text-left text-xs font-medium uppercase tracking-[0.12em] text-muted-foreground", column.className)}>
                  {column.title}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {data.map((row, index) => (
              <tr
                key={rowKey(row, index)}
                className={cn("border-t border-border transition-colors hover:bg-[var(--bg-subtle)]/70", onRowClick && "cursor-pointer")}
                onClick={onRowClick ? () => onRowClick(row) : undefined}
              >
                {columns.map((column) => (
                  <td key={column.key} className={cn("px-4 py-3 align-top text-sm text-foreground", column.className)}>
                    {column.render ? column.render(row) : String((row as Record<string, unknown>)[column.key] ?? "-")}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </Table>
      </div>
    </div>
  );
}

export function Pagination({
  page,
  limit,
  total,
  onChange,
  labels
}: {
  page: number;
  limit: number;
  total: number;
  onChange: (page: number) => void;
  labels?: {
    summary?: (page: number, pages: number, total: number) => React.ReactNode;
    previous?: React.ReactNode;
    next?: React.ReactNode;
  };
}) {
  const pages = Math.max(1, Math.ceil(total / limit));
  return (
    <div className="flex items-center justify-between gap-3 text-sm text-muted-foreground">
      <span>{labels?.summary ? labels.summary(page, pages, total) : `Page ${page} / ${pages}, ${total} total`}</span>
      <div className="flex items-center gap-2">
        <Button variant="secondary" size="sm" disabled={page <= 1} onClick={() => onChange(page - 1)}><ChevronLeft className="h-4 w-4" />{labels?.previous ?? "Previous"}</Button>
        <Button variant="secondary" size="sm" disabled={page >= pages} onClick={() => onChange(page + 1)}>{labels?.next ?? "Next"}<ChevronRight className="h-4 w-4" /></Button>
      </div>
    </div>
  );
}
