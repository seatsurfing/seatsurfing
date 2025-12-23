import React from "react";
import { Table, Form, Col, Row, Button, Alert } from "react-bootstrap";
import {
  Search as IconSearch,
  Download as IconDownload,
  Check as IconCheck,
} from "react-feather";
import FullLayout from "@/components/FullLayout";
import Loading from "@/components/Loading";
import { NextRouter } from "next/router";
import withReadyRouter from "@/components/withReadyRouter";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import Formatting from "@/util/Formatting";
import Ajax from "@/util/Ajax";
import Location from "@/types/Location";
import RedirectUtil from "@/util/RedirectUtil";
import DateTimeButtonPicker from "@/components/DateTimeButtonPicker";
import "react-calendar/dist/Calendar.css";
import AjaxError from "@/util/AjaxError";
import ErrorText from "@/types/ErrorText";

interface State {
  loading: boolean;
  start: Date;
  end: Date;
  locationId: string;
  error: boolean;
  errorCode: number;
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
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
      start,
      end,
      locationId: "",
      error: false,
      errorCode: 0,
    };
  }

  componentDidMount = () => {
    if (!Ajax.hasAccessToken()) {
      RedirectUtil.toLogin(this.props.router);
      return;
    }
    Location.list().then((locations) => (this.locations = locations));
    import("excellentexport").then(
      (imp) => (this.ExcellentExport = imp.default),
    );
    this.loadItems();
  };

  loadItems = () => {
    const end = new Date(this.state.end);
    end.setHours(23, 59, 59);
    let params =
      "start=" +
      encodeURIComponent(
        Formatting.convertToFakeUTCDate(this.state.start).toISOString(),
      );
    params +=
      "&end=" +
      encodeURIComponent(Formatting.convertToFakeUTCDate(end).toISOString());
    params += "&locationId=" + encodeURIComponent(this.state.locationId);
    Ajax.get("/booking/report/presence/?" + params)
      .then((res) => {
        this.data = res.json;
        this.setState({ loading: false });
      })
      .catch((e: any) => {
        let errorCode: number = 0;
        if (e instanceof AjaxError) {
          errorCode = e.appErrorCode;
        }
        this.setState({ loading: false, errorCode, error: true });
      });
  };

  getRows = () => {
    return this.data.users.map((user: any, i: number) => {
      let j = 0;
      let cols = this.data.presences[i].map((num: number) => {
        let val = num > 0 ? <IconCheck className="feather" /> : "-";
        j++;
        return (
          <td key={"row-" + user.userId + "-" + j} className="center">
            {val}
          </td>
        );
      });
      return (
        <tr key={user.userId}>
          <td className="no-wrap">{user.email}</td>
          {cols}
        </tr>
      );
    });
  };

  onFilterSubmit = (e: any) => {
    e.preventDefault();
    this.setState({ loading: true, error: false });
    this.loadItems();
  };

  exportTable = (e: any) => {
    const fixFn = (value: string, row: number, col: number) => {
      if (value.startsWith("<")) {
        return "1";
      }
      if (value === "-") {
        return "0";
      }
      return value;
    };
    return this.ExcellentExport.convert(
      { anchor: e.target, filename: "seatsurfing-analysis", format: "xlsx" },
      [
        {
          name: "Seatsurfing Analysis",
          from: { table: "datatable" },
          fixValue: fixFn,
        },
      ],
    );
  };

  render() {
    const searchButton = (
      <Button
        className="btn-sm"
        variant="outline-secondary"
        type="submit"
        form="form"
      >
        <IconSearch className="feather" /> {this.props.t("search")}
      </Button>
    );
    // eslint-disable-next-line
    const downloadButton = (
      <a
        download="seatsurfing-analysis.xlsx"
        href="#"
        className="btn btn-sm btn-outline-secondary"
        onClick={this.exportTable}
      >
        <IconDownload className="feather" /> {this.props.t("download")}
      </a>
    );
    const buttons = (
      <>
        {this.data &&
        this.data.users &&
        this.data.dates &&
        this.data.users.length > 0 &&
        this.data.dates.length > 0 ? (
          downloadButton
        ) : (
          <></>
        )}
        {searchButton}
      </>
    );
    const form = (
      <Form onSubmit={this.onFilterSubmit} id="form">
        <Form.Group as={Row}>
          <Form.Label column sm="2">
            {this.props.t("enter")}
          </Form.Label>
          <Col sm="4">
            <DateTimeButtonPicker
              value={this.state.start}
              onChange={(d: Date) => this.setState({ start: d })}
              dailyOnly={true}
            />
          </Col>
        </Form.Group>
        <Form.Group as={Row}>
          <Form.Label column sm="2">
            {this.props.t("leave")}
          </Form.Label>
          <Col sm="4">
            <DateTimeButtonPicker
              value={this.state.end}
              onChange={(d: Date) => this.setState({ end: d })}
              dailyOnly={true}
              showTodayButton={false}
            />
          </Col>
        </Form.Group>
        <Form.Group as={Row}>
          <Form.Label column sm="2">
            {this.props.t("area")}
          </Form.Label>
          <Col sm="4">
            <Form.Select
              value={this.state.locationId}
              onChange={(e: any) =>
                this.setState({ locationId: e.target.value })
              }
            >
              <option value="">({this.props.t("all")})</option>
              {this.locations.map((location) => (
                <option key={location.id} value={location.id}>
                  {location.name}
                </option>
              ))}
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

    if (this.data.users.length === 0 || this.data.dates.length === 0) {
      return (
        <FullLayout headline={this.props.t("analysis")} buttons={buttons}>
          {form}
          <p>{this.props.t("noRecords")}</p>
        </FullLayout>
      );
    }

    if (this.state.error) {
      return (
        <FullLayout headline={this.props.t("analysis")} buttons={buttons}>
          {form}
          <Alert variant="danger">
            {ErrorText.getTextForAppCode(this.state.errorCode, this.props.t)}
          </Alert>
        </FullLayout>
      );
    }

    return (
      <FullLayout headline={this.props.t("analysis")} buttons={buttons}>
        {form}
        <Table
          striped={true}
          hover={true}
          className="clickable-table"
          id="datatable"
          responsive={true}
        >
          <thead>
            <tr>
              <th className="no-wrap">{this.props.t("user")}</th>
              {this.data.dates.map((date: string) => (
                <th key={"date-" + date} className="no-wrap">
                  {date}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>{this.getRows()}</tbody>
        </Table>
      </FullLayout>
    );
  }
}

export default withTranslation(withReadyRouter(ReportAnalysis as any));
