import React from "react";
import {
  Table,
  Form,
  Col,
  Row,
  Button,
  Modal,
  Badge,
  Pagination,
} from "react-bootstrap";
import { Search as IconSearch } from "react-feather";
import FullLayout from "@/components/FullLayout";
import { NextRouter } from "next/router";
import Loading from "@/components/Loading";
import withReadyRouter from "@/components/withReadyRouter";
import withPermission from "@/components/withPermission";
import { Permission, PermissionLevel } from "@/types/Permission";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import AuthAttempt from "@/types/AuthAttempt";
import DateUtil from "@/util/DateUtil";
import Formatting from "@/util/Formatting";
import DateTimePicker from "@/components/DateTimePicker";

const PAGE_SIZE = 50;

interface State {
  loading: boolean;
  start: Date;
  end: Date;
  filterUser: string;
  filterOutcome: "" | "success" | "failure";
  page: number;
  total: number;
  selectedItem: AuthAttempt | null;
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
}

class Audit extends React.Component<Props, State> {
  data: AuthAttempt[];

  constructor(props: any) {
    super(props);
    this.data = [];

    const getDateFromQuery = (
      paramName: string,
      defaultOffsetDays: number,
    ): Date => {
      const queryValue = this.props.router.query[paramName] as string;
      if (queryValue && DateUtil.isValidDateTime(queryValue)) {
        return new Date(queryValue);
      }
      const defaultDate = new Date();
      defaultDate.setDate(defaultDate.getDate() + defaultOffsetDays);
      return defaultOffsetDays < 0
        ? DateUtil.setHoursToMin(defaultDate)
        : DateUtil.setHoursToMax(defaultDate);
    };

    const outcomeQuery = this.props.router.query["outcome"] as string;
    this.state = {
      loading: true,
      start: getDateFromQuery("start", -7),
      end: getDateFromQuery("end", 0),
      filterUser: (this.props.router.query["user"] as string) ?? "",
      filterOutcome:
        outcomeQuery === "success" || outcomeQuery === "failure"
          ? outcomeQuery
          : "",
      page: 0,
      total: 0,
      selectedItem: null,
    };
  }

  componentDidMount = () => {
    this.loadItems();
  };

  updateUrlParams = () => {
    const currentQuery = {
      start: DateUtil.formatToDateTimeString(this.state.start),
      end: DateUtil.formatToDateTimeString(this.state.end),
      ...(this.state.filterUser && { user: this.state.filterUser }),
      ...(this.state.filterOutcome && { outcome: this.state.filterOutcome }),
    };
    this.props.router.replace(
      {
        pathname: this.props.router.pathname,
        query: currentQuery,
      },
      undefined,
      { shallow: true },
    );
  };

  loadItems = async (page: number = 0) => {
    const result = await AuthAttempt.list({
      start: this.state.start,
      end: DateUtil.setSecondsToMax(this.state.end),
      user: this.state.filterUser,
      success:
        this.state.filterOutcome === ""
          ? undefined
          : this.state.filterOutcome === "success",
      limit: PAGE_SIZE,
      offset: page * PAGE_SIZE,
    });
    this.data = result.items;
    this.setState({ loading: false, total: result.total, page: page });
    this.updateUrlParams();
  };

  onFilterSubmit = (e: any) => {
    e.preventDefault();
    this.setState({ loading: true });
    this.loadItems();
  };

  onPageSelect = (page: number) => {
    this.setState({ loading: true });
    this.loadItems(page);
  };

  getErrorCodeLabel = (errorCode: string): string => {
    if (!errorCode) {
      return "";
    }
    if (AuthAttempt.ERROR_CODES.indexOf(errorCode) < 0) {
      return errorCode;
    }
    return this.props.t("autherror_" + errorCode);
  };

  getMethodLabel = (method: string): string => {
    if (!method || AuthAttempt.METHODS.indexOf(method) < 0) {
      return this.props.t("unknown");
    }
    return this.props.t("authmethod_" + method);
  };

  renderItem = (item: AuthAttempt) => {
    return (
      <tr key={item.id} onClick={() => this.setState({ selectedItem: item })}>
        <td>{Formatting.getFormatterShort().format(item.timestamp)}</td>
        <td>{item.email}</td>
        <td>{this.getMethodLabel(item.method)}</td>
        <td>{item.authProviderName}</td>
        <td>
          {item.successful ? (
            <Badge bg="success">{this.props.t("successful")}</Badge>
          ) : (
            <Badge bg="danger">{this.props.t("failed")}</Badge>
          )}
        </td>
        <td>{this.getErrorCodeLabel(item.errorCode)}</td>
      </tr>
    );
  };

  renderPagination = () => {
    const numPages = Math.ceil(this.state.total / PAGE_SIZE);
    if (numPages <= 1) {
      return <></>;
    }
    const items = [];
    for (let i = 0; i < numPages; i++) {
      items.push(
        <Pagination.Item
          key={i}
          active={i === this.state.page}
          onClick={() => this.onPageSelect(i)}
        >
          {i + 1}
        </Pagination.Item>,
      );
    }
    return (
      <Pagination>
        <Pagination.Prev
          disabled={this.state.page === 0}
          onClick={() => this.onPageSelect(this.state.page - 1)}
        />
        {items}
        <Pagination.Next
          disabled={this.state.page >= numPages - 1}
          onClick={() => this.onPageSelect(this.state.page + 1)}
        />
      </Pagination>
    );
  };

