import React from "react";
import { Table, Button } from "react-bootstrap";
import {
  Download as IconDownload,
  X as IconX,
  Check as IconOK,
  RefreshCw as IconRecurring,
} from "react-feather";
import FullLayout from "@/components/FullLayout";
import { NextRouter } from "next/router";
import Loading from "@/components/Loading";
import withReadyRouter from "@/components/withReadyRouter";
import type * as CSS from "csstype";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import RuntimeConfig from "@/components/RuntimeConfig";
import CloudFeatureHint from "@/components/CloudFeatureHint";
import Booking from "@/types/Booking";
import Ajax from "@/util/Ajax";
import Formatting from "@/util/Formatting";
import RedirectUtil from "@/util/RedirectUtil";

interface State {
  data: Booking[];
  loading: boolean;
  updating: boolean;
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
}

class Approvals extends React.Component<Props, State> {
  ExcellentExport: any;

  constructor(props: any) {
    super(props);
    this.state = {
      data: [],
      updating: false,
      loading: true,
    };
  }

  componentDidMount = () => {
    if (!Ajax.hasAccessToken()) {
      RedirectUtil.toLogin(this.props.router);
      return;
    }
    import("excellentexport").then(
      (imp) => (this.ExcellentExport = imp.default),
    );
    this.loadItems();
  };

  loadItems = () => {
    Booking.listPendingApprovals().then((list) => {
      this.setState({
        data: list,
        loading: false,
      });
    });
  };

  approveBooking = (booking: Booking, approve: boolean) => {
    this.setState({
      updating: true,
    });
    booking
      .approve(approve)
      .then(() => {
        this.setState({
          updating: false,
          data: this.state.data.filter((b) => b.id !== booking.id),
        });
      })
      .catch(() => {
        this.setState({ updating: false });
        // Re-Load items as another approver may have approved/declined the booking
        this.loadItems();
      });
  };

  renderItem = (booking: Booking) => {
    const btnStyle: CSS.Properties = {
      ["padding" as any]: "0.1rem 0.3rem",
      ["font-size" as any]: "0.875rem",
      ["border-radius" as any]: "0.2rem",
    };
    return (
      <tr key={booking.id}>
        <td>
          {booking.recurringId ? <IconRecurring className="feather" /> : <></>}
        </td>
        <td>{booking.user.email}</td>
        <td>{booking.space.location.name}</td>
        <td>{booking.space.name}</td>
        <td>{Formatting.getFormatterShort().format(booking.enter)}</td>
        <td>{Formatting.getFormatterShort().format(booking.leave)}</td>
        <td>{booking.subject}</td>
        <td>
          <Button
            variant="success"
            id="approveBookingButton"
            disabled={this.state.updating}
            style={btnStyle}
            onClick={(e) => {
              this.approveBooking(booking, true);
            }}
          >
            <IconOK className="feather" />
          </Button>
        </td>
        <td>
          <Button
            variant="danger"
            id="cancelBookingButton"
            disabled={this.state.updating}
            style={btnStyle}
            onClick={(e) => {
              this.approveBooking(booking, false);
            }}
          >
            <IconX className="feather" />
          </Button>
        </td>
      </tr>
    );
  };

  onFilterSubmit = (e: any) => {
    e.preventDefault();
    this.setState({ loading: true });
    this.loadItems();
  };

  exportTable = (e: any) => {
    return this.ExcellentExport.convert(
      { anchor: e.target, filename: "seatsurfing-bookings", format: "xlsx" },
      [{ name: "Seatsurfing Bookings", from: { table: "datatable" } }],
    );
  };

  render() {
    if (
      RuntimeConfig.INFOS.cloudHosted &&
      !RuntimeConfig.INFOS.subscriptionActive
    ) {
      return (
        <FullLayout headline={this.props.t("approvals")}>
          <CloudFeatureHint />
        </FullLayout>
      );
    }

    // eslint-disable-next-line
    let downloadButton = (
      <a
        download="seatsurfing-approvals.xlsx"
        href="#"
        className="btn btn-sm btn-outline-secondary"
        onClick={this.exportTable}
      >
        <IconDownload className="feather" /> {this.props.t("download")}
      </a>
    );
    let buttons = (
      <>
        {this.state.data && this.state.data.length > 0 ? downloadButton : <></>}
      </>
    );

    if (this.state.loading) {
      return (
        <FullLayout headline={this.props.t("approvals")}>
          <Loading />
        </FullLayout>
      );
    }

    let rows = this.state.data.map((item) => this.renderItem(item));
    if (rows.length === 0) {
      return (
        <FullLayout headline={this.props.t("approvals")} buttons={buttons}>
          <p>{this.props.t("noRecords")}</p>
        </FullLayout>
      );
    }
    return (
      <FullLayout headline={this.props.t("approvals")} buttons={buttons}>
        <Table
          striped={true}
          hover={true}
          className="clickable-table"
          id="datatable"
        >
          <thead>
            <tr>
              <th></th>
              <th>{this.props.t("user")}</th>
              <th>{this.props.t("area")}</th>
              <th>{this.props.t("space")}</th>
              <th>{this.props.t("enter")}</th>
              <th>{this.props.t("leave")}</th>
              <th>{this.props.t("subject")}</th>
              <th></th>
              <th></th>
            </tr>
          </thead>
          <tbody>{rows}</tbody>
        </Table>
      </FullLayout>
    );
  }
}

export default withTranslation(withReadyRouter(Approvals as any));
