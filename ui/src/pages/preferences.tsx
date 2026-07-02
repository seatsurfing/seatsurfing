import React from "react";
import Loading from "../components/Loading";
import { Alert, Button, ButtonGroup, Form, Modal, Nav, ToggleButton } from "react-bootstrap";
import { NextRouter } from "next/router";
import { IoLinkOutline } from "react-icons/io5";
import NavBar from "@/components/NavBar";
import withReadyRouter from "@/components/withReadyRouter";
import RuntimeConfig from "@/components/RuntimeConfig";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import Ajax from "@/util/Ajax";
import UserPreference from "@/types/UserPreference";
import Location from "@/types/Location";

import Session from "@/types/Session";
import JwtDecoder from "@/util/JwtDecoder";
import Formatting from "@/util/Formatting";
import Validation from "@/util/Validation";
import TotpSettings from "@/components/TotpSettings";
import PasskeySettings from "@/components/PasskeySettings";
import SaveButton from "@/components/SaveButton";
import UrlInput from "@/components/form/UrlInput";
import Passkey from "@/types/Passkey";
import RendererUtils from "@/util/RendererUtils";
import { PreferencesTab } from "@/util/Navigation";
import CONSTANT from "@/util/Contant";

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
  mailReminder: boolean;
  mailLanguage: string;
  use24HourTime: boolean;
  dateFormat: string;
  weekStartDay: number;
  activeSessions: Session[];
  currentSessionId: string;
  showPasswordChangedModal: boolean;
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
}

