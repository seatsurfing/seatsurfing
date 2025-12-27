import React from "react";
import Loading from "../components/Loading";
import { Button, Form, ListGroup, Modal } from "react-bootstrap";
import Link from "next/link";
import {
  Loader as IconLoad,
  Calendar as IconCalendar,
  LogIn as IconEnter,
  LogOut as IconLeave,
  MapPin as IconLocation,
  Clock as IconPending,
  RefreshCw as IconRecurring,
  Trello as IconTrello,
  ArrowLeft as IconArrowLeft,
  ArrowRight as IconArrowRight,
} from "react-feather";
import { NextRouter } from "next/router";
import NavBar from "@/components/NavBar";
import withReadyRouter from "@/components/withReadyRouter";
import RuntimeConfig from "@/components/RuntimeConfig";
import ErrorText from "@/types/ErrorText";
import { getIcal } from "@/components/Ical";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import Booking from "@/types/Booking";
import Ajax from "@/util/Ajax";
import RecurringBooking from "@/types/RecurringBooking";
import Formatting from "@/util/Formatting";
import AjaxError from "@/util/AjaxError";
import RedirectUtil from "@/util/RedirectUtil";
import { Calendar, momentLocalizer, ToolbarProps } from "react-big-calendar";
import moment from "moment-timezone";
import "react-big-calendar/lib/css/react-big-calendar.css";
import { IoCalendarNumber as CalendarIcon } from "react-icons/io5";

interface State {
  loading: boolean;
  deletingItem: boolean;
  selectedItem: Booking | null;
  cancelSeries: boolean;
  calendarDate: Date;
  calendarShow: boolean;
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
}

class Bookings extends React.Component<Props, State> {
  data: Booking[];

  constructor(props: any) {
    super(props);
    this.data = [];
    this.state = {
      loading: true,
      deletingItem: false,
      selectedItem: null,
      cancelSeries: false,
      calendarDate: new Date(),
      calendarShow: true,
    };
  }

  componentDidMount = () => {
    if (!Ajax.hasAccessToken()) {
      RedirectUtil.toLogin(this.props.router);
      return;
    }
    this.loadData();
  };

  loadData = () => {
    Booking.list().then((list) => {
      this.data = list;
      this.setState({ loading: false });
    });
  };

