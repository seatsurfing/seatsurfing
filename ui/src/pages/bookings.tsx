import React from "react";
import Loading from "../components/Loading";
import { Button, Form, Modal } from "react-bootstrap";
import {
  Loader as IconLoad,
  Calendar as IconCalendar,
  MapPin as IconLocation,
  Clock as IconTime,
  AlertCircle as IconPending,
  RefreshCw as IconRecurring,
  Download as IconDownload,
  Trash2 as IconCancel,
} from "react-feather";
import { NextRouter } from "next/router";
import NavBar from "@/components/NavBar";
import withReadyRouter from "@/components/withReadyRouter";
import ErrorText from "@/types/ErrorText";
import { getIcal } from "@/components/Ical";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import Booking from "@/types/Booking";
import RecurringBooking from "@/types/RecurringBooking";
import Formatting from "@/util/Formatting";
import AjaxError from "@/util/AjaxError";
import AlertModal from "@/components/AlertModal";

import { Calendar, View } from "react-big-calendar";
import CustomToolbar from "@/components/calendar/CustomToolbar";
import createCustomEvent, {
  createCustomMonthEvent,
  bookingToCalendarEvent,
  CalendarEvent,
} from "@/components/calendar/CustomEvent";
import moment from "moment-timezone";
import "react-big-calendar/lib/css/react-big-calendar.css";
import { IoCalendarNumber as CalendarIcon } from "react-icons/io5";
import DateUtil from "@/util/DateUtil";
import CalendarUtil from "@/util/CalendarUtil";
import BrowserUtil from "@/util/BrowserUtil";
import UserPreference from "@/types/UserPreference";
import Validation from "@/util/Validation";
import RuntimeConfig from "@/components/RuntimeConfig";

interface State {
  loading: boolean;
  deletingItem: boolean;
  selectedItem: Booking | null;
  selectedFromCalendar: boolean;
  cancelSeries: boolean;
  calendarDate: Date;
  calendarShow: boolean;
  calendarView: View;
  workdays: number[];
  alertMessage: string | null;
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
}

class Bookings extends React.Component<Props, State> {
  data: Booking[];
  workdayStart: string;

  constructor(props: any) {
    super(props);
    this.workdayStart = UserPreference.DEFAULT_WORKDAY_START;
    this.data = [];
    this.state = {
      loading: true,
      deletingItem: false,
      selectedItem: null,
      selectedFromCalendar: false,
      cancelSeries: false,
      calendarDate: new Date(),
      calendarShow: (() =>
        BrowserUtil.tryLocalStorageGetItem(
          BrowserUtil.LOCAL_STORAGE_KEY_MY_BOOKINGS_VIEW,
          "1",
        ) === "1")(),
      calendarView: "week",
      workdays: [],
      alertMessage: null,
    };
  }

  componentDidMount = () => {
    this.loadData();
  };

  loadData = () => {
    Promise.all([
      Booking.list(),
      UserPreference.getOne(UserPreference.PREF_WORKDAY_START),
      UserPreference.getOne(UserPreference.PREF_WORKDAYS),
    ]).then(([list, workDayStart, workdays]) => {
      this.data = list;
      this.workdayStart = Validation.isValidTimeString(workDayStart)
        ? workDayStart
        : UserPreference.DEFAULT_WORKDAY_START;
      const parsedWorkdays = workdays
        ? workdays.split(",").map((v: string) => parseInt(v))
        : [];
      this.setState({ loading: false, workdays: parsedWorkdays });
    });
  };

  onItemPress = (item: Booking, fromCalendar = false) => {
    this.setState({
      selectedItem: item,
      selectedFromCalendar: fromCalendar,
      cancelSeries: false,
    });
  };

  cancelBooking = async () => {
    if (!this.state.selectedItem) {
      return;
    }
    this.setState({
      deletingItem: true,
    });
    let item: any;
    item = this.state.selectedItem;
    if (this.state.cancelSeries && item.isRecurring()) {
      item = await RecurringBooking.get(item.recurringId);
    }
    item.delete().then(
      () => {
        this.setState(
          {
            selectedItem: null,
            deletingItem: false,
            loading: true,
          },
          this.loadData,
        );
      },
      (reason: any) => {
        if (reason instanceof AjaxError && reason.httpStatusCode === 403) {
          this.setState({
            alertMessage: ErrorText.getTextForAppCode(
              reason.appErrorCode,
              this.props.t,
            ),
          });
        } else {
          this.setState({ alertMessage: this.props.t("errorDeleteBooking") });
        }
        this.setState(
          {
            selectedItem: null,
            deletingItem: false,
            loading: true,
          },
          this.loadData,
        );
      },
    );
  };

