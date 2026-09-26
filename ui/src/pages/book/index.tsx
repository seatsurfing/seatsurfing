import React from "react";
import { Form, Button, Alert } from "react-bootstrap";
import { NextRouter } from "next/router";
import withReadyRouter from "@/components/withReadyRouter";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import SeatsurfingAppLogo from "@/components/SeatsurfingAppLogo";
import Loading from "@/components/Loading";
import Ajax from "@/util/Ajax";
import BrowserUtil from "@/util/BrowserUtil";
import DateUtil from "@/util/DateUtil";
import DateTimePicker from "@/components/DateTimePicker";
import Validation from "@/util/Validation";
import CopyrightFooter from "@/components/CopyrightFooter";

interface PublicBookableSpace {
  spaceId: string;
  spaceName: string;
  locationId: string;
  locationName: string;
  requireSubject: boolean;
  bookableDays: number[];
}

interface State {
  loading: boolean;
  spaces: PublicBookableSpace[];
  spaceId: string;
  enter: Date;
  leave: Date;
  maxDaysInAdvance: number;
  name: string;
  email: string;
  subject: string;
  submitting: boolean;
  submitted: boolean;
  error: boolean;
  customLogoUrl: string;
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
  lang: string;
}

class PublicBooking extends React.Component<Props, State> {
  orgId: string = "";

  constructor(props: any) {
    super(props);
    const enter = new Date();
    enter.setDate(enter.getDate() + 1);
    enter.setHours(9, 0, 0, 0);
    const leave = new Date(enter);
    leave.setHours(17, 0, 0, 0);
    this.state = {
      loading: true,
      spaces: [],
      spaceId: "",
      enter: enter,
      leave: leave,
      maxDaysInAdvance: 0,
      name: "",
      email: "",
      subject: "",
      submitting: false,
      submitted: false,
      error: false,
      customLogoUrl: "",
    };
  }

  componentDidMount = () => {
    BrowserUtil.applyLanguageFromQuery();
    this.loadOrgAndSpaces();
  };

  loadOrgAndSpaces = async () => {
    const domain = window.location.host.split(":").shift();
    let res;
    try {
      res = await Ajax.get("/auth/org/" + domain, () => true);
    } catch {
      try {
        res = await Ajax.get("/auth/singleorg", () => true);
      } catch {
        this.props.router.replace("/404");
        return;
      }
    }
    this.setState({ customLogoUrl: res.json.customLogoUrl || "" });
    await this.onOrgResolved(res.json.organization.id);
  };

  onOrgResolved = async (orgId: string) => {
    this.orgId = orgId;
    try {
      const res = await Ajax.get(
        "/public-booking/" + encodeURIComponent(orgId) + "/spaces",
        () => true,
      );
      const spaces: PublicBookableSpace[] = res.json.spaces || [];
      if (spaces.length === 0) {
        this.setState({ loading: false });
        return;
      }
      const maxDaysInAdvance: number = res.json.maxDaysInAdvance || 0;
      const maxDate = new Date();
      maxDate.setDate(maxDate.getDate() + maxDaysInAdvance);
      maxDate.setHours(23, 59, 59, 999);
      let enter = this.state.enter;
      let leave = this.state.leave;
      if (enter > maxDate) {
        enter = DateUtil.copyDate(maxDate, enter);
        leave = DateUtil.copyDate(maxDate, leave);
      }
      const bookableDate = this.findBookableDate(
        spaces[0],
        enter,
        maxDaysInAdvance,
      );
      enter = DateUtil.copyDate(bookableDate, enter);
      leave = DateUtil.copyDate(bookableDate, leave);
      this.setState({
        loading: false,
        spaces: spaces,
        spaceId: spaces[0].spaceId,
        maxDaysInAdvance: maxDaysInAdvance,
        enter: enter,
        leave: leave,
      });
    } catch {
      this.props.router.replace("/404");
    }
  };

  getSelectedSpace = (): PublicBookableSpace | undefined => {
    return this.state.spaces.find((s) => s.spaceId === this.state.spaceId);
  };

  getMaxDate = (): Date => {
    const maxDate = DateUtil.getTodayStart();
    maxDate.setDate(maxDate.getDate() + this.state.maxDaysInAdvance);
    return maxDate;
  };