  onItemPress = (item: Booking) => {
    this.setState({
      selectedItem: item,
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
          this.loadData
        );
      },
      (reason: any) => {
        if (reason instanceof AjaxError && reason.httpStatusCode === 403) {
          window.alert(
            ErrorText.getTextForAppCode(reason.appErrorCode, this.props.t)
          );
        } else {
          window.alert(this.props.t("errorDeleteBooking"));
        }
        this.setState(
          {
            selectedItem: null,
            deletingItem: false,
            loading: true,
          },
          this.loadData
        );
      }
    );
  };

  renderItem = (item: Booking) => {
    let formatter = Formatting.getFormatter();
    if (RuntimeConfig.INFOS.dailyBasisBooking) {
      formatter = Formatting.getFormatterNoTime();
    }
    let pending = <></>;
    if (item.approved === false) {
      pending = (
        <>
          <IconPending className="feather" />
          &nbsp;{this.props.t("approval")}: {this.props.t("pending")}
          <br />
        </>
      );
    }
    let recurringIcon = <></>;
    if (item.isRecurring()) {
      recurringIcon = (
        <IconRecurring className="feather recurring-booking-icon" />
      );
    }
    return (
      <ListGroup.Item
        key={item.id}
        action={true}
        onClick={(e) => {
          e.preventDefault();
          this.onItemPress(item);
        }}
      >
        <h5>{Formatting.getDateOffsetText(item.enter, item.leave)}</h5>
        {recurringIcon}
        <h6 hidden={!item.subject}>{item.subject}</h6>
        <p>
          {pending}
          <IconLocation className="feather" />
          &nbsp;{item.space.location.name}, {item.space.name}
          <br />
          <IconEnter className="feather" />
          &nbsp;{formatter.format(item.enter)}
          <br />
          <IconLeave className="feather" />
          &nbsp;{formatter.format(item.leave)}
        </p>
      </ListGroup.Item>
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
            <Form className="form-signin">
              <p>{this.props.t("noBookings")}</p>
            </Form>
          </div>
        </>
      );
    }

    type Event = {
      start: Date;
      end: Date;
      title: string;
      booking: Booking;
    };

    const calendarEvents: Event[] = [];
    for (const item of this.data) {
      let title = `${item.space.location.name} (${item.space.name})`;
      if (item.subject) {
        title += `, ${item.subject}`;
      }
      if (item.isRecurring()) {
        title += ` (${this.props.t("recurring")})`;
      }

      calendarEvents.push({
        start: item.enter,
        end: item.leave,
        title, // used in tooltip
        booking: item,
      });
    }

    let formatter = Formatting.getFormatter();
    if (RuntimeConfig.INFOS.dailyBasisBooking) {
      formatter = Formatting.getFormatterNoTime();
    }

    const CustomEvent = ({ event }: { event: Event }) => {
      // show no information for events < 1 hr
      if (
        event.booking.leave.getTime() - event.booking.enter.getTime() <=
        60 * 60 * 1000
      ) {
        return null;
      }

      let pending = <></>;
      if (event.booking.approved === false) {
        pending = (
          <>
            <IconPending
              className="feather"
              style={{ width: "12px", height: "12px" }}
            />
            &nbsp;{this.props.t("approval")}: {this.props.t("pending")}
            <br />
          </>
        );
      }
      let recurringIcon = <></>;
      if (event.booking.isRecurring()) {
        recurringIcon = (
          <IconRecurring
            className="feather recurring-booking-icon"
            style={{ width: "12px", height: "12px", top: "4px", right: "4px" }}
          />
        );
      }

      return (
        <div style={{ fontSize: "12px" }}>
          {recurringIcon}
          <p hidden={!event.booking.subject}>
            <strong>{event.booking.subject}</strong>
          </p>
          {pending}
          <IconLocation
            className="feather"
            style={{ width: "12px", height: "12px" }}
          />{" "}
          {event.booking.space.location.name}, {event.booking.space.name}
          <br />
        </div>
      );
    };

    const CustomToolbar = (toolbar: ToolbarProps) => {
      const goToBack = () => {
        toolbar.onNavigate("PREV");
      };

      const goToNext = () => {
        toolbar.onNavigate("NEXT");
      };

      const goToToday = () => {
        toolbar.onNavigate("TODAY");
      };

      const weekStart = moment(toolbar.date).clone().startOf("week");
      const weekEnd = moment(toolbar.date).clone().endOf("week");
      const formatter = Formatting.getFormatterDate();

      return (
        <div
          className="custom-toolbar"
          style={{ marginBottom: "5px", textAlign: "left" }}
        >
          <Link
            href="#"
            className="btn btn-sm btn-outline-secondary"
            onClick={goToToday}
          >
            <IconTrello className="feather" /> {this.props.t("today")}
          </Link>{" "}
          <Link
            href="#"
            className="btn btn-sm btn-outline-secondary"
            onClick={goToBack}
          >
            <IconArrowLeft className="feather" />
          </Link>{" "}
          <Link
            href="#"
            className="btn btn-sm btn-outline-secondary"
            onClick={goToNext}
          >
            <IconArrowRight className="feather" />
          </Link>{" "}
          <span
            className="toolbar-label"
            style={{
              display: "flex",
              float: "right",
              height: "100%",
              alignItems: "center",
            }}
          >
            {formatter.format(weekStart.toDate())} -{" "}
            {formatter.format(weekEnd.toDate())}
          </span>
        </div>
      );
    };

    const calendarMessages = {
      today: this.props.t("today"),
      previous: this.props.t("previous"),
      next: this.props.t("next"),
      week: this.props.t("week"),
      date: this.props.t("date"),
      time: this.props.t("time"),
      event: this.props.t("bookings"),
      noEventsInRange: this.props.t("noEventsInRange"),
    };

    moment.tz.setDefault("UTC");
    moment.locale(Formatting.Language);
    const calendarLocalizer = momentLocalizer(moment);

    return (
      <>
        <NavBar />
        <div className="container-signin">
          <div className="d-lg-block d-none container-search-config">
            <div className="content" style={{ paddingTop: "5px" }}>
              <Form>
                <Form.Group className="d-flex margin-top-10">
                  <div className="me-2">
                    <CalendarIcon
                      title={this.props.t("map")}
                      color={"#555"}
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
                        this.setState({
                          calendarShow: !this.state.calendarShow,
                        });
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

          {/* classic view */}
          <Form
            className={
              !this.state.calendarShow ? "form-signin" : "form-signin d-lg-none"
            }
          >
            <ListGroup>
              {this.data.map((item) => this.renderItem(item))}
            </ListGroup>
          </Form>

          {/* calendar view */}
          <div
            className={this.state.calendarShow ? "d-none d-lg-block" : "d-none"}
            style={{ width: "100%" }}
          >
            <Calendar
              localizer={calendarLocalizer}
              events={calendarEvents}
              startAccessor="start"
              endAccessor="end"
              style={{ height: 500, width: "90%", margin: "auto" }}
              defaultView="week"
              date={this.state.calendarDate}
              onNavigate={(newDate: Date) => {
                const today = new Date();
                today.setHours(0, 0, 0, 0);
                const navigateDate = new Date(newDate);
                navigateDate.setHours(0, 0, 0, 0);
                if (navigateDate >= today) {
                  this.setState({ calendarDate: newDate });
                }
              }}
              onSelectEvent={(e) => {
                this.onItemPress(e.booking);
              }}
              culture={Formatting.Language}
              messages={calendarMessages}
              length={7}
              views={["week"]}
              components={{
                toolbar: CustomToolbar,
                event: CustomEvent,
              }}
              scrollToTime={new Date(Date.UTC(1970, 1, 1, 8, 0, 0))}
            ></Calendar>
          </div>
        </div>

        <Modal
          show={this.state.selectedItem != null}
          onHide={() => this.setState({ selectedItem: null })}
        >
          <Modal.Header closeButton>
            <Modal.Title>{this.props.t("cancelBooking")}</Modal.Title>
          </Modal.Header>
          <Modal.Body>
            <h6 hidden={!this.state.selectedItem?.subject}>
              {this.state.selectedItem?.subject}
            </h6>
            <p
              dangerouslySetInnerHTML={{
                __html: this.props.t("confirmCancelBooking", {
                  enter: formatter.format(this.state.selectedItem?.enter),
                }),
              }}
            ></p>
            <div hidden={!this.state.selectedItem?.isRecurring()}>
              <Form.Check
                type="checkbox"
                id="cancelAllUpcomingBookings"
                onChange={(e) =>
                  this.setState({ cancelSeries: e.target.checked })
                }
                checked={this.state.cancelSeries}
                label={this.props.t("cancelAllUpcomingBookings")}
              />
            </div>
          </Modal.Body>
          <Modal.Footer>
            <Button
              variant="secondary"
              onClick={() => this.setState({ selectedItem: null })}
              disabled={this.state.deletingItem}
            >
              {this.props.t("back")}
            </Button>
            <Button
              variant="secondary"
              onClick={() => {
                if (this.state.selectedItem?.isRecurring()) {
                  getIcal(this.state.selectedItem.recurringId, true);
                } else {
                  getIcal(
                    this.state.selectedItem ? this.state.selectedItem.id : ""
                  );
                }
              }}
            >
              <IconCalendar
                className="feather"
                style={{ marginRight: "5px" }}
              />{" "}
              Event
            </Button>
            <Button
              variant="danger"
              onClick={() => this.cancelBooking()}
              disabled={this.state.deletingItem}
            >
              {this.props.t("cancelBooking")}
              {this.state.deletingItem ? (
                <IconLoad
                  className="feather loader"
                  style={{ marginLeft: "5px" }}
                />
              ) : (
                <></>
              )}
            </Button>
          </Modal.Footer>
        </Modal>
      </>
    );
  }
}

export default withTranslation(withReadyRouter(Bookings as any));
