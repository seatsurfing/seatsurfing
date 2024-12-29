import React from 'react';
import { Form, Button, InputGroup, DropdownButton, Dropdown } from 'react-bootstrap';
import { Organization, AuthProvider, Ajax, JwtDecoder, User } from 'flexspace-commons';
import Loading from '../../components/Loading';
import { WithTranslation, withTranslation } from 'next-i18next';
import { NextRouter } from 'next/router';
import withReadyRouter from '@/components/withReadyRouter';
import Link from 'next/link';

interface State {
  email: string
  password: string
  invalid: boolean
  redirect: string | null
  requirePassword: boolean
  providers: AuthProvider[] | null
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
      invalid: false,
      redirect: null,
      requirePassword: false,
      providers: null,
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
    let payload = {
      email: this.state.email
    };
    Ajax.postData("/auth/preflight", payload).then((res) => {
      this.org = new Organization();
      this.org.deserialize(res.json.organization);
      this.setState({
        providers: res.json.authProviders,
        requirePassword: res.json.requirePassword
      });
    }).catch((e) => {
      this.setState({
        invalid: true
      });
    });
  }

  onPasswordSubmit = (e: any) => {
    e.preventDefault();
    let payload = {
      email: this.state.email,
      password: this.state.password,
      longLived: false
    };
    Ajax.postData("/auth/login", payload).then((res) => {
      let jwtPayload = JwtDecoder.getPayload(res.json.accessToken);
      if (jwtPayload.role < User.UserRoleSpaceAdmin) {
        this.setState({
          invalid: true
        });
        return;
      }
      Ajax.CREDENTIALS = {
        accessToken: res.json.accessToken,
        refreshToken: res.json.refreshToken,
        accessTokenExpiry: new Date(new Date().getTime() + Ajax.ACCESS_TOKEN_EXPIRY_OFFSET),
        logoutUrl: res.json.logoutUrl,
      };
      Ajax.PERSISTER.updateCredentialsSessionStorage(Ajax.CREDENTIALS).then(() => {
        this.setState({
          redirect: "/dashboard"
        });
      });
    }).catch((e) => {
      this.setState({
        invalid: true
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
        <Button variant="primary" className="btn-auth-provider" onClick={() => this.useProvider(provider.id)}>{provider.name}</Button>
      </p>
    );
  }

  useProvider = (providerId: string) => {
    let target = Ajax.getBackendUrl() + "/auth/" + providerId + "/login/web";
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
            <img src="/admin/seatsurfing.svg" alt="Seatsurfing" className="logo" />
            <p>{this.props.t("signinAsAt", { user: this.state.email, org: this.org?.name })}</p>
            <InputGroup>
              <Form.Control type="password" placeholder={this.props.t("password")} value={this.state.password} onChange={(e: any) => this.setState({ password: e.target.value, invalid: false })} required={true} isInvalid={this.state.invalid} minLength={8} autoFocus={true} />
              <Button variant="primary" type="submit">&#10148;</Button>
            </InputGroup>
            <Form.Control.Feedback type="invalid">{this.props.t("errorInvalidPassword")}</Form.Control.Feedback>
            <Button variant="secondary" className="btn-auth-provider btn-back" onClick={this.cancelPasswordLogin}>{this.props.t("back")}</Button>
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
            <img src="/admin/seatsurfing.svg" alt="Seatsurfing" className="logo" />
            {providerSelection}
            {buttons}
            <Button variant="secondary" className="btn-auth-provider" onClick={() => this.setState({ providers: null })}>{this.props.t("back")}</Button>
          </Form>
          {copyrightFooter}
        </div>
      );
    }

    return (
      <div className="container-signin">
        <Form className="form-signin" onSubmit={this.onSubmit}>
          <img src="/admin/seatsurfing.svg" alt="Seatsurfing" className="logo" />
          <h3>{this.props.t("mangageOrgHeadline")}</h3>
          <InputGroup>
            <Form.Control type="email" placeholder={this.props.t("emailAddress")} value={this.state.email} onChange={(e: any) => this.setState({ email: e.target.value, invalid: false })} required={true} isInvalid={this.state.invalid} autoFocus={true} />
            <Button variant="primary" type="submit">&#10148;</Button>
          </InputGroup>
          <Form.Control.Feedback type="invalid">{this.props.t("errorInvalidEmail")}</Form.Control.Feedback>
        </Form>
        {copyrightFooter}
      </div>
    );
  }
}

export default withTranslation(['admin'])(withReadyRouter(Login as any));
