export function clashSubscriptionUrl(subId: string | undefined, baseURI: string): string | null {
  if (!subId) return null;
  try {
    const url = new URL(baseURI);
    if (url.protocol !== 'http:' && url.protocol !== 'https:') return null;
    url.pathname = `${url.pathname.replace(/\/+$/, '')}/${encodeURIComponent(subId)}`;
    return url.href;
  } catch {
    return null;
  }
}
