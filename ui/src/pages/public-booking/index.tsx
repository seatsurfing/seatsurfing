import React from "react";
import { Form, Button, Alert } from "react-bootstrap";
import { NextRouter } from "next/router";
import withReadyRouter from "@/components/withReadyRouter";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import SeatsurfingLogo from "@/components/SeatsurfingLogo";
import Loading from "@/components/Loading";
import Ajax from "@/util/Ajax";
import DateUtil from "@/util/DateUtil";
import DateTimePicker from "@/components/DateTimePicker";
import Validation from "@/util/Validation";

interface AnonymousBookableSpace {
  spaceId: string;
  spaceName: string;
  locationId: string;
  locationName: string;
}

interface State {
  loading: boolean;
  spaces: AnonymousBookableSpace[];
  spaceId: string;
  enter: Date;
  leave: Date;
  name: string;
  email: string;
  subject: string;
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
    const enter = new Date();
    enter.setMinutes(0, 0, 0);
    enter.setHours(enter.getHours() + 1);
    const leave = new Date(enter);
    leave.setHours(leave.getHours() + 1);
    this.state = {
      loading: true,
      spaces: [],
      spaceId: "",
      enter: enter,
      leave: leave,
      name: "",
      email: "",
      subject: "",
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
          .catch(() => this.props.router.replace("/404"));
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
        if (spaces.length === 0) {
          this.props.router.replace("/404");
          return;
        }
        this.setState({
          loading: false,
          spaces: spaces,
          spaceId: spaces[0].spaceId,
        });
      })
      .catch(() => this.props.router.replace("/404"));
  };

  onSubmit = (e: any) => {
    e.preventDefault();
    if (!this.state.spaceId) {
      return;
    }
    if (
      this.state.leave <= this.state.enter ||
      !DateUtil.isSameDay(this.state.enter, this.state.leave)
    ) {
      this.setState({ error: true });
      return;
    }
    this.setState({ submitting: true, error: false });
    const payload = {
      spaceId: this.state.spaceId,
      enter: DateUtil.convertToFakeUTCDate(this.state.enter).toISOString(),
      leave: DateUtil.convertToFakeUTCDate(this.state.leave).toISOString(),
      name: this.state.name,
      email: this.state.email,
      subject: this.state.subject,
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

    if (this.state.spaces.length === 0) {
      return <Loading />;
    }

    if (this.state.submitted) {
      return (
        <div className="container-center">
          <div className="container-center-inner-wide">
            <SeatsurfingLogo />
            <p>{this.props.t("anonymousBookingRequestSubmitted")}</p>
          </div>
        </div>
      );
    }

    return (
      <div className="container-center">
        <Form className="container-center-inner-wide" onSubmit={this.onSubmit}>
          <SeatsurfingLogo />
          <p>{this.props.t("anonymousBookingIntro")}</p>
          {this.state.error && (
            <Alert variant="danger">
              {this.props.t("anonymousBookingRequestError")}
            </Alert>
          )}
          <Form.Group className="mb-3">
            <Form.Label>{this.props.t("name")}</Form.Label>
            <Form.Control
              type="text"
              value={this.state.name}
              onChange={(e: any) => this.setState({ name: e.target.value })}
              required={true}
              minLength={2}
              maxLength={64}
              pattern={Validation.HUMAN_NAME_PATTERN}
              title={this.props.t("nameRequirements")}
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
            <Form.Label>{this.props.t("subject")}</Form.Label>
            <Form.Control
              type="text"
              value={this.state.subject}
              onChange={(e: any) => this.setState({ subject: e.target.value })}
              maxLength={256}
            />
          </Form.Group>
          <Form.Group className="mb-3">
            <Form.Label>{this.props.t("date")}</Form.Label>
            <DateTimePicker
              value={this.state.enter}
              onChange={(value: Date) =>
                this.setState({
                  enter: DateUtil.copyDate(value, this.state.enter),
                  leave: DateUtil.copyDate(value, this.state.leave),
                })
              }
              required={true}
              enableTime={false}
            />
          </Form.Group>
          <Form.Group className="mb-3">
            <Form.Label>{this.props.t("enter")}</Form.Label>
            <DateTimePicker
              value={this.state.enter}
              onChange={(value: Date) =>
                this.setState({
                  enter: DateUtil.copyTime(value, this.state.enter),
                })
              }
              required={true}
              noCalendar={true}
              enableTime={true}
            />
          </Form.Group>
          <Form.Group className="mb-3">
            <Form.Label>{this.props.t("leave")}</Form.Label>
            <DateTimePicker
              value={this.state.leave}
              onChange={(value: Date) =>
                this.setState({
                  leave: DateUtil.copyTime(value, this.state.leave),
                })
              }
              required={true}
              noCalendar={true}
              enableTime={true}
              minDate={
                new Date(this.state.enter.getTime() + 60 * 1000)
              }
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
