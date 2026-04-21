"use client";

import { Button } from "@/components/ui/button";
import { useEffect } from "react";

/* biome-ignore lint/suspicious/noShadowRestrictedNames: Next.js error boundary requires 'error' parameter */
export default function Error({
  error,
  reset,
}: {
  error: globalThis.Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error("Page error:", error);
  }, [error]);

  return (
    <div className="flex min-h-[50vh] flex-col items-center justify-center gap-4">
      <div className="text-center space-y-2">
        <h2 className="text-2xl font-semibold">Something went wrong</h2>
        <p className="text-muted-foreground text-sm max-w-md">
          {error.message || "An unexpected error occurred. Please try again."}
        </p>
        {error.digest && (
          <p className="text-xs text-muted-foreground font-mono">
            Error ID: {error.digest}
          </p>
        )}
      </div>
      <div className="flex gap-2">
        <Button onClick={reset} variant="default">
          Try again
        </Button>
        <Button
          onClick={() => {
            window.location.href = "/console/";
          }}
          variant="outline"
        >
          Go to dashboard
        </Button>
      </div>
    </div>
  );
}
