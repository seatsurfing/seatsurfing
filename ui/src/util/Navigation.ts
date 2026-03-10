export const PATH_SEARCH: string = "/search";

export function isAdminPath(url: string): boolean {
  return url.startsWith("/admin/");
}

export function adminLocationDetails(locationId: string): string {
  return `/admin/locations/${locationId}`;
}

export function adminUserDetails(userId: string): string {
  return `/admin/users/${userId}`;
}

export function adminLocations(): string {
  return "/admin/locations/";
}

export function adminUsers(): string {
  return "/admin/users/";
}

export function adminBookings(query: string): string {
  return `/admin/bookings/?${query}`;
}

export function adminGroupDetails(groupId: string): string {
  return `/admin/groups/${groupId}`;
}
