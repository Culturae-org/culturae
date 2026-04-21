"use client";

import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from "@/components/ui/hover-card";
import { IconInfoCircle } from "@tabler/icons-react";

interface InfoHoverProps {
  description: string;
  docLink?: string;
}

export function InfoHover({ description, docLink }: InfoHoverProps) {
  return (
    <HoverCard openDelay={10} closeDelay={100}>
      <HoverCardTrigger asChild>
        <button type="button">
          <IconInfoCircle className="h-4 w-4 text-muted-foreground hover:text-foreground" />
        </button>
      </HoverCardTrigger>
      <HoverCardContent className="flex w-64 flex-col gap-0.5">
        <div className="text-sm text-foreground">{description}</div>
        {docLink && (
          <div className="mt-1 text-xs text-blue-500 hover:underline flex items-center gap-1">
            <a href={docLink} target="_blank" rel="noopener noreferrer">
              Learn more in docs →
            </a>
          </div>
        )}
      </HoverCardContent>
    </HoverCard>
  );
}
