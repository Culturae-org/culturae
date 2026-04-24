import { Badge } from "@/components/ui/badge";
import { TableCell, TableRow } from "@/components/ui/table";
import type { PodInfo } from "@/lib/types/metrics.types";
import { PodStatusIndicator } from "./pod-status-indicator";
import { PodTypeBadge } from "./pod-type-badge";

interface PodRowProps {
  pod: PodInfo;
}

function getUptime(startedAt: Date): string {
  const diff = Date.now() - startedAt.getTime();
  const hours = Math.floor(diff / (1000 * 60 * 60));
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
  if (hours > 24) {
    const days = Math.floor(hours / 24);
    return `${days}d ${hours % 24}h`;
  }
  return `${hours}h ${minutes}m`;
}

function getHeartbeatAge(lastHeartbeat: string): { label: string; stale: boolean } {
  const diff = Date.now() - new Date(lastHeartbeat).getTime();
  const secs = Math.floor(diff / 1000);
  const stale = secs > 12; // heartbeat TTL is 15s, warn after 12s
  if (secs < 60) return { label: `${secs}s ago`, stale };
  return { label: `${Math.floor(secs / 60)}m ago`, stale };
}

export function PodRow({ pod }: PodRowProps) {
  const startedDate = new Date(pod.started_at);
  const uptime = getUptime(startedDate);
  const heartbeat = getHeartbeatAge(pod.last_heartbeat);
  const loadScore = pod.connected_clients + pod.active_games * 2;

  return (
    <TableRow className={pod.is_current ? "bg-muted/40" : ""}>
      <TableCell>
        <PodStatusIndicator status={pod.status} />
      </TableCell>
      <TableCell>
        <div className="flex items-center gap-2">
          <code className="text-xs font-mono text-muted-foreground">
            {pod.pod_id.slice(0, 8)}
          </code>
          {pod.is_current && (
            <Badge variant="outline" className="text-xs h-4 px-1.5">
              current
            </Badge>
          )}
        </div>
      </TableCell>
      <TableCell>
        <PodTypeBadge type={pod.pod_type} />
      </TableCell>
      <TableCell className="text-right tabular-nums text-sm">
        {pod.connected_clients}
      </TableCell>
      <TableCell className="text-right tabular-nums text-sm">
        {pod.online_users}
      </TableCell>
      <TableCell className="text-right tabular-nums text-sm">
        {pod.active_games ?? 0}
      </TableCell>
      <TableCell className="text-right tabular-nums text-sm font-medium">
        {loadScore}
      </TableCell>
      <TableCell className="text-sm text-muted-foreground tabular-nums">
        {uptime}
      </TableCell>
      <TableCell className={`text-xs tabular-nums ${heartbeat.stale ? "text-amber-500" : "text-muted-foreground"}`}>
        {heartbeat.label}
      </TableCell>
    </TableRow>
  );
}
