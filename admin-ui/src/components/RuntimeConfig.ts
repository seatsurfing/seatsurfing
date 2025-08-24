import {
  Ajax,
  AjaxCredentials,
  User,
  Settings as OrgSettings,
} from "seatsurfing-commons";

interface RuntimeUserInfos {
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
}

export default class RuntimeConfig {
  static INFOS: RuntimeUserInfos = {
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
  };

  static verifyToken = async (resolve: Function) => {
    let credentials: AjaxCredentials =
      Ajax.PERSISTER.readCredentialsFromSessionStorage();
    if (!credentials.accessToken) {
      const refreshToken: string =
        Ajax.PERSISTER.readRefreshTokenFromLocalStorage();
      if (refreshToken) {
        await Ajax.refreshAccessToken(refreshToken);
        credentials = Ajax.PERSISTER.readCredentialsFromSessionStorage();
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
      //this.setState({ isLoading: false });
    }
  };

  static loadSettings = async (): Promise<void> => {
    return new Promise<void>(function (resolve, reject) {
      OrgSettings.list().then((settings) => {
        settings.forEach((s) => {
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
        });
        resolve();
      });
    });
  };

  static loadUserAndSettings = async (): Promise<void> => {
    return User.getSelf().then((user) => {
      RuntimeConfig.INFOS.organizationId = user.organizationId;
      RuntimeConfig.INFOS.superAdmin = user.superAdmin;
      RuntimeConfig.INFOS.spaceAdmin = user.spaceAdmin;
      RuntimeConfig.INFOS.orgAdmin = user.admin;
      return RuntimeConfig.loadSettings();
    });
  };

  static getLanguage(): string {
    if (typeof window !== "undefined") {
      let curLang = window.localStorage.getItem("next-export-i18n-lang");
      if (curLang) {
        return curLang;
      }
    }
    return "en";
  }

  static getAvailableLanguages(): string[] {
    return ["en", "de", "et", "fr", "he", "hu", "it", "nl", "pl", "pt", "ro", "es"];
  }
}
