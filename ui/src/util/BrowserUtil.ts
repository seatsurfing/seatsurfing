export default class BrowserUtil {
  static LOCAL_STORAGE_KEY_SEARCH_VIEW = "searchListView";
  static LOCAL_STORAGE_KEY_MY_BOOKINGS_VIEW = "myBookingsListView";

  static tryLocalStorageSetItem(key: string, value: string): boolean {
    if (window === undefined || window.localStorage === undefined) return false;
    try {
      window.localStorage.setItem(key, value);
    } catch {
      return false;
    }
    return true;
  }

  static tryLocalStorageGetItem(key: string, defaultValue: any): any {
    if (window === undefined || window.localStorage === undefined)
      return defaultValue;
    try {
      return window.localStorage.getItem(key) ?? defaultValue;
    } catch {}
    return defaultValue;
  }
}
