import { Badge } from "@/components/ui/badge";
import type { PodInfo } from "@/lib/types/metrics.types";

interface PodTypeBadgeProps {
  type: PodInfo["pod_type"];
}

export function PodTypeBadge({ type }: PodTypeBadgeProps) {
  const className =
    type === "main"
      ? "text-xs bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-950 dark:text-blue-300 dark:border-blue-800"
      : "text-xs bg-violet-50 text-violet-700 border-violet-200 dark:bg-violet-950 dark:text-violet-300 dark:border-violet-800";

  return (
    <Badge variant="outline" className={className}>
      {type}
    </Badge>
  );
}
