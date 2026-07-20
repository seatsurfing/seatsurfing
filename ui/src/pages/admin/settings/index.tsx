import React from "react";
import {
  Form,
  Col,
  Row,
  Table,
  Button,
  Alert,
  InputGroup,
  Popover,
  OverlayTrigger,
  Badge,
} from "react-bootstrap";
import {
  Plus as IconPlus,
  Save as IconSave,
  AlertTriangle as IconAlert,
  Check as IconCheck,
  RefreshCw as IconRefresh,
} from "react-feather";
import { NextRouter } from "next/router";
import FullLayout from "@/components/FullLayout";
import Link from "next/link";
import Loading from "@/components/Loading";
import withReadyRouter from "@/components/withReadyRouter";
import RuntimeConfig from "@/components/RuntimeConfig";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import PremiumFeatureIcon from "@/components/PremiumFeatureIcon";
import CloudFeatureHint from "@/components/CloudFeatureHint";
import UrlInput from "@/components/form/UrlInput";
import Domain from "@/types/Domain";
import Organization from "@/types/Organization";
import AuthProvider from "@/types/AuthProvider";
import Ajax from "@/util/Ajax";
import User from "@/types/User";
import OrgSettings from "@/types/Settings";

import CopyToClipboardButton from "@/components/CopyToClipboardButton";
import ReloadModal from "@/components/ReloadModal";
import Validation from "@/util/Validation";
import RendererUtils from "@/util/RendererUtils";
import UpdateChecker from "@/util/UpdateChecker";

interface State {
  allowAnyUser: boolean;
  defaultTimezone: string;
  confluenceServerSharedSecret: string;
  customLogoUrl: string;
  maxBookingsPerUser: number;
  maxConcurrentBookingsPerUser: number;
  maxDaysInAdvance: number;
  bookingRetentionEnabled: boolean;
  bookingRetentionDays: number;
  subjectDefault: number;
  enableMaxHoursBeforeDelete: boolean;
  maxHoursBeforeDelete: number;
  maxHoursPartiallyBooked: number;
  maxHoursPartiallyBookedEnabled: boolean;
  maxBookingDuration: number;
  minBookingDuration: number;
  targetUtilizationHoursPerWeek: number;
  dailyBasisBooking: boolean;
  noAdminRestrictions: boolean;
  showNames: boolean;
  allowBookingNonExistUsers: boolean;
  allowOrgDelete: boolean;
  selectedAuthProvider: string;
  disableBuddies: boolean;
  loading: boolean;
  submitting: boolean;
  showSavedModal: boolean;
  error: boolean;
  newDomain: string;
  domains: Domain[];
  latestVersion: any;
  featureNoUserLimit: boolean;
  featureCustomDomains: boolean;
  allowRecurringBookings: boolean;
  newUserDefaultMailNotification: boolean;
  enforceTOTP: number;
  kioskSecret: string;
  kioskModeEnabled: boolean;
  hideReports: boolean;
  hideStats: boolean;
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
}

class Settings extends React.Component<Props, State> {
  org: Organization | null;
  authProviders: AuthProvider[];
  timezones: string[];
  maxConcurrentBookingsPerUserLastValue: number;

  constructor(props: any) {
    super(props);
    this.org = null;
    this.authProviders = [];
    this.maxConcurrentBookingsPerUserLastValue = 1;
    this.timezones = [];
    this.state = {
      allowAnyUser: true,
      defaultTimezone: "",
      confluenceServerSharedSecret: "",
      customLogoUrl: "",
      maxBookingsPerUser: 0,
      maxConcurrentBookingsPerUser: 0,
      maxBookingDuration: 0,
      minBookingDuration: 0,
      targetUtilizationHoursPerWeek: 0,
      maxDaysInAdvance: 0,
      bookingRetentionEnabled: false,
      bookingRetentionDays: 0,
      subjectDefault: 0,
      enableMaxHoursBeforeDelete: false,
      maxHoursBeforeDelete: 0,
      maxHoursPartiallyBooked: 0,
      maxHoursPartiallyBookedEnabled: false,
      dailyBasisBooking: false,
      noAdminRestrictions: false,
      showNames: false,
      allowBookingNonExistUsers: false,
      allowOrgDelete: false,
      selectedAuthProvider: "",
      disableBuddies: false,
      loading: true,
      submitting: false,
      showSavedModal: false,
      error: false,
      newDomain: "",
      domains: [],
      latestVersion: null,
      featureNoUserLimit: false,
      featureCustomDomains: false,
      allowRecurringBookings: true,
      newUserDefaultMailNotification: false,
      enforceTOTP: Organization.ENFORCE_TOTP_DISABLED,
      kioskSecret: "",
      kioskModeEnabled: false,
      hideReports: false,
      hideStats: false,
    };
  }

  componentDidMount = () => {
    const promises = [
      this.loadSettings(),
      this.loadItems(),
      this.loadAuthProviders(),
      this.loadTimezones(),
      this.checkUpdates(),
    ];
    Promise.all(promises).then(() => {
      this.setState({ loading: false });
    });
  };

  loadItems = async (): Promise<void> => {
    return User.getSelf().then((user) => {
      return Organization.get(user.organizationId).then((org) => {
        this.org = org;
        return Domain.list(org.id).then((domains) => {
          this.setState({
            domains: domains,
          });
        });
      });
    });
  };

  checkUpdates = async (): Promise<void> => {
    if (RuntimeConfig.INFOS.cloudHosted) return;
    this.setState({ latestVersion: await UpdateChecker.check() });
  };

