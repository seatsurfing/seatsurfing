export default class Navigation {
  static PATH_SEARCH: string = "/search";

  static isAdminPath(url: string): boolean {
    return url.startsWith("/admin/");
  }

  static adminLocationDetails(locationId: string): string {
    return `/admin/locations/${locationId}`;
  }

  static adminUserDetails(userId: string): string {
    return `/admin/users/${userId}`;
  }

  static adminLocations(): string {
    return "/admin/locations/";
  }

  static adminUsers(): string {
    return "/admin/users/";
  }

  static adminBookings(query: string): string {
    return `/admin/bookings/?${query}`;
  }

  static adminGroupDetails(groupId: string): string {
    return `/admin/groups/${groupId}`;
  }

  static spaceAbsolute(locationId: string, spaceId: string): string {
    return `${window.location.origin}/ui/search/?lid=${encodeURIComponent(locationId)}&sid=${encodeURIComponent(spaceId)}`;
  }

  static locationAbsolute(locationId: string): string {
    return `${window.location.origin}/ui/search/?lid=${encodeURIComponent(locationId)}`;
  }
}
