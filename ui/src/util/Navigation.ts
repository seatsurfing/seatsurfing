export const PATH_SEARCH: string = "/search";

export function isAdminPath(url: string): boolean {
  return url.startsWith("/admin/");
}
