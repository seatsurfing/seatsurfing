export function isAbsoluteUrl(url: string): boolean {
  return /^https?:\/\//i.test(url);
}
