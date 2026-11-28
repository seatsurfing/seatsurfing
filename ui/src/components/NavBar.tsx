import React from "react";
import dynamic from "next/dynamic";
import { Navbar, Nav, Container, NavLink } from "react-bootstrap";
import RuntimeConfig from "./RuntimeConfig";
import {
  Settings as IconSettings,
  Calendar as IconCalendar,
  PlusSquare as IconPlus,
  User as IconUser,
  Heart as IconBuddies,
  Shield as IconAdmin,
  Grid as IconIntegration,
} from "react-feather";
import { NextRouter } from "next/router";
import withReadyRouter from "./withReadyRouter";
import Link from "next/link";
import { TranslationFunc, withTranslation } from "./withTranslation";
import User from "@/types/User";
import Ajax from "@/util/Ajax";
import LanguageSelector from "./LanguageSelector";
import ThemeSelector from "./ThemeSelector";
import RendererUtils from "@/util/RendererUtils";
import PreferencesModal from "./PreferencesModal";
import Navigation, { PreferencesTab } from "@/util/Navigation";

interface State {
  allowAdmin: boolean;
  showPreferences: boolean;
  preferencesTab: PreferencesTab | null;
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
  activeIntegrationId?: string | null;
  onToggleIntegration?: (id: string) => void;
}

class NavBar extends React.Component<Props, State> {
  dynamicIcons: Map<string, any> = new Map();

  constructor(props: any) {
    super(props);
    this.state = {
      allowAdmin: false,
      showPreferences: false,
      preferencesTab: null,
    };
  }

  componentDidMount = () => {
    if (!Ajax.hasAccessToken()) {
      return;
    }
    this.loadData();
    this.openPreferencesFromQuery();
  };

  componentDidUpdate = (prevProps: Props) => {
    if (
      prevProps.router.query[Navigation.QUERY_PREFERENCES] !==
      this.props.router.query[Navigation.QUERY_PREFERENCES]
    ) {
      this.openPreferencesFromQuery();
    }
  };

  openPreferencesFromQuery = () => {
    const param = this.props.router.query[Navigation.QUERY_PREFERENCES];
    if (param === undefined) {
      return;
    }
    const tab = typeof param === "string" && param ? param : null;
    this.setState({
      showPreferences: true,
      preferencesTab: tab as PreferencesTab | null,
    });
    const query = { ...this.props.router.query };
    delete query[Navigation.QUERY_PREFERENCES];
    this.props.router.replace(
      { pathname: this.props.router.pathname, query },
      undefined,
      {
        shallow: true,
      },
    );
  };

  openPreferences = (e: React.MouseEvent) => {
    e.preventDefault();
    this.setState({ showPreferences: true, preferencesTab: null });
  };

  loadData = () => {
    User.getSelf().then((user) => {
      if (Object.values(user.permissions).some((level) => level > 0)) {
        this.setState({ allowAdmin: true });
      }
    });
  };

  logOut = (e: any) => {
    e.preventDefault();
    RuntimeConfig.logOut();
  };

