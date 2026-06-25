import React from "react";
import { Info } from "react-feather";
import {
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
import WeekdayChart from "@/components/WeekdayChart";
import FullLayout from "@/components/FullLayout";
import Loading from "@/components/Loading";
import withReadyRouter from "@/components/withReadyRouter";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import RuntimeConfig from "@/components/RuntimeConfig";
import Link from "next/link";
import PremiumFeatureIcon from "@/components/PremiumFeatureIcon";
import Stats from "@/types/Stats";
import StatsLoad from "@/types/StatsLoad";
import Location from "@/types/Location";
import User from "@/types/User";
import DateUtil from "@/util/DateUtil";

import Navigation from "@/util/Navigation";
import UpdateChecker from "@/util/UpdateChecker";
import CloudHint from "@/components/CloudHint";

interface State {
  loading: boolean;
  redirect: string;
  spaceAdmin: boolean;
  orgAdmin: boolean;
  latestVersion: any;
  selectedUtilizationLocationId: string | null;
  selectedWeekdayLocationId: string | null;
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
      spaceAdmin: false,
      orgAdmin: false,
      latestVersion: null,
      selectedUtilizationLocationId: null,
      selectedWeekdayLocationId: null,
      stats: null,
    };
  }

  componentDidMount = async () => {
    await Promise.all([
      this.loadItems(),
      this.loadLocations(),
      this.getUserInfo(),
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

  getUserInfo = async (): Promise<void> => {
    const user = await User.getSelf();
    this.setState({ spaceAdmin: user.spaceAdmin, orgAdmin: user.admin });
  };

  loadLocations = async (): Promise<void> => {
    this.locations = await Location.list();
  };

  loadItems = async (): Promise<void> => {
    if (RuntimeConfig.INFOS.hideStats) {
      return;
    }
    const stats = await Stats.get();
    this.setState({ stats });
  };

  updateLoad = async (locationId: string): Promise<void> => {
    if (RuntimeConfig.INFOS.hideStats) {
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

  updateWeekdayChart = async (locationId: string): Promise<void> => {
    if (RuntimeConfig.INFOS.hideStats) {
      return;
    }
    const bookingsByWeekday = await StatsLoad.getWeekday(locationId);
    const stats = Object.assign(new Stats(), this.state.stats ?? {});
    stats.bookingsByWeekday = bookingsByWeekday;
    this.setState({ stats, selectedWeekdayLocationId: locationId });
  };

  renderStatsCard = (num: number | undefined, title: string, link?: string) => {
    return (
      <Col sm="3" xl="2">
        <Card
          {...(link
            ? {
                className: "dashboard-card-clickable",
                onClick: () => this.setState({ redirect: link }),
              }
            : {})}
        >
          <Card.Body>
            <Card.Title className="dashboard-number text-center">
              {num}
            </Card.Title>
            <Card.Subtitle className="text-center mb-2 text-muted">
              {title}
            </Card.Subtitle>
          </Card.Body>
        </Card>
      </Col>
    );
  };

  renderWeekdayChart = (bookingsByWeekday: number[]) => {
    const labels = bookingsByWeekday.map((_, index) =>
      this.props.t(`workday-${index}`),
    );
    return <WeekdayChart data={bookingsByWeekday} labels={labels} />;
  };

  renderProgressBar = (num: number | undefined, title: string) => {
    if (!num) {
      num = 0;
    }
    const label = `${title}: ${num} %`;
    let variant = "success";
    if (num >= 75) {
      variant = "warning";
    }
    if (num >= 100) {
      variant = "danger";
    }
    return (
      <div>
        {label} <ProgressBar now={num} className="mb-3" variant={variant} />
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

    let updateHint = <></>;
    if (
      this.state.latestVersion &&
      this.state.latestVersion.updateAvailable &&
      !RuntimeConfig.INFOS.cloudHosted
    ) {
      updateHint = (
        <Row className="mb-4">
          <Col sm="8">
            <Alert variant="warning">
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
      RuntimeConfig.INFOS.orgAdmin &&
      RuntimeConfig.INFOS.cloudHosted &&
      !RuntimeConfig.INFOS.subscriptionActive ? (
        <Row className="mb-4">
          <Col sm="8">
            <Alert variant="info">
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
    if (RuntimeConfig.INFOS.hideStats) {
      statsContent = <p>{this.props.t("statsHiddenByAdmin")}</p>;
    } else {
      statsContent = (
        <>
          <Row className="mb-4">
            {this.renderStatsCard(
              this.state.stats?.numUsers,
              this.props.t("users"),
              this.state.orgAdmin ? Navigation.adminUsers() : "",
            )}
            {this.renderStatsCard(
              this.state.stats?.numLocations,
              this.props.t("areas"),
              Navigation.adminLocations(),
            )}
            {this.renderStatsCard(
              this.state.stats?.numSpaces,
              this.props.t("spaces"),
              Navigation.adminLocations(),
            )}
            {this.renderStatsCard(
              this.state.stats?.numBookings,
              this.props.t("bookings"),
              Navigation.adminBookings(
                "enter=2000-01-01T00:00&leave=2999-12-31T23:59&filter=enter_leave",
              ),
            )}
          </Row>
          <Row className="mb-4">
            {this.renderStatsCard(
              this.state.stats?.numBookingsCurrent,
              this.props.t("current"),
              Navigation.adminBookings("filter=current"),
            )}
            {this.renderStatsCard(
              this.state.stats?.numBookingsToday,
              this.props.t("today"),
              Navigation.adminBookings("filter=today"),
            )}
            {this.renderStatsCard(
              this.state.stats?.numBookingsYesterday,
              this.props.t("yesterday"),
              Navigation.adminBookings(
                `enter=${yesterdayDateString}T00:00&leave=${yesterdayDateString}T23:59&filter=enter_leave`,
              ),
            )}
            {this.renderStatsCard(
              this.state.stats?.numBookingsThisWeek,
              this.props.t("thisWeek"),
              Navigation.adminBookings(
                `enter=${DateUtil.getThisWeekMondayDateString()}T00:00&leave=${DateUtil.getThisWeekSundayDateString()}T23:59&filter=enter_leave`,
              ),
            )}
          </Row>
          <Row className="mb-4">
            <Col sm="12" xl="8">
              <Card>
                <Card.Body>
                  <div className="d-flex justify-content-between align-items-center mb-3">
                    <Card.Title className="mb-0">
                      {this.props.t("bookingsByWeekday")}
                    </Card.Title>
                    <Dropdown>
                      <Dropdown.Toggle variant="outline-secondary" size="sm">
                        {this.state.selectedWeekdayLocationId
                          ? this.locations.find(
                              (e) =>
                                e.id == this.state.selectedWeekdayLocationId,
                            )?.name
                          : this.props.t("allAreas")}
                      </Dropdown.Toggle>
                      <Dropdown.Menu align="end">
                        <Dropdown.Item
                          onClick={() => {
                            this.updateWeekdayChart("");
                          }}
                        >
                          {this.props.t("allAreas")}
                        </Dropdown.Item>
                        {this.locations.map((location) => (
                          <Dropdown.Item
                            key={location.id}
                            onClick={() => {
                              this.updateWeekdayChart(location.id);
                            }}
                          >
                            {location.name}
                          </Dropdown.Item>
                        ))}
                      </Dropdown.Menu>
                    </Dropdown>
                  </div>
                  {this.renderWeekdayChart(
                    this.state.stats?.bookingsByWeekday ?? [
                      0, 0, 0, 0, 0, 0, 0,
                    ],
                  )}
                </Card.Body>
              </Card>
            </Col>
          </Row>
          <Row className="mb-4">
            <Col sm="12" xl="8">
              <Card>
                <Card.Body>
                  <div className="d-flex justify-content-between align-items-center mb-3">
                    <Card.Title className="mb-0 d-flex align-items-center gap-2">
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
                          style={{ cursor: "pointer", color: "#6c757d" }}
                        />
                      </OverlayTrigger>
                    </Card.Title>
                    <Dropdown>
                      <Dropdown.Toggle variant="outline-secondary" size="sm">
                        {this.state.selectedUtilizationLocationId
                          ? this.locations.find(
                              (e) =>
                                e.id ==
                                this.state.selectedUtilizationLocationId,
                            )?.name
                          : this.props.t("allAreas")}
                      </Dropdown.Toggle>
                      <Dropdown.Menu align="end">
                        <Dropdown.Item
                          onClick={() => {
                            this.updateLoad("");
                          }}
                        >
                          {this.props.t("allAreas")}
                        </Dropdown.Item>
                        {this.locations.map((location) => (
                          <Dropdown.Item
                            key={location.id}
                            onClick={() => {
                              this.updateLoad(location.id);
                            }}
                          >
                            {location.name}
                          </Dropdown.Item>
                        ))}
                      </Dropdown.Menu>
                    </Dropdown>
                  </div>

                  {this.renderProgressBar(
                    this.state.stats?.spaceLoadNextWeek,
                    this.props.t("nextWeek"),
                  )}
                  {this.renderProgressBar(
                    this.state.stats?.spaceLoadThisWeek,
                    this.props.t("thisWeek"),
                  )}
                  {this.renderProgressBar(
                    this.state.stats?.spaceLoadLastWeek,
                    this.props.t("lastWeek"),
                  )}
                  {this.renderProgressBar(
                    this.state.stats?.spaceLoadLastMonth,
                    this.props.t("lastMonth"),
                  )}
                </Card.Body>
              </Card>
            </Col>
          </Row>
        </>
      );
    }

    return (
      <FullLayout headline="Dashboard">
        {cloudUpgradeHint}
        {updateHint}
        <CloudHint />
        {statsContent}
      </FullLayout>
    );
  }
}

export default withTranslation(withReadyRouter(Dashboard as any));
