import React from "react";
import {
  BarChart2,
  Calendar,
  ChevronRight,
  Grid,
  Info,
  MapPin,
  PieChart,
  Users,
} from "react-feather";
import {
  Button,
  ButtonGroup,
  Card,
  Row,
  Col,
  ProgressBar,
  Alert,
  Dropdown,
  OverlayTrigger,
  Tooltip,
} from "react-bootstrap";
import { NextRouter } from "next/router";
import WeekdayChart, { WeekdayChartType } from "@/components/WeekdayChart";
import FullLayout from "@/components/FullLayout";
import Loading from "@/components/Loading";
import withReadyRouter from "@/components/withReadyRouter";
import withPermission from "@/components/withPermission";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import RuntimeConfig from "@/components/RuntimeConfig";
import { Permission, PermissionLevel } from "@/types/Permission";
import Link from "next/link";
import PremiumFeatureIcon from "@/components/PremiumFeatureIcon";
import Stats from "@/types/Stats";
import StatsLoad from "@/types/StatsLoad";
import Location from "@/types/Location";
import DateUtil from "@/util/DateUtil";
import Formatting from "@/util/Formatting";

import Navigation from "@/util/Navigation";
import UpdateChecker from "@/util/UpdateChecker";
import SelfHostedHint from "@/components/SelfHostedHint";
import OrgAnnouncement from "@/components/OrgAnnouncement";

interface State {
  loading: boolean;
  redirect: string;
  latestVersion: any;
  selectedUtilizationLocationId: string | null;
  selectedWeekdayLocationId: string | null;
  selectedWeekdayPeriod: string | null;
  weekdayChartType: WeekdayChartType;
  stats: Stats | null;
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
}

class Dashboard extends React.Component<Props, State> {
  locations: Location[];

  constructor(props: any) {
    super(props);
    this.locations = [];
    this.state = {
      loading: true,
      redirect: "",
      latestVersion: null,
      selectedUtilizationLocationId: null,
      selectedWeekdayLocationId: null,
      selectedWeekdayPeriod: null,
      weekdayChartType: "bar",
      stats: null,
    };
  }

  componentDidMount = async () => {
    await Promise.all([
      this.loadItems(),
      this.loadLocations(),
      this.checkUpdates(),
    ]);
    this.setState({ loading: false });
  };

  checkUpdates = async (): Promise<void> => {
    if (RuntimeConfig.INFOS.cloudHosted) {
      return;
    }
    this.setState({ latestVersion: await UpdateChecker.check() });
  };

  loadLocations = async (): Promise<void> => {
    this.locations = await Location.list();
  };

  canViewStats = (): boolean => {
    return RuntimeConfig.hasPermission(
      Permission.Analytics,
      PermissionLevel.Read,
    );
  };

  loadItems = async (): Promise<void> => {
    if (RuntimeConfig.INFOS.hideStats || !this.canViewStats()) {
      return;
    }
    const stats = await Stats.get();
    this.setState({ stats });
  };

  updateLoad = async (locationId: string | null = null): Promise<void> => {
    if (RuntimeConfig.INFOS.hideStats || !this.canViewStats()) {
      return;
    }
    const statsLoad = await StatsLoad.getLoad(locationId);
    const stats = Object.assign(new Stats(), this.state.stats ?? {});
    stats.spaceLoadNextWeek = statsLoad.spaceLoadNextWeek;
    stats.spaceLoadThisWeek = statsLoad.spaceLoadThisWeek;
    stats.spaceLoadLastWeek = statsLoad.spaceLoadLastWeek;
    stats.spaceLoadLastMonth = statsLoad.spaceLoadLastMonth;
    this.setState({ stats, selectedUtilizationLocationId: locationId });
  };

