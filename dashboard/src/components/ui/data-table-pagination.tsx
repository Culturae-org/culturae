"use client";

import { Field, FieldLabel } from "@/components/ui/field";
import {
  Pagination,
  PaginationContent,
  PaginationItem,
  PaginationNext,
  PaginationPrevious,
} from "@/components/ui/pagination";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import * as React from "react";

export interface DataTablePaginationProps {
  pageIndex: number;
  pageCount: number;
  pageSize: number;
  totalCount?: number;
  onPageChange: (pageIndex: number) => void;
  onPageSizeChange?: (pageSize: number) => void;
  onPageReset?: () => void;
  loading?: boolean;
  pageLabel?: (page: number, total: number) => string;
}

export function DataTablePagination({
  pageIndex,
  pageCount,
  pageSize,
  totalCount,
  onPageChange,
  onPageSizeChange,
  onPageReset,
  loading = false,
  pageLabel,
}: DataTablePaginationProps) {
  const currentPage = pageIndex + 1;

  const formatPageLabel = (page: number, total: number) => {
    if (pageLabel) return pageLabel(page, total);
    return `Page ${page} of ${total}`;
  };

  const handlePageSizeChange = (value: string) => {
    if (onPageReset) onPageReset();
    onPageSizeChange?.(Number(value));
  };

  return (
    <div className="flex flex-col gap-2 py-4 sm:flex-row sm:items-center sm:justify-between">
      <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:gap-4">
        {onPageSizeChange && (
          <Field orientation="horizontal" className="w-fit">
            <FieldLabel htmlFor="select-rows-per-page">
              Rows per page
            </FieldLabel>
            <Select
              value={String(pageSize)}
              onValueChange={handlePageSizeChange}
            >
              <SelectTrigger
                className="w-20"
                id="select-rows-per-page"
                size="sm"
              >
                <SelectValue />
              </SelectTrigger>
              <SelectContent align="start">
                <SelectGroup>
                  {[10, 20, 30, 40, 50].map((s) => (
                    <SelectItem key={s} value={String(s)}>
                      {s}
                    </SelectItem>
                  ))}
                </SelectGroup>
              </SelectContent>
            </Select>
          </Field>
        )}
        <span className="text-sm text-muted-foreground">
          {formatPageLabel(currentPage, pageCount)}
          {totalCount !== undefined && (
            <span className="ml-2">— {totalCount} items</span>
          )}
        </span>
      </div>
      <Pagination className="mx-0 w-auto">
        <PaginationContent>
          <PaginationItem>
            <PaginationPrevious
              onClick={() => onPageChange(pageIndex - 1)}
              disabled={pageIndex === 0 || loading}
            />
          </PaginationItem>
          <PaginationItem>
            <PaginationNext
              onClick={() => onPageChange(pageIndex + 1)}
              disabled={pageIndex >= pageCount - 1 || loading}
            />
          </PaginationItem>
        </PaginationContent>
      </Pagination>
    </div>
  );
}
