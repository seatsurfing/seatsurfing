import React from "react";
import Loading from "../components/Loading";
import { Alert, Button, ButtonGroup, Form, Nav } from "react-bootstrap";
import { NextRouter } from "next/router";
import { IoLinkOutline } from "react-icons/io5";
import NavBar from "@/components/NavBar";
import withReadyRouter from "@/components/withReadyRouter";
import RuntimeConfig from "@/components/RuntimeConfig";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import Ajax from "@/util/Ajax";
import UserPreference from "@/types/UserPreference";
import Location from "@/types/Location";
import RedirectUtil from "@/util/RedirectUtil";
import Session from "@/types/Session";
import JwtDecoder from "@/util/JwtDecoder";
import Formatting from "@/util/Formatting";
import { PASSWORD_PATTERN } from "@/util/Validation";
import TotpSettings from "@/components/TotpSettings";
import PasskeySettings from "@/components/PasskeySettings";
import SaveButton from "@/components/SaveButton";
import Passkey from "@/types/Passkey";
import RendererUtils from "@/util/RendererUtils";

interface State {
  loading: boolean;
  submitting: boolean;
  saved: boolean;
  error: boolean;
  enterTime: number;
  workdayStart: number;
  workdayEnd: number;
  workdays: boolean[];
  booked: string;
  notBooked: string;
  selfBooked: string;
  partiallyBooked: string;
  buddyBooked: string;
  disallowed: string;
  locationId: string;
  changePassword: boolean;
  password: string;
  activeTab: string;
  caldavUrl: string;
  caldavUser: string;
  caldavPass: string;
  caldavCalendar: string;
  caldavCalendars: any[];
  caldavCalendarsLoaded: boolean;
  caldavError: boolean;
  mailNotifications: boolean;
  use24HourTime: boolean;
  dateFormat: string;
  activeSessions: Session[];
  currentSessionId: string;
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
}

const COLOR_BOOKED: string = "#ff453a";
const COLOR_NOT_BOOKED: string = "#30d158";
const COLOR_SELF_BOOKED: string = "#b825de";
const COLOR_PARTIALLY_BOOKED: string = "#ff9100";
const COLOR_BUDDY_BOOKED: string = "#2415c5";
const COLOR_DISALLOWED: string = "#eeeeee";

class Preferences extends React.Component<Props, State> {
  locations: Location[];

  constructor(props: any) {
    super(props);
    this.locations = [];
    this.state = {
      loading: true,
      submitting: false,
      saved: false,
      error: false,
      enterTime: 0,
      workdayStart: 0,
      workdayEnd: 0,
      workdays: [],
      booked: COLOR_BOOKED,
      notBooked: COLOR_NOT_BOOKED,
      selfBooked: COLOR_SELF_BOOKED,
      partiallyBooked: COLOR_PARTIALLY_BOOKED,
      buddyBooked: COLOR_BUDDY_BOOKED,
      disallowed: COLOR_DISALLOWED,
      locationId: "",
      changePassword: false,
      password: "",
      activeTab: "tab-bookings",
      caldavUrl: "",
      caldavUser: "",
      caldavPass: "",
      caldavCalendar: "",
      caldavCalendars: [],
      caldavCalendarsLoaded: false,
      caldavError: false,
      mailNotifications: false,
      use24HourTime: true,
      dateFormat: "Y-m-d",
      activeSessions: [],
      currentSessionId: "",
    };
  }

  componentDidMount = () => {
    if (!Ajax.hasAccessToken()) {
      RedirectUtil.toLogin(this.props.router);
      return;
    }
    const tabParam = this.props.router.query.tab;
    if (tabParam === "security") {
      this.setState({ activeTab: "tab-security" });
    }
    let promises = [
      this.loadPreferences(),
      this.loadLocations(),
      this.loadActiveSessions(),
    ];
    Promise.all(promises).then(() => {
      this.setState({ loading: false });
    });
  };

