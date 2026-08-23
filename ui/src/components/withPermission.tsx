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
  permission: PermissionKey,
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
      const allowed = RuntimeConfig.hasPermission(permission, level);
      this.setState({ checked: true, allowed });
      if (!allowed && this.props.router) {
        this.props.router.push("/admin/dashboard/");
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
