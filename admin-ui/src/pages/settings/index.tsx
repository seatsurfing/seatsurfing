import React from 'react';
import { User, Organization, AuthProvider, Settings as OrgSettings, Domain, Ajax, AjaxCredentials } from 'seatsurfing-commons';
import { Form, Col, Row, Table, Button, Alert, InputGroup, Popover, OverlayTrigger, Badge } from 'react-bootstrap';
import { Plus as IconPlus, Save as IconSave } from 'react-feather';
import { NextRouter } from 'next/router';
import { WithTranslation, withTranslation } from 'next-i18next';
import FullLayout from '@/components/FullLayout';
import Link from 'next/link';
import Loading from '@/components/Loading';
import withReadyRouter from '@/components/withReadyRouter';

interface State {
  allowAnyUser: boolean
  defaultTimezone: string
  confluenceServerSharedSecret: string
  customLogoUrl: string
  maxBookingsPerUser: number
  maxConcurrentBookingsPerUser: number
  maxDaysInAdvance: number
  enableMaxHoursBeforeDelete: boolean
  maxHoursBeforeDelete: number
  maxHoursPartiallyBooked: number
  maxHoursPartiallyBookedEnabled: boolean
  maxBookingDurationHours: number
  minBookingDurationHours: number
  dailyBasisBooking: boolean
  noAdminRestrictions: boolean
  showNames: boolean
  allowBookingNonExistUsers: boolean
  subscriptionActive: boolean
  subscriptionMaxUsers: number
  allowOrgDelete: boolean
  selectedAuthProvider: string
  disableBuddies: boolean
  loading: boolean
  submitting: boolean
  saved: boolean
  error: boolean
  newDomain: string
  domains: Domain[]
}

interface Props extends WithTranslation {
  router: NextRouter
}

class Settings extends React.Component<Props, State> {
  org: Organization | null;
  authProviders: AuthProvider[];
  timezones: string[];

  constructor(props: any) {
    super(props);
    this.org = null;
    this.authProviders = [];
    this.timezones = [];
    this.state = {
      allowAnyUser: true,
      defaultTimezone: "",
      confluenceServerSharedSecret: "",
      customLogoUrl: "",
      maxBookingsPerUser: 0,
      maxConcurrentBookingsPerUser: 0,
      maxBookingDurationHours: 0,
      minBookingDurationHours: 0,
      maxDaysInAdvance: 0,
      enableMaxHoursBeforeDelete: false,
      maxHoursBeforeDelete: 0,
      maxHoursPartiallyBooked: 0,
      maxHoursPartiallyBookedEnabled: false,
      dailyBasisBooking: false,
      noAdminRestrictions: false,
      showNames: false,
      allowBookingNonExistUsers: false,
      subscriptionActive: false,
      subscriptionMaxUsers: 0,
      allowOrgDelete: false,
      selectedAuthProvider: "",
      disableBuddies: false,
      loading: true,
      submitting: false,
      saved: false,
      error: false,
      newDomain: "",
      domains: []
    };
  }

  componentDidMount = () => {
    if (!Ajax.CREDENTIALS.accessToken) {
      this.props.router.push("/login");
      return;
    }
    let promises = [
      this.loadSettings(),
      this.loadItems(),
      this.loadAuthProviders(),
      this.loadTimezones(),
    ];
    Promise.all(promises).then(() => {
      this.setState({ loading: false });
    });
  }

  loadItems = async (): Promise<void> => {
    return User.getSelf().then(user => {
      return Organization.get(user.organizationId).then(org => {
        this.org = org;
        return Domain.list(org.id).then(domains => {
          this.setState({
            domains: domains,
          });
        });
      });
    });
  }

  loadAuthProviders = async (): Promise<void> => {
    return AuthProvider.list().then(list => {
      this.authProviders = list;
    });
  }

