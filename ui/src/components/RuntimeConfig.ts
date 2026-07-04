import User from "@/types/User";
import OrgSettings from "@/types/Settings";
import Ajax from "@/util/Ajax";
import UserPreference from "@/types/UserPreference";

interface RuntimeUserInfos {
  username: string;
  userId: string;
  firstname: string;
  lastname: string;
  idpLogin: boolean;
  isLoading: boolean;
  maxBookingsPerUser: number;
  maxConcurrentBookingsPerUser: number;
  maxDaysInAdvance: number;
  maxBookingDurationHours: number;
  maxHoursBeforeDelete: number;
  minBookingDurationHours: number;
  dailyBasisBooking: boolean;
  noAdminRestrictions: boolean;
  showNames: boolean;
  customLogoUrl: string;
  defaultTimezone: string;
  orgLanguage: string;
  disableBuddies: boolean;
  maxHoursPartiallyBooked: number;
  maxHoursPartiallyBookedEnabled: boolean;
  featureRecurringBookings: boolean;
  organizationId: string;
  orgName: string;
  superAdmin: boolean;
  spaceAdmin: boolean;
  orgAdmin: boolean;
  pluginMenuItems: any[];
  pluginWelcomeScreens: any[];
  featureGroups: boolean;
  featureAuthProviders: boolean;
  featureKioskMode: boolean;
  kioskModeEnabled: boolean;
  cloudHosted: boolean;
  subscriptionActive: boolean;
  orgPrimaryDomain: string;
  disablePasswordLogin: boolean;
  allowRecurringBookings: boolean;
  subjectDefault: number;
  use24HourTime: boolean;
  dateFormat: string;
  weekStartDay: number;
  totpEnabled: boolean;
  enforceTOTP: boolean;
  hideReports: boolean;
  hideStats: boolean;
  hasPasskeys: boolean;
  isPrimaryDomain: boolean;
  targetUtilizationHoursPerWeek: number;
}

export default class RuntimeConfig {
  static EMBEDDED: boolean = false;
  static INFOS: RuntimeUserInfos;

  static resetInfos = () => {
    RuntimeConfig.INFOS = {
      username: "",
      firstname: "",
      lastname: "",
      userId: "",
      idpLogin: false,
      isLoading: true,
      maxBookingsPerUser: 0,
      maxConcurrentBookingsPerUser: 0,
      maxDaysInAdvance: 0,
      maxBookingDurationHours: 0,
      maxHoursBeforeDelete: 0,
      minBookingDurationHours: 0,
      dailyBasisBooking: false,
      noAdminRestrictions: false,
      disableBuddies: false,
      customLogoUrl: "",
      maxHoursPartiallyBooked: 0,
      maxHoursPartiallyBookedEnabled: false,
      showNames: false,
      defaultTimezone: "",
      orgLanguage: "",
      featureRecurringBookings: false,
      organizationId: "",
      orgName: "",
      superAdmin: false,
      spaceAdmin: false,
      orgAdmin: false,
      pluginMenuItems: [],
      pluginWelcomeScreens: [],
      featureGroups: false,
      featureAuthProviders: false,
      featureKioskMode: false,
      kioskModeEnabled: false,
      cloudHosted: false,
      subscriptionActive: false,
      orgPrimaryDomain: "",
      disablePasswordLogin: false,
      allowRecurringBookings: true,
      subjectDefault: 2,
      use24HourTime: true,
      dateFormat: "Y-m-d",
      weekStartDay: 1,
      totpEnabled: false,
      enforceTOTP: false,
      hideReports: false,
      hideStats: false,
      hasPasskeys: false,
      isPrimaryDomain: false,
      targetUtilizationHoursPerWeek: 0,
    };
  };

  static verifyToken = async (resolve: Function) => {
    let credentials = Ajax.PERSISTER.readCredentialsFromLocalStorage();
    if (!credentials.accessToken) {
      const refreshToken = Ajax.PERSISTER.readRefreshTokenFromLocalStorage();
      if (refreshToken) {
        await Ajax.refreshAccessToken(refreshToken);
        credentials = Ajax.PERSISTER.readCredentialsFromLocalStorage();
      }
    }
    if (credentials.accessToken) {
      try {
        await RuntimeConfig.loadUserAndSettings();
      } catch {
        Ajax.PERSISTER.deleteCredentialsFromStorage();
      }
    }
    resolve();
  };