  updateWeekdayChart = async (
    locationId: string | null = null,
    period: string | null = null,
  ): Promise<void> => {
    if (RuntimeConfig.INFOS.hideStats || !this.canViewStats()) {
      return;
    }
    const bookingsByWeekday = await StatsLoad.getWeekday(locationId, period);
    const stats = Object.assign(new Stats(), this.state.stats ?? {});
    stats.bookingsByWeekday = bookingsByWeekday;
    this.setState({
      stats,
      selectedWeekdayLocationId: locationId,
      selectedWeekdayPeriod: period,
    });
  };

  getClickableProps = (link?: string) => {
    if (!link) {
      return {};
    }
    const onActivate = () => this.setState({ redirect: link });
    return {
      role: "button",
      tabIndex: 0,
      onClick: onActivate,
      onKeyDown: (e: React.KeyboardEvent) => {
        if (e.key === "Enter" || e.key === " ") {
          e.preventDefault();
          onActivate();
        }
      },
    };
  };

  renderStatsCard = (
    num: number | undefined,
    title: string,
    icon: React.ReactNode,
    link?: string,
  ) => {
    const clickable = !!link;
    return (
      <Col xs="12" sm="4">
        <Card
          className={
            "dashboard-stat h-100" +
            (clickable ? " dashboard-card-clickable" : "")
          }
          {...this.getClickableProps(link)}
        >
          <Card.Body className="d-flex align-items-center gap-3">
            <div className="dashboard-stat-icon" aria-hidden="true">
              {icon}
            </div>
            <div className="dashboard-stat-text flex-grow-1">
              <div className="dashboard-stat-label text-truncate">{title}</div>
              <div className="dashboard-number">
                {(num ?? 0).toLocaleString()}
              </div>
            </div>
            {clickable && (
              <ChevronRight
                size={18}
                className="dashboard-stat-chevron"
                aria-hidden="true"
              />
            )}
          </Card.Body>
        </Card>
      </Col>
    );
  };

  renderBookingMetric = (
    num: number | undefined,
    title: string,
    link: string,
  ) => {
    return (
      <div
        className="dashboard-booking-metric"
        aria-label={`${this.props.t("bookings")} ${title}: ${num ?? 0}`}
        {...this.getClickableProps(link)}
      >
        <div className="dashboard-stat-label text-truncate">{title}</div>
        <div className="dashboard-number">{(num ?? 0).toLocaleString()}</div>
      </div>
    );
  };

  renderLocationDropdown = (
    selectedLocationId: string | null,
    onSelect: (locationId: string | null) => void,
  ) => {
    return (
      <Dropdown>
        <Dropdown.Toggle variant="outline-secondary" size="sm">
          {selectedLocationId
            ? this.locations.find((e) => e.id == selectedLocationId)?.name
            : this.props.t("allAreas")}
        </Dropdown.Toggle>
        <Dropdown.Menu align="end">
          <Dropdown.Item
            active={!selectedLocationId}
            onClick={() => onSelect(null)}
          >
            {this.props.t("allAreas")}
          </Dropdown.Item>
          {this.locations.map((location) => (
            <Dropdown.Item
              key={location.id}
              active={location.id == selectedLocationId}
              onClick={() => onSelect(location.id)}
            >
              {location.name}
            </Dropdown.Item>
          ))}
        </Dropdown.Menu>
      </Dropdown>
    );
  };

  renderPeriodDropdown = () => {
    const periods: (string | null)[] = [
      null,
      "lastMonth",
      "lastWeek",
      "thisWeek",
      "nextWeek",
    ];
    const selected = this.state.selectedWeekdayPeriod;
    return (
      <Dropdown>
        <Dropdown.Toggle variant="outline-secondary" size="sm">
          {selected ? this.props.t(selected) : this.props.t("allTime")}
        </Dropdown.Toggle>
        <Dropdown.Menu align="end">
          {periods.map((period) => (
            <Dropdown.Item
              key={period ?? "allTime"}
              active={period === selected}
              onClick={() =>
                this.updateWeekdayChart(
                  this.state.selectedWeekdayLocationId,
                  period,
                )
              }
            >
              {this.props.t(period ?? "allTime")}
            </Dropdown.Item>
          ))}
        </Dropdown.Menu>
      </Dropdown>
    );
  };

