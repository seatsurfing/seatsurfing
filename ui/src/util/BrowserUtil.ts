import i18n from "../../i18n";

export default class BrowserUtil {
  static LOCAL_STORAGE_KEY_SEARCH_VIEW = "searchListView";
  static LOCAL_STORAGE_KEY_MY_BOOKINGS_VIEW = "myBookingsListView";
  static LOCAL_STORAGE_KEY_SEARCH_BOOKER_NAMES = "searchBookerNames";

  static tryLocalStorageSetItem(key: string, value: string): boolean {
    if (typeof window === "undefined" || window.localStorage === undefined)
      return false;
    try {
      window.localStorage.setItem(key, value);
    } catch {
      return false;
    }
    return true;
  }

  static tryLocalStorageGetItem(key: string, defaultValue: any): any {
    if (typeof window === "undefined" || window.localStorage === undefined)
      return defaultValue;
    try {
      return window.localStorage.getItem(key) ?? defaultValue;
    } catch {}
    return defaultValue;
  }

  static applyLanguageFromQuery(notify: boolean = true) {
    if (typeof window === "undefined") return;
    try {
      const lang = new URLSearchParams(window.location.search).get("lang");
      if (!lang || !Object.hasOwn(i18n.translations, lang)) return;
      window.localStorage.setItem("next-export-i18n-lang", lang);
      if (notify) document.dispatchEvent(new Event("localStorageLangChange"));
    } catch {
      // ignore
    }
  }
}
