import React from "react";
import { ChevronLeft as IconBack } from "react-feather";
import { NextRouter } from "next/router";
import Link from "next/link";
import FullLayout from "@/components/FullLayout";
import Loading from "@/components/Loading";
import withReadyRouter from "@/components/withReadyRouter";
import withPermission from "@/components/withPermission";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import UserBookingCalendar from "@/components/calendar/UserBookingCalendar";
import User from "@/types/User";
import { Permission, PermissionLevel } from "@/types/Permission";
import RendererUtils from "@/util/RendererUtils";

interface State {
  user: User | null;
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
}

class UserBookingCalendarPage extends React.Component<Props, State> {
  constructor(props: any) {
    super(props);
    this.state = {
      user: null,
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
    this.setState({ user });
  };

  render() {
    const user = this.state.user;
    const backHref = user ? `/admin/users/${user.id}` : "/admin/users";
    const buttons = (
      <Link href={backHref} className="btn btn-sm btn-outline-secondary">
        <IconBack className="feather" /> {this.props.t("back")}
      </Link>
    );
    const headline = user
      ? `${this.props.t("bookingCalendar")}: ${RendererUtils.fullname(
          user.firstname,
          user.lastname,
          user.email,
        )}`
      : this.props.t("bookingCalendar");

    return (
      <FullLayout headline={headline} buttons={buttons}>
        {user ? (
          <UserBookingCalendar
            t={this.props.t}
            userEmail={user.email}
            onSelectBooking={(id: string) =>
              this.props.router.push(`/admin/bookings/${id}`)
            }
          />
        ) : (
          <Loading />
        )}
      </FullLayout>
    );
  }
}

export default withTranslation(
  withReadyRouter(
    withPermission(
      withPermission(
        UserBookingCalendarPage as any,
        Permission.Bookings,
        PermissionLevel.Read,
      ) as any,
      Permission.Users,
      PermissionLevel.Read,
    ) as any,
  ),
);
