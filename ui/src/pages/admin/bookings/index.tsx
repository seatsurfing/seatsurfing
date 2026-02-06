import ProfilePicture from "@/components/ProfilePicture";
import React from "react";
import { Table, Form, Col, Row, Button } from "react-bootstrap";
import {
  Plus as IconPlus,
  Search as IconSearch,
  Download as IconDownload,
  X as IconX,
  RefreshCw as IconRecurring,
} from "react-feather";
import { AsyncTypeahead } from "react-bootstrap-typeahead";
import FullLayout from "@/components/FullLayout";
import { NextRouter } from "next/router";
import Link from "next/link";
import Loading from "@/components/Loading";
import withReadyRouter from "@/components/withReadyRouter";
import type * as CSS from "csstype";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import Booking from "@/types/Booking";
import DateUtil from "@/util/DateUtil";
import Formatting from "@/util/Formatting";
import OrgSettings from "@/types/Settings";
import Ajax from "@/util/Ajax";
import AjaxError from "@/util/AjaxError";
import RedirectUtil from "@/util/RedirectUtil";
import DateTimePicker from "@/components/DateTimePicker";
import Search, { SearchOptions } from "@/types/Search";
import RuntimeConfig from "@/components/RuntimeConfig";

interface State {
  selectedItem: string;
  loading: boolean;
  start: Date;
  end: Date;
  filterUser: string;
  filterOption: "enter_leave" | "current" | "today";
  typeaheadOptions: any[];
  typeaheadLoading: boolean;
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
}

class Bookings extends React.Component<Props, State> {
  data: Booking[];
  ExcellentExport: any;
  maxHoursBeforeDelete: number = 0;
  typeahead: any = null;

