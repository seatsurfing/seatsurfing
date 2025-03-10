import React from 'react';
import { Form, Col, Row, Button, Alert, InputGroup } from 'react-bootstrap';
import { ChevronLeft as IconBack, Save as IconSave, Trash2 as IconDelete } from 'react-feather';
import { Ajax, Domain, Organization, User } from 'seatsurfing-commons';
import { NextRouter } from 'next/router';
import { WithTranslation, withTranslation } from 'next-i18next';
import FullLayout from '@/components/FullLayout';
import Loading from '@/components/Loading';
import Link from 'next/link';
import withReadyRouter from '@/components/withReadyRouter';

interface State {
  loading: boolean
  submitting: boolean
  saved: boolean
  error: boolean
  goBack: boolean
  name: string
  firstname: string
  lastname: string
  email: string
  language: string
  domain: string
  password: string
}

interface Props extends WithTranslation {
  router: NextRouter
}

class EditOrganization extends React.Component<Props, State> {
  entity: Organization = new Organization();

  constructor(props: any) {
    super(props);
    this.state = {
      loading: true,
      submitting: false,
      saved: false,
      error: false,
      goBack: false,
      name: "",
      firstname: "",
      lastname: "",
      email: "",
      language: "de",
      domain: "",
      password: "",
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
      Organization.get(id).then(org => {
        this.entity = org;
        this.setState({
          name: org.name,
          firstname: org.contactFirstname,
          lastname: org.contactLastname,
          email: org.contactEmail,
          language: org.language,
          loading: false,
        });
      });
    } else {
      this.setState({ loading: false });
    }
  }

  onSubmit = (e: any) => {
    e.preventDefault();
    this.setState({
      error: false,
      saved: false
    });
    this.entity.name = this.state.name;
    this.entity.contactFirstname = this.state.firstname;
    this.entity.contactLastname = this.state.lastname;
    this.entity.contactEmail = this.state.email;
    this.entity.language = this.state.language;
    let createUser = (!this.entity.id);
    this.entity.save().then(() => {
      if (createUser) {
        Domain.add(this.entity.id, this.state.domain).then(() => {
          let user = new User();
          user.organizationId = this.entity.id;
          user.email = this.state.email;
          user.password = this.state.password;
          user.requirePassword = true;
          user.role = 20;
          user.save().then(() => {
            this.props.router.push("/organizations/" + this.entity.id);
            this.setState({ saved: true });
          });
        });
      } else {
        this.props.router.push("/organizations/" + this.entity.id);
        this.setState({ saved: true });
      }
    }).catch(() => {
      this.setState({ error: true });
    });
  }

  deleteItem = () => {
    if (window.confirm(this.props.t("confirmDeleteOrg"))) {
      this.entity.delete().then(() => {
        this.setState({ goBack: true });
      });
    }
  }

  render() {
    if (this.state.goBack) {
      this.props.router.push('/organizations');
      return <></>
    }

    let backButton = <Link href="/organizations" className="btn btn-sm btn-outline-secondary"><IconBack className="feather" /> {this.props.t("back")}</Link>;
    let buttons = backButton;

    if (this.state.loading) {
      return (
        <FullLayout headline={this.props.t("editOrg")} buttons={buttons}>
          <Loading />
        </FullLayout>
      );
    }

    let hint = <></>;
    if (this.state.saved) {
      hint = <Alert variant="success">{this.props.t("entryUpdated")}</Alert>
    } else if (this.state.error) {
      hint = <Alert variant="danger">{this.props.t("errorSave")}</Alert>
    }

    let buttonDelete = <Button className="btn-sm" variant="outline-secondary" onClick={this.deleteItem} disabled={false}><IconDelete className="feather" /> {this.props.t("delete")}</Button>;
    let buttonSave = <Button className="btn-sm" variant="outline-secondary" type="submit" form="form"><IconSave className="feather" /> {this.props.t("save")}</Button>;
    if (this.entity.id) {
      buttons = <>{backButton} {buttonDelete} {buttonSave}</>;
    } else {
      buttons = <>{backButton} {buttonSave}</>;
    }

    let countries = ["BE", "BG", "DK", "DE", "EE", "FJ", "FR", "GR", "IE", "IL", "IT", "HR", "LV", "LT", "LU", "MT", "NL", "AT", "PL", "PT", "RO", "SE", "SK", "SI", "ES", "CY", "CZ", "HU"];
    let languages = ["de", "en", "he"];

    let adminSection = <></>;
    if (!this.entity.id) {
      adminSection = (
        <>
          <Form.Group as={Row}>
            <Form.Label column sm="6" className="lead text-uppercase">{this.props.t("domain")}</Form.Label>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">{this.props.t("domain")}</Form.Label>
            <Col sm="4">
              <Form.Control type="text" placeholder={this.props.t("yourDomainPlaceholder")} value={this.state.domain} onChange={(e: any) => this.setState({ domain: e.target.value })} required={true} />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="6" className="lead text-uppercase">{this.props.t("admin")}</Form.Label>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">{this.props.t("password")}</Form.Label>
            <Col sm="4">
              <Form.Control type="password" value={this.state.password} onChange={(e: any) => this.setState({ password: e.target.value })} required={true} minLength={8} />
            </Col>
          </Form.Group>
        </>
      );
    }

    return (
      <FullLayout headline={this.props.t("editOrg")} buttons={buttons}>
        <Form onSubmit={this.onSubmit} id="form">
          {hint}
          <Form.Group as={Row}>
            <Form.Label column sm="2">{this.props.t("org")}</Form.Label>
            <Col sm="4">
              <Form.Control type="text" value={this.state.name} onChange={(e: any) => this.setState({ name: e.target.value })} required={true} autoFocus={true} />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">{this.props.t("language")}</Form.Label>
            <Col sm="4">
              <Form.Select value={this.state.language} onChange={(e: any) => this.setState({ language: e.target.value })} required={true}>
                {languages.map(lc => <option key={lc}>{lc}</option>)}
              </Form.Select>
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="6" className="lead text-uppercase">{this.props.t("primaryContact")}</Form.Label>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">{this.props.t("firstname")}</Form.Label>
            <Col sm="4">
              <Form.Control type="text" value={this.state.firstname} onChange={(e: any) => this.setState({ firstname: e.target.value })} required={true} />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">{this.props.t("lastname")}</Form.Label>
            <Col sm="4">
              <Form.Control type="text" value={this.state.lastname} onChange={(e: any) => this.setState({ lastname: e.target.value })} required={true} />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">{this.props.t("emailAddress")}</Form.Label>
            <Col sm="4">
              <Form.Control type="email" value={this.state.email} onChange={(e: any) => this.setState({ email: e.target.value })} required={true} />
            </Col>
          </Form.Group>
          {adminSection}
        </Form>
      </FullLayout>
    );
  }
}

export default withTranslation(['admin'])(withReadyRouter(EditOrganization as any));