  isDateBookable = (
    space: PublicBookableSpace | undefined,
    date: Date,
  ): boolean => {
    if (!space || !space.bookableDays || space.bookableDays.length === 0) {
      return true;
    }
    return space.bookableDays.includes(date.getDay());
  };

  findBookableDate = (
    space: PublicBookableSpace,
    start: Date,
    maxDaysInAdvance: number,
  ): Date => {
    const today = DateUtil.getTodayStart();
    const date = new Date(start);
    for (let i = 0; i < 7; i++) {
      const daysAhead = Math.round(
        (DateUtil.setHoursToMin(date).getTime() - today.getTime()) /
          (24 * 60 * 60 * 1000),
      );
      if (daysAhead > maxDaysInAdvance) {
        break;
      }
      if (this.isDateBookable(space, date)) {
        return date;
      }
      date.setDate(date.getDate() + 1);
    }
    return start;
  };

  onSpaceChange = (spaceId: string) => {
    const space = this.state.spaces.find((s) => s.spaceId === spaceId);
    if (!space || this.isDateBookable(space, this.state.enter)) {
      this.setState({ spaceId: spaceId });
      return;
    }
    const date = this.findBookableDate(
      space,
      this.state.enter,
      this.state.maxDaysInAdvance,
    );
    this.setState({
      spaceId: spaceId,
      enter: DateUtil.copyDate(date, this.state.enter),
      leave: DateUtil.copyDate(date, this.state.leave),
    });
  };

  onSubmit = async (e: any) => {
    e.preventDefault();
    if (!this.state.spaceId) {
      return;
    }
    if (
      this.state.leave <= this.state.enter ||
      !DateUtil.isSameDay(this.state.enter, this.state.leave) ||
      !this.isDateBookable(this.getSelectedSpace(), this.state.enter)
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
      language: this.props.lang,
    };
    try {
      await Ajax.postData(
        "/public-booking/" + encodeURIComponent(this.orgId) + "/request",
        payload,
        () => true,
      );
      this.setState({ submitting: false, submitted: true });
    } catch {
      this.setState({ submitting: false, error: true });
    }
  };

  render() {
    if (this.state.loading) {
      return <Loading />;
    }

    if (this.state.spaces.length === 0) {
      return (
        <div className="container-center">
          <div className="container-center-inner-wide">
            <SeatsurfingAppLogo customLogoUrl={this.state.customLogoUrl} />
            <p>{this.props.t("publicBookingNoSpaces")}</p>
          </div>
          <CopyrightFooter />
        </div>
      );
    }

    if (this.state.submitted) {
      return (
        <div className="container-center">
          <div className="container-center-inner-wide">
            <SeatsurfingAppLogo customLogoUrl={this.state.customLogoUrl} />
            <p>{this.props.t("publicBookingRequestSubmitted")}</p>
          </div>
          <CopyrightFooter />
        </div>
      );
    }

    return (
      <div className="container-center">
        <Form className="container-center-inner-wide" onSubmit={this.onSubmit}>
          <SeatsurfingAppLogo customLogoUrl={this.state.customLogoUrl} />
          <p>{this.props.t("publicBookingIntro")}</p>
          {this.state.error && (
            <Alert variant="danger">
              {this.props.t("publicBookingRequestError")}
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
              pattern={Validation.EMAIL_PATTERN}
              maxLength={256}
            />
          </Form.Group>
          <Form.Group className="mb-3">
            <Form.Label>{this.props.t("space")}</Form.Label>
            <Form.Select
              value={this.state.spaceId}
              onChange={(e: any) => this.onSpaceChange(e.target.value)}
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
              required={this.getSelectedSpace()?.requireSubject}
              minLength={this.getSelectedSpace()?.requireSubject ? 3 : 0}
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
              minDate={DateUtil.getTodayStart()}
              maxDate={this.getMaxDate()}
              isDateDisabled={(date: Date) =>
                !this.isDateBookable(this.getSelectedSpace(), date)
              }
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
              minDate={new Date(this.state.enter.getTime() + 60 * 1000)}
            />
          </Form.Group>
          <Button
            variant="primary"
            type="submit"
            disabled={this.state.submitting}
          >
            {this.props.t("publicBookingSubmit")}
          </Button>
        </Form>
        <CopyrightFooter />
      </div>
    );
  }
}

export default withTranslation(withReadyRouter(PublicBooking as any));
