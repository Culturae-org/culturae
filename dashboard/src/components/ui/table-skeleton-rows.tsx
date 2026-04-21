import { Skeleton } from "@/components/ui/skeleton";
import { TableCell, TableRow } from "@/components/ui/table";

interface TableSkeletonRowsProps {
  columnCount: number;
  rowCount?: number;
}

export function TableSkeletonRows({
  columnCount,
  rowCount = 8,
}: TableSkeletonRowsProps) {
  const rows = Array.from({ length: rowCount }, (_, i) => `row-${i}`);
  const cells = Array.from({ length: columnCount }, (_, j) => `cell-${j}`);

  return (
    <>
      {rows.map((rowKey) => (
        <TableRow key={rowKey}>
          {cells.map((cellKey) => (
            <TableCell key={cellKey}>
              <Skeleton className="h-4 w-full" />
            </TableCell>
          ))}
        </TableRow>
      ))}
    </>
  );
}