  getTimeRangeText = (item: Booking): string => {
    const dateFormatter = Formatting.getFormatterDate();
    const multiDay =
      dateFormatter.format(item.enter) !== dateFormatter.format(item.leave);
    if (RuntimeConfig.INFOS.dailyBasisBooking) {
      return multiDay
        ? `${this.props.t("allDay")} · ${dateFormatter.format(item.enter)} – ${dateFormatter.format(item.leave)}`
        : this.props.t("allDay");
    }
    if (multiDay) {
      const formatter = Formatting.getFormatterShort();
      return `${formatter.format(item.enter)} – ${formatter.format(item.leave)}`;
    }
    const timeFormatter = Formatting.getFormatterTime();
    return `${timeFormatter.format(item.enter)} – ${timeFormatter.format(item.leave)}`;
  };

  downloadIcal = (item: Booking) => {
    if (item.isRecurring()) {
      getIcal(item.recurringId, true);
    } else {
      getIcal(item.id);
    }
  };

  renderItem = (item: Booking) => {
    const weekday = Formatting.getFormatterPart({ weekday: "short" });
    const day = Formatting.getFormatterPart({ day: "numeric" });
    const month = Formatting.getFormatterPart({ month: "short" });
    const offsetText = Formatting.getDateOffsetText(item.enter, item.leave);
    const isToday = offsetText === this.props.t("today");
    return (
      <li
        key={item.id}
        className={"my-booking-card" + (isToday ? " my-booking-today" : "")}
      >
        <div className="my-booking-date" aria-hidden="true">
          <span className="my-booking-date-weekday">
            {weekday.format(item.enter)}
          </span>
          <span className="my-booking-date-day">{day.format(item.enter)}</span>
          <span className="my-booking-date-month">
            {month.format(item.enter)}
          </span>
        </div>
        <div className="my-booking-body">
          <div className="my-booking-meta">
            <span className="my-booking-offset">{offsetText}</span>
            {item.approved === false && (
              <span className="badge my-booking-badge my-booking-badge-pending">
                <IconPending className="feather" />
                {this.props.t("approval")}: {this.props.t("pending")}
              </span>
            )}
            {item.isRecurring() && (
              <span className="badge my-booking-badge my-booking-badge-recurring">
                <IconRecurring className="feather" />
                {this.props.t("recurring")}
              </span>
            )}
          </div>
          <div className="my-booking-space">
            {item.space.name}
            {item.subject && (
              <span className="my-booking-subject"> · {item.subject}</span>
            )}
          </div>
          <div className="my-booking-detail">
            <IconLocation className="feather" />
            {item.space.location.name}
          </div>
          <div className="my-booking-detail">
            <IconTime className="feather" />
            {this.getTimeRangeText(item)}
          </div>
        </div>
        <div className="my-booking-actions">
          <button
            type="button"
            className="my-booking-action"
            title={this.props.t("download")}
            aria-label={this.props.t("download")}
            onClick={() => this.downloadIcal(item)}
          >
            <IconDownload className="feather" />
          </button>
          <button
            type="button"
            className="my-booking-action my-booking-action-danger"
            title={this.props.t("cancelBooking")}
            aria-label={this.props.t("cancelBooking")}
            onClick={() => this.onItemPress(item)}
          >
            <IconCancel className="feather" />
          </button>
        </div>
      </li>
    );
  };