  renderDetailsModal = () => {
    const item = this.state.selectedItem;
    if (!item) {
      return <></>;
    }
    return (
      <Modal
        show={true}
        onHide={() => this.setState({ selectedItem: null })}
        size="lg"
      >
        <Modal.Header closeButton>
          <Modal.Title>{this.props.t("details")}</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <Table>
            <tbody>
              <tr>
                <th>{this.props.t("timestamp")}</th>
                <td>{Formatting.getFormatterShort().format(item.timestamp)}</td>
              </tr>
              <tr>
                <th>{this.props.t("username")}</th>
                <td>{item.email}</td>
              </tr>
              <tr>
                <th>{this.props.t("authMethod")}</th>
                <td>{this.getMethodLabel(item.method)}</td>
              </tr>
              {item.authProviderName ? (
                <tr>
                  <th>{this.props.t("authProvider")}</th>
                  <td>{item.authProviderName}</td>
                </tr>
              ) : null}
              <tr>
                <th>{this.props.t("outcome")}</th>
                <td>
                  {item.successful
                    ? this.props.t("successful")
                    : this.props.t("failed")}
                </td>
              </tr>
              {item.errorCode ? (
                <tr>
                  <th>{this.props.t("errorCode")}</th>
                  <td>
                    {this.getErrorCodeLabel(item.errorCode)} ({item.errorCode})
                  </td>
                </tr>
              ) : null}
              {item.errorDetail ? (
                <tr>
                  <th>{this.props.t("errorDetail")}</th>
                  <td>
                    <pre
                      style={{ whiteSpace: "pre-wrap", wordBreak: "break-all" }}
                    >
                      {item.errorDetail}
                    </pre>
                  </td>
                </tr>
              ) : null}
              <tr>
                <th>{this.props.t("device")}</th>
                <td>{item.device}</td>
              </tr>
              {item.userId ? (
                <tr>
                  <th>{this.props.t("userId")}</th>
                  <td>{item.userId}</td>
                </tr>
              ) : null}
            </tbody>
          </Table>
        </Modal.Body>
        <Modal.Footer>
          <Button
            variant="secondary"
            onClick={() => this.setState({ selectedItem: null })}
          >
            {this.props.t("close")}
          </Button>
        </Modal.Footer>
      </Modal>
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
    const form = (
      <Form onSubmit={this.onFilterSubmit} id="form">
        <Form.Group as={Row}>
          <Form.Label column sm="2" htmlFor="input-start">
            {this.props.t("from")}
          </Form.Label>
          <Col sm="4">
            <DateTimePicker
              id="input-start"
              value={this.state.start}
              onChange={(value: Date | null) => {
                if (value != null) this.setState({ start: value });
              }}
              clearIcon={null}
              required={true}
              enableTime={true}
            />
          </Col>
        </Form.Group>
        <Form.Group as={Row}>
          <Form.Label column sm="2" htmlFor="input-end">
            {this.props.t("end")}
          </Form.Label>
          <Col sm="4">
            <DateTimePicker
              id="input-end"
              value={this.state.end}
              onChange={(value: Date | null) => {
                if (value != null) this.setState({ end: value });
              }}
              clearIcon={null}
              required={true}
              enableTime={true}
            />
          </Col>
        </Form.Group>
        <Form.Group as={Row}>
          <Form.Label column sm="2" htmlFor="input-user">
            {this.props.t("user")}
          </Form.Label>
          <Col sm="4">
            <Form.Control
              id="input-user"
              type="text"
              value={this.state.filterUser}
              placeholder={this.props.t("emailPlaceholder")}
              onChange={(e: any) =>
                this.setState({ filterUser: e.target.value })
              }
            />
          </Col>
        </Form.Group>
        <Form.Group as={Row}>
          <Form.Label column sm="2" htmlFor="outcome-select">
            {this.props.t("outcome")}
          </Form.Label>
          <Col sm="4">
            <Form.Select
              id="outcome-select"
              value={this.state.filterOutcome}
              onChange={(e: any) =>
                this.setState({ filterOutcome: e.target.value })
              }
            >
              <option value="">({this.props.t("all")})</option>
              <option value="success">{this.props.t("successful")}</option>
              <option value="failure">{this.props.t("failed")}</option>
            </Form.Select>
          </Col>
        </Form.Group>
      </Form>
    );

    if (this.state.loading) {
      return (
        <FullLayout headline={this.props.t("audit")}>
          {form}
          <Loading />
        </FullLayout>
      );
    }

    const rows = this.data.map((item) => this.renderItem(item));
    if (rows.length === 0) {
      return (
        <FullLayout headline={this.props.t("audit")} buttons={searchButton}>
          {form}
          <p>{this.props.t("noRecords")}</p>
        </FullLayout>
      );
    }
    return (
      <FullLayout headline={this.props.t("audit")} buttons={searchButton}>
        {form}
        <Table
          striped={true}
          hover={true}
          className="clickable-table caption-top"
          id="datatable"
        >
          <caption>
            {this.props.t("numRecords")}: {this.state.total}
          </caption>
          <thead>
            <tr>
              <th>{this.props.t("timestamp")}</th>
              <th>{this.props.t("username")}</th>
              <th>{this.props.t("authMethod")}</th>
              <th>{this.props.t("authProvider")}</th>
              <th>{this.props.t("outcome")}</th>
              <th>{this.props.t("errorCode")}</th>
            </tr>
          </thead>
          <tbody>{rows}</tbody>
        </Table>
        {this.renderPagination()}
        {this.renderDetailsModal()}
      </FullLayout>
    );
  }
}

export default withTranslation(
  withReadyRouter(
    withPermission(
      Audit as any,
      Permission.AuditLog,
      PermissionLevel.Read,
    ) as any,
  ),
);
