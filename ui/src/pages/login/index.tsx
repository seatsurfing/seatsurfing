import React from "react";
import { Form, Button, InputGroup, Alert } from "react-bootstrap";
import { NextRouter } from "next/router";
import Link from "next/link";
import withReadyRouter from "@/components/withReadyRouter";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import AuthProvider from "@/types/AuthProvider";
import Organization from "@/types/Organization";
import Ajax from "@/util/Ajax";
import AjaxCredentials from "@/util/AjaxCredentials";
import RuntimeConfig from "@/components/RuntimeConfig";
import Loading from "@/components/Loading";
import LanguageSelector from "@/components/LanguageSelector";

interface State {
  email: string;
  password: string;
  invalid: boolean;
  redirect: string | null;
  requirePassword: boolean;
  disablePasswordLogin: boolean;
  providers: AuthProvider[] | null;
  inPasswordSubmit: boolean;
  inAuthProviderLogin: boolean;
  singleOrgMode: boolean;
  noPasswords: boolean;
  loading: boolean;
  orgDomain: string;
  domainNotFound: boolean;
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
}

class Login extends React.Component<Props, State> {
  org: Organization | null;

  constructor(props: any) {
    super(props);
    this.org = null;
    this.state = {
      email: "",
      password: "",
      invalid: false,
      redirect: null,
      requirePassword: false,
      disablePasswordLogin: false,
      providers: null,
      inPasswordSubmit: false,
      inAuthProviderLogin: false,
      singleOrgMode: false,
      noPasswords: false,
      loading: true,
      orgDomain: "",
      domainNotFound: false,
    };
  }

  componentDidMount = () => {
    if (this.state.email === "") {
      const emailParam = this.props.router.query["email"];
      if (emailParam !== "") {
        this.setState({
          email: emailParam as string,
        });
      }
    }
    this.loadOrgDetails();
  };

  applyOrg = (res: any) => {
    this.org = new Organization();
    this.org.deserialize(res.json.organization);
    this.setState(
      {
        providers: res.json.authProviders,
        noPasswords: !res.json.requirePassword,
        disablePasswordLogin: res.json.disablePasswordLogin,
        singleOrgMode: true,
        loading: false,
      },
      () => {
        const noRedirect = this.props.router.query["noredirect"];
        if (
          noRedirect !== "1" &&
          this.state.noPasswords &&
          this.state.providers &&
          this.state.providers.length === 1
        ) {
          this.useProvider(this.state.providers[0].id);
        } else {
          this.setState({ loading: false });
        }
      },
    );
  };

  loadOrgDetails = () => {
    const domain = window.location.host.split(":").shift();
    Ajax.get("/auth/org/" + domain)
      .then((res) => {
        this.applyOrg(res);
      })
      .catch(() => {
        // No org for domain found
        this.checkSingleOrg();
      });
  };

  checkSingleOrg = () => {
    Ajax.get("/auth/singleorg")
      .then((res) => {
        this.applyOrg(res);
      })
      .catch(() => {
        this.setState({
          domainNotFound: true,
          loading: false,
        });
      });
  };

  onPasswordSubmit = (e: any) => {
    e.preventDefault();
    this.setState({
      inPasswordSubmit: true,
    });
    const payload = {
      email: this.state.email,
      password: this.state.password,
      organizationId: this.org?.id,
    };
    Ajax.postData("/auth/login", payload)
      .then((res) => {
        const credentials: AjaxCredentials = {
          accessToken: res.json.accessToken,
          accessTokenExpiry: new Date(
            new Date().getTime() + Ajax.ACCESS_TOKEN_EXPIRY_OFFSET,
          ),
          logoutUrl: res.json.logoutUrl,
          profilePageUrl: "",
        };
        Ajax.PERSISTER.updateCredentialsSessionStorage(credentials);
        Ajax.PERSISTER.persistRefreshTokenInLocalStorage(res.json.refreshToken);
        RuntimeConfig.loadUserAndSettings().then(() => {
          this.setState({ redirect: this.getRedirectUrl() });
        });
      })
      .catch(() => {
        this.setState({
          invalid: true,
          inPasswordSubmit: false,
        });
      });
  };

  cancelPasswordLogin = (e: any) => {
    e.preventDefault();
    this.setState({
      requirePassword: false,
      providers: null,
      invalid: false,
    });
  };

  getRedirectUrl = () => {
    return (this.props.router.query["redir"] as string) || "/search";
  };

  renderAuthProviderButton = (provider: AuthProvider) => {
    return (
      <p key={provider.id}>
        <Button
          variant="primary"
          className="btn-auth-provider"
          onClick={() => this.useProvider(provider.id)}
        >
          {this.state.inAuthProviderLogin ? (
            <Loading showText={false} paddingTop={false} />
          ) : (
            provider.name
          )}
        </Button>
      </p>
    );
  };

