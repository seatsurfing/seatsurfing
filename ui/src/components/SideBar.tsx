import React from "react";
import {
  Home as IconHome,
  User as IconUsers,
  Users as IconGroups,
  Map as IconMap,
  Book as IconBook,
  Settings as IconSettings,
  Box as IconBox,
  Activity as IconAnalysis,
  Clipboard as IconClipboard,
  Icon,
  Clock as IconApproval,
} from "react-feather";
import { Badge, Nav } from "react-bootstrap";
import { NextRouter } from "next/router";
import withReadyRouter from "./withReadyRouter";
import Link from "next/link";
import dynamic from "next/dynamic";
import RuntimeConfig from "./RuntimeConfig";
import { TranslationFunc, withTranslation } from "./withTranslation";
import PremiumFeatureIcon from "./PremiumFeatureIcon";
import Ajax from "@/util/Ajax";
import Booking from "@/types/Booking";
import AjaxError from "@/util/AjaxError";
import RedirectUtil from "@/util/RedirectUtil";

interface State {
  approvalCount: number;
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
}

class SideBar extends React.Component<Props, State> {
  dynamicIcons: Map<string, any> = new Map();

  constructor(props: any) {
    super(props);
    this.state = {
      approvalCount: 0,
    };
  }

  componentDidMount = () => {
    if (!RuntimeConfig.INFOS.spaceAdmin) {
      Ajax.PERSISTER.deleteCredentialsFromStorage();
      RedirectUtil.toLogin(this.props.router);
      return;
    }
    Booking.getPendingApprovalsCount()
      .then((count) => {
        this.setState({ approvalCount: count });
      })
      .catch((error) => {
        console.error("Error fetching pending approvals count:", error);
      });
    this.updateApprovalCount();
  };

  updateApprovalCount = () => {
    if (!Ajax.hasAccessToken()) {
      // Do nothing if we don't have an access token
      return;
    }
    window.setTimeout(() => {
      Booking.getPendingApprovalsCount()
        .then((count) => {
          // Successfully fetched pending approvals count, update state & continue polling
          this.setState({ approvalCount: count });
          this.updateApprovalCount();
        })
        .catch((error) => {
          if (error instanceof AjaxError && error.httpStatusCode === 401) {
            // Not authenticated anymore, stop polling
            return;
          }
          // Some other error occurred, try again in next interval
          console.error("Error fetching pending approvals count:", error);
          this.updateApprovalCount();
        });
    }, 30 * 1000); // Poll every 30 seconds
  };

  getActiveKey = () => {
    let path = this.props.router.pathname;
    if (path.startsWith("/plugin/")) {
      path = window.location.pathname.replace("/admin", "");
    }
    let startPaths = [
      "/admin/organizations",
      "/admin/users",
      "/admin/groups",
      "/admin/settings",
      "/admin/locations",
      "/admin/bookings",
      "/admin/approvals",
      ...RuntimeConfig.INFOS.pluginMenuItems.map((item) => {
        return "/admin/plugin/" + item.id;
      }),
    ];
    let result = path;
    startPaths.forEach((startPath) => {
      if (path.startsWith(startPath)) {
        result = startPath;
      }
    });
    return result;
  };

