import React from 'react';
import { Form, Col, Row, Button, Alert, ButtonGroup } from 'react-bootstrap';
import { ChevronLeft as IconBack, Save as IconSave, Trash2 as IconDelete } from 'react-feather';
import { Ajax, AuthProvider } from 'flexspace-commons';
import { WithTranslation, withTranslation } from 'next-i18next';
import { NextRouter } from 'next/router';
import FullLayout from '@/components/FullLayout';
import Link from 'next/link';
import Loading from '@/components/Loading';
import withReadyRouter from '@/components/withReadyRouter';

interface State {
  loading: boolean
  submitting: boolean
  saved: boolean
  goBack: boolean
  name: string
  providerType: number;
  authUrl: string;
  tokenUrl: string;
  authStyle: number;
  scopes: string;
  userInfoUrl: string;
  userInfoEmailField: string;
  clientId: string;
  clientSecret: string;
  logoutUrl: string;
}

interface Props extends WithTranslation {
  router: NextRouter
}

class EditAuthProvider extends React.Component<Props, State> {
  entity: AuthProvider = new AuthProvider();

  constructor(props: any) {
    super(props);
    this.state = {
      loading: true,
      submitting: false,
      saved: false,
      goBack: false,
      name: "",
      providerType: 0,
      authUrl: "",
      tokenUrl: "",
      authStyle: 0,
      scopes: "",
      userInfoUrl: "",
      userInfoEmailField: "",
      clientId: "",
      clientSecret: "",
      logoutUrl: "",
    };
  }

  componentDidMount = () => {
    if (!Ajax.CREDENTIALS.accessToken) {
      this.props.router.push("/login");
      return;
    }
    this.loadData();
  }

  loadData = () => {
    const { id } = this.props.router.query;
    if (id && (typeof id === "string") && (id !== 'add')) {
      AuthProvider.get(id).then(authProvider => {
        this.entity = authProvider;
        this.setState({
          name: authProvider.name,
          providerType: authProvider.providerType,
          authUrl: authProvider.authUrl,
          tokenUrl: authProvider.tokenUrl,
          authStyle: authProvider.authStyle,
          scopes: authProvider.scopes,
          userInfoUrl: authProvider.userInfoUrl,
          userInfoEmailField: authProvider.userInfoEmailField,
          clientId: authProvider.clientId,
          clientSecret: authProvider.clientSecret,
          logoutUrl: authProvider.logoutUrl,
          loading: false
        });
      });
    } else {
      this.setState({
        loading: false
      });
    }
  }

  onSubmit = (e: any) => {
    e.preventDefault();
    this.entity.name = this.state.name;
    this.entity.providerType = this.state.providerType;
    this.entity.authUrl = this.state.authUrl;
    this.entity.tokenUrl = this.state.tokenUrl;
    this.entity.authStyle = this.state.authStyle;
    this.entity.scopes = this.state.scopes;
    this.entity.userInfoUrl = this.state.userInfoUrl;
    this.entity.userInfoEmailField = this.state.userInfoEmailField;
    this.entity.clientId = this.state.clientId;
    this.entity.clientSecret = this.state.clientSecret;
    this.entity.logoutUrl = this.state.logoutUrl;
    this.entity.save().then(() => {
      this.props.router.push("/settings/auth-providers/" + this.entity.id);
      this.setState({
        saved: true
      });
    });
  }

  deleteItem = () => {
    if (window.confirm(this.props.t("confirmDeleteAuthProvider"))) {
      this.entity.delete().then(() => {
        this.setState({ goBack: true });
      });
    }
  }

  templateGoogle = () => {
    this.setState({
      name: "Google",
      providerType: 1,
      authUrl: "https://accounts.google.com/o/oauth2/auth",
      tokenUrl: "https://oauth2.googleapis.com/token",
      authStyle: 1,
      scopes: "https://www.googleapis.com/auth/userinfo.email",
      userInfoUrl: "https://www.googleapis.com/oauth2/v3/userinfo",
      userInfoEmailField: "email",
      logoutUrl: ""
    });
  }

  templateMicrosoft = () => {
    this.setState({
      name: "Microsoft",
      providerType: 1,
      authUrl: "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
      tokenUrl: "https://login.microsoftonline.com/common/oauth2/v2.0/token",
      authStyle: 1,
      scopes: "openid,email",
      userInfoUrl: "https://graph.microsoft.com/oidc/userinfo",
      userInfoEmailField: "email",
      logoutUrl: "https://login.microsoftonline.com/common/oauth2/v2.0/logout?post_logout_redirect_uri={logoutRedirectUri}"
    });
  }

  templateKeycloak = () => {
    this.setState({
      name: "Keycloak",
      providerType: 1,
      authUrl: "https://keycloakhost.sample/auth/realms/master/protocol/openid-connect/auth",
      tokenUrl: "https://keycloakhost.sample/auth/realms/master/protocol/openid-connect/token",
      authStyle: 1,
      scopes: "openid,email",
      userInfoUrl: "https://keycloakhost.sample/auth/realms/master/protocol/openid-connect/userinfo",
      userInfoEmailField: "email",
      logoutUrl: "https://keycloakhost.sample/auth/realms/master/protocol/openid-connect/logout?post_logout_redirect_uri={logoutRedirectUri}"
    });
  }