  loadActiveSessions = async (): Promise<void> => {
    const accessTokenPayload = JwtDecoder.getPayload(
      Ajax.PERSISTER.readCredentialsFromLocalStorage().accessToken,
    );
    let self = this;
    return new Promise<void>(function (resolve, reject) {
      Session.list()
        .then((sessions) => {
          self.setState({
            activeSessions: sessions,
            currentSessionId: accessTokenPayload.sid,
          });
          resolve();
        })
        .catch((e) => reject(e));
    });
  };

  loadPreferences = async (): Promise<void> => {
    let self = this;
    return new Promise<void>(function (resolve, reject) {
      UserPreference.list()
        .then((list) => {
          let state: any = {};
          list.forEach((s) => {
            if (typeof window !== "undefined") {
              if (s.name === UserPreference.PREF_ENTER_TIME)
                state.enterTime = window.parseInt(s.value);
              if (s.name === UserPreference.PREF_WORKDAY_START)
                state.workdayStart = window.parseInt(s.value);
              if (s.name === UserPreference.PREF_WORKDAY_END)
                state.workdayEnd = window.parseInt(s.value);
            }
            if (s.name === UserPreference.PREF_WORKDAYS) {
              state.workdays = [];
              for (let i = 0; i <= 6; i++) {
                state.workdays[i] = false;
              }
              s.value.split(",").forEach((val) => (state.workdays[val] = true));
            }
            if (s.name === UserPreference.PREF_BOOKED_COLOR)
              state.booked = s.value;
            if (s.name === UserPreference.PREF_NOT_BOOKED_COLOR)
              state.notBooked = s.value;
            if (s.name === UserPreference.PREF_SELF_BOOKED_COLOR)
              state.selfBooked = s.value;
            if (s.name === UserPreference.PREF_PARTIALLY_BOOKED_COLOR)
              state.partiallyBooked = s.value;
            if (s.name === UserPreference.PREF_BUDDY_BOOKED_COLOR)
              state.buddyBooked = s.value;
            if (s.name === UserPreference.PREF_DISALLOWED_COLOR)
              state.disallowedColor = s.value;
            if (s.name === UserPreference.PREF_LOCATION_ID)
              state.locationId = s.value;
            if (s.name === UserPreference.PREF_CALDAV_URL)
              state.caldavUrl = s.value;
            if (s.name === UserPreference.PREF_CALDAV_USER)
              state.caldavUser = s.value;
            if (s.name === UserPreference.PREF_CALDAV_PASS)
              state.caldavPass = s.value;
            if (s.name === UserPreference.PREF_CALDAV_PATH)
              state.caldavCalendar = s.value;
            if (s.name === UserPreference.PREF_MAIL_NOTIFICATIONS)
              state.mailNotifications = s.value === "1";
            if (s.name === UserPreference.PREF_USE_24_HOUR_TIME)
              state.use24HourTime = s.value === "1";
            if (s.name === UserPreference.PREF_DATE_FORMAT)
              state.dateFormat = s.value;
          });
          self.setState(
            {
              ...self.state,
              ...state,
            },
            () => resolve(),
          );
        })
        .catch((e) => reject(e));
    });
  };

  loadLocations = async (): Promise<void> => {
    let self = this;
    return new Promise<void>(function (resolve, reject) {
      Location.list()
        .then((list) => {
          self.locations = list;
          resolve();
        })
        .catch((e) => reject(e));
    });
  };

  onSubmit = (e: any) => {
    e.preventDefault();
    this.setState({
      submitting: true,
      saved: false,
      error: false,
      caldavError: false,
    });
    let workdays: string[] = [];
    this.state.workdays.forEach((val, day) => {
      if (val) {
        workdays.push(day.toString());
      }
    });
    let payload = [
      new UserPreference("enter_time", this.state.enterTime.toString()),
      new UserPreference("workday_start", this.state.workdayStart.toString()),
      new UserPreference("workday_end", this.state.workdayEnd.toString()),
      new UserPreference("workdays", workdays.join(",")),
      new UserPreference(
        "mail_notifications",
        this.state.mailNotifications ? "1" : "0",
      ),
      new UserPreference(
        "use_24_hour_time",
        this.state.use24HourTime ? "1" : "0",
      ),
      new UserPreference("location_id", this.state.locationId),
      new UserPreference("date_format", this.state.dateFormat),
    ];
    UserPreference.setAll(payload)
      .then(() => {
        RuntimeConfig.loadUserPreferences().then(() => {
          this.setState({
            submitting: false,
            saved: true,
          });
        });
      })
      .catch(() => {
        this.setState({
          submitting: false,
          error: true,
        });
      });
  };