  render() {
    if (this.state.loading) {
      return <Loading />;
    }

    if (this.data.length === 0) {
      return (
        <>
          <NavBar />
          <div className="container-signin">
            <div className="my-bookings-empty">
              <IconCalendar className="feather" />
              <p>{this.props.t("noBookings")}</p>
            </div>
          </div>
        </>
      );
    }

    const calendarEvents: CalendarEvent[] = [];
    for (const item of this.data) {
      calendarEvents.push(bookingToCalendarEvent(item, "user", this.props.t));
    }

    const toolbar = (props: object) => (
      <CustomToolbar
        toolbar={props as any}
        t={this.props.t}
        events={calendarEvents}
        minDate={DateUtil.getTodayStart()}
      />
    );

    const calendarLocalizer = CalendarUtil.getLocalizer();

    return (
      <>
        <NavBar />
        <div className="container-signin">
          <div className="d-lg-block d-none search-config-outer">
            <div className="container-search-config">
              <div className="content" style={{ paddingTop: "5px" }}>
                <Form>
                  <Form.Group className="d-flex margin-top-10">
                    <div className="me-2">
                      <CalendarIcon
                        title={this.props.t("map")}
                        color={"var(--bs-secondary-color)"}
                        height="20px"
                        width="20px"
                      />
                    </div>
                    <div className="ms-2 w-100">
                      <Form.Check
                        style={{ textAlign: "start" }}
                        type="switch"
                        checked={this.state.calendarShow}
                        onChange={() => {
                          this.setState(
                            {
                              calendarShow: !this.state.calendarShow,
                            },
                            () => {
                              BrowserUtil.tryLocalStorageSetItem(
                                BrowserUtil.LOCAL_STORAGE_KEY_MY_BOOKINGS_VIEW,
                                this.state.calendarShow ? "1" : "0",
                              );
                            },
                          );
                        }}
                        label={this.props.t("calendar")}
                        aria-label={this.props.t("calendar")}
                        id="switch-control"
                      />
                    </div>
                  </Form.Group>
                </Form>
              </div>
            </div>
          </div>

          {/* classic view */}
          <div
            className={
              !this.state.calendarShow
                ? "my-bookings-list"
                : "my-bookings-list d-lg-none"
            }
          >
            <h1 className="my-bookings-title">
              {this.props.t("myBookings")}
              <span className="my-bookings-count">{this.data.length}</span>
            </h1>
            <ul aria-label={this.props.t("myBookings")}>
              {this.data.map((item) => this.renderItem(item))}
            </ul>
          </div>

          {/* calendar view */}
          <div
            className={this.state.calendarShow ? "d-none d-lg-block" : "d-none"}
            style={{ width: "100%" }}
          >
            <Calendar
              showMultiDayTimes={true}
              getNow={CalendarUtil.getNow}
              localizer={calendarLocalizer}
              formats={CalendarUtil.getFormats()}
              events={calendarEvents}
              startAccessor={CalendarUtil.startAccessor}
              endAccessor={CalendarUtil.endAccessor}
              style={{
                height: "calc(100vh - 160px)",
                width: "100%",
                padding: "10px",
                margin: "auto",
              }}
              view={this.state.calendarView}
              onView={(view: View) => this.setState({ calendarView: view })}
              date={this.state.calendarDate}
              drilldownView="week"
              onNavigate={(newDate: Date, view: View) => {
                const today = DateUtil.getTodayStart();
                const navigateDate = DateUtil.setHoursToMin(new Date(newDate));
                if (navigateDate >= today) {
                  this.setState({ calendarDate: newDate });
                } else if (
                  moment(newDate)
                    .endOf(view === "month" ? "month" : "week")
                    .toDate() >= today
                ) {
                  this.setState({ calendarDate: new Date() });
                }
              }}
              onSelectEvent={(e) => {
                const booking = this.data.find((b) => b.id === e.bookingId);
                if (booking) this.onItemPress(booking, true);
              }}
              culture={Formatting.Language}
              length={7}
              views={["week", "month"]}
              eventPropGetter={CalendarUtil.eventPropGetter}
              components={{
                toolbar,
                event: createCustomEvent(),
                month: { event: createCustomMonthEvent() },
              }}
              scrollToTime={DateUtil.convertToFakeUTCDate(
                DateUtil.getTodayTimeFromTimeString(this.workdayStart),
              )}
              dayPropGetter={CalendarUtil.getDayPropGetter(this.state.workdays)}
            ></Calendar>
          </div>
        </div>

        <Modal
          show={this.state.selectedItem != null}
          onHide={() => this.setState({ selectedItem: null })}
          centered
          className="booking-modal"
          dialogClassName="cancel-booking-modal"
        >
          <Modal.Header closeButton>
            <div className="booking-modal-heading">
              <div className="booking-modal-eyebrow">
                {this.state.selectedItem?.space.location.name}
              </div>
              <Modal.Title>{this.state.selectedItem?.space.name}</Modal.Title>
            </div>
          </Modal.Header>
          <Modal.Body>
            {this.state.selectedItem && (
              <div className="cancel-booking-summary">
                {this.state.selectedItem.subject && (
                  <div className="my-booking-space">
                    {this.state.selectedItem.subject}
                  </div>
                )}
                <div className="my-booking-detail">
                  <IconCalendar className="feather" />
                  {Formatting.getFormatterNoTime().format(
                    this.state.selectedItem.enter,
                  )}
                </div>
                <div className="my-booking-detail">
                  <IconTime className="feather" />
                  {this.getTimeRangeText(this.state.selectedItem)}
                </div>
              </div>
            )}
            {this.state.selectedItem?.isRecurring() && (
              <Form.Check
                type="switch"
                id="cancelAllUpcomingBookings"
                className="cancel-booking-series"
                onChange={(e) =>
                  this.setState({ cancelSeries: e.target.checked })
                }
                checked={this.state.cancelSeries}
                label={this.props.t("cancelAllUpcomingBookings")}
              />
            )}
          </Modal.Body>
          <Modal.Footer
            className={
              "cancel-booking-footer" +
              (this.state.selectedFromCalendar
                ? " cancel-booking-footer-download"
                : "")
            }
          >
            {this.state.selectedFromCalendar && (
              <Button
                variant="light"
                title={this.props.t("download")}
                aria-label={this.props.t("download")}
                onClick={() => {
                  if (this.state.selectedItem) {
                    this.downloadIcal(this.state.selectedItem);
                  }
                }}
              >
                <IconDownload className="feather" />
              </Button>
            )}
            <Button
              variant="danger"
              onClick={() => this.cancelBooking()}
              disabled={this.state.deletingItem}
            >
              {this.state.deletingItem ? (
                <IconLoad className="feather loader" />
              ) : (
                <IconCancel className="feather" />
              )}
              {this.props.t("cancelBooking")}
            </Button>
          </Modal.Footer>
        </Modal>
        <AlertModal
          show={this.state.alertMessage !== null}
          message={this.state.alertMessage || ""}
          onConfirm={() => this.setState({ alertMessage: null })}
        />
      </>
    );
  }
}

export default withTranslation(withReadyRouter(Bookings as any));