  render() {
    let orgItem = <></>;
    if (RuntimeConfig.INFOS.superAdmin) {
      orgItem = (
        <li className="nav-item">
          <Nav.Link
            as={Link}
            eventKey="/admin/organizations"
            href="/admin/organizations"
          >
            <IconBox className="feather" />{" "}
            <span className="d-none d-md-inline">
              {this.props.t("organizations")}
            </span>
          </Nav.Link>
        </li>
      );
    }
    let orgAdminItems = <></>;
    if (RuntimeConfig.INFOS.orgAdmin) {
      orgAdminItems = (
        <>
          <li className="nav-item">
            <Nav.Link as={Link} eventKey="/admin/users" href="/admin/users">
              <IconUsers className="feather" />{" "}
              <span className="d-none d-md-inline">
                {this.props.t("users")}
              </span>
            </Nav.Link>
          </li>
          <li className="nav-item">
            <Nav.Link
              as={Link}
              eventKey="/admin/groups"
              href="/admin/groups"
              disabled={
                !RuntimeConfig.INFOS.featureGroups &&
                !RuntimeConfig.INFOS.cloudHosted
              }
            >
              <IconGroups className="feather" />{" "}
              <span className="d-none d-md-inline">
                {this.props.t("groups")}
              </span>
              <PremiumFeatureIcon className="d-none d-md-inline" />
            </Nav.Link>
          </li>
          <li className="nav-item">
            <Nav.Link
              as={Link}
              eventKey="/admin/settings"
              href="/admin/settings"
            >
              <IconSettings className="feather" />{" "}
              <span className="d-none d-md-inline">
                {this.props.t("settings")}
              </span>
            </Nav.Link>
          </li>
          {RuntimeConfig.INFOS.pluginMenuItems.map((item) => {
            if (item.visibility !== "admin") {
              return;
            }
            let PluginIcon = this.dynamicIcons.get(item.icon);
            if (!PluginIcon) {
              PluginIcon = dynamic(
                () =>
                  import("react-feather/dist/icons/" + item.icon.toLowerCase()),
                { ssr: true },
              ) as Icon;
              this.dynamicIcons.set(item.icon, PluginIcon);
            }
            return (
              <li className="nav-item" key={"plugin-" + item.id}>
                <Nav.Link
                  as={Link}
                  eventKey={"/admin/plugin/" + item.id}
                  href={"/admin/plugin/" + item.id}
                >
                  <PluginIcon className="feather" />{" "}
                  <span className="d-none d-md-inline">{item.title}</span>
                </Nav.Link>
              </li>
            );
          })}
        </>
      );
    }
    return (
      <Nav
        id="sidebarMenu"
        className="col-1 col-md-3 col-lg-2 d-md-block bg-light sidebar"
        activeKey={this.getActiveKey()}
      >
        <div className="sidebar-sticky pt-3">
          <ul className="nav flex-column">
            <li className="nav-item">
              <Nav.Link
                as={Link}
                eventKey="/admin/dashboard"
                href="/admin/dashboard"
              >
                <IconClipboard className="feather" />{" "}
                <span className="d-none d-md-inline">
                  {this.props.t("dashboard")}
                </span>
              </Nav.Link>
            </li>
            <li className="nav-item">
              <Nav.Link
                as={Link}
                eventKey="/admin/locations"
                href="/admin/locations"
              >
                <IconMap className="feather" />{" "}
                <span className="d-none d-md-inline">
                  {this.props.t("areas")}
                </span>
              </Nav.Link>
            </li>
            <li className="nav-item">
              <Nav.Link
                as={Link}
                eventKey="/admin/bookings"
                href="/admin/bookings"
              >
                <IconBook className="feather" />{" "}
                <span className="d-none d-md-inline">
                  {this.props.t("bookings")}
                </span>
              </Nav.Link>
            </li>
            <li className="nav-item">
              <Nav.Link
                as={Link}
                eventKey="/admin/approvals"
                href="/admin/approvals"
                disabled={
                  !RuntimeConfig.INFOS.featureGroups &&
                  !RuntimeConfig.INFOS.cloudHosted
                }
              >
                <IconApproval className="feather" />
                <span className="d-none d-md-inline">
                  {" "}
                  {this.props.t("approvals")}
                </span>
                <PremiumFeatureIcon className="d-none d-md-inline" />
                <Badge
                  bg="primary"
                  hidden={this.state.approvalCount === 0}
                  style={{ marginLeft: "5px" }}
                >
                  {this.state.approvalCount}
                </Badge>
              </Nav.Link>
            </li>
            <li className="nav-item">
              <Nav.Link
                as={Link}
                eventKey="/admin/report/analysis"
                href="/admin/report/analysis"
              >
                <IconAnalysis className="feather" />
                <span className="d-none d-md-inline">
                  {" "}
                  {this.props.t("analysis")}
                </span>
              </Nav.Link>
            </li>
            {RuntimeConfig.INFOS.pluginMenuItems.map((item) => {
              if (item.visibility !== "spaceadmin") {
                return;
              }
              let PluginIcon = this.dynamicIcons.get(item.icon);
              if (!PluginIcon) {
                PluginIcon = dynamic(
                  () =>
                    import(
                      "react-feather/dist/icons/" + item.icon.toLowerCase()
                    ),
                  { ssr: true },
                ) as Icon;
                this.dynamicIcons.set(item.icon, PluginIcon);
              }
              return (
                <li className="nav-item" key={"plugin-" + item.id}>
                  <Nav.Link
                    as={Link}
                    eventKey={"/admin/plugin/" + item.id}
                    href={"/admin/plugin/" + item.id}
                  >
                    <PluginIcon className="feather" />{" "}
                    <span className="d-none d-md-inline">{item.title}</span>
                  </Nav.Link>
                </li>
              );
            })}
            {orgAdminItems}
            {orgItem}
            <li className="nav-item">
              <Nav.Link as={Link} href="/search/">
                <IconHome className="feather" />{" "}
                <span className="d-none d-md-inline">
                  {this.props.t("bookingui")}
                </span>
              </Nav.Link>
            </li>
          </ul>
        </div>
      </Nav>
    );
  }
}

export default withTranslation(withReadyRouter(SideBar as any));
