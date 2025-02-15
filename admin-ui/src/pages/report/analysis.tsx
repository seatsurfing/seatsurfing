import React from 'react';
import { Ajax, Formatting, Location } from 'flexspace-commons';
import { Table, Form, Col, Row, Button } from 'react-bootstrap';
import { Search as IconSearch, Download as IconDownload, Check as IconCheck } from 'react-feather';
import { WithTranslation, withTranslation } from 'next-i18next';
import FullLayout from '@/components/FullLayout';
import Loading from '@/components/Loading';
import { NextRouter } from 'next/router';
import withReadyRouter from '@/components/withReadyRouter';

interface State {
  loading: boolean
  start: string
  end: string
  locationId: string
}

interface Props extends WithTranslation {
  router: NextRouter
}

class ReportAnalysis extends React.Component<Props, State> {
  locations: Location[];
  data: any;
  ExcellentExport: any;

  constructor(props: any) {
    super(props);
    this.locations = [];
    this.data = [];
    let end = new Date();
    let start = new Date();
    start.setDate(end.getDate() - 7);
    this.state = {
      loading: true,
      start: Formatting.getISO8601(start),
      end: Formatting.getISO8601(end),
      locationId: "",
    };
  }

  componentDidMount = () => {
    if (!Ajax.CREDENTIALS.accessToken) {
      this.props.router.push("/login");
      return;
    }
    Location.list().then(locations => this.locations = locations);
    import('excellentexport').then(imp => this.ExcellentExport = imp.default);
    this.loadItems();
  }

  loadItems = () => {
    let end = new Date(this.state.end);
    end.setUTCHours(23, 59, 59);
    let payload = {
      start: new Date(this.state.start),
      end: end,
      locationId: this.state.locationId,
    };
    Ajax.postData("/booking/report/presence/", payload).then((res) => {
      this.data = res.json;
      this.setState({ loading: false });
    });
  }

  getRows = () => {
    return this.data.users.map((user: any, i: number) => {
      let j = 0;
      let cols = this.data.presences[i].map((num: number) => {
        let val = num > 0 ? <IconCheck className="feather" /> : "-";
        j++;
        return <td key={'row-' + user.userId + '-' + j} className="center">{val}</td>;
      });
      return (
        <tr key={user.userId}>
          <td className="no-wrap">{user.email}</td>
          {cols}
        </tr>
      );
    });
  }

  onFilterSubmit = (e: any) => {
    e.preventDefault();
    this.setState({ loading: true });
    this.loadItems();
  }

  exportTable = (e: any) => {
    let fixFn = (value: string, row: number, col: number) => {
      if (value.startsWith("<")) {
        return "1";
      }
      if (value === "-") {
        return "0";
      }
      return value;
    }
    return this.ExcellentExport.convert(
      { anchor: e.target, filename: "seatsurfing-analysis", format: "xlsx" },
      [{ name: "Seatsurfing Analysis", from: { table: "datatable" }, fixValue: fixFn }]
    );
  }

  render() {
    let searchButton = <Button className="btn-sm" variant="outline-secondary" type="submit" form="form"><IconSearch className="feather" /> {this.props.t("search")}</Button>;
    // eslint-disable-next-line
    let downloadButton = <a download="seatsurfing-analysis.xlsx" href="#" className="btn btn-sm btn-outline-secondary" onClick={this.exportTable}><IconDownload className="feather" /> {this.props.t("download")}</a>;
    let buttons = (
      <>
        {this.data && this.data.users && this.data.dates && this.data.users.length > 0 && this.data.dates.length > 0 ? downloadButton : <></>}
        {searchButton}
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
        <Form.Group as={Row}>
          <Form.Label column sm="2">{this.props.t("area")}</Form.Label>
          <Col sm="4">
            <Form.Select value={this.state.locationId} onChange={(e: any) => this.setState({ locationId: e.target.value })}>
              <option value="">({this.props.t("all")})</option>
              {this.locations.map(location => <option key={location.id} value={location.id}>{location.name}</option>)}
            </Form.Select>
          </Col>
        </Form.Group>
      </Form>
    );

    if (this.state.loading) {
      return (
        <FullLayout headline={this.props.t("analysis")}>
          {form}
          <Loading />
        </FullLayout>
      );
    }

    if ((this.data.users.length === 0) || (this.data.dates.length === 0)) {
      return (
        <FullLayout headline={this.props.t("analysis")} buttons={buttons}>
          {form}
          <p>{this.props.t("noRecords")}</p>
        </FullLayout>
      );
    }
    return (
      <FullLayout headline={this.props.t("analysis")} buttons={buttons}>
        {form}
        <Table striped={true} hover={true} className="clickable-table" id="datatable" responsive={true}>
          <thead>
            <tr>
              <th className="no-wrap">{this.props.t("user")}</th>
              {this.data.dates.map((date: string) => <th key={'date-' + date} className="no-wrap">{date}</th>)}
            </tr>
          </thead>
          <tbody>
            {this.getRows()}
          </tbody>
        </Table>
      </FullLayout>
    );
  }
}

export default withTranslation(['admin'])(withReadyRouter(ReportAnalysis as any));
