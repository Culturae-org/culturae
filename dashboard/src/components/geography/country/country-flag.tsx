"use client";

import { DATASETS_ENDPOINTS } from "@/lib/api/endpoints";
import * as React from "react";

interface CountryFlagProps {
  countryCode: string;
  size?: "sm" | "md" | "lg";
  className?: string;
  fallback?: React.ReactNode;
}

const sizePx = {
  sm: { width: 24, height: 16 },
  md: { width: 32, height: 24 },
  lg: { width: 48, height: 32 },
};

const sizeClasses = {
  sm: "w-6 h-4",
  md: "w-8 h-6",
  lg: "w-12 h-8",
};

export function CountryFlag({
  countryCode,
  size = "md",
  className = "",
  fallback,
}: CountryFlagProps) {
  const [error, setError] = React.useState(false);
  const [loading, setLoading] = React.useState(true);

  const flagUrl = DATASETS_ENDPOINTS.GET_FLAG(countryCode.toLowerCase());

  if (error) {
    return fallback ? (
      fallback
    ) : (
      <span
        className={`${sizeClasses[size]} inline-flex items-center justify-center text-muted-foreground ${className}`}
      >
        🏳️
      </span>
    );
  }

  return (
    <span className={`inline-flex items-center ${className}`}>
      {loading && (
        <span
          className={`${sizeClasses[size]} animate-pulse bg-muted rounded`}
        />
      )}
      {/* eslint-disable-next-line @next/next/no-img-element */}
      <img
        src={flagUrl}
        alt={`${countryCode} flag`}
        width={sizePx[size].width}
        height={sizePx[size].height}
        className={`object-contain rounded-sm shadow-sm ${loading ? "hidden" : ""}`}
        onLoad={() => setLoading(false)}
        onError={() => {
          setError(true);
          setLoading(false);
        }}
      />
    </span>
  );
}

export function EmojiFlag({
  flag,
  className = "",
}: { flag?: string; className?: string }) {
  if (!flag) {
    return <span className={`text-muted-foreground ${className}`}>🏳️</span>;
  }
  return <span className={className}>{flag}</span>;
}