const TAB_MAP: Record<PreferencesTab, string> = {
  booking: "tab-bookings",
  style: "tab-style",
  security: "tab-security",
  integration: "tab-integrations",
};

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
      mailReminder: false,
      mailLanguage: "",
      use24HourTime: true,
      dateFormat: "Y-m-d",
      weekStartDay: 1,
      activeSessions: [],
      currentSessionId: "",
      showPasswordChangedModal: false,
    };
  }

  componentDidMount = async () => {
    const tabParam = this.props.router.query.tab as PreferencesTab;
    if (tabParam && TAB_MAP[tabParam]) {
      this.setState({ activeTab: TAB_MAP[tabParam] });
    }
    await Promise.all([
      this.loadPreferences(),
      this.loadLocations(),
      this.loadActiveSessions(),
    ]);
    this.setState({ loading: false });
  };

  loadActiveSessions = async (): Promise<void> => {
    const accessTokenPayload = JwtDecoder.getPayload(
      Ajax.PERSISTER.readCredentialsFromLocalStorage().accessToken,
    );
    const sessions = await Session.list();
    this.setState({
      activeSessions: sessions,
      currentSessionId: accessTokenPayload.sid,
    });
  };

  loadPreferences = async (): Promise<void> => {
    const list = await UserPreference.list();
    const state: Partial<State> = {};
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
        s.value
          .split(",")
          .forEach((val) => (state.workdays![parseInt(val)] = true));
      }
      if (s.name === UserPreference.PREF_BOOKED_COLOR) state.booked = s.value;
      if (s.name === UserPreference.PREF_NOT_BOOKED_COLOR)
        state.notBooked = s.value;
      if (s.name === UserPreference.PREF_SELF_BOOKED_COLOR)
        state.selfBooked = s.value;
      if (s.name === UserPreference.PREF_PARTIALLY_BOOKED_COLOR)
        state.partiallyBooked = s.value;
      if (s.name === UserPreference.PREF_BUDDY_BOOKED_COLOR)
        state.buddyBooked = s.value;
      if (s.name === UserPreference.PREF_DISALLOWED_COLOR)
        state.disallowed = s.value;
      if (s.name === UserPreference.PREF_LOCATION_ID)
        state.locationId = s.value;
      if (s.name === UserPreference.PREF_CALDAV_URL) state.caldavUrl = s.value;
      if (s.name === UserPreference.PREF_CALDAV_USER)
        state.caldavUser = s.value;
      if (s.name === UserPreference.PREF_CALDAV_PASS)
        state.caldavPass = s.value;
      if (s.name === UserPreference.PREF_CALDAV_PATH)
        state.caldavCalendar = s.value;
      if (s.name === UserPreference.PREF_MAIL_NOTIFICATIONS)
        state.mailNotifications = s.value === "1";
      if (s.name === UserPreference.PREF_MAIL_REMINDER)
        state.mailReminder = s.value === "1";
      if (s.name === UserPreference.PREF_MAIL_LANGUAGE)
        state.mailLanguage = s.value;
      if (s.name === UserPreference.PREF_USE_24_HOUR_TIME)
        state.use24HourTime = s.value === "1";
      if (s.name === UserPreference.PREF_DATE_FORMAT)
        state.dateFormat = s.value;
      if (s.name === UserPreference.PREF_WEEK_START_DAY) {
        const v = parseInt(s.value);
        state.weekStartDay = CONSTANT.WEEK_START_DAYS.includes(v) ? v : 1;
      }
    });
    await new Promise<void>((resolve) =>
      this.setState({ ...this.state, ...state }, resolve),
    );
  };

  loadLocations = async (): Promise<void> => {
    this.locations = await Location.list();
  };

  onSubmit = async (e: any) => {
    e.preventDefault();
    this.setState({
      submitting: true,
      saved: false,
      error: false,
      caldavError: false,
    });
    const workdays: string[] = [];
    this.state.workdays.forEach((val, day) => {
      if (val) {
        workdays.push(day.toString());
      }
    });
    const payload = [
      new UserPreference(
        UserPreference.PREF_ENTER_TIME,
        this.state.enterTime.toString(),
      ),
      new UserPreference(
        UserPreference.PREF_WORKDAY_START,
        this.state.workdayStart.toString(),
      ),
      new UserPreference(
        UserPreference.PREF_WORKDAY_END,
        this.state.workdayEnd.toString(),
      ),
      new UserPreference(UserPreference.PREF_WORKDAYS, workdays.join(",")),
      new UserPreference(
        UserPreference.PREF_MAIL_NOTIFICATIONS,
        this.state.mailNotifications ? "1" : "0",
      ),
      new UserPreference(
        UserPreference.PREF_MAIL_REMINDER,
        this.state.mailReminder ? "1" : "0",
      ),
      new UserPreference(
        UserPreference.PREF_MAIL_LANGUAGE,
        this.state.mailLanguage,
      ),
      new UserPreference(
        UserPreference.PREF_USE_24_HOUR_TIME,
        this.state.use24HourTime ? "1" : "0",
      ),
      new UserPreference(
        UserPreference.PREF_LOCATION_ID,
        this.state.locationId,
      ),
      new UserPreference(
        UserPreference.PREF_DATE_FORMAT,
        this.state.dateFormat,
      ),
      new UserPreference(
        UserPreference.PREF_WEEK_START_DAY,
        this.state.weekStartDay.toString(),
      ),
    ];
    try {
      await UserPreference.setAll(payload);
      await RuntimeConfig.loadUserPreferences();
      this.setState({ submitting: false, saved: true });
    } catch {
      this.setState({ submitting: false, error: true });
    }
  };

  onSubmitSecurity = async (e: any) => {
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

    await Ajax.putData("/user/me/password", payload);
    this.setState({ submitting: false, showPasswordChangedModal: true });
  };

  onSubmitColors = async (e: any) => {
    e.preventDefault();
    this.setState({
      submitting: true,
      saved: false,
      error: false,
      caldavError: false,
    });
    const payload = [
      new UserPreference(UserPreference.PREF_BOOKED_COLOR, this.state.booked),
      new UserPreference(
        UserPreference.PREF_NOT_BOOKED_COLOR,
        this.state.notBooked,
      ),
      new UserPreference(
        UserPreference.PREF_SELF_BOOKED_COLOR,
        this.state.selfBooked,
      ),
      new UserPreference(
        UserPreference.PREF_PARTIALLY_BOOKED_COLOR,
        this.state.partiallyBooked,
      ),
      new UserPreference(
        UserPreference.PREF_BUDDY_BOOKED_COLOR,
        this.state.buddyBooked,
      ),
      new UserPreference(
        UserPreference.PREF_DISALLOWED_COLOR,
        this.state.disallowed,
      ),
    ];
    try {
      await UserPreference.setAll(payload);
      this.setState({ submitting: false, saved: true });
    } catch {
      this.setState({ submitting: false, error: true });
    }
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
    const workdays = this.state.workdays.map((val, i) =>
      i === day ? checked : val,
    );
    this.setState({
      workdays: workdays,
    });
  };

  connectCalDav = async () => {
    this.setState({
      submitting: true,
      saved: false,
      error: false,
      caldavError: false,
      caldavCalendarsLoaded: false,
    });
    const payload = {
      url: this.state.caldavUrl,
      username: this.state.caldavUser,
      password: this.state.caldavPass,
    };
    try {
      const res = await Ajax.postData(
        "/preference/caldav/listCalendars",
        payload,
      );
      this.setState({
        caldavCalendarsLoaded: true,
        caldavCalendars: res.json,
        caldavCalendar: res.json && res.json.length > 0 ? res.json[0].path : "",
        submitting: false,
      });
    } catch {
      this.setState({ submitting: false, caldavError: true });
    }
  };

  disconnectCalDav = async () => {
    this.setState({
      submitting: true,
      saved: false,
      error: false,
      caldavError: false,
      caldavCalendarsLoaded: false,
    });
    const payload = [
      new UserPreference(UserPreference.PREF_CALDAV_URL, ""),
      new UserPreference(UserPreference.PREF_CALDAV_USER, ""),
      new UserPreference(UserPreference.PREF_CALDAV_PASS, ""),
      new UserPreference(UserPreference.PREF_CALDAV_PATH, ""),
    ];
    try {
      await UserPreference.setAll(payload);
      this.setState({
        submitting: false,
        saved: true,
        caldavUrl: "",
        caldavUser: "",
        caldavPass: "",
        caldavCalendar: "",
        caldavCalendars: [],
      });
    } catch {
      this.setState({ submitting: false, error: true });
    }
  };

  saveCaldavSettings = async (e: any) => {
    e.preventDefault();
    this.setState({
      submitting: true,
      saved: false,
      error: false,
      caldavError: false,
    });
    const payload = [
      new UserPreference(UserPreference.PREF_CALDAV_URL, this.state.caldavUrl),
      new UserPreference(
        UserPreference.PREF_CALDAV_USER,
        this.state.caldavUser,
      ),
      new UserPreference(
        UserPreference.PREF_CALDAV_PASS,
        this.state.caldavPass,
      ),
      new UserPreference(
        UserPreference.PREF_CALDAV_PATH,
        this.state.caldavCalendar,
      ),
    ];
    try {
      await UserPreference.setAll(payload);
      this.setState({ submitting: false, saved: true });
    } catch {
      this.setState({ submitting: false, error: true });
    }
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
                if (key) {
                  this.setState({ activeTab: key, error: false, saved: false });
                  const tabParam = Object.entries(TAB_MAP).find(
                    ([, v]) => v === key,
                  )?.[0] as PreferencesTab;
                  this.props.router.replace(
                    { query: { ...this.props.router.query, tab: tabParam } },
                    undefined,
                    { shallow: true },
                  );
                }
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
                {!RuntimeConfig.INFOS.dailyBasisBooking &&
                  this.state.workdayEnd - this.state.workdayStart >
                    RuntimeConfig.INFOS.maxBookingDurationHours && (
                    <Form.Text muted>
                      {this.props.t("workingHoursHintExceedsMaxDuration", {
                        num: RuntimeConfig.INFOS.maxBookingDurationHours,
                      })}
                    </Form.Text>
                  )}
                {RuntimeConfig.INFOS.dailyBasisBooking && (
                  <Form.Text muted>
                    {this.props.t("workingHoursHintOnlyDaily")}
                  </Form.Text>
                )}
              </Form.Group>
              <Form.Group className="margin-top-15">
                <Form.Label>{this.props.t("workdays")}</Form.Label>
                <div className="text-left">
                  <ButtonGroup>
                    {[0, 1, 2, 3, 4, 5, 6]
                      .map((offset) => (this.state.weekStartDay + offset) % 7)
                      .map((day) => (
                        <ToggleButton
                          type="checkbox"
                          variant={
                            this.state.workdays[day]
                              ? "primary"
                              : "outline-secondary"
                          }
                          key={"workday-" + day}
                          id={"workday-" + day}
                          value={day}
                          checked={this.state.workdays[day]}
                          onChange={(e: any) =>
                            this.onWorkdayCheck(day, e.target.checked)
                          }
                        >
                          {this.props.t("workday-short-" + day)}
                        </ToggleButton>
                      ))}
                  </ButtonGroup>
                </div>
              </Form.Group>
              <Form.Group className="margin-top-15">
                <Form.Label htmlFor="weekStartDay">
                  {this.props.t("weekStartDay")}
                </Form.Label>
                <Form.Select
                  id="weekStartDay"
                  value={this.state.weekStartDay}
                  onChange={(e: any) =>
                    this.setState({
                      weekStartDay: window.parseInt(e.target.value),
                    })
                  }
                >
                  {CONSTANT.WEEK_START_DAYS.map((day) => (
                    <option key={"week-start-" + day} value={day}>
                      {this.props.t("workday-" + day)}
                    </option>
                  ))}
                </Form.Select>
              </Form.Group>
              <Form.Group className="margin-top-15">
                <Form.Label htmlFor="mailNotifications">
                  {this.props.t("mailNotifications")}
                </Form.Label>
                <div className="text-left">
                  <Form.Check
                    type="checkbox"
                    id="mailNotifications"
                    label={this.props.t("mailNotificationsBookingInfo")}
                    checked={this.state.mailNotifications}
                    onChange={(e: any) =>
                      this.setState({ mailNotifications: e.target.checked })
                    }
                  />
                  <Form.Check
                    type="checkbox"
                    id="mailReminder"
                    label={this.props.t("mailReminderBookingInfo")}
                    checked={this.state.mailReminder}
                    onChange={(e: any) =>
                      this.setState({ mailReminder: e.target.checked })
                    }
                  />
                </div>
              </Form.Group>
              <Form.Group className="margin-top-15">
                <Form.Label htmlFor="mailLanguage">
                  {this.props.t("mailLanguage")}
                </Form.Label>
                <Form.Select
                  id="mailLanguage"
                  value={this.state.mailLanguage}
                  onChange={(e: any) =>
                    this.setState({ mailLanguage: e.target.value })
                  }
                >
                  <option value="">
                    ({this.props.t("default")} -{" "}
                    {this.props.t(
                      "language-" + RuntimeConfig.INFOS.orgLanguage,
                    )}
                    )
                  </option>
                  {["de", "en"].map((lc) => (
                    <option key={lc} value={lc}>
                      {this.props.t("language-" + lc)}
                    </option>
                  ))}
                </Form.Select>
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
                  minLength={Validation.PASSWORD_MIN_LENGTH}
                  maxLength={Validation.PASSWORD_MAX_LENGTH}
                  pattern={Validation.PASSWORD_PATTERN}
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
              onPasskeyDeleted={async () => {
                const passkeys = await Passkey.list();
                RuntimeConfig.INFOS.hasPasskeys = passkeys.length > 0;
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
                              onClick={async (e) => {
                                e.preventDefault();
                                try {
                                  await session.delete();
                                  await this.loadActiveSessions();
                                } catch {
                                  RuntimeConfig.logOut();
                                }
                              }}
                            >
                              {this.props.t("logout")}
                            </a>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                  <p>* {this.props.t("thisSession")}</p>
                  <Button
                    hidden={this.state.activeSessions?.length <= 1}
                    type="button"
                    variant="secondary"
                    onClick={async () => {
                      const others = this.state.activeSessions.filter(
                        (s) => s.id !== this.state.currentSessionId,
                      );
                      try {
                        await Promise.all(others.map((s) => s.delete()));
                        await this.loadActiveSessions();
                      } catch {
                        RuntimeConfig.logOut();
                      }
                    }}
                  >
                    {this.props.t("logoutOthers")}
                  </Button>
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
                sandbox="allow-scripts allow-same-origin allow-forms"
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
                <UrlInput
                  id="caldavUrl"
                  placeholder="https://…"
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
        <Modal
          show={this.state.showPasswordChangedModal}
          onHide={() => {}}
          backdrop="static"
          keyboard={false}
        >
          <Modal.Body>
            <p>{this.props.t("passwordChangedLoginAgain")}</p>
          </Modal.Body>
          <Modal.Footer>
            <Button variant="primary" onClick={() => window.location.reload()}>
              {this.props.t("ok")}
            </Button>
          </Modal.Footer>
        </Modal>
      </>
    );
  }
}

export default withTranslation(withReadyRouter(Preferences as any));
