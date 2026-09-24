import React from "react";
import { ChevronLeft as IconBack } from "react-feather";
import { NextRouter } from "next/router";
import Link from "next/link";
import { Calendar, View } from "react-big-calendar";
import "react-big-calendar/lib/css/react-big-calendar.css";
import FullLayout from "@/components/FullLayout";
import Loading from "@/components/Loading";
import withReadyRouter from "@/components/withReadyRouter";
import withPermission from "@/components/withPermission";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import createCustomEvent, {
  bookingToCalendarEvent,
  CalendarEvent,
} from "@/components/calendar/CustomEvent";
import Booking from "@/types/Booking";
import User from "@/types/User";
import { Permission, PermissionLevel } from "@/types/Permission";
import DateUtil from "@/util/DateUtil";
import CalendarUtil from "@/util/CalendarUtil";
import Formatting from "@/util/Formatting";
import RendererUtils from "@/util/RendererUtils";

interface State {
  loading: boolean;
  date: Date;
  view: View;
  events: CalendarEvent[];
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
}

class UserBookingCalendar extends React.Component<Props, State> {
  user: User | null = null;
  requestCounter = 0;

  constructor(props: any) {
    super(props);
    this.state = {
      loading: true,
      date: DateUtil.getNowFakeUTC(),
      view: "week",
      events: [],
    };
  }

  componentDidMount = async () => {
    const { id } = this.props.router.query;
    if (typeof id !== "string" || !id) {
      this.props.router.push("/404");
      return;
    }
    const user = await User.get(id);
    if (
      user.accountType === User.AccountTypeServiceAccountRO ||
      user.accountType === User.AccountTypeServiceAccountRW
    ) {
      this.props.router.push("/404");
      return;
    }
    this.user = user;
    await this.loadBookings(this.state.date, this.state.view);
  };

  loadBookings = async (date: Date, view: View) => {
    if (!this.user) {
      return;
    }
    const requestId = ++this.requestCounter;
    const range = CalendarUtil.getRange(date, view);
    const list = await Booking.listFiltered(
      DateUtil.convertFromFakeUTCDate(range.start),
      DateUtil.convertFromFakeUTCDate(range.end),
      this.user.email,
      "",
      true,
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
    const backHref = this.user
      ? `/admin/users/${this.user.id}`
      : "/admin/users";
    const buttons = (
      <Link href={backHref} className="btn btn-sm btn-outline-secondary">
        <IconBack className="feather" /> {this.props.t("back")}
      </Link>
    );
    const headline = this.user
      ? `${this.props.t("bookingCalendar")}: ${RendererUtils.fullname(
          this.user.firstname,
          this.user.lastname,
          this.user.email,
        )}`
      : this.props.t("bookingCalendar");

    if (this.state.loading) {
      return (
        <FullLayout headline={headline} buttons={buttons}>
          <Loading />
        </FullLayout>
      );
    }

    return (
      <FullLayout headline={headline} buttons={buttons}>
        <Calendar
          showMultiDayTimes={true}
          getNow={() => DateUtil.getNowFakeUTC()}
          localizer={CalendarUtil.getLocalizer()}
          culture={Formatting.Language}
          events={this.state.events}
          startAccessor={(event: CalendarEvent) => event.enter}
          endAccessor={(event: CalendarEvent) => event.leave}
          style={{ height: "calc(100vh - 220px)", width: "100%" }}
          date={this.state.date}
          onNavigate={this.onNavigate}
          view={this.state.view}
          views={["week", "month"]}
          onView={this.onView}
          onSelectEvent={(e: CalendarEvent) =>
            this.props.router.push(`/admin/bookings/${e.bookingId}`)
          }
          messages={{
            today: this.props.t("today"),
            previous: this.props.t("previous"),
            next: this.props.t("next"),
            week: this.props.t("week"),
            month: this.props.t("month"),
          }}
          eventPropGetter={(event: CalendarEvent) =>
            event.approved === false ? { style: { opacity: 0.5 } } : {}
          }
          components={{ event: createCustomEvent() }}
        />
      </FullLayout>
    );
  }
}

export default withTranslation(
  withReadyRouter(
    withPermission(
      withPermission(
        UserBookingCalendar as any,
        Permission.Bookings,
        PermissionLevel.Read,
      ) as any,
      Permission.Users,
      PermissionLevel.Read,
    ) as any,
  ),
);
