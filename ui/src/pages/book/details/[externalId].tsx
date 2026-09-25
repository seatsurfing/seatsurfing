import React from "react";
import { Button } from "react-bootstrap";
import { NextRouter } from "next/router";
import { LogIn as IconEnter, LogOut as IconLeave } from "react-feather";
import withReadyRouter from "@/components/withReadyRouter";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import SeatsurfingAppLogo from "@/components/SeatsurfingAppLogo";
import Loading from "@/components/Loading";
import ConfirmModal from "@/components/ConfirmModal";
import Ajax from "@/util/Ajax";
import BrowserUtil from "@/util/BrowserUtil";
import Formatting from "@/util/Formatting";

interface PublicBookingDetails {
  enter: Date;
  leave: Date;
  subject: string;
  spaceName: string;
  locationName: string;
}

interface State {
  loading: boolean;
  notFound: boolean;
  deleted: boolean;
  deleting: boolean;
  showDeleteConfirm: boolean;
  booking: PublicBookingDetails | null;
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
}

class PublicBookingDetailsPage extends React.Component<Props, State> {
  constructor(props: any) {
    super(props);
    this.state = {
      loading: true,
      notFound: false,
      deleted: false,
      deleting: false,
      showDeleteConfirm: false,
      booking: null,
    };
  }

  componentDidMount = async () => {
    BrowserUtil.applyLanguageFromQuery();
    const { externalId } = this.props.router.query;
    if (typeof externalId !== "string" || externalId.length === 0) {
      return;
    }
    try {
      const res = await Ajax.get(
        "/public-booking/details/" + encodeURIComponent(externalId),
        () => true,
      );
      this.setState({
        loading: false,
        booking: {
          enter: new Date(Formatting.stripTimezoneDetails(res.json.enter)),
          leave: new Date(Formatting.stripTimezoneDetails(res.json.leave)),
          subject: res.json.subject || "",
          spaceName: res.json.spaceName,
          locationName: res.json.locationName,
        },
      });
    } catch {
      this.setState({ loading: false, notFound: true });
    }
  };

  deleteBooking = () => {
    this.setState({ showDeleteConfirm: true });
  };

  confirmDeleteBooking = async () => {
    const { externalId } = this.props.router.query;
    if (typeof externalId !== "string" || externalId.length === 0) {
      return;
    }
    this.setState({ showDeleteConfirm: false, deleting: true });
    try {
      await Ajax.delete(
        "/public-booking/details/" + encodeURIComponent(externalId),
        () => true,
      );
      this.setState({ deleting: false, deleted: true });
    } catch {
      this.setState({ deleting: false, notFound: true });
    }
  };

  render() {
    if (this.state.loading) {
      return <Loading />;
    }

    if (this.state.notFound || !this.state.booking) {
      return (
        <div className="container-center">
          <div className="container-center-inner">
            <SeatsurfingAppLogo />
            <p>{this.props.t("publicBookingDetailsInvalid")}</p>
          </div>
        </div>
      );
    }

    if (this.state.deleted) {
      return (
        <div className="container-center">
          <div className="container-center-inner">
            <SeatsurfingAppLogo />
            <p>{this.props.t("publicBookingDetailsDeleted")}</p>
          </div>
        </div>
      );
    }

    const formatter = Formatting.getBookingDateFormatter();
    const booking = this.state.booking;

    return (
      <div className="container-center">
        <div className="container-center-inner">
          <SeatsurfingAppLogo />
          <p>
            {booking.locationName} / {booking.spaceName}
            {booking.subject && (
              <>
                <br />
                {booking.subject}
              </>
            )}
          </p>
          <p>
            <IconEnter className="feather" />
            &nbsp;{formatter.format(booking.enter)}
            <br />
            <IconLeave className="feather" />
            &nbsp;{formatter.format(booking.leave)}
          </p>
          <Button
            variant="danger"
            disabled={this.state.deleting}
            onClick={this.deleteBooking}
          >
            {this.props.t("publicBookingDetailsDelete")}
          </Button>
          <ConfirmModal
            show={this.state.showDeleteConfirm}
            message={this.props.t("publicBookingDetailsDeleteConfirm")}
            onCancel={() => this.setState({ showDeleteConfirm: false })}
            onConfirm={this.confirmDeleteBooking}
          />
        </div>
      </div>
    );
  }
}

export default withTranslation(
  withReadyRouter(PublicBookingDetailsPage as any),
);
