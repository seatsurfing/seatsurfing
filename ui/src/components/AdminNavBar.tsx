import React from "react";
import { Nav, Button, Form } from "react-bootstrap";
import { NextRouter } from "next/router";
import Link from "next/link";
import Ajax from "@/util/Ajax";
import { TranslationFunc, withTranslation } from "./withTranslation";
import withReadyRouter from "./withReadyRouter";
import RuntimeConfig from "./RuntimeConfig";

interface State {
  search: string;
  redirect: string | null;
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
}

class AdminNavBar extends React.Component<Props, State> {
  constructor(props: any) {
    super(props);
    this.state = {
      search: "",
      redirect: null,
    };
  }

  logout = (e: any) => {
    e.preventDefault();
    RuntimeConfig.logOut();
  };

  submitSearchForm = (e: any) => {
    e.preventDefault();
    window.sessionStorage.setItem("searchKeyword", this.state.search);
    this.setState({
      redirect: "/admin/search/" + window.encodeURIComponent(this.state.search),
    });
  };

  componentDidMount = () => {
    let isSearchPage = window.location.href.indexOf("/admin/search/") > -1;
    let keyword = window.sessionStorage.getItem("searchKeyword");
    if (isSearchPage && keyword) {
      this.setState({ search: keyword });
    } else {
      window.sessionStorage.removeItem("searchKeyword");
    }
  };

  render() {
    if (this.state.redirect != null) {
      let target = this.state.redirect;
      this.setState({ redirect: null });
      this.props.router.push(target);
      return <></>;
    }

    return (
      <Nav className="admin-navbar navbar navbar-dark sticky-top bg-dark flex-nowrap p-0 shadow">
        <Link
          className="navbar-brand col-1 col-md-3 col-lg-2 me-0 px-3"
          href="/admin/dashboard"
        >
          <img
            src="/ui/seatsurfing_white.svg"
            alt="Seatsurfing"
            className="d-none d-md-block"
          />
          <img
            src="/ui/seatsurfing_white_logo.svg"
            alt="Seatsurfing"
            className="d-block d-md-none"
          />
        </Link>
        <Form onSubmit={this.submitSearchForm} className="w-100">
          <Form.Control
            type="text"
            className="form-control form-control-dark w-100"
            placeholder={this.props.t("search")}
            aria-label="Suchen"
            value={this.state.search}
            onChange={(e: any) => this.setState({ search: e.target.value })}
            required={true}
          />
        </Form>
        <ul className="navbar-nav px-3">
          <li className="nav-item text-nowrap">
            <Button variant="link" className="nav-link" onClick={this.logout}>
              {" "}
              {this.props.t("logout")}
            </Button>
          </li>
        </ul>
      </Nav>
    );
  }
}

export default withTranslation(withReadyRouter(AdminNavBar as any));