  onSubmitSecurity = (e: any) => {
    e.preventDefault();
    if (!this.state.changePassword) {
      return;
    }
    this.setState({
      submitting: true,
      saved: false,
      error: false,
      caldavError: false,
    });
    const payload = {
      password: this.state.password,
    };
    Ajax.putData("/user/me/password", payload).then(() => {
      this.setState({
        submitting: false,
        saved: true,
      });
    });
  };

  onSubmitColors = (e: any) => {
    e.preventDefault();
    this.setState({
      submitting: true,
      saved: false,
      error: false,
      caldavError: false,
    });
    let workdays: string[] = [];
    this.state.workdays.forEach((val, day) => {
      if (val) {
        workdays.push(day.toString());
      }
    });
    let payload = [
      new UserPreference("booked_color", this.state.booked),
      new UserPreference("not_booked_color", this.state.notBooked),
      new UserPreference("self_booked_color", this.state.selfBooked),
      new UserPreference("partially_booked_color", this.state.partiallyBooked),
      new UserPreference("buddy_booked_color", this.state.buddyBooked),
      new UserPreference("disallowed_color", this.state.disallowed),
    ];
    UserPreference.setAll(payload)
      .then(() => {
        this.setState({
          submitting: false,
          saved: true,
        });
      })
      .catch(() => {
        this.setState({
          submitting: false,
          error: true,
        });
      });
  };

  resetColors = () => {
    this.setState({
      booked: COLOR_BOOKED,
      notBooked: COLOR_NOT_BOOKED,
      selfBooked: COLOR_SELF_BOOKED,
      partiallyBooked: COLOR_PARTIALLY_BOOKED,
      buddyBooked: COLOR_BUDDY_BOOKED,
      disallowed: COLOR_DISALLOWED,
    });
  };

  onWorkdayCheck = (day: number, checked: boolean) => {
    let workdays = this.state.workdays.map((val, i) =>
      i === day ? checked : val,
    );
    this.setState({
      workdays: workdays,
    });
  };

  connectCalDav = () => {
    this.setState({
      submitting: true,
      saved: false,
      error: false,
      caldavError: false,
      caldavCalendarsLoaded: false,
    });
    let payload = {
      url: this.state.caldavUrl,
      username: this.state.caldavUser,
      password: this.state.caldavPass,
    };
    Ajax.postData("/preference/caldav/listCalendars", payload)
      .then((res) => {
        this.setState({
          caldavCalendarsLoaded: true,
          caldavCalendars: res.json,
          caldavCalendar:
            res.json && res.json.length > 0 ? res.json[0].path : "",
          submitting: false,
        });
      })
      .catch(() => {
        this.setState({
          submitting: false,
          caldavError: true,
        });
      });
  };

  disconnectCalDav = () => {
    this.setState({
      submitting: true,
      saved: false,
      error: false,
      caldavError: false,
      caldavCalendarsLoaded: false,
    });
    let payload = [
      new UserPreference("caldav_url", ""),
      new UserPreference("caldav_user", ""),
      new UserPreference("caldav_pass", ""),
      new UserPreference("caldav_path", ""),
    ];
    UserPreference.setAll(payload)
      .then(() => {
        this.setState({
          submitting: false,
          saved: true,
          caldavUrl: "",
          caldavUser: "",
          caldavPass: "",
          caldavCalendar: "",
          caldavCalendars: [],
        });
      })
      .catch(() => {
        this.setState({
          submitting: false,
          error: true,
        });
      });
  };