  loadSettings = async (): Promise<void> => {
    return OrgSettings.list().then(settings => {
      let state: any = {};
      settings.forEach(s => {
        if (s.name === "allow_any_user") state.allowAnyUser = (s.value === "1");
        if (s.name === "default_timezone") state.defaultTimezone = s.value;
        if (s.name === "confluence_server_shared_secret") state.confluenceServerSharedSecret = s.value;
        if (s.name === "custom_logo_url") state.customLogoUrl = s.value;
        if (s.name === "max_bookings_per_user") state.maxBookingsPerUser = window.parseInt(s.value);
        if (s.name === "max_concurrent_bookings_per_user") state.maxConcurrentBookingsPerUser = window.parseInt(s.value);
        if (s.name === "max_days_in_advance") state.maxDaysInAdvance = window.parseInt(s.value);
        if (s.name === "enable_max_hours_before_delete") state.enableMaxHoursBeforeDelete = window.parseInt(s.value);
        if (s.name === "max_hours_before_delete") state.maxHoursBeforeDelete = window.parseInt(s.value);
        if (s.name === "max_booking_duration_hours") state.maxBookingDurationHours = window.parseInt(s.value);
        if (s.name === "min_booking_duration_hours") state.minBookingDurationHours = window.parseInt(s.value);
        if (s.name === "daily_basis_booking") state.dailyBasisBooking = (s.value === "1");
        if (s.name === "no_admin_restrictions") state.noAdminRestrictions = (s.value === "1");
        if (s.name === "show_names") state.showNames = (s.value === "1");
        if (s.name === "allow_booking_nonexist_users") state.allowBookingNonExistUsers = (s.value === "1");
        if (s.name === "subscription_active") state.subscriptionActive = (s.value === "1");
        if (s.name === "subscription_max_users") state.subscriptionMaxUsers = window.parseInt(s.value);
        if (s.name === "disable_buddies") state.disableBuddies = (s.value === "1");
        if (s.name === "max_hours_partially_booked_enabled") state.maxHoursPartiallyBookedEnabled = (s.value === "1");
        if (s.name === "max_hours_partially_booked") state.maxHoursPartiallyBooked = window.parseInt(s.value);
        if (s.name === "_sys_org_signup_delete") state.allowOrgDelete = (s.value === "1");
      });
      if (state.dailyBasisBooking && (state.maxBookingDurationHours%24 !== 0)) {
        state.maxBookingDurationHours += (24-state.maxBookingDurationHours%24);
      }
      this.setState({
        ...this.state,
        ...state
      });
    });
  }

  loadTimezones = async (): Promise<void> => {
    return Ajax.get("/setting/timezones").then(res => {
      this.timezones = res.json;
    });
  }

  onSubmit = (e: any) => {
    e.preventDefault();
    this.setState({
      submitting: true,
      saved: false,
      error: false
    });
    let payload = [
      new OrgSettings("allow_any_user", this.state.allowAnyUser ? "1" : "0"),
      new OrgSettings("default_timezone", this.state.defaultTimezone),
      new OrgSettings("confluence_server_shared_secret", this.state.confluenceServerSharedSecret),
      new OrgSettings("custom_logo_url", this.state.customLogoUrl),
      new OrgSettings("daily_basis_booking", this.state.dailyBasisBooking ? "1" : "0"),
      new OrgSettings("no_admin_restrictions", this.state.noAdminRestrictions  ? "1" : "0"),
      new OrgSettings("show_names", this.state.showNames ? "1" : "0"),
      new OrgSettings("allow_booking_nonexist_users", this.state.allowBookingNonExistUsers ? "1" : "0"),
      new OrgSettings("disable_buddies", this.state.disableBuddies ? "1" : "0"),
      new OrgSettings("max_bookings_per_user", this.state.maxBookingsPerUser.toString()),
      new OrgSettings("max_concurrent_bookings_per_user", this.state.maxConcurrentBookingsPerUser.toString()),
      new OrgSettings("max_days_in_advance", this.state.maxDaysInAdvance.toString()),
      new OrgSettings("enable_max_hours_before_delete", this.state.enableMaxHoursBeforeDelete ? "1" : "0"),
      new OrgSettings("max_hours_before_delete", this.state.maxHoursBeforeDelete.toString()),
      new OrgSettings("max_booking_duration_hours", this.state.maxBookingDurationHours.toString()),
      new OrgSettings("max_hours_partially_booked_enabled", this.state.maxHoursPartiallyBookedEnabled ? "1" : "0"),
      new OrgSettings("max_hours_partially_booked", this.state.maxHoursPartiallyBooked.toString()),
      new OrgSettings("min_booking_duration_hours", this.state.minBookingDurationHours.toString())
    ];
    OrgSettings.setAll(payload).then(() => {
      this.setState({
        submitting: false,
        saved: true
      });
    }).catch(() => {
      this.setState({
        submitting: false,
        error: true
      });
    });
  }