  render() {
    if (this.state.goBack) {
      this.props.router.push(`/settings`);
      return <></>
    }

    let backButton = <Link href="/settings" className="btn btn-sm btn-outline-secondary"><IconBack className="feather" /> {this.props.t("back")}</Link>;
    let buttons = backButton;

    if (this.state.loading) {
      return (
        <FullLayout headline={this.props.t("editAuthProvider")} buttons={buttons}>
          <Loading />
        </FullLayout>
      );
    }

    let hint = <></>;
    if (this.state.saved) {
      hint = <Alert variant="success">{this.props.t("entryUpdated")}</Alert>
    }
    
    let urlInfo = <></>;
    let buttonDelete = <Button className="btn-sm" variant="outline-secondary" onClick={this.deleteItem}><IconDelete className="feather" /> {this.props.t("delete")}</Button>;
    let buttonSave = <Button className="btn-sm" variant="outline-secondary" type="submit" form="form"><IconSave className="feather" /> {this.props.t("save")}</Button>;
    if (this.entity.id) {
      buttons = <>{backButton} {buttonDelete} {buttonSave}</>;
      urlInfo = (
        <Form.Group as={Row}>
          <Form.Label column sm="2">Callback URL</Form.Label>
          <Col sm="9">
            <Form.Control plaintext={true} readOnly={true} onClick={(e: any) => e.target.select()} defaultValue={`/auth/${this.entity.id}/callback`} />
          </Col>
        </Form.Group>
      );
    } else {
      buttons = <>{backButton} {buttonSave}</>;
    }
    return (
      <FullLayout headline={this.props.t("editAuthProvider")} buttons={buttons}>
        <Form onSubmit={this.onSubmit} id="form">
          {hint}
          <Form.Group as={Row}>
            <Form.Label column sm="2">{this.props.t("name")}</Form.Label>
            <Col sm="9">
              <Form.Control type="text" placeholder="Name" value={this.state.name} onChange={(e: any) => this.setState({ name: e.target.value })} required={true} />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">{this.props.t("type")}</Form.Label>
            <Col sm="9">
              <Form.Select value={this.state.providerType} onChange={(e: any) => this.setState({ providerType: parseInt(e.target.value) })} required={true}>
                <option value="0">({this.props.t("pleaseSelect")})</option>
                <option value="1">OAuth 2</option>
              </Form.Select>
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">Auth URL</Form.Label>
            <Col sm="9">
              <Form.Control type="url" placeholder="https://..." value={this.state.authUrl} onChange={(e: any) => this.setState({ authUrl: e.target.value })} required={true} />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">Token URL</Form.Label>
            <Col sm="9">
              <Form.Control type="url" placeholder="https://..." value={this.state.tokenUrl} onChange={(e: any) => this.setState({ tokenUrl: e.target.value })} required={true} />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">Auth Style</Form.Label>
            <Col sm="9">
              <Form.Select value={this.state.authStyle} onChange={(e: any) => this.setState({ authStyle: parseInt(e.target.value) })} required={true}>
                <option value="0">{this.props.t("automatic")}</option>
                <option value="1">Parameter (HTTP POST body)</option>
                <option value="2">Header (HTTP Basic Authorization)</option>
              </Form.Select>
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">Scopes</Form.Label>
            <Col sm="9">
              <Form.Control type="text" placeholder="scope1,scope2,..." value={this.state.scopes} onChange={(e: any) => this.setState({ scopes: e.target.value })} required={true} />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">Client ID</Form.Label>
            <Col sm="9">
              <Form.Control type="text" placeholder="Client ID" value={this.state.clientId} onChange={(e: any) => this.setState({ clientId: e.target.value })} required={true} />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">Client Secret</Form.Label>
            <Col sm="9">
              <Form.Control type="text" placeholder="Client Secret" value={this.state.clientSecret} onChange={(e: any) => this.setState({ clientSecret: e.target.value })} required={true} />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">Userinfo URL</Form.Label>
            <Col sm="9">
              <Form.Control type="url" placeholder="https://..." value={this.state.userInfoUrl} onChange={(e: any) => this.setState({ userInfoUrl: e.target.value })} required={true} />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">{this.props.t("userinfoEmailField")}</Form.Label>
            <Col sm="9">
              <Form.Control type="text" placeholder="email" value={this.state.userInfoEmailField} onChange={(e: any) => this.setState({ userInfoEmailField: e.target.value })} required={true} />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">Logout URL</Form.Label>
            <Col sm="9">
              <Form.Control type="url" placeholder="https://..." value={this.state.logoutUrl} onChange={(e: any) => this.setState({ logoutUrl: e.target.value })} />
            </Col>
          </Form.Group>
          {urlInfo}
          <Form.Group as={Row}>
            <Form.Label column sm="2">{this.props.t("templates")}</Form.Label>
            <Col sm="9">
              <ButtonGroup>
                <Button variant="outline-secondary" onClick={this.templateGoogle}>Google</Button>
                <Button variant="outline-secondary" onClick={this.templateMicrosoft}>Microsoft</Button>
                <Button variant="outline-secondary" onClick={this.templateKeycloak}>Keycloak</Button>
              </ButtonGroup>
            </Col>
          </Form.Group>
        </Form>
      </FullLayout>
    );
  }
}

export default withTranslation(['admin'])(withReadyRouter(EditAuthProvider as any));
