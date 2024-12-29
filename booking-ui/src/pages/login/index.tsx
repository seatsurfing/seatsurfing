import React from 'react';
import { Form, Button, InputGroup, Dropdown, DropdownButton } from 'react-bootstrap';
import { Organization, AuthProvider, Ajax } from 'flexspace-commons';
import { withTranslation, WithTranslation } from 'next-i18next';
import RuntimeConfig from '../../components/RuntimeConfig';
import Loading from '../../components/Loading';
import { NextRouter } from 'next/router';
import Link from 'next/link';
import withReadyRouter from '@/components/withReadyRouter';

interface State {
  email: string
  password: string
  rememberMe: boolean
  invalid: boolean
  redirect: string | null
  requirePassword: boolean
  providers: AuthProvider[] | null
  inPreflight: boolean
  inPasswordSubmit: boolean
  inAuthProviderLogin: boolean
  singleOrgMode: boolean
  noPasswords: boolean
  loading: boolean
}

interface Props extends WithTranslation {
  router: NextRouter
}

class Login extends React.Component<Props, State> {
  org: Organization | null;

  constructor(props: any) {
    super(props);
    this.org = null;
    this.state = {
      email: "",
      password: "",
      rememberMe: false,
      invalid: false,
      redirect: null,
      requirePassword: false,
      providers: null,
      inPreflight: false,
      inPasswordSubmit: false,
      inAuthProviderLogin: false,
      singleOrgMode: false,
      noPasswords: false,
      loading: true
    };
  }

  componentDidMount = () => {
    this.checkSingleOrg();
  }

  checkSingleOrg = () => {
    Ajax.get("/auth/singleorg").then((res) => {
      this.org = new Organization();
      this.org.deserialize(res.json.organization);
      if ((res.json.authProviders) && (res.json.authProviders.length > 0)) {
        this.setState({
          providers: res.json.authProviders,
          noPasswords: !res.json.requirePassword,
          singleOrgMode: true,
          loading: false
        }, () => {
          if ((this.state.noPasswords) && (this.state.providers) && (this.state.providers.length === 1)) {
            this.useProvider(this.state.providers[0].id);
          } else {
            this.setState({ loading: false });
          }
        });
      } else {
        this.setState({ loading: false });
      }
    }).catch(() => {
      this.setState({ loading: false });
    });
  }

  onSubmit = (e: any) => {
    e.preventDefault();
    let email = this.state.email.split("@");
    if (email.length !== 2) {
      // Error
      return;
    }
    this.setState({
      inPreflight: true
    });
    let payload = {
      email: this.state.email
    };
    Ajax.postData("/auth/preflight", payload).then((res) => {
      this.org = new Organization();
      this.org.deserialize(res.json.organization);
      this.setState({
        providers: res.json.authProviders,
        requirePassword: res.json.requirePassword,
        inPreflight: false
      });
    }).catch(() => {
      this.setState({
        invalid: true,
        inPreflight: false
      });
    });
  }

  onPasswordSubmit = (e: any) => {
    e.preventDefault();
    this.setState({
      inPasswordSubmit: true
    });
    let payload = {
      email: this.state.email,
      password: this.state.password,
      longLived: this.state.rememberMe
    };
    Ajax.postData("/auth/login", payload).then((res) => {
      Ajax.CREDENTIALS = {
        accessToken: res.json.accessToken,
        refreshToken: res.json.refreshToken,
        accessTokenExpiry: new Date(new Date().getTime() + Ajax.ACCESS_TOKEN_EXPIRY_OFFSET),
        logoutUrl: res.json.logoutUrl,
      };
      Ajax.PERSISTER.updateCredentialsSessionStorage(Ajax.CREDENTIALS).then(() => {
        if (this.state.rememberMe) {
          Ajax.PERSISTER.persistRefreshTokenInLocalStorage(Ajax.CREDENTIALS);
        }
        RuntimeConfig.setLoginDetails().then(() => {
          let redirect = this.props.router.query["redir"] as string || "/search";

          this.setState({ redirect });
        });
      });
    }).catch(() => {
      this.setState({
        invalid: true,
        inPasswordSubmit: false
      });
    });
  }

  cancelPasswordLogin = (e: any) => {
    e.preventDefault();
    this.setState({
      requirePassword: false,
      providers: null,
      invalid: false
    });
  }

  renderAuthProviderButton = (provider: AuthProvider) => {
    return (
      <p key={provider.id}>
        <Button variant="primary" className="btn-auth-provider" onClick={() => this.useProvider(provider.id)}>{this.state.inAuthProviderLogin ? <Loading showText={false} paddingTop={false} /> : provider.name}</Button>
      </p>
    );
  }