  static loadSettings = async (): Promise<void> => {
    const settings = await OrgSettings.list();
    settings.forEach((s) => {
      if (typeof window !== "undefined") {
        if (s.name === "max_bookings_per_user")
          RuntimeConfig.INFOS.maxBookingsPerUser = window.parseInt(s.value);
        if (s.name === "max_concurrent_bookings_per_user")
          RuntimeConfig.INFOS.maxConcurrentBookingsPerUser = window.parseInt(
            s.value,
          );
        if (s.name === "max_days_in_advance")
          RuntimeConfig.INFOS.maxDaysInAdvance = window.parseInt(s.value);
        if (s.name === "max_booking_duration_hours")
          RuntimeConfig.INFOS.maxBookingDurationHours = window.parseInt(
            s.value,
          );
        if (s.name === "max_hours_before_delete")
          RuntimeConfig.INFOS.maxHoursBeforeDelete = window.parseInt(s.value);
        if (s.name === "max_hours_partially_booked")
          RuntimeConfig.INFOS.maxHoursPartiallyBooked = window.parseInt(
            s.value,
          );
        if (s.name === "min_booking_duration_hours")
          RuntimeConfig.INFOS.minBookingDurationHours = window.parseInt(
            s.value,
          );
      }
      if (s.name === "daily_basis_booking")
        RuntimeConfig.INFOS.dailyBasisBooking = s.value === "1";
      if (s.name === "no_admin_restrictions")
        RuntimeConfig.INFOS.noAdminRestrictions = s.value === "1";
      if (s.name === "max_hours_partially_booked_enabled")
        RuntimeConfig.INFOS.maxHoursPartiallyBookedEnabled = s.value === "1";
      if (s.name === "show_names")
        RuntimeConfig.INFOS.showNames = s.value === "1";
      if (s.name === "disable_buddies")
        RuntimeConfig.INFOS.disableBuddies = s.value === "1";
      if (s.name === "custom_logo_url")
        RuntimeConfig.INFOS.customLogoUrl = s.value;
      if (s.name === "default_timezone")
        RuntimeConfig.INFOS.defaultTimezone = s.value;
      if (s.name === "_sys_org_language")
        RuntimeConfig.INFOS.orgLanguage = s.value;
      if (s.name === "feature_recurring_bookings")
        RuntimeConfig.INFOS.featureRecurringBookings = s.value === "1";
      if (s.name === "allow_recurring_bookings")
        RuntimeConfig.INFOS.allowRecurringBookings = s.value === "1";
      if (s.name === "_sys_admin_menu_items")
        RuntimeConfig.INFOS.pluginMenuItems = s.value
          ? JSON.parse(s.value)
          : [];
      if (s.name === "_sys_admin_welcome_screens")
        RuntimeConfig.INFOS.pluginWelcomeScreens = s.value
          ? JSON.parse(s.value)
          : [];
      if (s.name === "feature_groups")
        RuntimeConfig.INFOS.featureGroups = s.value ? JSON.parse(s.value) : [];
      if (s.name === "feature_auth_providers")
        RuntimeConfig.INFOS.featureAuthProviders = s.value
          ? JSON.parse(s.value)
          : [];
      if (s.name === "feature_kiosk_mode")
        RuntimeConfig.INFOS.featureKioskMode = s.value === "1";
      if (s.name === "kiosk_mode_enabled")
        RuntimeConfig.INFOS.kioskModeEnabled = s.value === "1";
      if (s.name === "cloud_hosted")
        RuntimeConfig.INFOS.cloudHosted = s.value ? JSON.parse(s.value) : [];
      if (s.name === "subscription_active")
        RuntimeConfig.INFOS.subscriptionActive = s.value
          ? JSON.parse(s.value)
          : [];
      if (s.name === "_sys_org_primary_domain")
        RuntimeConfig.INFOS.orgPrimaryDomain = s.value;
      if (s.name === "_sys_disable_password_login")
        RuntimeConfig.INFOS.disablePasswordLogin = s.value === "1";
      if (s.name === "subject_default")
        RuntimeConfig.INFOS.subjectDefault = window.parseInt(s.value);
      if (s.name === "enforce_totp")
        RuntimeConfig.INFOS.enforceTOTP = s.value === "1";
      if (s.name === "hide_reports")
        RuntimeConfig.INFOS.hideReports = s.value === "1";
      if (s.name === "hide_stats")
        RuntimeConfig.INFOS.hideStats = s.value === "1";
      if (s.name === "target_utilization_hours_per_week")
        RuntimeConfig.INFOS.targetUtilizationHoursPerWeek = window.parseInt(
          s.value,
        );
    });
  };

