const isSecure = (): boolean =>
  typeof window !== "undefined" && window.location.protocol === "https:";

export function setCookie(name: string, value: string, maxAgeDays = 30): void {
  if (typeof document === "undefined") return;

  const maxAge = maxAgeDays * 24 * 60 * 60;
  const secure = isSecure() ? "; Secure" : "";

  const encodedName = encodeURIComponent(name);
  const encodedValue = encodeURIComponent(value);

  document.cookie = `${encodedName}=${encodedValue}; path=/; SameSite=Strict; Max-Age=${maxAge}${secure}`;
}

export function getCookie(name: string): string | null {
  if (typeof document === "undefined") return null;

  const escapedName = name.replaceAll(/[.*+?^${}()|[\]\\]/g, "\\$&");

  const match = document.cookie.match(
    new RegExp(`(?:^|; )${escapedName}=([^;]*)`),
  );

  return match ? decodeURIComponent(match[1]) : null;
}

export function deleteCookie(name: string): void {
  if (typeof document === "undefined") return;

  const secure = isSecure() ? "; Secure" : "";
  const encodedName = encodeURIComponent(name);

  document.cookie = `${encodedName}=; path=/; SameSite=Strict; Max-Age=0${secure}`;
}