  useProvider = (providerId: string) => {
    this.setState({
      inAuthProviderLogin: true
    });
    let target = Ajax.getBackendUrl() + "/auth/" + providerId + "/login/ui";
    if (this.state.rememberMe) {
      target += "/1"
    }
    let redir = this.props.router.query["redir"] as string;
    if (redir) {
      target += "?redir=" + encodeURIComponent(redir);
    }
    window.location.href = target;
  }

  changeLanguage = (lng: string) => {
    this.props.i18n.changeLanguage(lng);
  }

  render() {
    if (this.state.redirect != null) {
      this.props.router.push(this.state.redirect);
      return <></>
    }
    if (Ajax.CREDENTIALS.accessToken) {
      this.props.router.push("/search");
      return <></>
    }

    if (this.state.loading || !this.props.tReady) {
      return (
        <>
          <Loading />
        </>
      );
    }

    let languageSelectDropdown = (
      <DropdownButton title={this.props.i18n.language} className='lng-selector' size='sm' variant='outline-secondary' drop='up'>
        {this.props.router.locales?.sort().map(l => <Dropdown.Item key={'lng-btn-' + l} onClick={() => this.changeLanguage(l)} active={l === this.props.i18n.language}>{l}</Dropdown.Item>)}
      </DropdownButton>
    );

    let copyrightFooter = (
      <p className="copyright-footer">
        &copy; Seatsurfing &#183; Version {process.env.NEXT_PUBLIC_PRODUCT_VERSION}
        {languageSelectDropdown}
      </p>
    );

    if (this.state.requirePassword) {
      return (
        <div className="container-signin">
          <Form className="form-signin" onSubmit={this.onPasswordSubmit}>
            <img src="/ui/seatsurfing.svg" alt="Seatsurfing" className="logo" />
            <p>{this.props.t("signinAsAt", { user: this.state.email, org: this.org?.name })}</p>
            <InputGroup>
              <Form.Control type="password" readOnly={this.state.inPasswordSubmit} placeholder={this.props.t("password")} value={this.state.password} onChange={(e: any) => this.setState({ password: e.target.value, invalid: false })} required={true} isInvalid={this.state.invalid} minLength={8} autoFocus={true} />
              <Button variant="primary" type="submit">{this.state.inPasswordSubmit ? <Loading showText={false} paddingTop={false} /> : <div className="feather-btn">&#10148;</div>}</Button>
            </InputGroup>
            <Form.Control.Feedback type="invalid">{this.props.t("errorInvalidPassword")}</Form.Control.Feedback>
            <p className="margin-top-50"><Button variant="link" onClick={this.cancelPasswordLogin}>{this.props.t("back")}</Button></p>
          </Form>
          {copyrightFooter}
        </div>
      );
    }

    if (this.state.providers != null) {
      let buttons = this.state.providers.map(provider => this.renderAuthProviderButton(provider));
      let providerSelection = <p>{this.props.t("signinAsAt", { user: this.state.email, org: this.org?.name })}</p>;
      if (this.state.singleOrgMode) {
        providerSelection = <p>{this.props.t("signinAt", { org: this.org?.name })}</p>;
      }
      if (buttons.length === 0) {
        providerSelection = <p>{this.props.t("errorNoAuthProviders")}</p>
      }
      return (
        <div className="container-signin">
          <Form className="form-signin">
            <img src="/ui/seatsurfing.svg" alt="Seatsurfing" className="logo" />
            {providerSelection}
            {buttons}
            <p className="margin-top-50"><Button variant="link" onClick={() => this.setState({ providers: null })}>{this.props.t("back")}</Button></p>
          </Form>
          {copyrightFooter}
        </div>
      );
    }

    return (
      <div className="container-signin">
        <Form className="form-signin" onSubmit={this.onSubmit}>
          <img src="/ui/seatsurfing.svg" alt="Seatsurfing" className="logo" />
          <h3>{this.props.t("findYourPlace")}</h3>
          <InputGroup>
            <Form.Control type="email" readOnly={this.state.inPreflight} placeholder={this.props.t("emailPlaceholder")} value={this.state.email} onChange={(e: any) => this.setState({ email: e.target.value, invalid: false })} required={true} isInvalid={this.state.invalid} autoFocus={true} />
            <Button variant="primary" type="submit">{this.state.inPreflight ? <Loading showText={false} paddingTop={false} /> : <div className="feather-btn">&#10148;</div>}</Button>
          </InputGroup>
          <Form.Control.Feedback type="invalid">{this.props.t("errorInvalidEmail")}</Form.Control.Feedback>
          <Form.Check type="checkbox" id="check-rememberme" label={this.props.t("rememberMe")} checked={this.state.rememberMe} onChange={(e: any) => this.setState({ rememberMe: e.target.checked })} />
          <p className="margin-top-50"><Link href="/resetpw">{this.props.t("forgotPassword")}</Link></p>
        </Form>
        {copyrightFooter}
      </div>
    );
  }
}

export default withTranslation()(withReadyRouter(Login as any));
