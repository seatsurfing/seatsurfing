import React from "react";
import { Navbar, Nav, Container, NavLink } from "react-bootstrap";
import RuntimeConfig from "./RuntimeConfig";
import {
  Settings as IconSettings,
  Calendar as IconCalendar,
  PlusSquare as IconPlus,
  User as IconUser,
  Heart as IconBuddies,
  Shield as IconAdmin,
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

interface State {
  allowAdmin: boolean;
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
}

class NavBar extends React.Component<Props, State> {
  constructor(props: any) {
    super(props);
    this.state = {
      allowAdmin: false,
    };
  }

  componentDidMount = () => {
    if (!Ajax.hasAccessToken()) {
      return;
    }
    this.loadData();
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
          <Nav.Link as={Link} eventKey="/preferences" href="/preferences">
            {RuntimeConfig.EMBEDDED ? (
              <IconSettings className="feather feather-lg" />
            ) : (
              <>
                <IconSettings className="feather" />{" "}
                {this.props.t("preferences")}
              </>
            )}
          </Nav.Link>
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
          bg="body-tertiary"
          fixed="top"
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
      </>
    );
  }
}

export default withTranslation(withReadyRouter(NavBar as any));
