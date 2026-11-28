import React from "react";
import {
  Alert,
  Button,
  ButtonGroup,
  Form,
  Modal,
  Nav,
  OverlayTrigger,
  Tooltip,
} from "react-bootstrap";
import { IoLinkOutline } from "react-icons/io5";
import {
  Settings as IconSettings,
  Calendar as IconBookings,
  Droplet as IconStyle,
  Lock as IconSecurity,
  Link as IconIntegrations,
  MapPin as IconCurrentSession,
} from "react-feather";
import RuntimeConfig from "@/components/RuntimeConfig";
import DateTimePicker from "@/components/DateTimePicker";
import DateUtil from "@/util/DateUtil";
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
import ReloadModal from "@/components/ReloadModal";
import { PreferencesTab } from "@/util/Navigation";
import CONSTANT from "@/util/Constant";
import WeekdaySelection from "@/components/WeekdaySelection";

type ActiveTab = PreferencesTab | "idp";

interface State {
  loading: boolean;
  submitting: boolean;
  saved: boolean;
  error: boolean;
  enterTime: number;
  workdayStart: string;
  workdayEnd: string;
  workdays: number[];
  booked: string;
  notBooked: string;
  selfBooked: string;
  partiallyBooked: string;
  buddyBooked: string;
  disallowed: string;
  locationId: string;
  changePassword: boolean;
  password: string;
  activeTab: ActiveTab;
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
  showReloadModal: boolean;
  pickerContainer: HTMLDivElement | null;
}

interface Props {
  show: boolean;
  initialTab?: PreferencesTab | null;
  onHide: () => void;
  t: TranslationFunc;
}

const COLOR_BOOKED: string = "#ff453a";
const COLOR_NOT_BOOKED: string = "#30d158";
const COLOR_SELF_BOOKED: string = "#b825de";
const COLOR_PARTIALLY_BOOKED: string = "#ff9100";
const COLOR_BUDDY_BOOKED: string = "#2415c5";
const COLOR_DISALLOWED: string = "#eeeeee";

class PreferencesModal extends React.Component<Props, State> {
  locations: Location[];

  constructor(props: any) {
    super(props);
    this.locations = [];
    this.state = this.getInitialState();
  }

  getInitialState = (): State => {
    return {
      loading: true,
      submitting: false,
      saved: false,
      error: false,
      enterTime: 0,
      workdayStart: UserPreference.DEFAULT_WORKDAY_START,
      workdayEnd: UserPreference.DEFAULT_WORKDAY_END,
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
      activeTab: this.resolveTab(this.props.initialTab),
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
      showReloadModal: false,
      pickerContainer: null,
    };
  };

  resolveTab = (tab: PreferencesTab | null | undefined): ActiveTab => {
    if (tab === "security" && RuntimeConfig.INFOS.idpLogin) {
      return "idp";
    }
    if (
      tab === "booking" ||
      tab === "style" ||
      tab === "security" ||
      tab === "integration"
    ) {
      return tab;
    }
    return "booking";
  };

  componentDidMount = () => {
    if (this.props.show) {
      this.loadData();
    }
  };

  componentDidUpdate = (prevProps: Props) => {
    if (this.props.show && !prevProps.show) {
      // Keep the picker container: its callback ref may already have queued
      // an update in this commit and won't be called again for the same node.
      this.setState(
        (prevState) => ({
          ...this.getInitialState(),
          pickerContainer: prevState.pickerContainer,
        }),
        () => this.loadData(),
      );
    } else if (
      this.props.show &&
      this.props.initialTab !== prevProps.initialTab &&
      this.props.initialTab
    ) {
      this.setState({ activeTab: this.resolveTab(this.props.initialTab) });
    }
  };

  loadData = async () => {
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
      }
      if (s.name === UserPreference.PREF_WORKDAY_START)
        state.workdayStart =
          DateUtil.parseTimeString(s.value) ??
          UserPreference.DEFAULT_WORKDAY_START;
      if (s.name === UserPreference.PREF_WORKDAY_END)
        state.workdayEnd =
          DateUtil.parseTimeString(s.value) ??
          UserPreference.DEFAULT_WORKDAY_END;
      if (s.name === UserPreference.PREF_WORKDAYS) {
        state.workdays = s.value
          .split(",")
          .map((val) => parseInt(val))
          .filter((val) => !isNaN(val));
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
    this.locations = await Location.list(true);
  };

