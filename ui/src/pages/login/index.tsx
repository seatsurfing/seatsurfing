import React from "react";
import { Form, Button, InputGroup } from "react-bootstrap";
import { NextRouter } from "next/router";
import Link from "next/link";
import withReadyRouter from "@/components/withReadyRouter";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import AuthProvider from "@/types/AuthProvider";
import Organization from "@/types/Organization";
import Ajax from "@/util/Ajax";
import AjaxCredentials from "@/util/AjaxCredentials";
import RuntimeConfig from "@/components/RuntimeConfig";
import JwtDecoder from "@/util/JwtDecoder";
import Loading from "@/components/Loading";
import LanguageSelector from "@/components/LanguageSelector";
import * as Validation from "@/util/Validation";
import * as Navigation from "@/util/Navigation";
import AjaxError from "@/util/AjaxError";
import TotpInput from "@/components/TotpInput";
import Passkey, {
  prepareRequestOptions,
  serializeAssertionResponse,
  PasskeyChallengeResponse,
} from "@/types/Passkey";

interface State {
  email: string;
  password: string;
  code: string;
  requireTotp: boolean;
  requirePasskey: boolean;
  passkeyStateId: string;
  passkeyOptions: any;
  allowTotpFallback: boolean;
  inPasskeyLogin: boolean;
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
  passkeyAvailable: boolean;
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
      code: "",
      requireTotp: false,
      requirePasskey: false,
      passkeyStateId: "",
      passkeyOptions: null,
      allowTotpFallback: false,
      inPasskeyLogin: false,
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
      // Sync pre-check; refined by the async call in componentDidMount (Finding #14)
      passkeyAvailable: Passkey.isSupported(),
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
    // Refine the platform authenticator availability check asynchronously (Finding #14)
    Passkey.isPlatformAuthAvailable().then((available) =>
      this.setState({ passkeyAvailable: available }),
    );
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
    const payload: any = {
      email: this.state.email,
      password: this.state.password,
      organizationId: this.org?.id,
      code: this.state.code,
    };
    // If we have a pending passkey challenge response, attach it
    if (this.state.requirePasskey && this.state.passkeyStateId) {
      payload.passkeyStateId = this.state.passkeyStateId;
    }
    Ajax.postData("/auth/login", payload)
      .then((res) => {
        const credentials: AjaxCredentials = {
          accessToken: res.json.accessToken,
          accessTokenExpiry: JwtDecoder.getExpiryDate(res.json.accessToken),
          logoutUrl: res.json.logoutUrl,
          profilePageUrl: "",
        };
        Ajax.PERSISTER.updateCredentialsLocalStorage(credentials);
        Ajax.PERSISTER.persistRefreshTokenInLocalStorage(res.json.refreshToken);
        RuntimeConfig.loadUserAndSettings().then(() => {
          this.setState({ redirect: this.getRedirectUrl() });
        });
      })
      .catch((err) => {
        if (
          err instanceof AjaxError &&
          (err as AjaxError).httpStatusCode === 401
        ) {
          // Check if the body contains a passkey challenge
          if ((err as any).responseBody) {
            try {
              const body: PasskeyChallengeResponse = JSON.parse(
                (err as any).responseBody,
              );
              if (body.requirePasskey) {
                const rawOpts =
                  body.passkeyChallenge?.publicKey ?? body.passkeyChallenge;
                this.setState({
                  requirePasskey: true,
                  passkeyStateId: body.stateId,
                  passkeyOptions: rawOpts,
                  allowTotpFallback: body.allowTotpFallback,
                  invalid: false,
                  inPasswordSubmit: false,
                });
                // Automatically trigger the passkey ceremony
                this.performPasskeyAssertion(body.stateId, rawOpts);
                return;
              }
            } catch (_) {}
          }
          this.setState({
            requireTotp: true,
            invalid: false,
            inPasswordSubmit: false,
          });
          return;
        }
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

  performPasskeyAssertion = async (stateId: string, rawOptions: any) => {
    this.setState({ inPasskeyLogin: true, invalid: false });
    try {
      const publicKeyOptions = prepareRequestOptions(rawOptions);
      const credential = await navigator.credentials.get({
        publicKey: publicKeyOptions,
      });
      if (!credential) {
        throw new Error("No credential returned");
      }
      const serialized = serializeAssertionResponse(
        credential as PublicKeyCredential,
      );
      // Submit password + passkey together
      const payload: any = {
        email: this.state.email,
        password: this.state.password,
        organizationId: this.org?.id,
        code: this.state.code,
        passkeyStateId: stateId,
        passkeyCredential: serialized,
      };
      const res = await Ajax.postData("/auth/login", payload);
      const credentials: AjaxCredentials = {
        accessToken: res.json.accessToken,
        accessTokenExpiry: JwtDecoder.getExpiryDate(res.json.accessToken),
        logoutUrl: res.json.logoutUrl,
        profilePageUrl: "",
      };
      Ajax.PERSISTER.updateCredentialsLocalStorage(credentials);
      Ajax.PERSISTER.persistRefreshTokenInLocalStorage(res.json.refreshToken);
      await RuntimeConfig.loadUserAndSettings();
      this.setState({ redirect: this.getRedirectUrl(), inPasskeyLogin: false });
    } catch (err: any) {
      // On passkey failure, offer TOTP fallback if available (spec §6.1)
      if (this.state.allowTotpFallback) {
        this.setState({
          inPasskeyLogin: false,
          requirePasskey: false,
          requireTotp: true,
          invalid: false,
        });
      } else {
        this.setState({
          invalid: true,
          inPasskeyLogin: false,
          requirePasskey: false,
        });
      }
    }
  };

  loginWithPasskey = async () => {
    if (!this.state.passkeyAvailable) return;
    this.setState({ inPasskeyLogin: true, invalid: false });
    try {
      const beginResponse = await Passkey.beginLogin(this.org?.id ?? "");
      const rawOpts =
        beginResponse.challenge?.publicKey ?? beginResponse.challenge;
      const publicKeyOptions = prepareRequestOptions(rawOpts);
      const credential = await navigator.credentials.get({
        publicKey: publicKeyOptions,
      });
      if (!credential) {
        throw new Error("No credential returned");
      }
      const serialized = serializeAssertionResponse(
        credential as PublicKeyCredential,
      );
      const res = await Passkey.finishLogin(beginResponse.stateId, serialized);
      const credentials: AjaxCredentials = {
        accessToken: res.accessToken,
        accessTokenExpiry: JwtDecoder.getExpiryDate(res.accessToken),
        logoutUrl: "",
        profilePageUrl: "",
      };
      Ajax.PERSISTER.updateCredentialsLocalStorage(credentials);
      Ajax.PERSISTER.persistRefreshTokenInLocalStorage(res.refreshToken);
      await RuntimeConfig.loadUserAndSettings();
      this.setState({ redirect: this.getRedirectUrl(), inPasskeyLogin: false });
    } catch (err: any) {
      // User cancelled or credential not found
      this.setState({ inPasskeyLogin: false, invalid: false });
    }
  };

  getRedirectUrl = () => {
    // prevent (open) redirect to absolute URLs
    const redirectUrl = this.props.router.query["redir"] as string;
    if (!redirectUrl || Validation.isAbsoluteUrl(redirectUrl)) {
      return Navigation.PATH_SEARCH;
    }

    // do not redirect to admin pages for non-admin users (to prevent "auto logout")
    if (
      Navigation.isAdminPath(redirectUrl) &&
      !RuntimeConfig.INFOS?.spaceAdmin
    ) {
      return Navigation.PATH_SEARCH;
    }

    return redirectUrl;
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
        <a
          href="https://seatsurfing.io"
          target="_blank"
          rel="noopener noreferrer"
        >
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
        {/* Passkey 2FA prompt – shown when password verified but passkey required */}
        <Form
          className="form-signin"
          name="passkey-2fa"
          hidden={!this.state.requirePasskey}
        >
          <img src="/ui/seatsurfing.svg" alt="Seatsurfing" className="logo" />
          <h3>{this.org?.name}</h3>
          <p>{this.props.t("passkeyRequired")}</p>
          <Button
            variant="link"
            onClick={() =>
              this.performPasskeyAssertion(
                this.state.passkeyStateId,
                this.state.passkeyOptions,
              )
            }
            disabled={this.state.inPasskeyLogin}
          >
            {this.props.t("signInWithPasskey")}
          </Button>
          {this.state.allowTotpFallback && (
            <Button
              variant="link"
              onClick={() =>
                this.setState({ requirePasskey: false, requireTotp: true })
              }
              disabled={this.state.inPasskeyLogin}
            >
              {this.props.t("useTotpInstead")}
            </Button>
          )}
          {this.state.invalid && (
            <p className="text-danger">{this.props.t("errorUnknown")}</p>
          )}
        </Form>
        <Form
          className="form-signin"
          onSubmit={this.onPasswordSubmit}
          name="totp-login"
          hidden={!this.state.requireTotp}
        >
          <img src="/ui/seatsurfing.svg" alt="Seatsurfing" className="logo" />
          <h3>{this.org?.name}</h3>
          <p>{this.props.t("enterTotpCode")}</p>
          <Form.Group>
            <InputGroup>
              <TotpInput
                value={this.state.code}
                onChange={(value: string) =>
                  this.setState({ code: value, invalid: false })
                }
                onComplete={(value: string) => {
                  this.setState({ code: value, invalid: false }, () => {
                    this.onPasswordSubmit(new Event("submit") as any);
                  });
                }}
                disabled={this.state.inPasswordSubmit}
                invalid={this.state.invalid}
                hidden={!this.state.requireTotp}
                required={true}
              />
            </InputGroup>
          </Form.Group>
        </Form>
        <Form
          className="form-signin"
          onSubmit={this.onPasswordSubmit}
          name="password-login"
          hidden={this.state.requireTotp || this.state.requirePasskey}
        >
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
          {this.state.passkeyAvailable && (
            <p className="margin-top-25">
              <Button
                variant="link"
                onClick={this.loginWithPasskey}
                disabled={
                  this.state.inPasskeyLogin || this.state.inPasswordSubmit
                }
                type="button"
              >
                {this.props.t("signInWithPasskey")}
              </Button>
            </p>
          )}
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