  renderWeekdayChart = (bookingsByWeekday: number[]) => {
    if (bookingsByWeekday.every((num) => !num)) {
      return (
        <div
          className="d-flex align-items-center justify-content-center text-muted"
          style={{ height: "200px" }}
        >
          {this.props.t("noBookingsInPeriod")}
        </div>
      );
    }
    const weekStartDay = RuntimeConfig.INFOS.weekStartDay ?? 1;
    const order = [0, 1, 2, 3, 4, 5, 6].map(
      (offset) => (weekStartDay + offset) % 7,
    );
    const data = order.map((day) => bookingsByWeekday[day]);
    const labels = order.map((day) => this.props.t(`workday-${day}`));
    return (
      <WeekdayChart
        data={data}
        labels={labels}
        type={this.state.weekdayChartType}
      />
    );
  };

  getLoadDateRanges = () => {
    const now = new Date();
    const y = now.getUTCFullYear();
    const m = now.getUTCMonth();
    const monday = now.getUTCDate() - ((now.getUTCDay() + 6) % 7);
    const week = (offset: number): [Date, Date] => [
      new Date(Date.UTC(y, m, monday + offset)),
      new Date(Date.UTC(y, m, monday + offset + 6)),
    ];
    return {
      nextWeek: week(7),
      thisWeek: week(0),
      lastWeek: week(-7),
      lastMonth: [
        new Date(Date.UTC(y, m - 1, 1)),
        new Date(Date.UTC(y, m, 0)),
      ] as [Date, Date],
    };
  };

  renderProgressBar = (
    num: number | undefined,
    title: string,
    range: [Date, Date],
  ) => {
    if (!num) {
      num = 0;
    }
    const formatter = Formatting.getFormatterDate();
    const dateRange = `${formatter.format(range[0])} – ${formatter.format(range[1])}`;
    let variant = "success";
    if (num >= 75) {
      variant = "warning";
    }
    if (num >= 100) {
      variant = "danger";
    }
    return (
      <div className="dashboard-load">
        <div className="d-flex justify-content-between align-items-baseline mb-1">
          <div>
            <span className="fw-medium">{title}</span>
            <span className="dashboard-load-range ms-2">{dateRange}</span>
          </div>
          <span className="dashboard-load-value">{num} %</span>
        </div>
        <ProgressBar
          now={Math.min(num, 100)}
          variant={variant}
          className="dashboard-load-bar"
          aria-label={`${title}: ${num} %`}
        />
      </div>
    );
  };

