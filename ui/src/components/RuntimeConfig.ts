import User from "@/types/User";
import OrgSettings from "@/types/Settings";
import Ajax from "@/util/Ajax";
import UserPreference from "@/types/UserPreference";

interface RuntimeUserInfos {
  username: string;
  userId: string;
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
  disableBuddies: boolean;
  maxHoursPartiallyBooked: number;
  maxHoursPartiallyBookedEnabled: boolean;
  featureRecurringBookings: boolean;
  organizationId: string;
  superAdmin: boolean;
  spaceAdmin: boolean;
  orgAdmin: boolean;
  pluginMenuItems: any[];
  pluginWelcomeScreens: any[];
  featureGroups: boolean;
  featureAuthProviders: boolean;
  cloudHosted: boolean;
  subscriptionActive: boolean;
  orgPrimaryDomain: string;
  disablePasswordLogin: boolean;
  allowRecurringBookings: boolean;
  subjectDefault: number;
  use24HourTime: boolean;
  dateFormat: string;
  totpEnabled: boolean;
  enforceTOTP: boolean;
}

export default class RuntimeConfig {
  static EMBEDDED: boolean = false;
  static INFOS: RuntimeUserInfos;

  static resetInfos = () => {
    RuntimeConfig.INFOS = {
      username: "",
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
      featureRecurringBookings: false,
      organizationId: "",
      superAdmin: false,
      spaceAdmin: false,
      orgAdmin: false,
      pluginMenuItems: [],
      pluginWelcomeScreens: [],
      featureGroups: false,
      featureAuthProviders: false,
      cloudHosted: false,
      subscriptionActive: false,
      orgPrimaryDomain: "",
      disablePasswordLogin: false,
      allowRecurringBookings: true,
      subjectDefault: 2,
      use24HourTime: true,
      dateFormat: "Y-m-d",
      totpEnabled: false,
      enforceTOTP: false,
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
      RuntimeConfig.loadUserAndSettings()
        .then(() => {
          resolve();
        })
        .catch((e) => {
          Ajax.PERSISTER.deleteCredentialsFromStorage();
          resolve();
        });
    } else {
      resolve();
    }
  };

  static loadSettings = async (): Promise<void> => {
    return new Promise<void>(function (resolve, reject) {
      OrgSettings.list().then((settings) => {
        settings.forEach((s) => {
          if (typeof window !== "undefined") {
            if (s.name === "max_bookings_per_user")
              RuntimeConfig.INFOS.maxBookingsPerUser = window.parseInt(s.value);
            if (s.name === "max_concurrent_bookings_per_user")
              RuntimeConfig.INFOS.maxConcurrentBookingsPerUser =
                window.parseInt(s.value);
            if (s.name === "max_days_in_advance")
              RuntimeConfig.INFOS.maxDaysInAdvance = window.parseInt(s.value);
            if (s.name === "max_booking_duration_hours")
              RuntimeConfig.INFOS.maxBookingDurationHours = window.parseInt(
                s.value,
              );
            if (s.name === "max_hours_before_delete")
              RuntimeConfig.INFOS.maxHoursBeforeDelete = window.parseInt(
                s.value,
              );
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
            RuntimeConfig.INFOS.maxHoursPartiallyBookedEnabled =
              s.value === "1";
          if (s.name === "show_names")
            RuntimeConfig.INFOS.showNames = s.value === "1";
          if (s.name === "disable_buddies")
            RuntimeConfig.INFOS.disableBuddies = s.value === "1";
          if (s.name === "custom_logo_url")
            RuntimeConfig.INFOS.customLogoUrl = s.value;
          if (s.name === "default_timezone")
            RuntimeConfig.INFOS.defaultTimezone = s.value;
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
            RuntimeConfig.INFOS.featureGroups = s.value
              ? JSON.parse(s.value)
              : [];
          if (s.name === "feature_auth_providers")
            RuntimeConfig.INFOS.featureAuthProviders = s.value
              ? JSON.parse(s.value)
              : [];
          if (s.name === "cloud_hosted")
            RuntimeConfig.INFOS.cloudHosted = s.value
              ? JSON.parse(s.value)
              : [];
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
        });
        resolve();
      });
    });
  };

  static loadUserPreferences = async (): Promise<void> => {
    UserPreference.list()
      .then((list) => {
        list.forEach((pref) => {
          if (pref.name === "use_24_hour_time") {
            RuntimeConfig.INFOS.use24HourTime = pref.value === "1";
          }
          if (pref.name === "date_format") {
            RuntimeConfig.INFOS.dateFormat = pref.value;
          }
        });
      })
      .catch(() => {
        // Nothing to do
      });
  };

  static setDetails = (username: string, id: string) => {
    RuntimeConfig.loadSettings().then(() => {
      RuntimeConfig.INFOS.username = username;
      RuntimeConfig.INFOS.userId = id;
    });
  };

  static async setLoginDetails(): Promise<void> {
    return User.getSelf().then((user) => {
      RuntimeConfig.INFOS.idpLogin = !user.requirePassword;
      RuntimeConfig.setDetails(user.email, user.id);
    });
  }

  static loadUserAndSettings = async (): Promise<void> => {
    RuntimeConfig.resetInfos();
    return User.getSelf().then((user) => {
      RuntimeConfig.INFOS.organizationId = user.organizationId;
      RuntimeConfig.INFOS.superAdmin = user.superAdmin;
      RuntimeConfig.INFOS.spaceAdmin = user.spaceAdmin;
      RuntimeConfig.INFOS.orgAdmin = user.admin;
      RuntimeConfig.INFOS.idpLogin = !user.requirePassword;
      RuntimeConfig.INFOS.totpEnabled = user.totpEnabled;
      RuntimeConfig.setDetails(user.email, user.id);
      return RuntimeConfig.loadSettings().then(() => {
        return RuntimeConfig.loadUserPreferences();
      });
    });
  };

  static getLanguage(): string {
    if (typeof window !== "undefined") {
      let curLang = window.localStorage.getItem("next-export-i18n-lang");
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
      "fr",
      "he",
      "hu",
      "it",
      "nl",
      "pl",
      "pt",
      "ro",
      "es",
    ];
  }

  static logOut(): void {
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
    Ajax.get("/auth/logout/current")
      .then(() => proceed())
      .catch(() => proceed());
  }
}

RuntimeConfig.resetInfos();
