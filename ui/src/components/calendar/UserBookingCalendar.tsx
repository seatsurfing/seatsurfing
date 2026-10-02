import React from "react";
import { Calendar, View } from "react-big-calendar";
import "react-big-calendar/lib/css/react-big-calendar.css";
import Loading from "@/components/Loading";
import { TranslationFunc } from "@/components/withTranslation";
import CustomToolbar from "@/components/calendar/CustomToolbar";
import createCustomEvent, {
  bookingToCalendarEvent,
  CalendarEvent,
} from "@/components/calendar/CustomEvent";
import Booking from "@/types/Booking";
import DateUtil from "@/util/DateUtil";
import CalendarUtil from "@/util/CalendarUtil";
import Formatting from "@/util/Formatting";

interface Props {
  t: TranslationFunc;
  userEmail: string;
  onSelectBooking: (bookingId: string) => void;
}

interface State {
  loading: boolean;
  date: Date;
  view: View;
  events: CalendarEvent[];
}

class UserBookingCalendar extends React.Component<Props, State> {
  requestCounter = 0;

  constructor(props: Props) {
    super(props);
    this.state = {
      loading: true,
      date: DateUtil.getNowFakeUTC(),
      view: "week",
      events: [],
    };
  }

  componentDidMount = () => {
    this.loadBookings(this.state.date, this.state.view);
  };

  loadBookings = async (date: Date, view: View) => {
    const requestId = ++this.requestCounter;
    const range = CalendarUtil.getRange(view, date);
    const list = await Booking.listFiltered(
      DateUtil.convertFromFakeUTCDate(range.start),
      DateUtil.convertFromFakeUTCDate(range.end),
      this.props.userEmail,
      "",
    );
    if (requestId !== this.requestCounter) {
      return;
    }
    this.setState({
      loading: false,
      events: list.map((b) => bookingToCalendarEvent(b, "user", this.props.t)),
    });
  };

  onNavigate = (date: Date) => {
    this.setState({ date });
    this.loadBookings(date, this.state.view);
  };

  onView = (view: View) => {
    this.setState({ view });
    this.loadBookings(this.state.date, view);
  };

  render() {
    if (this.state.loading) {
      return <Loading />;
    }
    return (
      <Calendar
        showMultiDayTimes={true}
        getNow={CalendarUtil.getNow}
        localizer={CalendarUtil.getLocalizer()}
        culture={Formatting.Language}
        events={this.state.events}
        startAccessor={CalendarUtil.startAccessor}
        endAccessor={CalendarUtil.endAccessor}
        style={{ height: "calc(100vh - 220px)", width: "100%" }}
        date={this.state.date}
        onNavigate={this.onNavigate}
        view={this.state.view}
        onView={this.onView}
        views={["week", "month"]}
        drilldownView="week"
        onSelectEvent={(e: CalendarEvent) =>
          this.props.onSelectBooking(e.bookingId)
        }
        eventPropGetter={CalendarUtil.eventPropGetter}
        components={{
          toolbar: (props: object) => (
            <CustomToolbar toolbar={props as any} t={this.props.t} />
          ),
          event: createCustomEvent(),
        }}
      />
    );
  }
}

export default UserBookingCalendar;
