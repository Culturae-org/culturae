/**
 * Parses an entity_key returned by analytics endpoints.
 *
 * Format:
 *   "geo:<country-slug>:<variant>"  → geography / flag question
 *   "q:<uuid>"                      → trivia question
 */
export function parseEntityKey(key: string): {
  label: string;
  sub: string;
  isGeo: boolean;
} {
  if (key.startsWith("geo:")) {
    const parts = key.split(":");
    const slug = parts[1] ?? "";
    const variant = parts[2] ?? "";
    const name = slug
      ? slug.charAt(0).toUpperCase() + slug.slice(1).replace(/-/g, " ")
      : key;
    const sub = variant ? variant.replace(/_/g, " ") : "geography";
    return { label: name, sub, isGeo: true };
  }
  if (key.startsWith("q:")) {
    const id = key.slice(2);
    return { label: `${id.slice(0, 8)}…`, sub: "trivia", isGeo: false };
  }
  return { label: `${key.slice(0, 8)}…`, sub: "", isGeo: false };
}
