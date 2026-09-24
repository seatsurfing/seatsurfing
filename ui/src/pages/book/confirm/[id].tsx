import React from "react";
import { NextRouter } from "next/router";
import Link from "next/link";
import { LogIn as IconEnter, LogOut as IconLeave } from "react-feather";
import withReadyRouter from "@/components/withReadyRouter";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import SeatsurfingAppLogo from "@/components/SeatsurfingAppLogo";
import Loading from "@/components/Loading";
import Ajax from "@/util/Ajax";
import Formatting from "@/util/Formatting";

interface State {
  loading: boolean;
  status: "pending" | "unavailable" | "invalid" | null;
  enter: Date | null;
  leave: Date | null;
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
}

// The confirmation token is single-use, so concurrent mounts (e.g. React
// StrictMode mounting the component twice) must share one request per ID
// instead of firing a second one that creates a duplicate booking.
const confirmRequests: Map<
  string,
  ReturnType<typeof Ajax.postData>
> = new Map();

class ConfirmPublicBooking extends React.Component<Props, State> {
  constructor(props: any) {
    super(props);
    this.state = {
      loading: true,
      status: null,
      enter: null,
      leave: null,
    };
  }

  componentDidMount = async () => {
    const { id } = this.props.router.query;
    if (typeof id !== "string" || id.length === 0) {
      return;
    }
    let request = confirmRequests.get(id);
    if (!request) {
      request = Ajax.postData("/public-booking/confirm/" + id, {}, () => true);
      confirmRequests.set(id, request);
    }
    try {
      const res = await request;
      const status = res.json?.status === "pending" ? "pending" : "unavailable";
      this.setState({
        loading: false,
        status: status,
        enter: res.json?.enter
          ? new Date(Formatting.stripTimezoneDetails(res.json.enter))
          : null,
        leave: res.json?.leave
          ? new Date(Formatting.stripTimezoneDetails(res.json.leave))
          : null,
      });
    } catch {
      this.setState({ loading: false, status: "invalid" });
    }
  };

  render() {
    if (this.state.loading) {
      return <Loading />;
    }

    let message = this.props.t("publicBookingConfirmInvalid");
    if (this.state.status === "pending") {
      message = this.props.t("publicBookingConfirmPending");
    } else if (this.state.status === "unavailable") {
      message = this.props.t("publicBookingConfirmUnavailable");
    }

    const showBookingInfo =
      (this.state.status === "pending" ||
        this.state.status === "unavailable") &&
      this.state.enter &&
      this.state.leave;
    const formatter = Formatting.getBookingDateFormatter();

    return (
      <div className="container-center">
        <div className="container-center-inner">
          <SeatsurfingAppLogo />
          <p>{message}</p>
          {showBookingInfo && (
            <p>
              <IconEnter className="feather" />
              &nbsp;{formatter.format(this.state.enter as Date)}
              <br />
              <IconLeave className="feather" />
              &nbsp;{formatter.format(this.state.leave as Date)}
            </p>
          )}
          {this.state.status === "unavailable" && (
            <Link href="/book/" className="btn btn-primary">
              {this.props.t("publicBookingConfirmNewBooking")}
            </Link>
          )}
        </div>
      </div>
    );
  }
}

export default withTranslation(withReadyRouter(ConfirmPublicBooking as any));
