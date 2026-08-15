export type ThemeMode = "light" | "dark" | "auto";

export default class Theme {
  static readonly LIGHT: ThemeMode = "light";
  static readonly DARK: ThemeMode = "dark";
  static readonly AUTO: ThemeMode = "auto";

  static readonly STORAGE_KEY = "theme";
  static readonly CHANGE_EVENT = "themechange";

  static get(): ThemeMode {
    if (typeof window === "undefined") {
      return Theme.AUTO;
    }
    const stored = window.localStorage.getItem(Theme.STORAGE_KEY);
    if (
      stored === Theme.LIGHT ||
      stored === Theme.DARK ||
      stored === Theme.AUTO
    ) {
      return stored;
    }
    return Theme.AUTO;
  }

  static set(mode: ThemeMode): void {
    if (typeof window === "undefined") {
      return;
    }
    window.localStorage.setItem(Theme.STORAGE_KEY, mode);
    Theme.apply();
  }

  static getEffective(): "light" | "dark" {
    return Theme.resolve(Theme.get());
  }

  static resolve(mode: ThemeMode): "light" | "dark" {
    if (mode === Theme.AUTO) {
      if (typeof window !== "undefined" && window.matchMedia) {
        return window.matchMedia("(prefers-color-scheme: dark)").matches
          ? "dark"
          : "light";
      }
      return "light";
    }
    return mode as "light" | "dark";
  }

  static isAdminRoute(path?: string): boolean {
    if (typeof window === "undefined" && !path) {
      return false;
    }
    const pathname = (path ?? window.location.pathname).split("?")[0];
    return /\/admin(\/|$)/.test(pathname);
  }

  static apply(path?: string): void {
    if (typeof document === "undefined") {
      return;
    }
    if (Theme.isAdminRoute(path)) {
      // Dark mode only applies to the booking UI; the admin UI always
      // stays on the default (light) theme.
      document.documentElement.removeAttribute("data-bs-theme");
      window.dispatchEvent(new CustomEvent(Theme.CHANGE_EVENT));
      return;
    }
    const effective = Theme.getEffective();
    document.documentElement.setAttribute("data-bs-theme", effective);
    window.dispatchEvent(new CustomEvent(Theme.CHANGE_EVENT));
  }

  private static systemListenerInitialized = false;

  static initSystemListener(): void {
    if (Theme.systemListenerInitialized) {
      return;
    }
    if (typeof window === "undefined" || !window.matchMedia) {
      return;
    }
    Theme.systemListenerInitialized = true;
    const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
    const listener = () => {
      if (Theme.get() === Theme.AUTO) {
        Theme.apply();
      }
    };
    if (mediaQuery.addEventListener) {
      mediaQuery.addEventListener("change", listener);
    } else if ((mediaQuery as any).addListener) {
      (mediaQuery as any).addListener(listener);
    }
  }
}

if (typeof window !== "undefined") {
  Theme.initSystemListener();
}