  loadAuthProviders = async (): Promise<void> => {
    return AuthProvider.list().then((list) => {
      this.authProviders = list;
    });
  };

  loadSettings = async (): Promise<void> => {
    return OrgSettings.list().then((settings) => {
      const state: any = {};
      settings.forEach((s) => {
        if (s.name === Organization.PREF_ALLOW_ANY_USER)
          state.allowAnyUser = s.value === "1";
        if (s.name === Organization.PREF_DEFAULT_TIMEZONE)
          state.defaultTimezone = s.value;
        if (s.name === Organization.PREF_CONFLUENCE_SERVER_SHARED_SECRET)
          state.confluenceServerSharedSecret = s.value;
        if (s.name === Organization.PREF_CUSTOM_LOGO_URL)
          state.customLogoUrl = s.value;
        if (s.name === Organization.PREF_MAX_BOOKINGS_PER_USER)
          state.maxBookingsPerUser = window.parseInt(s.value);
        if (s.name === Organization.PREF_MAX_CONCURRENT_BOOKINGS_PER_USER)
          state.maxConcurrentBookingsPerUser = window.parseInt(s.value);
        if (s.name === Organization.PREF_MAX_DAYS_IN_ADVANCE)
          state.maxDaysInAdvance = window.parseInt(s.value);
        if (s.name === Organization.PREF_BOOKING_RETENTION_ENABLED)
          state.bookingRetentionEnabled = window.parseInt(s.value);
        if (s.name === Organization.PREF_BOOKING_RETENTION_DAYS)
          state.bookingRetentionDays = window.parseInt(s.value);
        if (s.name === Organization.PREF_ENABLE_MAX_HOURS_BEFORE_DELETE)
          state.enableMaxHoursBeforeDelete = window.parseInt(s.value);
        if (s.name === Organization.PREF_MAX_HOURS_BEFORE_DELETE)
          state.maxHoursBeforeDelete = window.parseInt(s.value);
        if (s.name === Organization.PREF_DAILY_BASIS_BOOKING)
          state.dailyBasisBooking = s.value === "1";
        if (s.name === Organization.PREF_MAX_BOOKING_DURATION_HOURS)
          state.maxBookingDuration = window.parseInt(s.value);
        if (s.name === Organization.PREF_MIN_BOOKING_DURATION_HOURS)
          state.minBookingDuration = window.parseInt(s.value);
        if (s.name === Organization.PREF_TARGET_UTILIZATION_HOURS_PER_WEEK)
          state.targetUtilizationHoursPerWeek = window.parseInt(s.value);
        if (s.name === Organization.PREF_SUBJECT_DEFAULT)
          state.subjectDefault = window.parseInt(s.value);
        if (s.name === Organization.PREF_NO_ADMIN_RESTRICTIONS)
          state.noAdminRestrictions = s.value === "1";
        if (s.name === Organization.PREF_SHOW_NAMES)
          state.showNames = s.value === "1";
        if (s.name === Organization.PREF_ALLOW_BOOKING_NONEXIST_USERS)
          state.allowBookingNonExistUsers = s.value === "1";
        if (s.name === Organization.PREF_DISABLE_BUDDIES)
          state.disableBuddies = s.value === "1";
        if (s.name === Organization.PREF_MAX_HOURS_PARTIALLY_BOOKED_ENABLED)
          state.maxHoursPartiallyBookedEnabled = s.value === "1";
        if (s.name === Organization.PREF_MAX_HOURS_PARTIALLY_BOOKED)
          state.maxHoursPartiallyBooked = window.parseInt(s.value);
        if (s.name === Organization.PREF_FEATURE_NO_USER_LIMIT)
          state.featureNoUserLimit = s.value === "1";
        if (s.name === Organization.PREF_FEATURE_CUSTOM_DOMAINS)
          state.featureCustomDomains = s.value === "1";
        if (s.name === Organization.PREF_ALLOW_RECURRING_BOOKINGS)
          state.allowRecurringBookings = s.value === "1";
        if (s.name === Organization.PREF_NEW_USER_DEFAULT_MAIL_NOTIFICATION)
          state.newUserDefaultMailNotification = s.value === "1";
        if (s.name === Organization.PREF_ENFORCE_TOTP)
          state.enforceTOTP = window.parseInt(s.value);
        if (s.name === Organization.PREF_KIOSK_MODE_ENABLED)
          state.kioskModeEnabled = s.value === "1";
        if (s.name === Organization.PREF_KIOSK_ACCESS_SECRET)
          state.kioskSecret =
            s.value === "1" ? RendererUtils.SECRET_PLACEHOLDER : "";
        if (s.name === Organization.PREF_SYS_ORG_SIGNUP_DELETE)
          state.allowOrgDelete = s.value === "1";
        if (s.name === Organization.PREF_HIDE_REPORTS)
          state.hideReports = s.value === "1";
        if (s.name === Organization.PREF_HIDE_STATS)
          state.hideStats = s.value === "1";
      });

      if (state.maxConcurrentBookingsPerUser > 0) {
        this.maxConcurrentBookingsPerUserLastValue =
          state.maxConcurrentBookingsPerUser;
      }

      const convert = (value: number): number => {
        return state.dailyBasisBooking ? value / 24 : value;
      };
      state.minBookingDuration = convert(state.minBookingDuration);
      state.maxBookingDuration = convert(state.maxBookingDuration);
      state.targetUtilizationHoursPerWeek = convert(
        state.targetUtilizationHoursPerWeek,
      );

      this.setState({
        ...this.state,
        ...state,
      });
    });
  };

  generateKioskSecret = () => {
    this.setState({ kioskSecret: Validation.generatePassword(32, true) });
  };

