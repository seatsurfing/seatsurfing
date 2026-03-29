export default class Navigation {
  // API
  static readonly PATH_API_USER_PREFERENCES = "/preference/";
  static readonly PATH_API_AUTH_INIT_PW_RESET = "/auth/initpwreset";
  static readonly PATH_API_AUTH_ORG = "/auth/org/";
  static readonly PATH_API_AUTH_SINGLE_ORG = "/auth/singleorg";

  // Pages
  static readonly PATH_PAGE_SEARCH: string = "/search";

  static isAdminPath(url: string): boolean {
    return url.startsWith("/admin/");
  }

  // -----------
  // ADMIN PAGES
  // -----------

  static adminLocations(): string {
    return "/admin/locations/";
  }

  static adminLocationDetails(locationId: string): string {
    return `/admin/locations/${locationId}`;
  }

  static adminUsers(): string {
    return "/admin/users/";
  }

  static adminUserDetails(userId: string): string {
    return `/admin/users/${userId}`;
  }

  static adminBookings(query: string): string {
    return `/admin/bookings/?${query}`;
  }

  static adminGroupDetails(groupId: string): string {
    return `/admin/groups/${groupId}`;
  }

  // -------------
  // BOOKING PAGES
  // -------------

  static spaceAbsolute(locationId: string, spaceId: string): string {
    return `${window.location.origin}/ui/search/?lid=${encodeURIComponent(locationId)}&sid=${encodeURIComponent(spaceId)}`;
  }

  static locationAbsolute(locationId: string): string {
    return `${window.location.origin}/ui/search/?lid=${encodeURIComponent(locationId)}`;
  }
}