  useProvider = (providerId: string) => {
    this.setState({
      inAuthProviderLogin: true,
    });
    let target = Ajax.getBackendUrl() + "/auth/" + providerId + "/login/ui/";
    const redir = this.props.router.query["redir"] as string;
    if (redir) {
      target += "?redir=" + encodeURIComponent(redir);
    }
    window.location.href = target;
  };

  render() {
    if (this.state.redirect != null) {
      this.props.router.push(this.state.redirect);
      return <></>;
    }
    if (Ajax.hasAccessToken()) {
      this.props.router.push(this.getRedirectUrl());
      return <></>;
    }

    if (this.state.loading) {
      return (
        <>
          <Loading />
        </>
      );
    }

    const copyrightFooter = (
      <div className="copyright-footer">
        &copy; Seatsurfing &#183;{" "}
        <a href="https://seatsurfing.io" target="_blank">
          https://seatsurfing.io
        </a>
        <LanguageSelector />
      </div>
    );

    if (this.state.domainNotFound) {
      return (
        <div className="container-signin">
          <Form className="form-signin">
            <img src="/ui/seatsurfing.svg" alt="Seatsurfing" className="logo" />
            <h3>Domain not found.</h3>
            <p>
              Please make sure your domain name is set up correctly in
              Seatsurfing&#39;s settings.
            </p>
            <p>If you believe this is an error, please contact support.</p>
          </Form>
        </div>
      );
    }

    if (this.state.providers != null && this.state.providers.length > 0) {
      const buttons = this.state.providers.map((provider) =>
        this.renderAuthProviderButton(provider),
      );
      let providerSelection = (
        <p>
          {this.props.t("signinAsAt", {
            user: this.state.email,
            org: this.org?.name,
          })}
        </p>
      );
      if (this.state.singleOrgMode) {
        providerSelection = <p></p>;
      }
      if (buttons.length === 0) {
        providerSelection = <p>{this.props.t("errorNoAuthProviders")}</p>;
      }
      return (
        <div className="container-signin">
          <Form className="form-signin">
            <img src="/ui/seatsurfing.svg" alt="Seatsurfing" className="logo" />
            <h3>{this.org?.name}</h3>
            {providerSelection}
            {buttons}
            <p
              className="margin-top-50"
              hidden={this.state.disablePasswordLogin}
            >
              <Button
                variant="link"
                onClick={() => this.setState({ providers: null })}
              >
                {this.props.t("loginUseUsernamePassword")}
              </Button>
            </p>
          </Form>
          {copyrightFooter}
        </div>
      );
    }

    if (this.state.disablePasswordLogin) {
      return (
        <div className="container-signin">
          <Form className="form-signin">
            <img src="/ui/seatsurfing.svg" alt="Seatsurfing" className="logo" />
            <h3>{this.org?.name}</h3>
            <p>
              Password Login is disabled, but no Auth Providers are configured.
              Please contact your administrator.
            </p>
          </Form>
        </div>
      );
    }

    return (
      <div className="container-signin">
        <Form className="form-signin" onSubmit={this.onPasswordSubmit}>
          <img src="/ui/seatsurfing.svg" alt="Seatsurfing" className="logo" />
          <h3>{this.org?.name}</h3>
          <Form.Group style={{ marginBottom: "5px" }}>
            <Form.Control
              type="email"
              readOnly={this.state.inPasswordSubmit}
              placeholder={this.props.t("emailPlaceholder")}
              value={this.state.email}
              onChange={(e: any) =>
                this.setState({ email: e.target.value, invalid: false })
              }
              required={true}
              isInvalid={this.state.invalid}
              autoFocus={true}
            />
          </Form.Group>
          <Form.Group>
            <InputGroup>
              <Form.Control
                type="password"
                readOnly={this.state.inPasswordSubmit}
                placeholder={this.props.t("password")}
                value={this.state.password}
                onChange={(e: any) =>
                  this.setState({ password: e.target.value, invalid: false })
                }
                required={true}
                isInvalid={this.state.invalid}
                minLength={8}
              />
              <Button variant="primary" type="submit">
                {this.state.inPasswordSubmit ? (
                  <Loading showText={false} paddingTop={false} />
                ) : (
                  <div className="feather-btn">&#10148;</div>
                )}
              </Button>
            </InputGroup>
          </Form.Group>
          <Form.Control.Feedback type="invalid">
            {this.props.t("errorInvalidEmail")}
          </Form.Control.Feedback>
          <p className="margin-top-50" hidden={!this.org}>
            <Link href="/resetpw">{this.props.t("forgotPassword")}</Link>
          </p>
        </Form>
        {copyrightFooter}
      </div>
    );
  }
}

export default withTranslation(withReadyRouter(Login as any) as any);