  saveKioskSecret = (e: any) => {
    e.preventDefault();
    OrgSettings.setOne(
      Organization.PREF_KIOSK_ACCESS_SECRET,
      this.state.kioskSecret,
    )
      .then(() => {
        this.setState({
          kioskSecret: RendererUtils.SECRET_PLACEHOLDER,
        });
      })
      .catch(() => {
        this.setState({ error: true });
      });
  };

  loadTimezones = async (): Promise<void> => {
    return Ajax.get("/setting/timezones").then((res) => {
      this.timezones = res.json;
    });
  };

  onSubmit = async (e: any) => {
    e.preventDefault();
    this.setState({
      submitting: true,
      error: false,
    });
    const payload = [
      new OrgSettings(
        Organization.PREF_ALLOW_ANY_USER,
        this.state.allowAnyUser ? "1" : "0",
      ),
      new OrgSettings(
        Organization.PREF_DEFAULT_TIMEZONE,
        this.state.defaultTimezone,
      ),
      new OrgSettings(
        Organization.PREF_CONFLUENCE_SERVER_SHARED_SECRET,
        this.state.confluenceServerSharedSecret,
      ),
      new OrgSettings(
        Organization.PREF_CUSTOM_LOGO_URL,
        this.state.customLogoUrl,
      ),
      new OrgSettings(
        Organization.PREF_DAILY_BASIS_BOOKING,
        this.state.dailyBasisBooking ? "1" : "0",
      ),
      new OrgSettings(
        Organization.PREF_NO_ADMIN_RESTRICTIONS,
        this.state.noAdminRestrictions ? "1" : "0",
      ),
      new OrgSettings(
        Organization.PREF_SHOW_NAMES,
        this.state.showNames ? "1" : "0",
      ),
      new OrgSettings(
        Organization.PREF_ALLOW_BOOKING_NONEXIST_USERS,
        this.state.allowBookingNonExistUsers ? "1" : "0",
      ),
      new OrgSettings(
        Organization.PREF_DISABLE_BUDDIES,
        this.state.disableBuddies ? "1" : "0",
      ),
      new OrgSettings(
        Organization.PREF_MAX_BOOKINGS_PER_USER,
        this.state.maxBookingsPerUser.toString(),
      ),
      new OrgSettings(
        Organization.PREF_MAX_CONCURRENT_BOOKINGS_PER_USER,
        this.state.maxConcurrentBookingsPerUser.toString(),
      ),
      new OrgSettings(
        Organization.PREF_MAX_DAYS_IN_ADVANCE,
        this.state.maxDaysInAdvance.toString(),
      ),
      new OrgSettings(
        Organization.PREF_BOOKING_RETENTION_ENABLED,
        this.state.bookingRetentionEnabled ? "1" : "0",
      ),
      new OrgSettings(
        Organization.PREF_BOOKING_RETENTION_DAYS,
        this.state.bookingRetentionDays.toString(),
      ),
      new OrgSettings(
        Organization.PREF_ENABLE_MAX_HOURS_BEFORE_DELETE,
        this.state.enableMaxHoursBeforeDelete ? "1" : "0",
      ),
      new OrgSettings(
        Organization.PREF_MAX_HOURS_BEFORE_DELETE,
        this.state.maxHoursBeforeDelete.toString(),
      ),
      new OrgSettings(
        Organization.PREF_MAX_BOOKING_DURATION_HOURS,
        (
          Math.max(Math.round(this.state.maxBookingDuration), 1) *
          (this.state.dailyBasisBooking ? 24 : 1)
        ).toString(),
      ),
      new OrgSettings(
        Organization.PREF_MAX_HOURS_PARTIALLY_BOOKED_ENABLED,
        this.state.maxHoursPartiallyBookedEnabled ? "1" : "0",
      ),
      new OrgSettings(
        Organization.PREF_MAX_HOURS_PARTIALLY_BOOKED,
        this.state.maxHoursPartiallyBooked.toString(),
      ),
      new OrgSettings(
        Organization.PREF_MIN_BOOKING_DURATION_HOURS,
        (
          Math.round(this.state.minBookingDuration) *
          (this.state.dailyBasisBooking ? 24 : 1)
        ).toString(),
      ),
      new OrgSettings(
        Organization.PREF_TARGET_UTILIZATION_HOURS_PER_WEEK,
        (
          Math.max(Math.round(this.state.targetUtilizationHoursPerWeek), 1) *
          (this.state.dailyBasisBooking ? 24 : 1)
        ).toString(),
      ),
      new OrgSettings(
        Organization.PREF_ALLOW_RECURRING_BOOKINGS,
        this.state.allowRecurringBookings ? "1" : "0",
      ),
      new OrgSettings(
        Organization.PREF_NEW_USER_DEFAULT_MAIL_NOTIFICATION,
        this.state.newUserDefaultMailNotification ? "1" : "0",
      ),
      new OrgSettings(
        Organization.PREF_ENFORCE_TOTP,
        this.state.enforceTOTP.toString(),
      ),
      new OrgSettings(
        Organization.PREF_SUBJECT_DEFAULT,
        this.state.subjectDefault.toString(),
      ),
      new OrgSettings(
        Organization.PREF_KIOSK_MODE_ENABLED,
        this.state.kioskModeEnabled ? "1" : "0",
      ),
      new OrgSettings(
        Organization.PREF_HIDE_REPORTS,
        this.state.hideReports ? "1" : "0",
      ),
      new OrgSettings(
        Organization.PREF_HIDE_STATS,
        this.state.hideStats ? "1" : "0",
      ),
    ];
    try {
      await OrgSettings.setAll(payload);
      this.setState({
        submitting: false,
        showSavedModal: true,
      });
    } catch {
      this.setState({
        submitting: false,
        error: true,
      });
    }
  };

