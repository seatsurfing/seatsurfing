import React from 'react';
import { Ajax, Booking, Formatting } from 'seatsurfing-commons';
import { Table, Button } from 'react-bootstrap';
import { Download as IconDownload, X as IconX, Check as IconOK } from 'react-feather';
import { WithTranslation, withTranslation } from 'next-i18next';
import FullLayout from '@/components/FullLayout';
import { NextRouter } from 'next/router';
import Loading from '@/components/Loading';
import withReadyRouter from '@/components/withReadyRouter';
import type * as CSS from 'csstype';

interface State {
  data: Booking[]
  loading: boolean
  updating: boolean
}

interface Props extends WithTranslation {
  router: NextRouter
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
    if (!Ajax.CREDENTIALS.accessToken) {
      this.props.router.push("/login");
      return;
    }
    import('excellentexport').then(imp => this.ExcellentExport = imp.default);
    this.loadItems();
  }

  loadItems = () => {
    Booking.listPendingApprovals().then(list => {
      this.setState({
        data: list,
        loading: false,
      });
    });
  }

  approveBooking = (booking: Booking, approve: boolean) => {
    this.setState({
      updating: true,
    });
    booking.approve(approve).then(() => {
      this.setState({
        updating: false,
        data: this.state.data.filter(b => b.id !== booking.id),
      });
    }
    ).catch(() => {
      this.setState({ updating: false });
    });
  }

  renderItem = (booking: Booking) => {
    const btnStyle: CSS.Properties = {
      ['padding' as any]: '0.1rem 0.3rem',
      ['font-size' as any]: '0.875rem',
      ['border-radius' as any]: '0.2rem',
    };
    return (
      <tr key={booking.id}>
        <td>{booking.user.email}</td>
        <td>{booking.space.location.name}</td>
        <td>{booking.space.name}</td>
        <td>{Formatting.getFormatterShort().format(booking.enter)}</td>
        <td>{Formatting.getFormatterShort().format(booking.leave)}</td>
        <td><Button variant="success" id="approveBookingButton" disabled={this.state.updating} style={btnStyle} onClick={e => { this.approveBooking(booking, true); }}><IconOK className="feather" /></Button></td>
        <td><Button variant="danger" id="cancelBookingButton" disabled={this.state.updating} style={btnStyle} onClick={e => { this.approveBooking(booking, false); }}><IconX className="feather" /></Button></td>
      </tr>
    );
  }

  onFilterSubmit = (e: any) => {
    e.preventDefault();
    this.setState({ loading: true });
    this.loadItems();
  }

  exportTable = (e: any) => {
    return this.ExcellentExport.convert(
      { anchor: e.target, filename: "seatsurfing-bookings", format: "xlsx" },
      [{ name: "Seatsurfing Bookings", from: { table: "datatable" } }]
    );
  }

  render() {
    // eslint-disable-next-line
    let downloadButton = <a download="seatsurfing-approvals.xlsx" href="#" className="btn btn-sm btn-outline-secondary" onClick={this.exportTable}><IconDownload className="feather" /> {this.props.t("download")}</a>;
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

    let rows = this.state.data.map(item => this.renderItem(item));
    if (rows.length === 0) {
      return (
        <FullLayout headline={this.props.t("approvals")} buttons={buttons}>
          <p>{this.props.t("noRecords")}</p>
        </FullLayout>
      );
    }
    return (
      <FullLayout headline={this.props.t("approvals")} buttons={buttons}>
        <Table striped={true} hover={true} className="clickable-table" id="datatable">
          <thead>
            <tr>
              <th>{this.props.t("user")}</th>
              <th>{this.props.t("area")}</th>
              <th>{this.props.t("space")}</th>
              <th>{this.props.t("enter")}</th>
              <th>{this.props.t("leave")}</th>
              <th></th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {rows}
          </tbody>
        </Table>
      </FullLayout>
    );
  }
}

export default withTranslation(['admin'])(withReadyRouter(Approvals as any));
