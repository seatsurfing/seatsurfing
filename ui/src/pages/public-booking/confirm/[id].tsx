import React from "react";
import { NextRouter } from "next/router";
import withReadyRouter from "@/components/withReadyRouter";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import SeatsurfingLogo from "@/components/SeatsurfingLogo";
import Loading from "@/components/Loading";
import Ajax from "@/util/Ajax";

interface State {
  loading: boolean;
  status: "pending" | "unavailable" | "invalid" | null;
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
}

class ConfirmAnonymousBooking extends React.Component<Props, State> {
  constructor(props: any) {
    super(props);
    this.state = {
      loading: true,
      status: null,
    };
  }

  componentDidMount = () => {
    const { id } = this.props.router.query;
    if (!id) {
      return;
    }
    Ajax.postData("/public-booking/confirm/" + id, {}, () => true)
      .then((res) => {
        const status =
          res.json?.status === "pending" ? "pending" : "unavailable";
        this.setState({ loading: false, status: status });
      })
      .catch(() => this.setState({ loading: false, status: "invalid" }));
  };

  render() {
    if (this.state.loading) {
      return <Loading />;
    }

    let message = this.props.t("anonymousBookingConfirmInvalid");
    if (this.state.status === "pending") {
      message = this.props.t("anonymousBookingConfirmPending");
    } else if (this.state.status === "unavailable") {
      message = this.props.t("anonymousBookingConfirmUnavailable");
    }

    return (
      <div className="container-center">
        <div className="container-center-inner">
          <SeatsurfingLogo />
          <p>{message}</p>
        </div>
      </div>
    );
  }
}

export default withTranslation(withReadyRouter(ConfirmAnonymousBooking as any));