  onAuthProviderSelect = (e: AuthProvider) => {
    this.setState({ selectedAuthProvider: e.id });
  }

  getAuthProviderTypeLabel = (providerType: number): string => {
    switch (providerType) {
      case 1: return "OAuth 2";
      default: return "Unknown";
    }
  }

  renderAuthProviderItem = (e: AuthProvider) => {
    return (
      <tr key={e.id} onClick={() => this.onAuthProviderSelect(e)}>
        <td>{e.name}</td>
        <td>{this.getAuthProviderTypeLabel(e.providerType)}</td>
      </tr>
    );
  }

  verifyDomain = (domainName: string) => {
    document.body.click();
    this.state.domains.forEach(domain => {
      if (domain.domain === domainName) {
        domain.verify().then(() => {
          Domain.list(domain.organizationId).then(domains => this.setState({ domains: domains }));
        }).catch(e => {
          alert(this.props.t("errorValidateDomain", {domain: domainName}));
        })
      }
    });
  }

  isValidDomain = () => {
    if (this.state.newDomain.indexOf(".") < 3) {
      return false;
    }
    let lastIndex = this.state.newDomain.length - 3;
    if (lastIndex < 3) {
      lastIndex = 3;
    }
    if (this.state.newDomain.lastIndexOf(".") > lastIndex) {
      return false;
    }
    return true;
  }

  addDomain = () => {
    if (!this.isValidDomain() || !this.org) {
      return;
    }
    Domain.add(this.org.id, this.state.newDomain).then(() => {
      Domain.list(this.org ? this.org.id : "").then(domains => this.setState({ domains: domains }));
      this.setState({ newDomain: "" });
    }).catch(() => {
      alert(this.props.t("errorAddDomain"));
    });
  }

  setPrimaryDomain = (domainName: string) => {
    this.state.domains.forEach(domain => {
      if (domain.domain === domainName) {
        domain.setPrimary().then(() => {
          Domain.list(this.org ? this.org.id : "").then(domains => this.setState({ domains: domains }));
        });
      }
    });
  }

  removeDomain = (domainName: string) => {
    if (!window.confirm(this.props.t("confirmDeleteDomain", {domain: domainName}))) {
      return;
    }
    this.state.domains.forEach(domain => {
      if (domain.domain === domainName) {
        domain.delete().then(() => {
          Domain.list(this.org ? this.org.id : "").then(domains => this.setState({ domains: domains }));
        }).catch(() => alert(this.props.t("errorDeleteDomain")));
      }
    });
  }

  handleNewDomainKeyDown = (target: any) => {
    if (target.key === "Enter") {
      target.preventDefault();
      this.addDomain();
    }
  }

