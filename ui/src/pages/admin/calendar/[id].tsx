import React from "react";
import { ChevronLeft as IconBack } from "react-feather";
import { NextRouter } from "next/router";
import Link from "next/link";
import { View } from "react-big-calendar";
import FullLayout from "@/components/FullLayout";
import Loading from "@/components/Loading";
import withReadyRouter from "@/components/withReadyRouter";
import withPermission from "@/components/withPermission";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import BookingCalendar, {
  getCalendarRange,
} from "@/components/calendar/BookingCalendar";
import {
  bookingToCalendarEvent,
  CalendarEvent,
} from "@/components/calendar/CustomEvent";
import Booking from "@/types/Booking";
import User from "@/types/User";
import { Permission, PermissionLevel } from "@/types/Permission";
import DateUtil from "@/util/DateUtil";
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

  componentDidMount = () => {
    const { id } = this.props.router.query;
    if (typeof id !== "string" || !id) {
      this.props.router.push("/404");
      return;
    }
    User.get(id).then((user) => {
      if (
        user.accountType === User.AccountTypeServiceAccountRO ||
        user.accountType === User.AccountTypeServiceAccountRW
      ) {
        this.props.router.push("/404");
        return;
      }
      this.user = user;
      this.loadBookings(this.state.date, this.state.view);
    });
  };

  loadBookings = (date: Date, view: View) => {
    if (!this.user) {
      return;
    }
    const requestId = ++this.requestCounter;
    const range = getCalendarRange(date, view);
    Booking.listFiltered(
      DateUtil.convertFromFakeUTCDate(range.start),
      DateUtil.convertFromFakeUTCDate(range.end),
      this.user.email,
      "",
      true,
    ).then((list) => {
      if (requestId !== this.requestCounter) {
        return;
      }
      this.setState({
        loading: false,
        events: list.map((b) =>
          bookingToCalendarEvent(b, "user", this.props.t),
        ),
      });
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
        <BookingCalendar
          t={this.props.t}
          events={this.state.events}
          date={this.state.date}
          onNavigate={this.onNavigate}
          view={this.state.view}
          views={["week", "month"]}
          onView={this.onView}
          onSelectEvent={(e) =>
            this.props.router.push(`/admin/bookings/${e.bookingId}`)
          }
          height="calc(100vh - 220px)"
          showBookingNavigation={false}
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
