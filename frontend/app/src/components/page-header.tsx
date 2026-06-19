export function PageHeader({ eyebrow, title, description }: { eyebrow: string; title: string; description: string }) {
  return (
    <div className="mb-8 flex flex-col gap-3 border-b border-border pb-6">
      <p className="text-xs font-medium uppercase tracking-[0.18em] text-[var(--clay)]">{eyebrow}</p>
      <h1 className="max-w-3xl font-serif text-3xl font-semibold tracking-tight text-foreground sm:text-[2.5rem] leading-[1.1]">{title}</h1>
      <p className="max-w-2xl text-[15px] leading-7 text-muted-foreground">{description}</p>
    </div>
  );
}