  saveCaldavSettings = (e: any) => {
    e.preventDefault();
    this.setState({
      submitting: true,
      saved: false,
      error: false,
      caldavError: false,
    });
    let payload = [
      new UserPreference("caldav_url", this.state.caldavUrl),
      new UserPreference("caldav_user", this.state.caldavUser),
      new UserPreference("caldav_pass", this.state.caldavPass),
      new UserPreference("caldav_path", this.state.caldavCalendar),
    ];
    UserPreference.setAll(payload)
      .then(() => {
        this.setState({
          submitting: false,
          saved: true,
        });
      })
      .catch(() => {
        this.setState({
          submitting: false,
          error: true,
        });
      });
  };

  renderBookingColor(
    stateKey: keyof Pick<
      State,
      | "notBooked"
      | "selfBooked"
      | "booked"
      | "partiallyBooked"
      | "buddyBooked"
      | "disallowed"
    >,
    labelKey: string,
  ) {
    const id = `color${RendererUtils.capitalize(stateKey)}`;
    return (
      <Form.Group className="margin-top-15 d-flex align-items-center">
        <Form.Control
          type="color"
          key={id}
          id={id}
          value={this.state[stateKey]}
          onChange={(e: any) =>
            this.setState({ [stateKey]: e.target.value } as any)
          }
        />
        <Form.Label htmlFor={id} className="mb-0">
          {this.props.t(labelKey)}
        </Form.Label>
      </Form.Group>
    );
  }