  constructor(props: any) {
    super(props);
    this.data = [];

    const getDateFromQuery = (
      paramName: string,
      defaultOffsetDays: number,
    ): Date => {
      const queryValue = this.props.router.query[paramName] as string;
      if (queryValue) {
        if (DateUtil.isValidDateTime(queryValue)) {
          return new Date(queryValue);
        } else if (DateUtil.isValidDate(queryValue)) {
          const date = new Date(queryValue);
          return defaultOffsetDays < 0
            ? DateUtil.setHoursToMin(date)
            : DateUtil.setHoursToMax(date);
        }
      }

      const defaultDate = new Date();
      defaultDate.setDate(defaultDate.getDate() + defaultOffsetDays);
      return defaultOffsetDays < 0
        ? DateUtil.setHoursToMin(defaultDate)
        : DateUtil.setHoursToMax(defaultDate);
    };

    this.state = {
      selectedItem: "",
      loading: true,
      start: getDateFromQuery("enter", -7), // default: 7 days in past
      end: getDateFromQuery("leave", +7), // default: 7 days in future
      filterUser: this.props.router.query["user"] as string,
      filterOption:
        this.props.router.query["filter"] === "current"
          ? "current"
          : this.props.router.query["filter"] === "today"
            ? "today"
            : "enter_leave",
      typeaheadOptions: [],
      typeaheadLoading: false,
    };
    this.loadSettings();
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

  updateUrlParams = (
    enter: string,
    leave: string,
    filter: string | null,
    filterUser: string | null,
  ) => {
    const currentPath = this.props.router.pathname;
    const {
      filter: _,
      user: __,
      ...queryWithoutFilter
    } = this.props.router.query;
    const currentQuery = {
      ...queryWithoutFilter,
      enter,
      leave,
      ...(filter !== null && { filter }),
      ...(filterUser && { user: filterUser }),
    };

    this.props.router.replace(
      {
        pathname: currentPath,
        query: currentQuery,
      },
      undefined,
      { shallow: true },
    );
  };

  loadItems = () => {
    const end = DateUtil.setSecondsToMax(this.state.end);

    const startOfToday = new Date();
    startOfToday.setHours(0, 0, 0, 0);
    const endOfToday = new Date();
    endOfToday.setHours(23, 59, 59, 999);

    const bookings =
      this.state.filterOption === "enter_leave"
        ? Booking.listFiltered(this.state.start, end, this.state.filterUser)
        : this.state.filterOption === "today"
          ? Booking.listFiltered(
              startOfToday,
              endOfToday,
              this.state.filterUser,
            )
          : Booking.listCurrent(this.state.filterUser);

    bookings.then((list) => {
      this.data = list;
      this.setState({ loading: false });
      this.updateUrlParams(
        DateUtil.formatToDateTimeString(this.state.start),
        DateUtil.formatToDateTimeString(this.state.end),
        this.state.filterOption === "current" ||
          this.state.filterOption === "today"
          ? this.state.filterOption
          : null,
        this.state.filterUser,
      );
    });
  };

  cancelBooking = (booking: Booking) => {
    let formatter = Formatting.getFormatter();
    if (RuntimeConfig.INFOS.dailyBasisBooking) {
      formatter = Formatting.getFormatterNoTime();
    }
    const confirmMessage = this.props.t("confirmCancelBooking", {
      enter: formatter.format(booking.enter),
    });
    if (!window.confirm(Formatting.decodeHtmlEntities(confirmMessage))) {
      return;
    }
    this.setState({
      loading: true,
    });
    booking.delete().then(
      () => {
        this.loadItems();
      },
      (reason: any) => {
        if (
          reason instanceof AjaxError &&
          reason.httpStatusCode === 403 &&
          reason.appErrorCode === 1007
        ) {
          window.alert(
            this.props.t("errorDeleteBookingBeforeMaxCancel", {
              num: this.maxHoursBeforeDelete,
            }),
          );
        } else {
          window.alert(this.props.t("errorDeleteBooking"));
        }
        this.loadItems();
      },
    );
  };

  loadSettings = async (): Promise<void> => {
    return OrgSettings.list().then((settings) => {
      settings.forEach((s) => {
        if (s.name === "max_hours_before_delete") {
          this.maxHoursBeforeDelete = window.parseInt(s.value);
        }
      });
    });
  };

  onItemSelect = (booking: Booking) => {
    this.setState({ selectedItem: booking.id });
  };

  canCancel = (booking: Booking) => {
    return !DateUtil.isInPast(booking.leave);
  };

  renderItem = (booking: Booking) => {
    const btnStyle: CSS.Properties = {
      ["padding" as any]: "0.1rem 0.3rem",
      ["font-size" as any]: "0.875rem",
      ["border-radius" as any]: "0.2rem",
    };
    return (
      <tr key={booking.id} onClick={() => this.onItemSelect(booking)}>
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
            variant="danger"
            id="cancelBookingButton"
            hidden={!this.canCancel(booking)}
            style={btnStyle}
            onClick={(e) => {
              e.stopPropagation();
              this.cancelBooking(booking);
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

  filterSearch = () => {
    return true;
  };

  onSearchSelected = (selected: any) => {
    this.setState({
      filterUser: selected[0]?.email,
    });
  };

  handleSearch = (query: string) => {
    this.setState({ typeaheadLoading: true });
    let options = new SearchOptions();
    options.includeUsers = true;
    Search.search(query ? query : "", options).then((res) => {
      this.setState({
        typeaheadOptions: res.users,
        typeaheadLoading: false,
      });
    });
  };

  render() {
    if (this.state.selectedItem) {
      this.props.router.push(`/admin/bookings/${this.state.selectedItem}`);
      return <></>;
    }
    let searchButton = (
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
    let downloadButton = (
      <a
        download="seatsurfing-bookings.xlsx"
        href="#"
        className="btn btn-sm btn-outline-secondary"
        onClick={this.exportTable}
      >
        <IconDownload className="feather" /> {this.props.t("download")}
      </a>
    );
    let buttons = (
      <>
        {this.data && this.data.length > 0 ? downloadButton : <></>}
        {searchButton}
        <Link
          href="/admin/bookings/add"
          className="btn btn-sm btn-outline-secondary"
        >
          <IconPlus className="feather" /> {this.props.t("add")}
        </Link>
      </>
    );
    const form = (
      <Form onSubmit={this.onFilterSubmit} id="form">
        <Form.Group as={Row}>
          <Form.Label column sm="2">
            {this.props.t("filter")}
          </Form.Label>
          <Col sm="4">
            <Form.Check
              type="radio"
              label={this.props.t("filterBookingCurrent")}
              name="radioGroup"
              id="radioCurrent"
              value="option2"
              checked={this.state.filterOption === "current"}
              onChange={(e) => this.setState({ filterOption: "current" })}
            />
            <Form.Check
              type="radio"
              label={this.props.t("today")}
              name="radioGroup"
              id="radioToday"
              value="option2"
              checked={this.state.filterOption === "today"}
              onChange={(e) => this.setState({ filterOption: "today" })}
            />
            <Form.Check
              type="radio"
              label={this.props.t("filterBookingEnterLeave")}
              name="radioGroup"
              id="radioEnterLeave"
              value="option1"
              checked={this.state.filterOption === "enter_leave"}
              onChange={(e) => this.setState({ filterOption: "enter_leave" })}
            />
          </Col>
        </Form.Group>

        <Form.Group as={Row}>
          <Form.Label column sm="2">
            {this.props.t("enter")}
          </Form.Label>
          <Col sm="4">
            <DateTimePicker
              value={this.state.start}
              onChange={(value: Date | null) => {
                if (value != null) this.setState({ start: value });
              }}
              disabled={this.state.filterOption !== "enter_leave"}
              clearIcon={null}
              required={true}
              enableTime={true}
            />
          </Col>
        </Form.Group>
        <Form.Group as={Row}>
          <Form.Label column sm="2">
            {this.props.t("leave")}
          </Form.Label>
          <Col sm="4">
            <DateTimePicker
              value={this.state.end}
              onChange={(value: Date | null) => {
                if (value != null) this.setState({ end: value });
              }}
              disabled={this.state.filterOption !== "enter_leave"}
              clearIcon={null}
              required={true}
              enableTime={true}
            />
          </Col>
        </Form.Group>
        <Form.Group as={Row}>
          <Form.Label column sm="2">
            {this.props.t("user")}
          </Form.Label>
          <Col sm="4">
            <AsyncTypeahead
              defaultSelected={[{ email: this.state.filterUser ?? "" }]}
              filterBy={this.filterSearch}
              isLoading={this.state.typeaheadLoading}
              labelKey="email"
              multiple={false}
              minLength={3}
              onChange={this.onSearchSelected}
              onSearch={this.handleSearch}
              options={this.state.typeaheadOptions}
              placeholder={this.props.t("searchForUser")}
              ref={(ref: any) => {
                this.typeahead = ref;
              }}
              renderMenuItemChildren={(option: any) => (
                <div className="d-flex">
                  <ProfilePicture width={24} height={24} />
                  <span style={{ marginLeft: "10px" }}>{option.email}</span>
                </div>
              )}
            />
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

    let rows = this.data.map((item) => this.renderItem(item));
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
        <Table
          striped={true}
          hover={true}
          className="clickable-table caption-top"
          id="datatable"
        >
          <caption>
            {this.props.t("numRecords")}: {rows.length}
          </caption>
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
            </tr>
          </thead>
          <tbody>{rows}</tbody>
        </Table>
      </FullLayout>
    );
  }
}

export default withTranslation(withReadyRouter(Bookings as any));
