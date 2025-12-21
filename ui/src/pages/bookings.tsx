import React from "react";
import Loading from "../components/Loading";
import { Button, Form, ListGroup, Modal } from "react-bootstrap";
import {
  LogIn as IconEnter,
  LogOut as IconLeave,
  MapPin as IconLocation,
  Clock as IconPending,
  RefreshCw as IconRecurring,
} from "react-feather";
import { NextRouter } from "next/router";
import NavBar from "@/components/NavBar";
import withReadyRouter from "@/components/withReadyRouter";
import RuntimeConfig from "@/components/RuntimeConfig";
import ErrorText from "@/types/ErrorText";
import { Loader as IconLoad, Calendar as IconCalendar } from "react-feather";
import { getIcal } from "@/components/Ical";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import Booking from "@/types/Booking";
import Ajax from "@/util/Ajax";
import RecurringBooking from "@/types/RecurringBooking";
import Formatting from "@/util/Formatting";
import AjaxError from "@/util/AjaxError";
import RedirectUtil from "@/util/RedirectUtil";
import { Calendar, momentLocalizer, View } from "react-big-calendar";
import moment from "moment";
import "react-big-calendar/lib/css/react-big-calendar.css";

interface State {
  loading: boolean;
  deletingItem: boolean;
  selectedItem: Booking | null;
  cancelSeries: boolean;
  calenderView: View;
  calenderDate: Date;
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
      calenderView: "week" as View,
      calenderDate: new Date(),
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

    const localizer = momentLocalizer(moment);

    type Event = {
      start: Date;
      end: Date;
      title: string;
      booking: Booking;
    };

    const events: Event[] = [];
    for (const item of this.data) {
      events.push({
        start: item.enter,
        end: item.leave,
        title: `${item.subject}\n${item.space.location.name}: ${item.space.name}`, // used in tooltip
        booking: item,
      });
    }

    let formatter = Formatting.getFormatter();

    if (RuntimeConfig.INFOS.dailyBasisBooking) {
      formatter = Formatting.getFormatterNoTime();
    }

    const CustomEvent = ({ event }: { event: Event }) => {
      return (
        <div>
          <IconLocation className="feather" />{" "}
          {event.booking.space.location.name}, {event.booking.space.name}
          <br />
          {event.booking.subject}
        </div>
      );
    };

    const calendarMessages = {
      today: this.props.t("today"),
      previous: this.props.t("back"),
      next: this.props.t("next"),
      agenda: this.props.t("agenda"),
      week: this.props.t("week"),
      date: this.props.t("date"),
      time: this.props.t("time"),
      event: this.props.t("bookings"),
    };

    return (
      <>
        <NavBar />
        <div className="container-signin">
          <Calendar
            localizer={localizer}
            events={events}
            startAccessor="start"
            endAccessor="end"
            style={{ height: 500, width: "90%", margin: "auto" }}
            defaultView="week"
            view={this.state.calenderView}
            onView={(view) => this.setState({ calenderView: view })}
            date={this.state.calenderDate}
            onNavigate={(date) => this.setState({ calenderDate: date })}
            onSelectEvent={(e) => {
              this.onItemPress(e.booking);
            }}
            culture={Formatting.Language}
            messages={calendarMessages}
            length={7}
            views={{
              agenda: true,
              week: true,
            }}
            components={{
              event: CustomEvent,
            }}
            scrollToTime={new Date(1970, 1, 1, 8, 0, 0)}
          ></Calendar>
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
            <p hidden={!this.state.selectedItem?.isRecurring()}>
              <Form.Check
                type="checkbox"
                id="cancelAllUpcomingBookings"
                onChange={(e) =>
                  this.setState({ cancelSeries: e.target.checked })
                }
                checked={this.state.cancelSeries}
                label={this.props.t("cancelAllUpcomingBookings")}
              />
            </p>
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
