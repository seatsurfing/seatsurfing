import CONSTANT from "@/util/Constant";
import RuntimeConfig from "@/components/RuntimeConfig";

export type PreferencesTab = "security" | "style" | "booking" | "integration";

export default class Navigation {
  // API
  static readonly PATH_API_SETTINGS = "/setting/";
  static readonly PATH_API_USER_PREFERENCES = "/preference/";
  static readonly PATH_API_AUTH_INIT_PW_RESET = "/auth/initpwreset";
  static readonly PATH_API_AUTH_ORG = "/auth/org/";
  static readonly PATH_API_AUTH_SINGLE_ORG = "/auth/singleorg";
  static readonly PATH_API_SEARCH = "/search";
  static readonly PATH_API_UC = "/uc";

  // Pages
  static readonly PATH_PAGE_SEARCH: string = "/search";
  static readonly PATH_PAGE_PREFERENCES: string = "/preferences";

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
    return `${window.location.origin}/ui${this.PATH_PAGE_SEARCH}/?lid=${encodeURIComponent(locationId)}&sid=${encodeURIComponent(spaceId)}`;
  }

  static locationAbsolute(locationId: string): string {
    return `${window.location.origin}/ui${this.PATH_PAGE_SEARCH}/?lid=${encodeURIComponent(locationId)}`;
  }

  static preferences(tab: PreferencesTab | null = null) {
    return `${this.PATH_PAGE_PREFERENCES}${tab ? `?tab=${tab}` : ""}`;
  }

  // ----------
  // KIOSK MODE
  // ----------

  static kioskUrl(spaceId: string, variant: "color" | "mono"): string {
    return `${window.location.origin}/ui/kiosk/${encodeURIComponent(spaceId)}/?variant=${encodeURIComponent(variant)}&lang=en&secret=${encodeURIComponent(CONSTANT.KIOSK_MODE_SECRET_PLACEHOLDER)}`;
  }

  // --------------
  // PUBLIC BOOKING
  // --------------

  static publicBookingUrl(): string {
    const origin = RuntimeConfig.INFOS.orgPrimaryDomain
      ? `${window.location.protocol}//${RuntimeConfig.INFOS.orgPrimaryDomain}${window.location.port ? `:${window.location.port}` : ""}`
      : window.location.origin;
    return `${origin}/ui/book/`;
  }
}