  render() {
    let signOffButton = <></>;
    let adminButton = <></>;
    let collapsable = <></>;
    let buddies = <></>;

    if (!RuntimeConfig.EMBEDDED) {
      if (this.state.allowAdmin) {
        adminButton = (
          <Nav.Link
            as={Link}
            eventKey="/admin/dashboard"
            href="/admin/dashboard"
          >
            <IconAdmin className="feather" /> {this.props.t("administration")}
          </Nav.Link>
        );
      }
      signOffButton = (
        <Nav.Link onClick={this.logOut}>{this.props.t("logout")}</Nav.Link>
      );
    }

    if (RuntimeConfig.INFOS.showNames && !RuntimeConfig.INFOS.disableBuddies) {
      buddies = (
        <Nav.Link as={Link} eventKey="/buddies" href="/buddies">
          {RuntimeConfig.EMBEDDED ? (
            <IconBuddies className="feather feather-lg" />
          ) : (
            <>
              <IconBuddies className="feather" /> {this.props.t("myBuddies")}
            </>
          )}
        </Nav.Link>
      );
    }

    const bookingUIIntegrationItems = RuntimeConfig.INFOS.bookingUIIntegrations
      .filter((item) => RuntimeConfig.canSeeBookingUIIntegration(item))
      .map((item) => {
        let PluginIcon = this.dynamicIcons.get(item.icon);
        if (!PluginIcon) {
          PluginIcon = item.icon
            ? dynamic(
                () =>
                  import("react-feather/dist/icons/" + item.icon.toLowerCase()),
                { ssr: true },
              )
            : IconIntegration;
          this.dynamicIcons.set(item.icon, PluginIcon);
        }
        const active = this.props.activeIntegrationId === item.id;
        const title = RuntimeConfig.pickBookingUIIntegrationTitle(item);
        return (
          <Nav.Link
            key={"integration-" + item.id}
            as={Link}
            href={"/search?openIntegration=" + encodeURIComponent(item.id)}
            active={active}
            onClick={(e: React.MouseEvent) => {
              // Already on the booking page: toggle the panel in place
              // instead of navigating (which would just reload it).
              if (this.props.onToggleIntegration) {
                e.preventDefault();
                this.props.onToggleIntegration(item.id);
              }
              // Otherwise, let the Link navigate to /search, which opens
              // the integration itself once it has mounted there.
            }}
          >
            {RuntimeConfig.EMBEDDED ? (
              <PluginIcon className="feather feather-lg" />
            ) : (
              <>
                <PluginIcon className="feather" /> {title}
              </>
            )}
          </Nav.Link>
        );
      });

    collapsable = (
      <>
        <Nav activeKey={this.props.router.pathname}>
          <Nav.Link as={Link} eventKey="/search" href="/search">
            {RuntimeConfig.EMBEDDED ? (
              <IconPlus className="feather feather-lg" />
            ) : (
              <>
                <IconPlus className="feather" /> {this.props.t("bookSeat")}
              </>
            )}
          </Nav.Link>
          <Nav.Link as={Link} eventKey="/bookings" href="/bookings">
            {RuntimeConfig.EMBEDDED ? (
              <IconCalendar className="feather feather-lg" />
            ) : (
              <>
                <IconCalendar className="feather" />{" "}
                {this.props.t("myBookings")}
              </>
            )}
          </Nav.Link>
          {buddies}
          {bookingUIIntegrationItems}
          {adminButton}
        </Nav>
        <Nav className="ms-auto">
          <Nav.Link as="span" className="icon-link d-none d-xl-flex pe-none">
            <IconUser className="feather feather-lg" />
            <span
              title={RuntimeConfig.INFOS.username}
              style={{ pointerEvents: "auto" }}
            >
              {RendererUtils.fullname(
                RuntimeConfig.INFOS.firstname,
                RuntimeConfig.INFOS.lastname,
                RuntimeConfig.INFOS.username,
              )}
            </span>
          </Nav.Link>
          <Nav.Link
            href="#"
            onClick={this.openPreferences}
            className="d-flex align-items-center nav-icon-link"
            title={this.props.t("preferences")}
            aria-label={this.props.t("preferences")}
          >
            <IconSettings className="feather" />
            {!RuntimeConfig.EMBEDDED && (
              <span className="d-lg-none ms-2">
                {this.props.t("preferences")}
              </span>
            )}
          </Nav.Link>
          <ThemeSelector inNavbar={true} compactBreakpoint="lg" align="end" />
          <LanguageSelector
            inNavbar={true}
            compactBreakpoint="lg"
            align="end"
          />
          {signOffButton}
        </Nav>
      </>
    );

    if (!RuntimeConfig.EMBEDDED) {
      collapsable = (
        <>
          <Navbar.Toggle aria-controls="basic-navbar-nav" />
          <Navbar.Collapse id="basic-navbar-nav">{collapsable}</Navbar.Collapse>
        </>
      );
    }

    const logoUrl = RuntimeConfig.INFOS.customLogoUrl || "/ui/seatsurfing.svg";
    const isDefaultLogo = !RuntimeConfig.INFOS.customLogoUrl;

    return (
      <>
        <Navbar
          bg="body"
          fixed="top"
          className="booking-navbar"
          expand={RuntimeConfig.EMBEDDED ? true : "lg"}
        >
          <Container fluid={true}>
            <Navbar.Brand as={NavLink} to="/search">
              <img
                src={logoUrl}
                alt="Seatsurfing"
                className={isDefaultLogo ? "default-logo" : undefined}
              />
            </Navbar.Brand>
            {collapsable}
          </Container>
        </Navbar>
        <PreferencesModal
          show={this.state.showPreferences}
          initialTab={this.state.preferencesTab}
          onHide={() => this.setState({ showPreferences: false })}
        />
      </>
    );
  }
}

export default withTranslation(withReadyRouter(NavBar as any));
