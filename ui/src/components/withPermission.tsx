import React from "react";
import RuntimeConfig from "./RuntimeConfig";
import Loading from "./Loading";
import { PermissionKey, PermissionLevel } from "@/types/Permission";

/**
 * Page-level guard for the administration UI.
 *
 * Admin pages previously rendered for anyone and relied on the backend
 * answering 403, which left a signed-in user staring at an empty page rather
 * than being sent somewhere useful. This sends them to the dashboard instead.
 *
 * The guard is a convenience, not a security boundary: every request is
 * checked again by the server.
 */
export default function withPermission<P extends object>(
  Page: React.ComponentType<P>,
  // Omitted for pages that are not tied to one functionality - the dashboard,
  // the admin search, a plugin's own screen - where holding any
  // administrative permission at all is the right test.
  permission?: PermissionKey,
  level: number = PermissionLevel.Admin,
) {
  return class WithPermission extends React.Component<
    P & { router?: any },
    { checked: boolean; allowed: boolean }
  > {
    constructor(props: any) {
      super(props);
      this.state = { checked: false, allowed: false };
    }

    componentDidMount() {
      const allowed = permission
        ? RuntimeConfig.hasPermission(permission, level)
        : RuntimeConfig.hasAnyPermission();
      this.setState({ checked: true, allowed });
      if (!allowed && this.props.router) {
        // Somebody with no administrative access at all belongs in the
        // booking UI, not on the dashboard - which is itself guarded, so
        // sending them there would bounce them straight back here.
        this.props.router.push(
          RuntimeConfig.hasAnyPermission() ? "/admin/dashboard/" : "/search/",
        );
      }
    }

    render() {
      if (!this.state.checked || !this.state.allowed) {
        return <Loading />;
      }
      return <Page {...(this.props as P)} />;
    }
  };
}
