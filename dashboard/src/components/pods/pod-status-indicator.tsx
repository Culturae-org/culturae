import type { PodInfo } from "@/lib/types/metrics.types";

interface PodStatusIndicatorProps {
  status: PodInfo["status"];
}

export function PodStatusIndicator({ status }: PodStatusIndicatorProps) {
  const dot =
    status === "healthy"
      ? "bg-emerald-500"
      : status === "degraded"
        ? "bg-amber-500"
        : "bg-red-500";

  const label =
    status === "healthy"
      ? "text-emerald-600 dark:text-emerald-400"
      : status === "degraded"
        ? "text-amber-600 dark:text-amber-400"
        : "text-red-600 dark:text-red-400";

  return (
    <div className="flex items-center gap-1.5">
      <span className={`inline-block h-1.5 w-1.5 rounded-full ${dot}`} />
      <span className={`text-sm capitalize ${label}`}>{status}</span>
    </div>
  );
}