  render() {
    if (this.state.redirect) {
      this.props.router.push(this.state.redirect);
      return <></>;
    }

    if (this.state.loading) {
      return (
        <FullLayout headline="Dashboard">
          <Loading />
        </FullLayout>
      );
    }

    const loadRanges = this.getLoadDateRanges();
    let updateHint = <></>;
    if (
      this.state.latestVersion &&
      this.state.latestVersion.updateAvailable &&
      !RuntimeConfig.INFOS.cloudHosted
    ) {
      updateHint = (
        <Row className="mb-4">
          <Col xs="12">
            <Alert variant="warning" className="mb-0">
              <a
                href="https://github.com/seatsurfing/seatsurfing/releases"
                target="_blank"
                rel="noopener noreferrer"
              >
                {this.props.t("updateAvailable", {
                  version: this.state.latestVersion.version,
                })}
              </a>
            </Alert>
          </Col>
        </Row>
      );
    }

    const cloudUpgradeHint =
      RuntimeConfig.hasPermission(
        Permission.OrgSettings,
        PermissionLevel.Admin,
      ) &&
      RuntimeConfig.INFOS.cloudHosted &&
      !RuntimeConfig.INFOS.subscriptionActive ? (
        <Row className="mb-4">
          <Col xs="12">
            <Alert variant="info" className="mb-0">
              <p style={{ fontWeight: "bold" }}>
                <PremiumFeatureIcon
                  style={{ marginLeft: "0px", marginRight: "5px" }}
                />
                {this.props.t("cloudUpgradeHintHeadline")}
              </p>
              <p>
                <Link href="/admin/plugin/subscription/">
                  {this.props.t("cloudUpgradeHintText")}
                </Link>{" "}
                🚀
              </p>
            </Alert>
          </Col>
        </Row>
      ) : (
        <></>
      );

    const yesterdayDateString = DateUtil.getDateString(-1);

    let statsContent = <></>;
    if (RuntimeConfig.INFOS.hideStats || !this.canViewStats()) {
      statsContent = <p>{this.props.t("statsHiddenByAdmin")}</p>;
    } else {
      const stats = this.state.stats;
      statsContent = (
        <>
          <Row className="g-3 mb-4">
            {this.renderStatsCard(
              stats?.numUsers,
              this.props.t("users"),
              <Users size={20} />,
              RuntimeConfig.hasPermission(
                Permission.Users,
                PermissionLevel.Read,
              )
                ? Navigation.adminUsers()
                : "",
            )}
            {this.renderStatsCard(
              stats?.numLocations,
              this.props.t("areas"),
              <MapPin size={20} />,
              Navigation.adminLocations(),
            )}
            {this.renderStatsCard(
              stats?.numSpaces,
              this.props.t("spaces"),
              <Grid size={20} />,
              Navigation.adminLocations(),
            )}
          </Row>
          <Row className="mb-4">
            <Col xs="12">
              <Card className="dashboard-panel">
                <Card.Body>
                  <div className="dashboard-panel-header">
                    <Card.Title
                      as="h2"
                      className="dashboard-panel-title d-flex align-items-center gap-2"
                    >
                      <span
                        className="dashboard-stat-icon dashboard-stat-icon-sm"
                        aria-hidden="true"
                      >
                        <Calendar size={16} />
                      </span>
                      {this.props.t("bookings")}
                    </Card.Title>
                  </div>
                  <div className="dashboard-booking-metrics">
                    {this.renderBookingMetric(
                      stats?.numBookings,
                      this.props.t("allTime"),
                      Navigation.adminBookings(
                        "enter=2000-01-01T00:00&leave=2999-12-31T23:59&filter=enter_leave",
                      ),
                    )}
                    {this.renderBookingMetric(
                      stats?.numBookingsCurrent,
                      this.props.t("current"),
                      Navigation.adminBookings("filter=current"),
                    )}
                    {this.renderBookingMetric(
                      stats?.numBookingsToday,
                      this.props.t("today"),
                      Navigation.adminBookings("filter=today"),
                    )}
                    {this.renderBookingMetric(
                      stats?.numBookingsYesterday,
                      this.props.t("yesterday"),
                      Navigation.adminBookings(
                        `enter=${yesterdayDateString}T00:00&leave=${yesterdayDateString}T23:59&filter=enter_leave`,
                      ),
                    )}
                    {this.renderBookingMetric(
                      stats?.numBookingsThisWeek,
                      this.props.t("thisWeek"),
                      Navigation.adminBookings(
                        `enter=${DateUtil.getThisWeekMondayDateString()}T00:00&leave=${DateUtil.getThisWeekSundayDateString()}T23:59&filter=enter_leave`,
                      ),
                    )}
                  </div>
                </Card.Body>
              </Card>
            </Col>
          </Row>
          <Row className="g-3 mb-4">
            <Col xs="12" xl="6">
              <Card className="dashboard-panel h-100">
                <Card.Body>
                  <div className="dashboard-panel-header">
                    <Card.Title as="h2" className="dashboard-panel-title">
                      {this.props.t("bookingsByWeekday")}
                    </Card.Title>
                    <div className="dashboard-panel-controls">
                      <ButtonGroup size="sm">
                        <Button
                          variant="outline-secondary"
                          active={this.state.weekdayChartType === "bar"}
                          aria-pressed={this.state.weekdayChartType === "bar"}
                          aria-label={this.props.t("chartTypeBar")}
                          title={this.props.t("chartTypeBar")}
                          onClick={() =>
                            this.setState({ weekdayChartType: "bar" })
                          }
                        >
                          <BarChart2 size={14} />
                        </Button>
                        <Button
                          variant="outline-secondary"
                          active={this.state.weekdayChartType === "pie"}
                          aria-pressed={this.state.weekdayChartType === "pie"}
                          aria-label={this.props.t("chartTypePie")}
                          title={this.props.t("chartTypePie")}
                          onClick={() =>
                            this.setState({ weekdayChartType: "pie" })
                          }
                        >
                          <PieChart size={14} />
                        </Button>
                      </ButtonGroup>
                      {this.renderPeriodDropdown()}
                      {this.renderLocationDropdown(
                        this.state.selectedWeekdayLocationId,
                        (locationId) =>
                          this.updateWeekdayChart(
                            locationId,
                            this.state.selectedWeekdayPeriod,
                          ),
                      )}
                    </div>
                  </div>
                  {this.renderWeekdayChart(
                    stats?.bookingsByWeekday ?? [0, 0, 0, 0, 0, 0, 0],
                  )}
                </Card.Body>
              </Card>
            </Col>
            <Col xs="12" xl="6">
              <Card className="dashboard-panel h-100">
                <Card.Body>
                  <div className="dashboard-panel-header">
                    <Card.Title
                      as="h2"
                      className="dashboard-panel-title d-flex align-items-center gap-2"
                    >
                      {this.props.t("utilization")}
                      <OverlayTrigger
                        placement="right"
                        overlay={
                          <Tooltip>
                            {this.props.t("targetUtilizationHoursPerWeek")}:{" "}
                            {RuntimeConfig.INFOS.dailyBasisBooking
                              ? DateUtil.hoursToDay(
                                  RuntimeConfig.INFOS
                                    .targetUtilizationHoursPerWeek,
                                )
                              : RuntimeConfig.INFOS
                                  .targetUtilizationHoursPerWeek}{" "}
                            {RuntimeConfig.INFOS.dailyBasisBooking
                              ? this.props.t("days")
                              : this.props.t("hours")}
                          </Tooltip>
                        }
                      >
                        <Info
                          size={16}
                          tabIndex={0}
                          aria-label={this.props.t(
                            "targetUtilizationHoursPerWeek",
                          )}
                          style={{
                            cursor: "pointer",
                            color: "var(--bs-secondary-color)",
                          }}
                        />
                      </OverlayTrigger>
                    </Card.Title>
                    {this.renderLocationDropdown(
                      this.state.selectedUtilizationLocationId,
                      (locationId) => this.updateLoad(locationId),
                    )}
                  </div>
                  <div className="d-flex flex-column gap-3">
                    {this.renderProgressBar(
                      stats?.spaceLoadNextWeek,
                      this.props.t("nextWeek"),
                      loadRanges.nextWeek,
                    )}
                    {this.renderProgressBar(
                      stats?.spaceLoadThisWeek,
                      this.props.t("thisWeek"),
                      loadRanges.thisWeek,
                    )}
                    {this.renderProgressBar(
                      stats?.spaceLoadLastWeek,
                      this.props.t("lastWeek"),
                      loadRanges.lastWeek,
                    )}
                    {this.renderProgressBar(
                      stats?.spaceLoadLastMonth,
                      this.props.t("lastMonth"),
                      loadRanges.lastMonth,
                    )}
                  </div>
                </Card.Body>
              </Card>
            </Col>
          </Row>
        </>
      );
    }

    return (
      <FullLayout headline="Dashboard">
        <OrgAnnouncement text={RuntimeConfig.INFOS.orgAnnouncement} />
        {cloudUpgradeHint}
        {updateHint}
        <SelfHostedHint />
        {statsContent}
      </FullLayout>
    );
  }
}

export default withTranslation(
  withReadyRouter(withPermission(Dashboard as any) as any),
);