  onAuthProviderSelect = (e: AuthProvider) => {
    if (e.readOnly) {
      return;
    }
    this.setState({ selectedAuthProvider: e.id });
  };

  getAuthProviderTypeLabel = (providerType: number): string => {
    switch (providerType) {
      case 1:
        return "OAuth 2";
      default:
        return "Unknown";
    }
  };

  renderAuthProviderItem = (e: AuthProvider) => {
    return (
      <tr key={e.id} onClick={() => this.onAuthProviderSelect(e)}>
        <td>{e.name}</td>
        <td>{this.getAuthProviderTypeLabel(e.providerType)}</td>
      </tr>
    );
  };

  verifyDomain = (domainName: string) => {
    document.body.click();
    this.state.domains.forEach((domain) => {
      if (domain.domain === domainName) {
        domain
          .verify()
          .then(() => {
            Domain.list(domain.organizationId).then((domains) =>
              this.setState({ domains: domains }),
            );
          })
          .catch((e) => {
            alert(this.props.t("errorValidateDomain", { domain: domainName }));
          });
      }
    });
  };

  isValidDomain = () => {
    return Validation.isValidDomain(this.state.newDomain);
  };

  addDomain = () => {
    if (!this.isValidDomain() || !this.org) {
      return;
    }
    Domain.add(this.org.id, this.state.newDomain)
      .then(() => {
        Domain.list(this.org ? this.org.id : "").then((domains) =>
          this.setState({ domains: domains }),
        );
        this.setState({ newDomain: "" });
      })
      .catch(() => {
        alert(this.props.t("errorAddDomain"));
      });
  };

  setPrimaryDomain = (domainName: string) => {
    this.state.domains.forEach((domain) => {
      if (domain.domain === domainName) {
        domain.setPrimary().then(() => {
          Domain.list(this.org ? this.org.id : "").then((domains) =>
            this.setState({ domains: domains }),
          );
        });
      }
    });
  };

  removeDomain = async (domainName: string) => {
    if (
      !window.confirm(
        this.props.t("confirmDeleteDomain", { domain: domainName }),
      )
    ) {
      return;
    }
    const domain = this.state.domains.find((d) => d.domain === domainName);
    if (!domain) {
      return;
    }
    try {
      await domain.delete();
      const domains = await Domain.list(this.org ? this.org.id : "");
      this.setState({ domains: domains });
    } catch {
      alert(this.props.t("errorDeleteDomain"));
    }
  };

  handleNewDomainKeyDown = (target: any) => {
    if (target.key === "Enter") {
      target.preventDefault();
      this.addDomain();
    }
  };

  deleteOrg = () => {
    if (window.confirm(this.props.t("confirmDeleteOrgQuestion1"))) {
      if (window.confirm(this.props.t("confirmDeleteOrgQuestion2"))) {
        this.org?.delete().then((code) => {
          window.alert(
            this.props.t("confirmDeleteOrgConfirmMailSent", { code }),
          );
        });
      }
    }
  };

  onDailyBasisBookingChange = (enabled: boolean) => {
    const convert = (value: number): number => {
      return enabled ? value / 24 : value * 24;
    };
    this.setState({
      minBookingDuration: convert(this.state.minBookingDuration),
      maxBookingDuration: convert(this.state.maxBookingDuration),
      targetUtilizationHoursPerWeek: convert(
        this.state.targetUtilizationHoursPerWeek,
      ),
      dailyBasisBooking: enabled,
    });
  };