  render() {
    if (this.state.loading) {
      return <Loading />;
    }

    let hint = <></>;
    if (this.state.saved) {
      hint = (
        <Alert variant="success" className="margin-top-15">
          {this.props.t("entryUpdated")}
        </Alert>
      );
    } else if (this.state.error) {
      hint = (
        <Alert variant="danger" className="margin-top-15">
          {this.props.t("errorSave")}
        </Alert>
      );
    } else if (this.state.caldavError) {
      hint = (
        <Alert variant="danger" className="margin-top-15">
          {this.props.t("errorCaldav")}
        </Alert>
      );
    }

    const credentials = Ajax.PERSISTER.readCredentialsFromLocalStorage();
    const profilePageUrl = credentials.profilePageUrl;

    return (
      <>
        <NavBar />
        <div className="container-center-top">
          <div className="container-center-inner-wide">
            <Nav
              variant="underline"
              activeKey={this.state.activeTab}
              onSelect={(key) => {
                if (key)
                  this.setState({ activeTab: key, error: false, saved: false });
              }}
            >
              <Nav.Item>
                <Nav.Link eventKey="tab-bookings">
                  {this.props.t("bookings")}
                </Nav.Link>
              </Nav.Item>
              <Nav.Item>
                <Nav.Link eventKey="tab-style">
                  {this.props.t("style")}
                </Nav.Link>
              </Nav.Item>
              <Nav.Item hidden={RuntimeConfig.INFOS.idpLogin}>
                <Nav.Link eventKey="tab-security">
                  {this.props.t("security")}
                </Nav.Link>
              </Nav.Item>
              <Nav.Item
                hidden={!RuntimeConfig.INFOS.idpLogin || !profilePageUrl}
              >
                <Nav.Link eventKey="tab-idp">
                  {this.props.t("security")}
                </Nav.Link>
              </Nav.Item>
              <Nav.Item>
                <Nav.Link eventKey="tab-integrations">
                  {this.props.t("integrations")}
                </Nav.Link>
              </Nav.Item>
            </Nav>
            {hint}

            {/* -------- */}
            {/* BOOKINGS */}
            {/* -------- */}

            <Form
              onSubmit={this.onSubmit}
              hidden={this.state.activeTab !== "tab-bookings"}
            >
              <Form.Group className="margin-top-15">
                <Form.Label htmlFor="enterTime">
                  {this.props.t("notice")}
                </Form.Label>
                <Form.Select
                  id="enterTime"
                  value={this.state.enterTime}
                  onChange={(e: any) =>
                    this.setState({ enterTime: e.target.value })
                  }
                >
                  <option value="1">{this.props.t("earliestPossible")}</option>
                  <option value="2">{this.props.t("nextDay")}</option>
                  <option value="3">{this.props.t("nextWorkday")}</option>
                </Form.Select>
              </Form.Group>
              <Form.Group className="margin-top-15">
                <Form.Label htmlFor="workdayStart">
                  {this.props.t("workingHours")}
                </Form.Label>
                <div>
                  <Form.Control
                    type="number"
                    id="workdayStart"
                    value={this.state.workdayStart}
                    onChange={(e: any) =>
                      this.setState({
                        workdayStart:
                          typeof window !== "undefined"
                            ? window.parseInt(e.target.value)
                            : 0,
                      })
                    }
                    min="0"
                    max="23"
                    style={{ display: "inline", width: "40%" }}
                  />
                  <span
                    style={{
                      width: "20%",
                      display: "inline-block",
                      textAlign: "center",
                    }}
                  >
                    <Form.Label htmlFor="workdayEnd">
                      {this.props.t("to").toString()}
                    </Form.Label>
                  </span>
                  <Form.Control
                    type="number"
                    id="workdayEnd"
                    value={this.state.workdayEnd}
                    onChange={(e: any) =>
                      this.setState({ workdayEnd: e.target.value })
                    }
                    min={this.state.workdayStart + 1}
                    max="23"
                    style={{ display: "inline", width: "40%" }}
                  />
                </div>
              </Form.Group>
              <Form.Group className="margin-top-15">
                <Form.Label>{this.props.t("workdays")}</Form.Label>
                <div className="text-left">
                  {[0, 1, 2, 3, 4, 5, 6].map((day) => (
                    <Form.Check
                      type="checkbox"
                      key={"workday-" + day}
                      id={"workday-" + day}
                      label={this.props.t("workday-" + day)}
                      checked={this.state.workdays[day]}
                      onChange={(e: any) =>
                        this.onWorkdayCheck(day, e.target.checked)
                      }
                    />
                  ))}
                </div>
              </Form.Group>
              <Form.Group className="margin-top-15">
                <Form.Label htmlFor="mailNotifications">
                  {this.props.t("mailNotifications")}
                </Form.Label>
                <div className="text-left">
                  <Form.Check
                    type="checkbox"
                    id="mailNotifications"
                    label={this.props.t("mailNotifications")}
                    checked={this.state.mailNotifications}
                    onChange={(e: any) =>
                      this.setState({ mailNotifications: e.target.checked })
                    }
                  />
                </div>
              </Form.Group>
              <Form.Group className="margin-top-15">
                <Form.Label htmlFor="use24HourTime">
                  {this.props.t("timeFormat")}
                </Form.Label>
                <div className="text-left">
                  <Form.Check
                    type="checkbox"
                    id="use24HourTime"
                    label={this.props.t("use24HourTime")}
                    checked={this.state.use24HourTime}
                    onChange={(e: any) =>
                      this.setState({ use24HourTime: e.target.checked })
                    }
                  />
                </div>
              </Form.Group>
              <Form.Group className="margin-top-15">
                <Form.Label htmlFor="dateFormat">
                  {this.props.t("dateFormat")}
                </Form.Label>
                <Form.Select
                  id="dateFormat"
                  value={this.state.dateFormat}
                  onChange={(e: any) =>
                    this.setState({ dateFormat: e.target.value })
                  }
                >
                  <option value="Y-m-d">Y-m-d</option>
                  <option value="d.m.Y">d.m.Y</option>
                  <option value="m/d/Y">m/d/Y</option>
                  <option value="d/m/Y">d/m/Y</option>
                </Form.Select>
              </Form.Group>
              <Form.Group className="margin-top-15">
                <Form.Label htmlFor="preferredLocation">
                  {this.props.t("preferredLocation")}
                </Form.Label>
                <Form.Select
                  id="preferredLocation"
                  value={this.state.locationId}
                  onChange={(e: any) =>
                    this.setState({ locationId: e.target.value })
                  }
                >
                  <option value="">({this.props.t("none")})</option>
                  {this.locations.map((location) => (
                    <option key={"location-" + location.id} value={location.id}>
                      {location.name}
                    </option>
                  ))}
                </Form.Select>
              </Form.Group>
              <SaveButton
                submitting={this.state.submitting}
                className="margin-top-15"
              />
            </Form>

            {/* ----- */}
            {/* STYLE */}
            {/* ----- */}

            <Form
              onSubmit={this.onSubmitColors}
              hidden={this.state.activeTab !== "tab-style"}
              className="form-colors"
            >
              <h5 className="margin-top-15">{this.props.t("bookingcolors")}</h5>
              {this.renderBookingColor("booked", "colorAlreadyBooked")}
              {this.renderBookingColor("notBooked", "colorNotBooked")}
              {this.renderBookingColor("selfBooked", "colorSelfBooked")}
              {RuntimeConfig.INFOS.maxHoursPartiallyBookedEnabled &&
                this.renderBookingColor(
                  "partiallyBooked",
                  "colorPartiallyBooked",
                )}
              {!RuntimeConfig.INFOS.disableBuddies &&
                this.renderBookingColor("buddyBooked", "colorBuddyBooked")}
              {this.renderBookingColor("disallowed", "colorDisallowed")}
              <ButtonGroup className="margin-top-15">
                <Button
                  type="button"
                  variant="secondary"
                  onClick={() => this.resetColors()}
                >
                  {this.props.t("reset")}
                </Button>
                <SaveButton submitting={this.state.submitting} />
              </ButtonGroup>
            </Form>

            {/* -------- */}
            {/* SECURITY */}
            {/* -------- */}

            <Form
              onSubmit={this.onSubmitSecurity}
              hidden={this.state.activeTab !== "tab-security"}
            >
              <h5 className="margin-top-15">{this.props.t("password")}</h5>
              <Form.Group className="margin-top-15">
                <Form.Check
                  type="checkbox"
                  inline={true}
                  id="check-changePassword"
                  label={this.props.t("passwordChange")}
                  checked={this.state.changePassword}
                  onChange={(e: any) =>
                    this.setState({ changePassword: e.target.checked })
                  }
                />
                <Form.Control
                  type="password"
                  value={this.state.password}
                  onChange={(e: any) =>
                    this.setState({ password: e.target.value })
                  }
                  required={this.state.changePassword}
                  disabled={!this.state.changePassword}
                  minLength={8}
                  pattern={PASSWORD_PATTERN}
                  title={this.props.t("passwordRequirements")}
                />
              </Form.Group>
              <SaveButton
                submitting={this.state.submitting}
                disabled={!this.state.changePassword}
                className="margin-top-15"
              />
            </Form>
            <TotpSettings
              hidden={
                this.state.activeTab !== "tab-security" ||
                RuntimeConfig.INFOS.idpLogin
              }
              t={this.props.t}
            />
            <PasskeySettings
              hidden={
                this.state.activeTab !== "tab-security" ||
                RuntimeConfig.INFOS.idpLogin
              }
              t={this.props.t}
              onPasskeyAdded={() => {
                RuntimeConfig.INFOS.hasPasskeys = true;
              }}
              onPasskeyDeleted={() => {
                Passkey.list().then((passkeys) => {
                  RuntimeConfig.INFOS.hasPasskeys = passkeys.length > 0;
                });
              }}
            />
            <div hidden={this.state.activeTab !== "tab-security"}>
              <h5 className="mt-5">{this.props.t("activeSessions")}</h5>
              {this.state.activeSessions.length === 0 ? (
                <p>{this.props.t("noActiveSessions")}</p>
              ) : (
                <div className="table-responsive">
                  <table className="table">
                    <thead>
                      <tr>
                        <th>{this.props.t("device")}</th>
                        <th>{this.props.t("created")} (UTC)</th>
                        <th></th>
                      </tr>
                    </thead>
                    <tbody>
                      {this.state.activeSessions.map((session) => (
                        <tr key={"session-" + session.id}>
                          <td>
                            {session.device}
                            {session.id === this.state.currentSessionId
                              ? " *"
                              : ""}
                          </td>
                          <td>
                            {Formatting.getFormatterShort(false).format(
                              new Date(session.created),
                            )}
                          </td>
                          <td>
                            <a
                              href="#"
                              onClick={(e) => {
                                e.preventDefault();
                                session
                                  .delete()
                                  .then(() => this.loadActiveSessions())
                                  .catch(() => RuntimeConfig.logOut());
                              }}
                            >
                              {this.props.t("logout")}
                            </a>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                  * {this.props.t("thisSession")}
                </div>
              )}
            </div>

            {/* --- */}
            {/* IDP */}
            {/* --- */}

            <div hidden={this.state.activeTab !== "tab-idp"}>
              <div className="text-end">
                <a
                  href={profilePageUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="btn btn-secondary btn-sm mb-2"
                >
                  <IoLinkOutline className="feather me-1" />
                  {this.props.t("manageProfile")}
                </a>
              </div>
              <iframe
                src={profilePageUrl}
                style={{ width: "100%", height: "100vh", borderWidth: 0 }}
                id="idp-profilepage-iframe"
              ></iframe>
            </div>

            {/* ------------ */}
            {/* INTEGRATIONS */}
            {/* ------------ */}

            <Form
              onSubmit={this.saveCaldavSettings}
              hidden={this.state.activeTab !== "tab-integrations"}
            >
              <h5 className="margin-top-15">
                {this.props.t("caldavCalendar")}
              </h5>
              <Form.Group className="margin-top-15">
                <Form.Label htmlFor="caldavUrl">
                  {this.props.t("caldavUrl")}
                </Form.Label>
                <Form.Control
                  id="caldavUrl"
                  type="url"
                  value={this.state.caldavUrl}
                  onChange={(e: any) =>
                    this.setState({
                      caldavUrl: e.target.value,
                      caldavCalendarsLoaded: false,
                    })
                  }
                />
              </Form.Group>
              <Form.Group className="margin-top-15">
                <Form.Label htmlFor="caldavUser">
                  {this.props.t("username")}
                </Form.Label>
                <Form.Control
                  id="caldavUser"
                  type="text"
                  value={this.state.caldavUser}
                  onChange={(e: any) =>
                    this.setState({
                      caldavUser: e.target.value,
                      caldavCalendarsLoaded: false,
                    })
                  }
                />
              </Form.Group>
              <Form.Group className="margin-top-15">
                <Form.Label htmlFor="caldavPass">
                  {this.props.t("password")}
                </Form.Label>
                <Form.Control
                  id="caldavPass"
                  type="password"
                  value={this.state.caldavPass}
                  onChange={(e: any) =>
                    this.setState({
                      caldavPass: e.target.value,
                      caldavCalendarsLoaded: false,
                    })
                  }
                />
              </Form.Group>
              <Form.Group className="margin-top-15">
                <Form.Label htmlFor="caldavCalendar">
                  {this.props.t("calendar")}
                </Form.Label>
                <Form.Select
                  id="caldavCalendar"
                  value={this.state.caldavCalendar}
                  onChange={(e: any) =>
                    this.setState({ caldavCalendar: e.target.value })
                  }
                  disabled={!this.state.caldavCalendarsLoaded}
                >
                  {this.state.caldavCalendars.map((cal) => (
                    <option key={cal.path} value={cal.path}>
                      {cal.name}
                    </option>
                  ))}
                </Form.Select>
              </Form.Group>
              <ButtonGroup className="margin-top-15">
                <Button
                  type="button"
                  variant="secondary"
                  disabled={
                    this.state.submitting ||
                    this.state.caldavUrl === "" ||
                    this.state.caldavUser === "" ||
                    this.state.caldavPass === ""
                  }
                  onClick={() => this.connectCalDav()}
                >
                  {this.props.t("connect")}
                </Button>
                <Button
                  type="button"
                  variant="secondary"
                  disabled={
                    this.state.submitting ||
                    this.state.caldavUrl === "" ||
                    this.state.caldavUser === "" ||
                    this.state.caldavPass === "" ||
                    this.state.caldavCalendar === ""
                  }
                  onClick={() => this.disconnectCalDav()}
                >
                  {this.props.t("disconnect")}
                </Button>
                <SaveButton
                  submitting={this.state.submitting}
                  disabled={
                    !(
                      this.state.caldavCalendarsLoaded &&
                      this.state.caldavCalendar != ""
                    ) || this.state.submitting
                  }
                />
              </ButtonGroup>
            </Form>
          </div>
        </div>
      </>
    );
  }
}

export default withTranslation(withReadyRouter(Preferences as any));
