import { NextRouter } from "next/router";

export default class RedirectUtil {
  static toLogin(router: NextRouter) {
    const currentPath = router.asPath;
    router.push(`/login?redir=${encodeURIComponent(currentPath)}`);
  }
}
