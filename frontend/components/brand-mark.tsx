import { cn } from "@/lib/utils";

/**
 * The "trilha" signature: a small connected path of nodes, like a study
 * route. Used as the product mark in the sidebar and on the login screen.
 */
export function BrandMark({ className }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 32 32"
      fill="none"
      aria-hidden="true"
      className={cn("size-8", className)}
    >
      <path
        d="M7 24c4-1 5-7 9-8s5-7 9-8"
        stroke="currentColor"
        strokeWidth="2.4"
        strokeLinecap="round"
        opacity="0.55"
      />
      <circle cx="7" cy="24" r="3.4" fill="currentColor" />
      <circle cx="16" cy="16" r="2.6" fill="currentColor" opacity="0.8" />
      <circle cx="25" cy="8" r="3.4" fill="currentColor" />
    </svg>
  );
}

export function BrandLockup({
  className,
  markClassName,
}: {
  className?: string;
  markClassName?: string;
}) {
  return (
    <div className={cn("flex items-center gap-2.5", className)}>
      <BrandMark className={cn("size-7", markClassName)} />
      <span className="font-[family-name:var(--font-display)] text-lg font-semibold tracking-tight">
        Planej<span className="text-sidebar-primary">AI</span>
      </span>
    </div>
  );
}
