import type { CorsOptions } from 'cors';

/**
 * Resolve the CORS configuration from the `CORS_ORIGIN` env variable.
 *
 * Accepted shapes:
 *   - `"*"`                          → allow any origin (no credentials).
 *   - `"https://a.com,https://b.com"` → comma-separated allow-list with credentials.
 *   - unset                          → defaults to `"http://localhost:3000"`.
 *
 * Production deployments must set `CORS_ORIGIN` to the explicit list of
 * trusted origins. Avoid `"*"` outside of local development; with credentials
 * enabled, browsers reject it anyway.
 */
export function corsOptions(): CorsOptions {
  const raw = process.env.CORS_ORIGIN?.trim();
  if (!raw) {
    return { origin: 'http://localhost:3000', credentials: true };
  }
  if (raw === '*') {
    return { origin: '*' };
  }
  const list = raw.split(',').map((s) => s.trim()).filter(Boolean);
  return { origin: list, credentials: true };
}
