import React from "react";
import { Card, Row, Col, ProgressBar, Alert } from "react-bootstrap";
import { NextRouter } from "next/router";
import FullLayout from "@/components/FullLayout";
import Loading from "@/components/Loading";
import withReadyRouter from "@/components/withReadyRouter";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import RuntimeConfig from "@/components/RuntimeConfig";
import Link from "next/link";
import PremiumFeatureIcon from "@/components/PremiumFeatureIcon";
import Stats from "@/types/Stats";
import Ajax from "@/util/Ajax";
import User from "@/types/User";
import DateUtil from "@/util/DateUtil";
import RedirectUtil from "@/util/RedirectUtil";
import * as Navigation from "@/util/Navigation";

interface State {
  loading: boolean;
  redirect: string;
  spaceAdmin: boolean;
  orgAdmin: boolean;
  latestVersion: any;
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
}

class Dashboard extends React.Component<Props, State> {
  stats: Stats | null;

  constructor(props: any) {
    super(props);
    this.stats = null;
    this.state = {
      loading: true,
      redirect: "",
      spaceAdmin: false,
      orgAdmin: false,
      latestVersion: null,
    };
  }

  componentDidMount = () => {
    if (!Ajax.hasAccessToken()) {
      RedirectUtil.toLogin(this.props.router);
      return;
    }
    const promises = [
      this.loadItems(),
      this.getUserInfo(),
      this.checkUpdates(),
    ];
    Promise.all(promises)
      .then(() => {
        this.setState({ loading: false });
      })
      .catch(() => {
        RedirectUtil.toLogin(this.props.router);
        return;
      });
  };

  checkUpdates = async (): Promise<void> => {
    const self = this;
    return new Promise<void>(function (resolve) {
      if (RuntimeConfig.INFOS.cloudHosted) {
        resolve();
        return;
      }
      Ajax.get("/uc/")
        .then((res) => {
          self.setState(
            {
              latestVersion: res.json,
            },
            () => resolve(),
          );
        })
        .catch(() => {
          console.warn("Could not check for updates.");
          const res = { version: "", updateAvailable: false };
          self.setState(
            {
              latestVersion: res,
            },
            () => resolve(),
          );
        });
    });
  };

  getUserInfo = async (): Promise<void> => {
    const self = this;
    return new Promise<void>(function (resolve, reject) {
      User.getSelf()
        .then((user) => {
          self.setState(
            {
              spaceAdmin: user.spaceAdmin,
              orgAdmin: user.admin,
            },
            () => resolve(),
          );
        })
        .catch((e) => reject(e));
    });
  };

  loadItems = async (): Promise<void> => {
    const self = this;
    return new Promise<void>(function (resolve, reject) {
      Stats.get()
        .then((stats) => {
          self.stats = stats;
          resolve();
        })
        .catch((e) => reject(e));
    });
  };

  renderStatsCard = (num: number | undefined, title: string, link?: string) => {
    const redirect = link ?? "";
    return (
      <Col sm="2">
        <Card
          className="dashboard-card-clickable"
          onClick={() => this.setState({ redirect })}
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

  renderProgressBar = (num: number | undefined, title: string) => {
    if (!num) {
      num = 0;
    }
    const label = `${title}: ${num} %`;
    let variant = "success";
    if (num >= 80) {
      variant = "danger";
    }
    if (num >= 60) {
      variant = "warning";
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

    return (
      <FullLayout headline="Dashboard">
        {cloudUpgradeHint}
        {updateHint}
        <Row className="mb-4">
          {this.renderStatsCard(
            this.stats?.numUsers,
            this.props.t("users"),
            this.state.orgAdmin ? Navigation.adminUsers() : "",
          )}
          {this.renderStatsCard(
            this.stats?.numLocations,
            this.props.t("areas"),
            Navigation.adminLocations(),
          )}
          {this.renderStatsCard(
            this.stats?.numSpaces,
            this.props.t("spaces"),
            Navigation.adminLocations(),
          )}
          {this.renderStatsCard(
            this.stats?.numBookings,
            this.props.t("bookings"),
            Navigation.adminBookings(
              "enter=2000-01-01T00:00&leave=2999-12-31T23:59&filter=enter_leave",
            ),
          )}
        </Row>
        <Row className="mb-4">
          {this.renderStatsCard(
            this.stats?.numBookingsCurrent,
            this.props.t("current"),
            Navigation.adminBookings("filter=current"),
          )}
          {this.renderStatsCard(
            this.stats?.numBookingsToday,
            this.props.t("today"),
            Navigation.adminBookings("filter=today"),
          )}
          {this.renderStatsCard(
            this.stats?.numBookingsYesterday,
            this.props.t("yesterday"),
            Navigation.adminBookings(
              `enter=${yesterdayDateString}T00:00&leave=${yesterdayDateString}T23:59&filter=enter_leave`,
            ),
          )}
          {this.renderStatsCard(
            this.stats?.numBookingsThisWeek,
            this.props.t("thisWeek"),
            Navigation.adminBookings(
              `/enter=${DateUtil.getThisWeekMondayDateString()}T00:00&leave=${DateUtil.getThisWeekSundayDateString()}T23:59&filter=enter_leave`,
            ),
          )}
        </Row>
        <Row className="mb-4">
          <Col sm="8">
            <Card>
              <Card.Body>
                <Card.Title>{this.props.t("utilization")}</Card.Title>
                {this.renderProgressBar(
                  this.stats?.spaceLoadToday,
                  this.props.t("today"),
                )}
                {this.renderProgressBar(
                  this.stats?.spaceLoadYesterday,
                  this.props.t("yesterday"),
                )}
                {this.renderProgressBar(
                  this.stats?.spaceLoadThisWeek,
                  this.props.t("thisWeek"),
                )}
                {this.renderProgressBar(
                  this.stats?.spaceLoadLastWeek,
                  this.props.t("lastWeek"),
                )}
              </Card.Body>
            </Card>
          </Col>
        </Row>
      </FullLayout>
    );
  }
}

export default withTranslation(withReadyRouter(Dashboard as any));
