import React from "react";
import { Form, Button, Alert } from "react-bootstrap";
import { NextRouter } from "next/router";
import withReadyRouter from "@/components/withReadyRouter";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import SeatsurfingLogo from "@/components/SeatsurfingLogo";
import Loading from "@/components/Loading";
import Ajax from "@/util/Ajax";
import DateUtil from "@/util/DateUtil";

interface AnonymousBookableSpace {
  spaceId: string;
  spaceName: string;
  locationId: string;
  locationName: string;
}

interface State {
  loading: boolean;
  notAvailable: boolean;
  spaces: AnonymousBookableSpace[];
  spaceId: string;
  date: string;
  startTime: string;
  endTime: string;
  name: string;
  email: string;
  submitting: boolean;
  submitted: boolean;
  error: boolean;
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
}

class PublicBooking extends React.Component<Props, State> {
  orgId: string = "";

  constructor(props: any) {
    super(props);
    this.state = {
      loading: true,
      notAvailable: false,
      spaces: [],
      spaceId: "",
      date: "",
      startTime: "",
      endTime: "",
      name: "",
      email: "",
      submitting: false,
      submitted: false,
      error: false,
    };
  }

  componentDidMount = () => {
    this.loadOrgAndSpaces();
  };

  loadOrgAndSpaces = () => {
    const domain = window.location.host.split(":").shift();
    Ajax.get("/auth/org/" + domain, () => true)
      .then((res) => this.onOrgResolved(res.json.organization.id))
      .catch(() => {
        Ajax.get("/auth/singleorg", () => true)
          .then((res) => this.onOrgResolved(res.json.organization.id))
          .catch(() => this.setState({ loading: false, notAvailable: true }));
      });
  };

  onOrgResolved = (orgId: string) => {
    this.orgId = orgId;
    Ajax.get(
      "/public-booking/" + encodeURIComponent(orgId) + "/spaces",
      () => true,
    )
      .then((res) => {
        const spaces: AnonymousBookableSpace[] = res.json || [];
        this.setState({
          loading: false,
          spaces: spaces,
          spaceId: spaces.length > 0 ? spaces[0].spaceId : "",
        });
      })
      .catch(() => this.setState({ loading: false, notAvailable: true }));
  };

  onSubmit = (e: any) => {
    e.preventDefault();
    if (
      !this.state.spaceId ||
      !this.state.date ||
      !this.state.startTime ||
      !this.state.endTime
    ) {
      return;
    }
    const [startHours, startMinutes] = this.state.startTime
      .split(":")
      .map(Number);
    const [endHours, endMinutes] = this.state.endTime.split(":").map(Number);
    const [year, month, day] = this.state.date.split("-").map(Number);
    const enter = new Date(year, month - 1, day, startHours, startMinutes, 0);
    const leave = new Date(year, month - 1, day, endHours, endMinutes, 0);
    if (leave <= enter) {
      this.setState({ error: true });
      return;
    }
    this.setState({ submitting: true, error: false });
    const payload = {
      spaceId: this.state.spaceId,
      enter: DateUtil.convertToFakeUTCDate(enter).toISOString(),
      leave: DateUtil.convertToFakeUTCDate(leave).toISOString(),
      name: this.state.name,
      email: this.state.email,
    };
    Ajax.postData(
      "/public-booking/" + encodeURIComponent(this.orgId) + "/request",
      payload,
      () => true,
    )
      .then(() => this.setState({ submitting: false, submitted: true }))
      .catch(() => this.setState({ submitting: false, error: true }));
  };

  render() {
    if (this.state.loading) {
      return <Loading />;
    }

    if (this.state.notAvailable || this.state.spaces.length === 0) {
      return (
        <div className="container-center">
          <div className="container-center-inner">
            <SeatsurfingLogo />
            <p>{this.props.t("anonymousBookingNotAvailable")}</p>
          </div>
        </div>
      );
    }

    if (this.state.submitted) {
      return (
        <div className="container-center">
          <div className="container-center-inner">
            <SeatsurfingLogo />
            <p>{this.props.t("anonymousBookingRequestSubmitted")}</p>
          </div>
        </div>
      );
    }

    return (
      <div className="container-center">
        <Form className="container-center-inner" onSubmit={this.onSubmit}>
          <SeatsurfingLogo />
          <p>{this.props.t("anonymousBookingIntro")}</p>
          {this.state.error && (
            <Alert variant="danger">
              {this.props.t("anonymousBookingRequestError")}
            </Alert>
          )}
          <Form.Group className="mb-3">
            <Form.Label>{this.props.t("space")}</Form.Label>
            <Form.Select
              value={this.state.spaceId}
              onChange={(e: any) => this.setState({ spaceId: e.target.value })}
              required={true}
            >
              {this.state.spaces.map((s) => (
                <option key={s.spaceId} value={s.spaceId}>
                  {s.locationName} / {s.spaceName}
                </option>
              ))}
            </Form.Select>
          </Form.Group>
          <Form.Group className="mb-3">
            <Form.Label>{this.props.t("date")}</Form.Label>
            <Form.Control
              type="date"
              value={this.state.date}
              onChange={(e: any) => this.setState({ date: e.target.value })}
              required={true}
            />
          </Form.Group>
          <Form.Group className="mb-3">
            <Form.Label>{this.props.t("enter")}</Form.Label>
            <Form.Control
              type="time"
              value={this.state.startTime}
              onChange={(e: any) =>
                this.setState({ startTime: e.target.value })
              }
              required={true}
            />
          </Form.Group>
          <Form.Group className="mb-3">
            <Form.Label>{this.props.t("leave")}</Form.Label>
            <Form.Control
              type="time"
              value={this.state.endTime}
              onChange={(e: any) => this.setState({ endTime: e.target.value })}
              required={true}
            />
          </Form.Group>
          <Form.Group className="mb-3">
            <Form.Label>{this.props.t("name")}</Form.Label>
            <Form.Control
              type="text"
              value={this.state.name}
              onChange={(e: any) => this.setState({ name: e.target.value })}
              required={true}
              maxLength={256}
            />
          </Form.Group>
          <Form.Group className="mb-3">
            <Form.Label>{this.props.t("emailAddress")}</Form.Label>
            <Form.Control
              type="email"
              value={this.state.email}
              onChange={(e: any) => this.setState({ email: e.target.value })}
              required={true}
              maxLength={256}
            />
          </Form.Group>
          <Button
            variant="primary"
            type="submit"
            disabled={this.state.submitting}
          >
            {this.props.t("anonymousBookingSubmit")}
          </Button>
        </Form>
      </div>
    );
  }
}

export default withTranslation(withReadyRouter(PublicBooking as any));