  static loadUserPreferences = async (): Promise<void> => {
    try {
      const list = await UserPreference.list();
      list.forEach((pref) => {
        if (pref.name === UserPreference.PREF_USE_24_HOUR_TIME) {
          RuntimeConfig.INFOS.use24HourTime = pref.value === "1";
        }
        if (pref.name === UserPreference.PREF_DATE_FORMAT) {
          RuntimeConfig.INFOS.dateFormat = pref.value;
        }
        if (pref.name === UserPreference.PREF_WEEK_START_DAY) {
          const v = parseInt(pref.value);
          RuntimeConfig.INFOS.weekStartDay = [0, 1, 6].includes(v) ? v : 1;
        }
      });
    } catch {
      // Nothing to do
    }
  };

  static setUserDetails = async (user: User) => {
    await RuntimeConfig.loadSettings();
    RuntimeConfig.INFOS.username = user.email;
    RuntimeConfig.INFOS.userId = user.id;
    RuntimeConfig.INFOS.firstname = user.firstname;
    RuntimeConfig.INFOS.lastname = user.lastname;
  };

  static async setLoginDetails(): Promise<void> {
    const user = await User.getSelf();
    RuntimeConfig.INFOS.idpLogin = !user.requirePassword;
    RuntimeConfig.setUserDetails(user);
  }

  static loadUserAndSettings = async (): Promise<void> => {
    RuntimeConfig.resetInfos();
    const user = await User.getSelf();
    RuntimeConfig.INFOS.organizationId = user.organizationId;
    RuntimeConfig.INFOS.superAdmin = user.superAdmin;
    RuntimeConfig.INFOS.spaceAdmin = user.spaceAdmin;
    RuntimeConfig.INFOS.orgAdmin = user.admin;
    RuntimeConfig.INFOS.idpLogin = !user.requirePassword;
    RuntimeConfig.INFOS.totpEnabled = user.totpEnabled;
    RuntimeConfig.INFOS.hasPasskeys = user.hasPasskeys;
    RuntimeConfig.INFOS.isPrimaryDomain = user.isPrimaryDomain;
    RuntimeConfig.setUserDetails(user);
    RuntimeConfig.INFOS.orgName = user.organization.name;
    await RuntimeConfig.loadSettings();
    await RuntimeConfig.loadUserPreferences();
  };

  static getLanguage(): string {
    if (typeof window !== "undefined") {
      const curLang = window.localStorage.getItem("next-export-i18n-lang");
      if (curLang) {
        return curLang;
      }
    }
    return "en-GB";
  }

  static getAvailableLanguages(): string[] {
    return [
      "en-GB",
      "en-US",
      "de",
      "et",
      "fi",
      "fr",
      "he",
      "hu",
      "it",
      "nl",
      "pl",
      "pt",
      "ro",
      "es",
      "zh-TW",
    ];
  }

  static async logOut(): Promise<void> {
    const credentials = Ajax.PERSISTER.readCredentialsFromLocalStorage();
    const logoutUrl = credentials.logoutUrl;
    const proceed = () => {
      Ajax.PERSISTER.deleteCredentialsFromStorage();
      RuntimeConfig.resetInfos();
      if (logoutUrl) {
        window.location.href = logoutUrl;
        return;
      }
      window.location.href = "/ui/login?noredirect=1";
    };
    try {
      await Ajax.get("/auth/logout/current");
    } finally {
      proceed();
    }
  }
}

RuntimeConfig.resetInfos();
