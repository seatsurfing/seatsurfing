import React from 'react';
import { Form, Col, Row, Button, Alert, InputGroup } from 'react-bootstrap';
import { ChevronLeft as IconBack, Save as IconSave, Trash2 as IconDelete } from 'react-feather';
import { User, Settings as OrgSettings, Domain, Ajax } from 'seatsurfing-commons';
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
  error: boolean
  goBack: boolean
  email: string
  requirePassword: boolean
  password: string
  changePassword: boolean
  role: number
  domain: string
}

interface Props extends WithTranslation {
  router: NextRouter
}

class EditUser extends React.Component<Props, State> {
  entity: User = new User();
  usersMax: number = 0;
  usersCur: number = -1;
  domains: Domain[] = [];
  adminUserRole: number = 0;

  constructor(props: any) {
    super(props);
    this.state = {
      loading: true,
      submitting: false,
      saved: false,
      error: false,
      goBack: false,
      email: "",
      requirePassword: false,
      password: "",
      changePassword: false,
      role: User.UserRoleUser,
      domain: ""
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
    let promises: Promise<any>[] = [
      OrgSettings.getOne("subscription_max_users"),
      User.getCount(),
      User.getSelf().then(me => {
        return Domain.list(me.organizationId).then(domains => {
          return [me, domains];
        });
      })
    ];
    const { id } = this.props.router.query;
    if (id && (typeof id === "string") && (id !== 'add')) {
      promises.push(User.get(id));
    }
    Promise.all(promises).then(values => {
      this.usersMax = window.parseInt(values[0]);
      this.usersCur = values[1];
      this.adminUserRole = values[2][0].role;
      this.domains = values[2][1];
      let selectedDomain = "";
      this.domains.forEach(domain => {
        if (!selectedDomain && domain.active) {
          selectedDomain = domain.domain;
        }
      });
      this.setState({
        domain: selectedDomain
      });
      if (values.length >= 4) {
        let user = values[3];
        let userDomain = user.email.substring(user.email.indexOf("@") + 1).toLowerCase();
        this.entity = user;
        this.setState({
          email: user.email.substring(0, user.email.indexOf("@")),
          requirePassword: user.requirePassword,
          role: user.role,
          domain: userDomain
        });
      }
      this.setState({
        loading: false
      });
    });
  }

  onSubmit = (e: any) => {
    e.preventDefault();
    this.setState({
      error: false,
      saved: false
    });
    this.entity.email = this.state.email + "@" + this.state.domain;
    this.entity.role = this.state.role;
    this.entity.save().then(() => {
      this.props.router.push("/users/" + this.entity.id);
      if (this.state.changePassword) {
        this.entity.setPassword(this.state.password).then(() => {
          this.setState({ saved: true });
        }).catch(() => {
          this.setState({ error: true });
        });
      } else {
        this.setState({ saved: true });
      }
    }).catch(() => {
      this.setState({ error: true });
    });
  }

  deleteItem = () => {
    if (window.confirm(this.props.t("confirmDeleteUser"))) {
      this.entity.delete().then(() => {
        this.setState({ goBack: true });
      });
    }
  }

  render() {
    if (this.state.goBack) {
      this.props.router.push('/users');
      return <></>
    }

    let backButton = <Link href="/users" className="btn btn-sm btn-outline-secondary"><IconBack className="feather" /> {this.props.t("back")}</Link>;
    let buttons = backButton;

    if (this.state.loading) {
      return (
        <FullLayout headline={this.props.t("editUser")} buttons={buttons}>
          <Loading />
        </FullLayout>
      );
    }

    if (this.usersCur >= this.usersMax && !this.entity.id) {
      return (
        <FullLayout headline={this.props.t("editUser")} buttons={buttons}>
          <p>{this.props.t("errorSubscriptionLimit")}</p>
          <Link href="/settings" className="btn btn-primary">{this.props.t("subscriptionManage")}</Link>
        </FullLayout>
      );
    }

    let hint = <></>;
    if (this.state.saved) {
      hint = <Alert variant="success">{this.props.t("entryUpdated")}</Alert>
    } else if (this.state.error) {
      hint = <Alert variant="danger">{this.props.t("errorSave")}</Alert>
    }

    let domainOptions = this.domains.map(domain => {
      return <option key={domain.domain} value={domain.domain} disabled={!domain.active}>@{domain.domain.toLowerCase()}</option>;
    });

    let buttonDelete = <Button className="btn-sm" variant="outline-secondary" onClick={this.deleteItem} disabled={false}><IconDelete className="feather" /> {this.props.t("delete")}</Button>;
    let buttonSave = <Button className="btn-sm" variant="outline-secondary" type="submit" form="form"><IconSave className="feather" /> {this.props.t("save")}</Button>;
    if (this.entity.id) {
      buttons = <>{backButton} {buttonDelete} {buttonSave}</>;
    } else {
      buttons = <>{backButton} {buttonSave}</>;
    }
    let changePasswordLabel = this.props.t("passwordLogin");
    if (this.entity.id) {
      changePasswordLabel = this.props.t("passwordChange");
    }
    let changePassword = (
      <Form.Group as={Row}>
        <Col sm="6">
          <Form.Check type="checkbox" id="check-changePassword" label={changePasswordLabel} checked={this.state.changePassword} onChange={(e: any) => this.setState({ changePassword: e.target.checked })} />
        </Col>
      </Form.Group>
    );
    let roleSelect = <></>;
    if (this.adminUserRole >= this.state.role) {
      roleSelect = (
        <Form.Select value={this.state.role} onChange={(e: any) => this.setState({ role: parseInt(e.target.value) })}>
          <option value={User.UserRoleUser}>{this.props.t("roleUser")}</option>
          { this.adminUserRole >= User.UserRoleSpaceAdmin ? <option value={User.UserRoleSpaceAdmin}>{this.props.t("roleSpaceAdmin")}</option> : <></> }
          { this.adminUserRole >= User.UserRoleOrgAdmin ? <option value={User.UserRoleOrgAdmin}>{this.props.t("roleOrgAdmin")}</option> : <></> }
          { this.adminUserRole >= User.UserRoleSuperAdmin ? <option value={User.UserRoleSuperAdmin}>{this.props.t("roleSuperAdmin")}</option> : <></> }
        </Form.Select>
      );
    } else {
      let role = this.props.t("roleUser");
      if (this.state.role === User.UserRoleSpaceAdmin) {
        role = this.props.t("roleSpaceAdmin");
      }
      if (this.state.role === User.UserRoleOrgAdmin) {
        role = this.props.t("roleOrgAdmin");
      }
      if (this.state.role === User.UserRoleSuperAdmin) {
        role = this.props.t("roleSuperAdmin");
      }
      roleSelect = <Form.Control plaintext={true} readOnly={true} defaultValue={role} />;
    }
    return (
      <FullLayout headline={this.props.t("editUser")} buttons={buttons}>
        <Form onSubmit={this.onSubmit} id="form">
          {hint}
          <Form.Group as={Row}>
            <Form.Label column sm="2">{this.props.t("emailAddress")}</Form.Label>
            <Col sm="4">
              <InputGroup>
                <Form.Control type="text" placeholder="max.mustermann" value={this.state.email} onChange={(e: any) => this.setState({ email: e.target.value })} required={true} />
                <Form.Select className="custom-select" value={this.state.domain} onChange={(e: any) => this.setState({ domain: e.target.value })}>
                  {domainOptions}
                </Form.Select>
              </InputGroup>
            </Col>
          </Form.Group>
          {changePassword}
          <Form.Group as={Row}>
            <Form.Label column sm="2">{this.props.t("password")}</Form.Label>
            <Col sm="4">
              <Form.Control type="password" value={this.state.password} onChange={(e: any) => this.setState({ password: e.target.value })} required={!this.entity.id || this.state.changePassword} disabled={!this.state.changePassword} minLength={8} />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">{this.props.t("role")}</Form.Label>
            <Col sm="4">
              {roleSelect}
            </Col>
          </Form.Group>
        </Form>
      </FullLayout>
    );
  }
}

export default withTranslation(['admin'])(withReadyRouter(EditUser as any));
