import React from 'react';
import {Ajax, AjaxError, Booking, Formatting, Settings as OrgSettings} from 'seatsurfing-commons';
import { Table, Form, Col, Row, Button } from 'react-bootstrap';
import { Plus as IconPlus, Search as IconSearch, Download as IconDownload, X as IconX } from 'react-feather';
import { WithTranslation, withTranslation } from 'next-i18next';
import FullLayout from '@/components/FullLayout';
import { NextRouter } from 'next/router';
import Link from 'next/link';
import Loading from '@/components/Loading';
import withReadyRouter from '@/components/withReadyRouter';
import type * as CSS from 'csstype';


interface State {
  selectedItem: string
  loading: boolean
  start: string
  end: string
}

interface Props extends WithTranslation {
  router: NextRouter
}

class Bookings extends React.Component<Props, State> {
  data: Booking[];
  ExcellentExport: any;
  maxHoursBeforeDelete: number = 0;

  constructor(props: any) {
    super(props);
    this.data = [];
    let end = new Date();
    let start = new Date();
    start.setDate(start.getDate() - 7);
    end.setDate(end.getDate() + 7);
    this.state = {
      selectedItem: "",
      loading: true,
      start: Formatting.getISO8601(start),
      end: Formatting.getISO8601(end),
    };
    this.loadSettings();
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
    let end = new Date(this.state.end);
    end.setUTCHours(23, 59, 59);
    Booking.listFiltered(new Date(this.state.start), end).then(list => {
      this.data = list;
      this.setState({ loading: false });
    });
  }

  cancelBooking = (booking: Booking) => {
      if (!window.confirm(this.props.t("confirmCancelBooking"))) {
        return;
      }
      this.setState({
        loading: true
      });
      booking.delete().then(() => {
        this.loadItems();
      }, (reason: any) => {
        if (reason instanceof AjaxError && reason.httpStatusCode === 403 && reason.appErrorCode === 1007) {
          window.alert(this.props.t("errorDeleteBookingBeforeMaxCancel", {
            num: this.maxHoursBeforeDelete
          }));
        } else {
          window.alert(this.props.t("errorDeleteBooking"));
        }
        this.loadItems();
      });
  }

  loadSettings = async (): Promise<void> => {
    return OrgSettings.list().then(settings => {
      settings.forEach(s => {
        if (s.name === "max_hours_before_delete") {this.maxHoursBeforeDelete = window.parseInt(s.value)};
        // this.setState({ loading: false });
      });
    });
  }

  onItemSelect = (booking: Booking) => {
      this.setState({ selectedItem: booking.id });
  }

  renderItem = (booking: Booking) => {
    const btnStyle: CSS.Properties = {
      ['padding' as any]: '0.1rem 0.3rem',
      ['font-size' as any]: '0.875rem',
      ['border-radius' as any]: '0.2rem',
    };
    return (
      <tr key={booking.id} onClick={() => this.onItemSelect(booking)}>
        <td>{booking.user.email}</td>
        <td>{booking.space.location.name}</td>
        <td>{booking.space.name}</td>
        <td>{Formatting.getFormatterShort().format(booking.enter)}</td>
        <td>{Formatting.getFormatterShort().format(booking.leave)}</td>
        <td><Button variant="danger" id="cancelBookingButton" style={btnStyle} onClick={e => { e.stopPropagation(); this.cancelBooking(booking); }}><IconX className="feather" /></Button></td>
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
    if (this.state.selectedItem) {
      this.props.router.push(`/bookings/${this.state.selectedItem}`);
      return <></>
    }
    let searchButton = <Button className="btn-sm" variant="outline-secondary" type="submit" form="form"><IconSearch className="feather" /> {this.props.t("search")}</Button>;
    // eslint-disable-next-line
    let downloadButton = <a download="seatsurfing-bookings.xlsx" href="#" className="btn btn-sm btn-outline-secondary" onClick={this.exportTable}><IconDownload className="feather" /> {this.props.t("download")}</a>;
    let buttons = (
      <>
        {this.data && this.data.length > 0 ? downloadButton : <></>}
        {searchButton}
        <Link href="/bookings/add" className="btn btn-sm btn-outline-secondary"><IconPlus className="feather" /> {this.props.t("add")}</Link>
      </>
    );
    let form = (
      <Form onSubmit={this.onFilterSubmit} id="form">
        <Form.Group as={Row}>
          <Form.Label column sm="2">{this.props.t("enter")}</Form.Label>
          <Col sm="4">
            <Form.Control type="date" value={this.state.start} onChange={(e: any) => this.setState({ start: e.target.value })} required={true} />
          </Col>
        </Form.Group>
        <Form.Group as={Row}>
          <Form.Label column sm="2">{this.props.t("leave")}</Form.Label>
          <Col sm="4">
            <Form.Control type="date" value={this.state.end} onChange={(e: any) => this.setState({ end: e.target.value })} required={true} />
          </Col>
        </Form.Group>
      </Form>
    );

    if (this.state.loading) {
      return (
        <FullLayout headline={this.props.t("bookings")}>
          {form}
          <Loading />
        </FullLayout>
      );
    }

    let rows = this.data.map(item => this.renderItem(item));
    if (rows.length === 0) {
      return (
        <FullLayout headline={this.props.t("bookings")} buttons={buttons}>
          {form}
          <p>{this.props.t("noRecords")}</p>
        </FullLayout>
      );
    }
    return (
      <FullLayout headline={this.props.t("bookings")} buttons={buttons}>
        {form}
        <Table striped={true} hover={true} className="clickable-table" id="datatable">
          <thead>
            <tr>
              <th>{this.props.t("user")}</th>
              <th>{this.props.t("area")}</th>
              <th>{this.props.t("space")}</th>
              <th>{this.props.t("enter")}</th>
              <th>{this.props.t("leave")}</th>
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

export default withTranslation(['admin'])(withReadyRouter(Bookings as any));