  deleteOrg = () => {
    if (window.confirm(this.props.t("confirmDeleteOrg"))) {
      if (window.confirm(this.props.t("confirmDeleteOrg2"))) {
        this.org?.delete().then(() => {
          Ajax.CREDENTIALS = new AjaxCredentials();
          Ajax.PERSISTER.deleteCredentialsFromSessionStorage().then(() => {
            window.location.href = "/admin/";
          });
        });
      }
    }
  }

  manageSubscription = () => {
    let windowRef = window.open();
    this.org?.getSubscriptionManagementURL().then(url => {
      if (windowRef) {
        windowRef.location.href = url;
      }
    }).catch(() => {
      if (windowRef) {
        windowRef?.close();
      }
      alert(this.props.t("errorTryAgain"));
    });
  }

  onDailyBasisBookingChange = (enabled: boolean) => {
    let maxBookingDurationHours: number = Number(this.state.maxBookingDurationHours);
    if (enabled && (maxBookingDurationHours%24 !== 0)) {
      maxBookingDurationHours += (24-maxBookingDurationHours%24);
    }
    this.setState({
      maxBookingDurationHours: maxBookingDurationHours,
      dailyBasisBooking: enabled
    });
  }

  render() {
    if (this.state.selectedAuthProvider) {
      this.props.router.push(`/settings/auth-providers/${this.state.selectedAuthProvider}`);
      return <></>
    }

    if (this.state.loading) {
      return (
        <FullLayout headline={this.props.t("settings")}>
          <Loading />
        </FullLayout>
      );
    }

    let domains = this.state.domains.map(domain => {
      let verify = <></>;
      let popoverId = "popover-domain-" + domain.domain;
      const popover = (
        <Popover id={popoverId}>
          <Popover.Header as="h3">{this.props.t("verifyDomain")}</Popover.Header>
          <Popover.Body>
            <div>{this.props.t("verifyDomainHowto", {domain: domain.domain})}</div>
            <div>&nbsp;</div>
            <div><strong>seatsurfing-verification={domain.verifyToken}</strong></div>
            <div>&nbsp;</div>
            <Button variant="primary" size="sm" onClick={() => this.verifyDomain(domain.domain)}>{this.props.t("verifyNow")}</Button>
          </Popover.Body>
        </Popover>
      );
      if (!domain.active) {
        verify = (
          <OverlayTrigger trigger="click" placement="auto" overlay={popover} rootClose={false}>
            <Button variant="primary" size="sm">{this.props.t("verify")}</Button>
          </OverlayTrigger>
        );
      }
      let key = "domain-" + domain.domain;
      return (
        <Form.Group key={key} className="domain-row">
          {domain.domain}
            &nbsp;
          <Badge hidden={!domain.primary}>Primary</Badge>
            &nbsp;
          <Button variant="secondary" size="sm" hidden={domain.primary} onClick={() => this.setPrimaryDomain(domain.domain)}>Primary</Button>
            &nbsp;
          <Button variant="danger" size="sm" onClick={() => this.removeDomain(domain.domain)}>{this.props.t("remove")}</Button>
            &nbsp;
          {verify}
        </Form.Group>
      );
    });

    let authProviderRows = this.authProviders.map(item => this.renderAuthProviderItem(item));
    let authProviderTable = <p>{this.props.t("noRecords")}</p>;
    if (authProviderRows.length > 0) {
      authProviderTable = (
        <Table striped={true} hover={true} className="clickable-table">
          <thead>
            <tr>
              <th>{this.props.t("name")}</th>
              <th>{this.props.t("type")}</th>
            </tr>
          </thead>
          <tbody>
            {authProviderRows}
          </tbody>
        </Table>
      );
    }

    let dangerZone = (
      <></>
    );
    if (this.state.allowOrgDelete) {
      dangerZone = (
        <>
          <div className="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
            <h1 className="h2">{this.props.t("dangerZone")}</h1>
          </div>
          <Button className="btn btn-danger" onClick={this.deleteOrg}>{this.props.t("deleteOrg")}</Button>
        </>
      );
    }

    let hint = <></>;
    if (this.state.saved) {
      hint = <Alert variant="success">{this.props.t("entryUpdated")}</Alert>
    } else if (this.state.error) {
      hint = <Alert variant="danger">{this.props.t("errorSave")}</Alert>
    }

    let buttonSave = <Button className="btn-sm" variant="outline-secondary" type="submit" form="form"><IconSave className="feather" /> {this.props.t("save")}</Button>;

    return (
      <FullLayout headline={this.props.t("settings")} buttons={buttonSave}>
        <Form onSubmit={this.onSubmit} id="form">
          {hint}
          <Form.Group as={Row}>
            <Form.Label column sm="2">{this.props.t("org")}</Form.Label>
            <Col sm="4">
              <Form.Control plaintext={true} readOnly={true} defaultValue={this.org?.name} />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">{this.props.t("orgId")}</Form.Label>
            <Col sm="4">
              <Form.Control plaintext={true} readOnly={true} defaultValue={this.org?.id} />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Col sm="6">
              <Form.Check type="checkbox" id="check-allowAnyUser" label={this.props.t("allowAnyUser")} checked={this.state.allowAnyUser} onChange={(e: any) => this.setState({ allowAnyUser: e.target.checked })} />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">{this.props.t("customLogoUrl")}</Form.Label>
            <Col sm="4">
              <Form.Control type="url" value={this.state.customLogoUrl} onChange={(e: any) => this.setState({ customLogoUrl: e.target.value })} />
              <Form.Text className="text-muted">{this.props.t("customLogoUrlHint")}</Form.Text>
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">{this.props.t("maxBookingsPerUser")}</Form.Label>
            <Col sm="4">
              <Form.Control type="number" value={this.state.maxBookingsPerUser} onChange={(e: any) => this.setState({ maxBookingsPerUser: e.target.value })} min="1" max="9999" />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">{this.props.t("maxConcurrentBookingsPerUser")}</Form.Label>
            <Col sm="4">
              <Form.Control type="number" value={this.state.maxConcurrentBookingsPerUser} onChange={(e: any) => this.setState({ maxConcurrentBookingsPerUser: e.target.value })} min="0" max="9999" />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">{this.props.t("maxDaysInAdvance")}</Form.Label>
            <Col sm="4">
              <InputGroup>
                <Form.Control type="number" value={this.state.maxDaysInAdvance} onChange={(e: any) => this.setState({ maxDaysInAdvance: e.target.value })} min="0" max="9999" />
                <InputGroup.Text>{this.props.t("days")}</InputGroup.Text>
              </InputGroup>
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">{this.props.t("maxHoursBeforeDelete")}</Form.Label>
            <Col sm="4">
              <InputGroup>
                <InputGroup.Checkbox id="check-maxHoursBeforeDelete" checked={this.state.enableMaxHoursBeforeDelete} onChange={(e: any) => this.setState({ enableMaxHoursBeforeDelete: e.target.checked })} />
                <Form.Control type="number" value={this.state.maxHoursBeforeDelete} onChange={(e: any) => this.setState({ maxHoursBeforeDelete: e.target.value })} min="0" max="9999" disabled={!this.state.enableMaxHoursBeforeDelete} />
              </InputGroup>
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">{this.props.t("maxHoursPartiallyBooked")}</Form.Label>
            <Col sm="4">
              <InputGroup>
                <InputGroup.Checkbox checked={this.state.maxHoursPartiallyBookedEnabled} onChange={(e: any) => this.setState({ maxHoursPartiallyBookedEnabled: e.target.checked })} />
                <Form.Control type="number" value={this.state.maxHoursPartiallyBooked} onChange={(e: any) => this.setState({ maxHoursPartiallyBooked: e.target.value })} min="0" max="9999" disabled={!this.state.maxHoursPartiallyBookedEnabled} />
                <InputGroup.Text>{this.props.t("hours")}</InputGroup.Text>
              </InputGroup>
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Col sm="6">
              <Form.Check type="checkbox" id="check-noAdminRestrictions" label={this.props.t("noAdminRestrictions")} checked={this.state.noAdminRestrictions} onChange={(e: any) => this.setState({ noAdminRestrictions: e.target.checked })} />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Col sm="6">
              <Form.Check type="checkbox" id="check-disableBuddies" label={this.props.t("disableBuddies")} checked={this.state.disableBuddies} onChange={(e: any) => this.setState({ disableBuddies: e.target.checked })} />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Col sm="6">
              <Form.Check type="checkbox" id="check-dailyBasisBooking" label={this.props.t("dailyBasisBooking")} checked={this.state.dailyBasisBooking} onChange={(e: any) => this.onDailyBasisBookingChange(e.target.checked)} />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">{this.props.t("maxBookingDurationHours")}</Form.Label>
            <Col sm="4">
              <InputGroup>
                <Form.Control type="number" value={this.state.maxBookingDurationHours} onChange={(e: any) => this.setState({ maxBookingDurationHours: e.target.value })} min="0" max="9999" />
                <InputGroup.Text>{this.props.t("hours")}</InputGroup.Text>
              </InputGroup>
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">{this.props.t("minBookingDurationHours")}</Form.Label>
            <Col sm="4">
              <InputGroup>
                <Form.Control type="number" value={this.state.minBookingDurationHours} onChange={(e: any) => this.setState({ minBookingDurationHours: e.target.value })} min="0" max="9999" />
                <InputGroup.Text>{this.props.t("hours")}</InputGroup.Text>
              </InputGroup>
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Col sm="6">
              <Form.Check type="checkbox" id="check-showNames" label={this.props.t("showNames")} checked={this.state.showNames} onChange={(e: any) => this.setState({ showNames: e.target.checked })} />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Col sm="6">
              <Form.Check type="checkbox" id="check-allowBookingNonExistUsers" label={this.props.t("allowBookingNonExistUsers")} checked={this.state.allowBookingNonExistUsers} onChange={(e: any) => this.setState({ allowBookingNonExistUsers: e.target.checked })} />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">{this.props.t("defaultTimezone")}</Form.Label>
            <Col sm="4">
              <Form.Select value={this.state.defaultTimezone} onChange={(e: any) => this.setState({ defaultTimezone: e.target.value })}>
                {this.timezones.map(tz => <option key={tz}>{tz}</option>)}
              </Form.Select>
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">{this.props.t("confluenceServerSharedSecret")}</Form.Label>
            <Col sm="4">
              <Form.Control type="text" value={this.state.confluenceServerSharedSecret} onChange={(e: any) => this.setState({ confluenceServerSharedSecret: e.target.value })} />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">{this.props.t("domains")}</Form.Label>
            <Col sm="4">
              {domains}
              <InputGroup size="sm">
                <Form.Control type="text" value={this.state.newDomain} onChange={(e: any) => this.setState({ newDomain: e.target.value })} placeholder={this.props.t("yourDomainPlaceholder")} onKeyDown={this.handleNewDomainKeyDown} />
                <Button variant="outline-secondary" onClick={this.addDomain} disabled={!this.isValidDomain()}>{this.props.t("addDomain")}</Button>
              </InputGroup>
            </Col>
          </Form.Group>
        </Form>
        <div className="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
          <h1 className="h2">{this.props.t("authProviders")}</h1>
          <div className="btn-toolbar mb-2 mb-md-0">
            <div className="btn-group me-2">
              <Link href="/settings/auth-providers/add" className="btn btn-sm btn-outline-secondary"><IconPlus className="feather" /> {this.props.t("add")}</Link>
            </div>
          </div>
        </div>
        {authProviderTable}
        {dangerZone}
      </FullLayout>
    );
  }
}

export default withTranslation(['admin'])(withReadyRouter(Settings as any));
