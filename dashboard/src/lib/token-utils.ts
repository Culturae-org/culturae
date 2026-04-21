import { decodeJwt } from "jose";

export interface TokenInfo {
  isValid: boolean;
  isExpired: boolean;
  expiresAt: number | null;
  timeUntilExpiry: number | null;
  shouldRefresh: boolean;
}

export interface TokenPayload {
  sub?: string;
  usn?: string;
  sid?: string;
  roles?: string[];
  [key: string]: unknown;
}

export function getUserIdFromToken(token: string | null): string | null {
  if (!token) return null;
  try {
    const payload = decodeJwt(token) as TokenPayload;
    return payload.sub ?? null;
  } catch {
    return null;
  }
}

export function getUsernameFromToken(token: string | null): string | null {
  if (!token) return null;
  try {
    const payload = decodeJwt(token) as TokenPayload;
    return payload.usn ?? null;
  } catch {
    return null;
  }
}

export function analyzeToken(token: string | null): TokenInfo {
  if (!token) {
    return {
      isValid: false,
      isExpired: true,
      expiresAt: null,
      timeUntilExpiry: null,
      shouldRefresh: false,
    };
  }

  try {
    const payload = decodeJwt(token);
    const now = Math.floor(Date.now() / 1000);
    const exp = payload.exp;

    if (!exp) {
      return {
        isValid: false,
        isExpired: true,
        expiresAt: null,
        timeUntilExpiry: null,
        shouldRefresh: false,
      };
    }

    const isExpired = exp <= now;
    const timeUntilExpiry = isExpired ? 0 : (exp - now) * 1000;
    const shouldRefresh = !isExpired && timeUntilExpiry < 300000;

    return {
      isValid: !isExpired,
      isExpired,
      expiresAt: exp * 1000,
      timeUntilExpiry,
      shouldRefresh,
    };
  } catch (error) {
    console.error("Token analysis failed:", error);
    return {
      isValid: false,
      isExpired: true,
      expiresAt: null,
      timeUntilExpiry: null,
      shouldRefresh: false,
    };
  }
}

export function parseExpirationDate(expiresAt: string | number): number {
  if (typeof expiresAt === "number") {
    return expiresAt > 10000000000 ? expiresAt : expiresAt * 1000;
  }

  if (typeof expiresAt === "string") {
    if (expiresAt.includes("T")) {
      return new Date(expiresAt).getTime();
    }
  }

  return 0;
}