  render() {
    if (this.state.selectedAuthProvider) {
      this.props.router.push(
        `/admin/settings/auth-providers/${this.state.selectedAuthProvider}`,
      );
      return <></>;
    }

    if (this.state.loading) {
      return (
        <FullLayout headline={this.props.t("settings")}>
          <Loading />
        </FullLayout>
      );
    }

    let domains = this.state.domains.map((domain) => {
      let verify = <></>;
      let popoverId = "popover-domain-" + domain.domain;
      const popover = (
        <Popover id={popoverId}>
          <Popover.Header as="h3">
            {this.props.t("verifyDomain")}
          </Popover.Header>
          <Popover.Body>
            <div>
              {this.props.t("verifyDomainHowto", { domain: domain.domain })}
            </div>
            <div>&nbsp;</div>
            <div>
              <strong>seatsurfing-verification={domain.verifyToken}</strong>
            </div>
            <div>&nbsp;</div>
            <Button
              variant="primary"
              size="sm"
              onClick={() => this.verifyDomain(domain.domain)}
            >
              {this.props.t("verifyNow")}
            </Button>
          </Popover.Body>
        </Popover>
      );
      if (!domain.active) {
        verify = (
          <OverlayTrigger
            trigger="click"
            placement="auto"
            overlay={popover}
            rootClose={false}
          >
            <Button variant="primary" size="sm">
              {this.props.t("verify")}
            </Button>
          </OverlayTrigger>
        );
      }
      let accessibleCheckmark = (
        <IconAlert className="feather" color="orange" />
      );
      if (domain.accessible) {
        accessibleCheckmark = <IconCheck className="feather" color="green" />;
      }
      let key = "domain-" + domain.domain;
      return (
        <Form.Group key={key} className="domain-row">
          {domain.domain}
          &nbsp;
          {accessibleCheckmark}
          &nbsp;
          <Badge hidden={!domain.primary}>Primary</Badge>
          &nbsp;
          <Button
            variant="secondary"
            size="sm"
            hidden={domain.primary}
            disabled={!domain.active}
            onClick={() => this.setPrimaryDomain(domain.domain)}
          >
            Primary
          </Button>
          &nbsp;
          <Button
            variant="danger"
            size="sm"
            hidden={domain.domain.endsWith(".seatsurfing.app")}
            disabled={this.state.domains.length <= 1}
            onClick={() => this.removeDomain(domain.domain)}
          >
            {this.props.t("remove")}
          </Button>
          &nbsp;
          {verify}
        </Form.Group>
      );
    });
    let authProviderRows = this.authProviders.map((item) =>
      this.renderAuthProviderItem(item),
    );
    let authProviderTable = <p>{this.props.t("noRecords")}</p>;
    if (
      RuntimeConfig.INFOS.cloudHosted &&
      !RuntimeConfig.INFOS.subscriptionActive
    ) {
      authProviderTable = <CloudFeatureHint />;
    } else if (authProviderRows.length > 0) {
      authProviderTable = (
        <Table striped={true} hover={true} className="clickable-table">
          <thead>
            <tr>
              <th>{this.props.t("name")}</th>
              <th>{this.props.t("type")}</th>
            </tr>
          </thead>
          <tbody>{authProviderRows}</tbody>
        </Table>
      );
    } else if (!RuntimeConfig.INFOS.featureAuthProviders) {
      authProviderTable = <p>Feature not enabled.</p>;
    }

    let dangerZone = <></>;
    if (this.state.allowOrgDelete) {
      dangerZone = (
        <>
          <div className="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
            <h1 className="h2">{this.props.t("dangerZone")}</h1>
          </div>
          <Button className="btn btn-danger" onClick={this.deleteOrg}>
            {this.props.t("deleteOrg")}
          </Button>
        </>
      );
    }

    const hint = this.state.error ? (
      <Alert variant="danger">{this.props.t("errorSave")}</Alert>
    ) : (
      <></>
    );

    const buttonSave = (
      <Button
        className="btn-sm"
        variant="outline-secondary"
        type="submit"
        form="form"
      >
        <IconSave className="feather" /> {this.props.t("save")}
      </Button>
    );

    let updateHint = (
      <span className="form-control-plaintext">
        {process.env.NEXT_PUBLIC_PRODUCT_VERSION}
      </span>
    );
    if (this.state.latestVersion?.updateAvailable) {
      updateHint = (
        <span className="form-control-plaintext">
          {process.env.NEXT_PUBLIC_PRODUCT_VERSION}
          &nbsp;(
          <a
            href="https://github.com/seatsurfing/seatsurfing/releases"
            target="_blank"
            rel="noopener noreferrer"
          >
            upgrade to {this.state.latestVersion.version}
          </a>
          )
        </span>
      );
    } else if (this.state.latestVersion) {
      updateHint = (
        <span className="form-control-plaintext">
          {process.env.NEXT_PUBLIC_PRODUCT_VERSION} (up to date)
        </span>
      );
    }

    return (
      <FullLayout headline={this.props.t("settings")} buttons={buttonSave}>
        <Form onSubmit={this.onSubmit} id="form">
          {hint}
          <Form.Group as={Row}>
            <Form.Label column sm="2">
              {this.props.t("org")}
            </Form.Label>
            <Col sm="4">
              <p className="form-control-plaintext">
                {this.org?.name} (
                {this.org?.language
                  ? this.props.t("language-" + this.org.language)
                  : "—"}
                )
                <br />
                {this.org?.contactFirstname} {this.org?.contactLastname} (
                {this.org?.contactEmail})
              </p>
              <Link href={`/admin/settings/org`}>{this.props.t("edit")}</Link>
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">
              {this.props.t("orgId")}
            </Form.Label>
            <Col sm="4">
              <Form.Control
                plaintext={true}
                readOnly={true}
                defaultValue={this.org?.id}
              />
            </Col>
          </Form.Group>
          <Form.Group as={Row} hidden={RuntimeConfig.INFOS.cloudHosted}>
            <Form.Label column sm="2">
              Version
            </Form.Label>
            <Col sm="4">{updateHint}</Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2" htmlFor="input-customLogoUrl">
              {this.props.t("customLogoUrl")}
            </Form.Label>
            <Col sm="4">
              <UrlInput
                id="input-customLogoUrl"
                placeholder="https://…"
                value={this.state.customLogoUrl}
                onChange={(e: any) =>
                  this.setState({ customLogoUrl: e.target.value })
                }
              />
              <Form.Text className="text-muted">
                {this.props.t("customLogoUrlHint")}
              </Form.Text>
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Col sm="6">
              <Form.Check
                type="checkbox"
                id="check-showNames"
                label={this.props.t("showNames")}
                checked={this.state.showNames}
                onChange={(e: any) =>
                  this.setState({ showNames: e.target.checked })
                }
              />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Col sm="6">
              <Form.Check
                type="checkbox"
                id="check-disableBuddies"
                label={this.props.t("disableBuddies")}
                disabled={!this.state.showNames}
                checked={this.state.disableBuddies}
                onChange={(e: any) =>
                  this.setState({ disableBuddies: e.target.checked })
                }
              />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Col sm="6">
              <Form.Check
                type="checkbox"
                id="check-newUserDefaultMailNotification"
                label={this.props.t("newUserDefaultMailNotification")}
                checked={this.state.newUserDefaultMailNotification}
                onChange={(e: any) =>
                  this.setState({
                    newUserDefaultMailNotification: e.target.checked,
                  })
                }
              />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2" htmlFor="input-enforceTOTP">
              {this.props.t("enforceTOTP")}
            </Form.Label>
            <Col sm="4">
              <Form.Select
                id="input-enforceTOTP"
                value={this.state.enforceTOTP}
                onChange={(e: any) =>
                  this.setState({
                    enforceTOTP: window.parseInt(e.target.value),
                  })
                }
              >
                <option value={Organization.ENFORCE_TOTP_DISABLED}>
                  {this.props.t("disabled")}
                </option>
                <option value={Organization.ENFORCE_TOTP_ALL_USERS}>
                  {this.props.t("enforceTOTPAllUsers")}
                </option>
                <option value={Organization.ENFORCE_TOTP_ADMINS_ONLY}>
                  {this.props.t("enforceTOTPAdminsOnly")}
                </option>
              </Form.Select>
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2" htmlFor="input-defaultTimezone">
              {this.props.t("defaultTimezone")}
            </Form.Label>
            <Col sm="4">
              <Form.Select
                id="input-defaultTimezone"
                value={this.state.defaultTimezone}
                onChange={(e: any) =>
                  this.setState({ defaultTimezone: e.target.value })
                }
              >
                {this.timezones.map((tz) => (
                  <option key={tz}>{tz}</option>
                ))}
              </Form.Select>
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label
              column
              sm="2"
              htmlFor="input-confluenceServerSharedSecret"
            >
              {this.props.t("confluenceServerSharedSecret")}
            </Form.Label>
            <Col sm="4">
              <Form.Control
                id="input-confluenceServerSharedSecret"
                type="text"
                value={this.state.confluenceServerSharedSecret}
                onChange={(e: any) =>
                  this.setState({
                    confluenceServerSharedSecret: e.target.value,
                  })
                }
              />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2" htmlFor="input-newDomain">
              {this.props.t("domains")}
              <PremiumFeatureIcon />
            </Form.Label>
            <Col sm="4">
              {domains}
              <InputGroup size="sm" hidden={!this.state.featureCustomDomains}>
                <Form.Control
                  id="input-newDomain"
                  type="text"
                  value={this.state.newDomain}
                  onChange={(e: any) =>
                    this.setState({ newDomain: e.target.value })
                  }
                  placeholder={this.props.t("yourDomainPlaceholder")}
                  onKeyDown={this.handleNewDomainKeyDown}
                />
                <Button
                  variant="outline-secondary"
                  onClick={this.addDomain}
                  disabled={!this.isValidDomain()}
                >
                  {this.props.t("addDomain")}
                </Button>
              </InputGroup>
            </Col>
          </Form.Group>

          {/* BOOKINGS */}

          <div className="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
            <h4>{this.props.t("bookings")}</h4>
          </div>
          <Form.Group as={Row}>
            <Form.Label column sm="2" htmlFor="input-maxBookingsPerUser">
              {this.props.t("maxBookingsPerUser")}
            </Form.Label>
            <Col sm="4">
              <Form.Control
                id="input-maxBookingsPerUser"
                type="number"
                value={this.state.maxBookingsPerUser}
                onChange={(e: any) =>
                  this.setState({ maxBookingsPerUser: e.target.value })
                }
                required={true}
                min="1"
                max="9999"
              />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label
              column
              sm="2"
              htmlFor="input-maxConcurrentBookingsPerUser"
            >
              {this.props.t("maxConcurrentBookingsPerUser")}
            </Form.Label>
            <Col sm="4">
              <InputGroup>
                <InputGroup.Text>
                  <Form.Check.Input
                    type="checkbox"
                    id="check-maxConcurrentBookingsPerUserUnlimited"
                    checked={this.state.maxConcurrentBookingsPerUser === 0}
                    onChange={(e: any) => {
                      if (!e.target.checked) {
                        this.setState({
                          maxConcurrentBookingsPerUser:
                            this.maxConcurrentBookingsPerUserLastValue,
                        });
                        return;
                      }
                      if (this.state.maxConcurrentBookingsPerUser > 0) {
                        this.maxConcurrentBookingsPerUserLastValue =
                          this.state.maxConcurrentBookingsPerUser;
                      }
                      this.setState({ maxConcurrentBookingsPerUser: 0 });
                    }}
                    className="mt-0 me-2"
                  />
                  <Form.Check.Label htmlFor="check-maxConcurrentBookingsPerUserUnlimited">
                    {this.props.t("unlimited")}
                  </Form.Check.Label>
                </InputGroup.Text>
                <Form.Control
                  id="input-maxConcurrentBookingsPerUser"
                  type="number"
                  value={
                    this.state.maxConcurrentBookingsPerUser === 0
                      ? ""
                      : this.state.maxConcurrentBookingsPerUser
                  }
                  onChange={(e: any) => {
                    const value = window.parseInt(e.target.value);
                    if (value > 0) {
                      this.maxConcurrentBookingsPerUserLastValue = value;
                    }
                    this.setState({
                      maxConcurrentBookingsPerUser: window.parseInt(
                        e.target.value,
                      ),
                    });
                  }}
                  min="1"
                  max="9999"
                  disabled={this.state.maxConcurrentBookingsPerUser === 0}
                />
              </InputGroup>
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2" htmlFor="input-maxDaysInAdvance">
              {this.props.t("maxDaysInAdvance")}
            </Form.Label>
            <Col sm="4">
              <InputGroup>
                <Form.Control
                  id="input-maxDaysInAdvance"
                  type="number"
                  value={this.state.maxDaysInAdvance}
                  onChange={(e: any) =>
                    this.setState({ maxDaysInAdvance: e.target.value })
                  }
                  required={true}
                  min="0"
                  max="9999"
                />
                <InputGroup.Text>{this.props.t("days")}</InputGroup.Text>
              </InputGroup>
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2" htmlFor="input-bookingRetentionDays">
              {this.props.t("bookingRetention")}
            </Form.Label>
            <Col sm="4">
              <InputGroup>
                <InputGroup.Checkbox
                  id="check-bookingRetentionDays"
                  checked={this.state.bookingRetentionEnabled}
                  onChange={(e: any) =>
                    this.setState({
                      bookingRetentionEnabled: e.target.checked,
                    })
                  }
                />
                <Form.Control
                  id="input-bookingRetentionDays"
                  type="number"
                  value={this.state.bookingRetentionDays}
                  onChange={(e: any) =>
                    this.setState({ bookingRetentionDays: e.target.value })
                  }
                  min="30"
                  max="999"
                  disabled={!this.state.bookingRetentionEnabled}
                />
                <InputGroup.Text>{this.props.t("days")}</InputGroup.Text>
              </InputGroup>
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2" htmlFor="input-maxHoursBeforeDelete">
              {this.props.t("maxHoursBeforeDelete")}
            </Form.Label>
            <Col sm="4">
              <InputGroup>
                <InputGroup.Checkbox
                  id="check-maxHoursBeforeDelete"
                  checked={this.state.enableMaxHoursBeforeDelete}
                  onChange={(e: any) =>
                    this.setState({
                      enableMaxHoursBeforeDelete: e.target.checked,
                    })
                  }
                />
                <Form.Control
                  id="input-maxHoursBeforeDelete"
                  type="number"
                  value={this.state.maxHoursBeforeDelete}
                  onChange={(e: any) =>
                    this.setState({ maxHoursBeforeDelete: e.target.value })
                  }
                  min="0"
                  max="9999"
                  disabled={!this.state.enableMaxHoursBeforeDelete}
                />
              </InputGroup>
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2" htmlFor="input-maxHoursPartiallyBooked">
              {this.props.t("maxHoursPartiallyBooked")}
            </Form.Label>
            <Col sm="4">
              <InputGroup>
                <InputGroup.Checkbox
                  checked={this.state.maxHoursPartiallyBookedEnabled}
                  onChange={(e: any) =>
                    this.setState({
                      maxHoursPartiallyBookedEnabled: e.target.checked,
                    })
                  }
                />
                <Form.Control
                  id="input-maxHoursPartiallyBooked"
                  type="number"
                  value={this.state.maxHoursPartiallyBooked}
                  onChange={(e: any) =>
                    this.setState({ maxHoursPartiallyBooked: e.target.value })
                  }
                  min="0"
                  max="9999"
                  disabled={!this.state.maxHoursPartiallyBookedEnabled}
                />
                <InputGroup.Text>{this.props.t("hours")}</InputGroup.Text>
              </InputGroup>
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Col sm="6">
              <Form.Check
                type="checkbox"
                id="check-noAdminRestrictions"
                label={this.props.t("noAdminRestrictions")}
                checked={this.state.noAdminRestrictions}
                onChange={(e: any) =>
                  this.setState({ noAdminRestrictions: e.target.checked })
                }
              />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Col sm="6">
              <Form.Check
                type="checkbox"
                id="check-dailyBasisBooking"
                label={this.props.t("dailyBasisBooking")}
                checked={this.state.dailyBasisBooking}
                onChange={(e: any) =>
                  this.onDailyBasisBookingChange(e.target.checked)
                }
              />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2" htmlFor="input-minBookingDuration">
              {this.props.t("minBookingDuration")}
            </Form.Label>
            <Col sm="4">
              <InputGroup>
                <Form.Control
                  id="input-minBookingDuration"
                  type="number"
                  value={Math.round(this.state.minBookingDuration)}
                  onChange={(e: any) =>
                    this.setState({ minBookingDuration: e.target.value })
                  }
                  required={true}
                  min="0"
                  max="9999"
                />
                <InputGroup.Text>
                  {this.state.dailyBasisBooking
                    ? this.props.t("days")
                    : this.props.t("hours")}
                </InputGroup.Text>
              </InputGroup>
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2" htmlFor="input-maxBookingDuration">
              {this.props.t("maxBookingDuration")}
            </Form.Label>
            <Col sm="4">
              <InputGroup>
                <Form.Control
                  id="input-maxBookingDuration"
                  type="number"
                  value={Math.max(Math.round(this.state.maxBookingDuration), 1)}
                  onChange={(e: any) =>
                    this.setState({ maxBookingDuration: e.target.value })
                  }
                  required={true}
                  min="0"
                  max="9999"
                />
                <InputGroup.Text>
                  {this.state.dailyBasisBooking
                    ? this.props.t("days")
                    : this.props.t("hours")}
                </InputGroup.Text>
              </InputGroup>
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label
              column
              sm="2"
              htmlFor="input-targetUtilizationHoursPerWeek"
            >
              {this.props.t("targetUtilizationHoursPerWeek")}
            </Form.Label>
            <Col sm="4">
              <InputGroup>
                <Form.Control
                  id="input-targetUtilizationHoursPerWeek"
                  type="number"
                  value={Math.max(
                    Math.round(this.state.targetUtilizationHoursPerWeek),
                    1,
                  )}
                  onChange={(e: any) =>
                    this.setState({
                      targetUtilizationHoursPerWeek: e.target.value,
                    })
                  }
                  min="0"
                  max="168"
                />
                <InputGroup.Text>
                  {this.state.dailyBasisBooking
                    ? this.props.t("days")
                    : this.props.t("hours")}
                </InputGroup.Text>
              </InputGroup>
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2" htmlFor="input-subjectDefault">
              {this.props.t("subjectDefault")}
            </Form.Label>
            <Col sm="4">
              <Form.Select
                id="input-subjectDefault"
                value={this.state.subjectDefault}
                onChange={(e: any) =>
                  this.setState({ subjectDefault: e.target.value })
                }
              >
                <option value="1">{this.props.t("disabled")}</option>
                <option value="2">{this.props.t("optional")}</option>
                <option value="3">{this.props.t("required")}</option>
              </Form.Select>
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Col sm="6">
              <Form.Check
                type="checkbox"
                id="check-allowRecurringBookings"
                label={this.props.t("allowRecurringBookings")}
                checked={this.state.allowRecurringBookings}
                onChange={(e: any) =>
                  this.setState({ allowRecurringBookings: e.target.checked })
                }
              />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Col sm="6">
              <Form.Check
                type="checkbox"
                id="check-allowBookingNonExistUsers"
                label={this.props.t("allowBookingNonExistUsers")}
                checked={this.state.allowBookingNonExistUsers}
                onChange={(e: any) =>
                  this.setState({ allowBookingNonExistUsers: e.target.checked })
                }
              />
            </Col>
          </Form.Group>

          {/* REPORTS */}

          <div className="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
            <h4>{this.props.t("reportSettings")}</h4>
          </div>
          <Form.Group as={Row}>
            <Col sm="6">
              <Form.Check
                type="checkbox"
                id="check-hideReports"
                label={this.props.t("hideReports")}
                checked={this.state.hideReports}
                onChange={(e: any) =>
                  this.setState({ hideReports: e.target.checked })
                }
              />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Col sm="6">
              <Form.Check
                type="checkbox"
                id="check-hideStats"
                label={this.props.t("hideStats")}
                checked={this.state.hideStats}
                onChange={(e: any) =>
                  this.setState({ hideStats: e.target.checked })
                }
              />
            </Col>
          </Form.Group>

          {/* KIOSK MODE */}

          <div className="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
            <h4>
              {this.props.t("kioskMode")}
              <PremiumFeatureIcon />
            </h4>
          </div>
          <Form.Group as={Row}>
            <Col sm="6">
              <Form.Check
                type="checkbox"
                id="check-kioskModeEnabled"
                label={this.props.t("kioskModeAvailable")}
                checked={
                  this.state.kioskModeEnabled &&
                  RuntimeConfig.INFOS.featureKioskMode
                }
                disabled={!RuntimeConfig.INFOS.featureKioskMode}
                onChange={(e: any) =>
                  this.setState({ kioskModeEnabled: e.target.checked })
                }
              />
              <Form.Text className="text-muted">
                {this.props.t("kioskModeAvailableHint")}
              </Form.Text>
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2" htmlFor="input-kioskSecret">
              {this.props.t("kioskSecret")}
            </Form.Label>
            <Col sm="4">
              <InputGroup>
                <Form.Control
                  id="input-kioskSecret"
                  type="text"
                  value={this.state.kioskSecret}
                  disabled={true}
                />
                <Button
                  variant="outline-secondary"
                  onClick={this.generateKioskSecret}
                  title={this.props.t("generatePassword")}
                  disabled={!RuntimeConfig.INFOS.featureKioskMode}
                >
                  <IconRefresh className="feather" />
                </Button>
                <CopyToClipboardButton
                  text={this.state.kioskSecret}
                  disabled={
                    !this.state.kioskSecret ||
                    this.state.kioskSecret === RendererUtils.SECRET_PLACEHOLDER
                  }
                />
              </InputGroup>
            </Col>
            <Col sm="2">
              <Button
                variant="outline-secondary"
                onClick={this.saveKioskSecret}
                disabled={
                  !RuntimeConfig.INFOS.featureKioskMode ||
                  !this.state.kioskSecret ||
                  this.state.kioskSecret === RendererUtils.SECRET_PLACEHOLDER
                }
              >
                {this.props.t("save")}
              </Button>
            </Col>
          </Form.Group>

          {/* AUTH PROVIDERS */}

          <div className="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
            <h4>
              {this.props.t("authProviders")}
              <PremiumFeatureIcon />
            </h4>
            <div className="btn-toolbar mb-2 mb-md-0">
              <div className="btn-group me-2">
                <Link
                  href="/admin/settings/auth-providers/add"
                  className={
                    "btn btn-sm btn-outline-secondary" +
                    (RuntimeConfig.INFOS.featureAuthProviders
                      ? ""
                      : " disabled")
                  }
                >
                  <IconPlus className="feather" /> {this.props.t("add")}
                </Link>
              </div>
            </div>
          </div>
          <Form.Group as={Row}>
            <Col sm="6">
              <Form.Check
                type="checkbox"
                id="check-allowAnyUser"
                title={this.props.t("allowAnyUserTooltip")}
                label={this.props.t("allowAnyUser")}
                checked={this.state.allowAnyUser}
                disabled={this.authProviders.length === 0}
                onChange={(e: any) =>
                  this.setState({ allowAnyUser: e.target.checked })
                }
              />
            </Col>
          </Form.Group>
          {authProviderTable}
          {dangerZone}
        </Form>
        <ReloadModal
          show={this.state.showSavedModal}
          title={this.props.t("settings")}
        />
      </FullLayout>
    );
  }
}

export default withTranslation(withReadyRouter(Settings as any));
