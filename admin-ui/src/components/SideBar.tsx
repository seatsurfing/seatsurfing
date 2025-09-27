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
  Globe as Globe,
  Icon,
  Clock as IconApproval,
} from "react-feather";
import { Ajax, AjaxError, Booking } from "seatsurfing-commons";
import { Badge, Nav } from "react-bootstrap";
import { NextRouter } from "next/router";
import withReadyRouter from "./withReadyRouter";
import Link from "next/link";
import dynamic from "next/dynamic";
import RuntimeConfig from "./RuntimeConfig";
import { TranslationFunc, withTranslation } from "./withTranslation";
import PremiumFeatureIcon from "./PremiumFeatureIcon";

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
      this.props.router.push("/login");
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
      "/organizations",
      "/users",
      "/groups",
      "/settings",
      "/locations",
      "/bookings",
      "/approvals",
      ...RuntimeConfig.INFOS.pluginMenuItems.map((item) => {
        return "/plugin/" + item.id;
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
          <Nav.Link as={Link} eventKey="/organizations" href="/organizations">
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
            <Nav.Link as={Link} eventKey="/users" href="/users">
              <IconUsers className="feather" />{" "}
              <span className="d-none d-md-inline">
                {this.props.t("users")}
              </span>
            </Nav.Link>
          </li>
          <li className="nav-item">
            <Nav.Link
              as={Link}
              eventKey="/groups"
              href="/groups"
              disabled={
                !RuntimeConfig.INFOS.featureGroups &&
                !RuntimeConfig.INFOS.cloudHosted
              }
            >
              <IconGroups className="feather" />{" "}
              <span className="d-none d-md-inline">
                {this.props.t("groups")}
              </span>
              <PremiumFeatureIcon />
            </Nav.Link>
          </li>
          <li className="nav-item">
            <Nav.Link as={Link} eventKey="/settings" href="/settings">
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
                { ssr: true }
              ) as Icon;
              this.dynamicIcons.set(item.icon, PluginIcon);
            }
            return (
              <li className="nav-item" key={"plugin-" + item.id}>
                <Nav.Link
                  as={Link}
                  eventKey={"/plugin/" + item.id}
                  href={"/plugin/" + item.id}
                >
                  <PluginIcon className="feather" /> {item.title}
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
              <Nav.Link as={Link} eventKey="/dashboard" href="/dashboard">
                <IconHome className="feather" />{" "}
                <span className="d-none d-md-inline">
                  {this.props.t("dashboard")}
                </span>
              </Nav.Link>
            </li>
            <li className="nav-item">
              <Nav.Link as={Link} eventKey="/locations" href="/locations">
                <IconMap className="feather" />{" "}
                <span className="d-none d-md-inline">
                  {this.props.t("areas")}
                </span>
              </Nav.Link>
            </li>
            <li className="nav-item">
              <Nav.Link as={Link} eventKey="/bookings" href="/bookings">
                <IconBook className="feather" />{" "}
                <span className="d-none d-md-inline">
                  {this.props.t("bookings")}
                </span>
              </Nav.Link>
            </li>
            <li className="nav-item">
              <Nav.Link
                as={Link}
                eventKey="/approvals"
                href="/approvals"
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
                <PremiumFeatureIcon />
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
                eventKey="/report/analysis"
                href="/report/analysis"
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
                  { ssr: true }
                ) as Icon;
                this.dynamicIcons.set(item.icon, PluginIcon);
              }
              return (
                <li className="nav-item" key={"plugin-" + item.id}>
                  <Nav.Link
                    as={Link}
                    eventKey={"/plugin/" + item.id}
                    href={"/plugin/" + item.id}
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
              <Nav.Link
                onClick={(e) => {
                  e.preventDefault();
                  window.location.href = "/ui/";
                }}
              >
                <Globe className="feather" />{" "}
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
