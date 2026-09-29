// Pages that must never be rendered inside a third-party frame (clickjacking).
// The server sends "frame-ancestors 'self'" for these paths, but client-side
// navigation (e.g. router.push("/login") from an embedded page) does not
// trigger a new document request, so the UI enforces the same list itself.
// Keep in sync with nonFramableUIPaths in server/router/routes.go. Paths are
// relative to Next's basePath (/ui), as returned by router.pathname.
const nonFramablePaths = ["/", "/login"];
const nonFramablePathPrefixes = ["/resetpw", "/setpw", "/book", "/admin"];

export default class FrameProtection {
  static isNonFramablePath(pathname: string): boolean {
    const path = pathname.split(/[?#]/)[0].replace(/(.)\/+$/, "$1");
    if (nonFramablePaths.includes(path)) {
      return true;
    }
    return nonFramablePathPrefixes.some(
      (prefix) => path === prefix || path.startsWith(prefix + "/"),
    );
  }

  static isFramed(): boolean {
    if (typeof window === "undefined") {
      return false;
    }
    try {
      return window.self !== window.top;
    } catch (e) {
      // Accessing window.top can throw for cross-origin frames.
      return true;
    }
  }
}