  onSubmit = async (e: any) => {
    e.preventDefault();
    this.setState({
      submitting: true,
      saved: false,
      error: false,
      caldavError: false,
    });
    const payload = [
      new UserPreference(
        UserPreference.PREF_ENTER_TIME,
        this.state.enterTime.toString(),
      ),
      new UserPreference(
        UserPreference.PREF_WORKDAY_START,
        this.state.workdayStart,
      ),
      new UserPreference(
        UserPreference.PREF_WORKDAY_END,
        this.state.workdayEnd,
      ),
      new UserPreference(
        UserPreference.PREF_WORKDAYS,
        [...this.state.workdays].sort((a, b) => a - b).join(","),
      ),
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
      this.setState({ submitting: false, showReloadModal: true });
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
      this.setState({ submitting: false, showReloadModal: true });
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

  onWorkdaysChange = (workdays: number[]) => {
    this.setState({ workdays: [...workdays].sort((a, b) => a - b) });
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

  renderNavItem(
    tab: ActiveTab,
    Icon: React.ElementType,
    labelKey: string,
    hidden: boolean = false,
  ) {
    return (
      <Nav.Item hidden={hidden}>
        <Nav.Link eventKey={tab} className="d-flex align-items-center gap-2">
          <Icon className="feather" />
          {this.props.t(labelKey)}
        </Nav.Link>
      </Nav.Item>
    );
  }

  renderContent() {
    if (this.state.loading) {
      return (
        <div className="d-flex justify-content-center py-5">
          <span
            className="spinner-border text-secondary"
            role="status"
            aria-label={this.props.t("loading")}
          />
        </div>
      );
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
        {hint}

        {/* -------- */}
        {/* BOOKINGS */}
        {/* -------- */}

        <Form
          onSubmit={this.onSubmit}
          hidden={this.state.activeTab !== "booking"}
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
            <div className="d-flex align-items-center">
              <div style={{ width: "40%" }}>
                <DateTimePicker
                  id="workdayStart"
                  appendTo={this.state.pickerContainer}
                  noCalendar={true}
                  enableTime={true}
                  value={DateUtil.getTodayTimeFromTimeString(
                    this.state.workdayStart,
                  )}
                  maxDate={DateUtil.getTodayTimeFromMinutes(23 * 60 - 1)}
                  onChange={(value: Date) => {
                    const workdayStart = DateUtil.formatTimeString(value);
                    const startMinutes =
                      DateUtil.timeStringToMinutes(workdayStart);
                    const endMinutes = DateUtil.timeStringToMinutes(
                      this.state.workdayEnd,
                    );
                    this.setState({
                      workdayStart: workdayStart,
                      workdayEnd:
                        endMinutes >= startMinutes + 60
                          ? this.state.workdayEnd
                          : DateUtil.formatTimeString(
                              DateUtil.getTodayTimeFromMinutes(
                                startMinutes + 60,
                              ),
                            ),
                    });
                  }}
                />
              </div>
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
              <div style={{ width: "40%" }}>
                <DateTimePicker
                  id="workdayEnd"
                  appendTo={this.state.pickerContainer}
                  noCalendar={true}
                  enableTime={true}
                  value={DateUtil.getTodayTimeFromTimeString(
                    this.state.workdayEnd,
                  )}
                  minDate={DateUtil.getTodayTimeFromMinutes(
                    DateUtil.timeStringToMinutes(this.state.workdayStart) + 60,
                  )}
                  onChange={(value: Date) => {
                    const workdayEnd = DateUtil.formatTimeString(value);
                    const endMinutes = DateUtil.timeStringToMinutes(workdayEnd);
                    const startMinutes = DateUtil.timeStringToMinutes(
                      this.state.workdayStart,
                    );
                    this.setState({
                      workdayEnd: workdayEnd,
                      workdayStart:
                        startMinutes <= endMinutes - 60
                          ? this.state.workdayStart
                          : DateUtil.formatTimeString(
                              DateUtil.getTodayTimeFromMinutes(endMinutes - 60),
                            ),
                    });
                  }}
                />
              </div>
            </div>
            {!RuntimeConfig.INFOS.dailyBasisBooking &&
              (DateUtil.timeStringToMinutes(this.state.workdayEnd) -
                DateUtil.timeStringToMinutes(this.state.workdayStart)) /
                60 >
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
              <WeekdaySelection
                id="workdays"
                value={this.state.workdays}
                weekStartDay={this.state.weekStartDay}
                onChange={this.onWorkdaysChange}
                preventEmpty={true}
              />
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
                type="switch"
                id="mailNotifications"
                label={this.props.t("mailNotificationsBookingInfo")}
                checked={this.state.mailNotifications}
                onChange={(e: any) =>
                  this.setState({ mailNotifications: e.target.checked })
                }
              />
              <Form.Check
                type="switch"
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
                {this.props.t("language-" + RuntimeConfig.INFOS.orgLanguage)})
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
                type="switch"
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
          hidden={this.state.activeTab !== "style"}
          className="form-colors"
        >
          <h5 className="margin-top-15">{this.props.t("bookingcolors")}</h5>
          {this.renderBookingColor("booked", "colorAlreadyBooked")}
          {this.renderBookingColor("notBooked", "colorNotBooked")}
          {this.renderBookingColor("selfBooked", "colorSelfBooked")}
          {RuntimeConfig.INFOS.maxHoursPartiallyBookedEnabled &&
            this.renderBookingColor("partiallyBooked", "colorPartiallyBooked")}
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
          hidden={this.state.activeTab !== "security"}
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
              onChange={(e: any) => this.setState({ password: e.target.value })}
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
            this.state.activeTab !== "security" || RuntimeConfig.INFOS.idpLogin
          }
          t={this.props.t}
        />
        <PasskeySettings
          hidden={
            this.state.activeTab !== "security" || RuntimeConfig.INFOS.idpLogin
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
        <div hidden={this.state.activeTab !== "security"}>
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
                        {session.id === this.state.currentSessionId && (
                          <OverlayTrigger
                            placement="top"
                            overlay={
                              <Tooltip>{this.props.t("thisSession")}</Tooltip>
                            }
                          >
                            <span className="preferences-session-icon">
                              <IconCurrentSession size={16} />
                              <span className="visually-hidden">
                                {this.props.t("thisSession")}
                              </span>
                            </span>
                          </OverlayTrigger>
                        )}
                      </td>
                      <td className="text-nowrap">
                        {Formatting.getFormatterShort(false).format(
                          new Date(session.created),
                        )}
                      </td>
                      <td className="text-nowrap">
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

        <div hidden={this.state.activeTab !== "idp"}>
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
            style={{ width: "100%", height: "65vh", borderWidth: 0 }}
            id="idp-profilepage-iframe"
            sandbox="allow-scripts allow-same-origin allow-forms"
          ></iframe>
        </div>

        {/* ------------ */}
        {/* INTEGRATIONS */}
        {/* ------------ */}

        <Form
          onSubmit={this.saveCaldavSettings}
          hidden={this.state.activeTab !== "integration"}
        >
          <h5 className="margin-top-15">{this.props.t("caldavCalendar")}</h5>
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
      </>
    );
  }

  // The search page listens for arrow keys on window to change the booking
  // date and area. React-Bootstrap renders this dialog in a portal, but its
  // synthetic events still bubble through the React tree, so stopping them
  // here keeps them from reaching that window listener. Handlers inside the
  // dialog (e.g. arrow key navigation in Nav) still run, since they fire first.
  // Container inside the modal for date/time picker popups, so they stay
  // within the modal's focus trap. Must be a stable callback ref.
  setPickerContainer = (el: HTMLDivElement | null) => {
    if (el !== this.state.pickerContainer) {
      this.setState({ pickerContainer: el });
    }
  };

  onKeyDown = (e: React.KeyboardEvent) => {
    if (
      e.key === "ArrowLeft" ||
      e.key === "ArrowRight" ||
      e.key === "ArrowUp" ||
      e.key === "ArrowDown"
    ) {
      e.stopPropagation();
    }
  };

  render() {
    const credentials = Ajax.PERSISTER.readCredentialsFromLocalStorage();
    const profilePageUrl = credentials.profilePageUrl;
    const showDialog =
      this.props.show &&
      !this.state.showReloadModal &&
      !this.state.showPasswordChangedModal;

    return (
      <div
        className="preferences-modal-keyboard-scope"
        onKeyDown={this.onKeyDown}
      >
        <Modal
          show={showDialog}
          onHide={this.props.onHide}
          size="lg"
          fullscreen="md-down"
          scrollable={true}
          className="preferences-modal"
          aria-labelledby="preferences-modal-title"
        >
          <Modal.Header closeButton={true}>
            <Modal.Title
              id="preferences-modal-title"
              className="d-flex align-items-center gap-2"
            >
              <IconSettings className="feather" />
              {this.props.t("preferences")}
            </Modal.Title>
          </Modal.Header>
          <Modal.Body className="p-0">
            <div className="preferences-layout">
              <Nav
                variant="pills"
                className="preferences-nav"
                activeKey={this.state.activeTab}
                onSelect={(key) => {
                  if (key) {
                    this.setState({
                      activeTab: key as ActiveTab,
                      error: false,
                      saved: false,
                      caldavError: false,
                    });
                  }
                }}
              >
                {this.renderNavItem("booking", IconBookings, "bookings")}
                {this.renderNavItem("style", IconStyle, "style")}
                {this.renderNavItem(
                  "security",
                  IconSecurity,
                  "security",
                  RuntimeConfig.INFOS.idpLogin,
                )}
                {this.renderNavItem(
                  "idp",
                  IconSecurity,
                  "security",
                  !RuntimeConfig.INFOS.idpLogin || !profilePageUrl,
                )}
                {this.renderNavItem(
                  "integration",
                  IconIntegrations,
                  "integrations",
                )}
              </Nav>
              <div className="preferences-content">{this.renderContent()}</div>
            </div>
          </Modal.Body>
          <div ref={this.setPickerContainer} />
        </Modal>
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
        <ReloadModal
          show={this.state.showReloadModal}
          title={this.props.t(
            this.state.activeTab === "style" ? "style" : "bookings",
          )}
        />
      </div>
    );
  }
}

export default withTranslation(PreferencesModal as any);
