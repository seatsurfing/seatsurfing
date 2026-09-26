import React from "react";
import { Form, Button, Alert, InputGroup, Modal, Nav } from "react-bootstrap";
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
import RendererUtils from "@/util/RendererUtils";

interface PublicBookableSpace {
  spaceId: string;
  spaceName: string;
  locationId: string;
  locationName: string;
  requireSubject: boolean;
  bookableDays: number[];
  x: number;
  y: number;
  width: number;
  height: number;
  rotation: number;
  shape: string;
  fontSize: string;
}

interface PublicLocationMap {
  width: number;
  height: number;
  scale: number;
  mimeType: string;
  data: string;
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
  showMap: boolean;
  mapData: PublicLocationMap | null;
  mapLoading: boolean;
  mapModalOpen: boolean;
  mapContainerWidth: number;
  mapContainerHeight: number;
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
  lang: string;
}

class PublicBooking extends React.Component<Props, State> {
  orgId: string = "";
  mapCache: { [locationId: string]: PublicLocationMap | null } = {};
  mapContainerObserver: ResizeObserver | null = null;

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
      showMap: false,
      mapData: null,
      mapLoading: false,
      mapModalOpen: false,
      mapContainerWidth: 0,
      mapContainerHeight: 0,
    };
  }

  componentDidMount = () => {
    BrowserUtil.applyLanguageFromQuery();
    this.loadOrgAndSpaces();
  };

  componentWillUnmount = () => {
    this.mapContainerObserver?.disconnect();
  };

  // Tracks the size of the map container in the modal so that the floor
  // plan can be scaled to fill it completely.
  setMapContainerRef = (el: HTMLDivElement | null) => {
    this.mapContainerObserver?.disconnect();
    this.mapContainerObserver = null;
    if (!el) {
      return;
    }
    this.mapContainerObserver = new ResizeObserver((entries) => {
      const rect = entries[0].contentRect;
      this.setState({
        mapContainerWidth: rect.width,
        mapContainerHeight: rect.height,
      });
    });
    this.mapContainerObserver.observe(el);
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
      const showMap: boolean = res.json.showMap === true;
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
        showMap: showMap,
      });
    } catch {
      this.props.router.replace("/404");
    }
  };

  loadMap = async (locationId: string) => {
    if (!(locationId in this.mapCache)) {
      this.setState({ mapData: null, mapLoading: true });
      try {
        const res = await Ajax.get(
          "/public-booking/" +
            encodeURIComponent(this.orgId) +
            "/location/" +
            encodeURIComponent(locationId) +
            "/map",
          () => true,
        );
        this.mapCache[locationId] = res.json;
      } catch {
        this.mapCache[locationId] = null;
      }
    }
    if (this.getSelectedSpace()?.locationId === locationId) {
      this.setState({
        mapData: this.mapCache[locationId],
        mapLoading: false,
      });
    }
  };

  getLocations = (): { id: string; name: string }[] => {
    const locations: { id: string; name: string }[] = [];
    this.state.spaces.forEach((s) => {
      if (!locations.find((l) => l.id === s.locationId)) {
        locations.push({ id: s.locationId, name: s.locationName });
      }
    });
    return locations;
  };

  onLocationChange = (locationId: string) => {
    const space = this.state.spaces.find((s) => s.locationId === locationId);
    if (space) {
      this.onSpaceChange(space.spaceId);
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
    const prevLocationId = this.getSelectedSpace()?.locationId;
    if (
      this.state.mapModalOpen &&
      space &&
      space.locationId !== prevLocationId
    ) {
      this.setState({ mapData: null }, () => this.loadMap(space.locationId));
    }
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

  renderMapSpace = (item: PublicBookableSpace) => {
    const selected = item.spaceId === this.state.spaceId;
    const boxStyle: React.CSSProperties = {
      position: "absolute",
      left: item.x,
      top: item.y,
      width: item.width,
      height: item.height,
      transform: `rotate(${item.rotation}deg)`,
      cursor: "pointer",
      backgroundColor: selected ? "var(--bs-primary)" : undefined,
      borderRadius: item.shape === "circle" ? "50%" : undefined,
      clipPath:
        item.shape === "trapezoid"
          ? "polygon(20% 0%, 80% 0%, 100% 100%, 0% 100%)"
          : undefined,
    };
    const innerStyle: React.CSSProperties = {
      transform: `rotate(${-item.rotation}deg)`,
      display: "flex",
      alignItems: "center",
      justifyContent: "center",
      width: "100%",
      height: "100%",
    };
    const textStyle: React.CSSProperties = {
      textAlign: "center",
      fontSize: RendererUtils.spaceFontSizePx(item.fontSize),
    };
    const className =
      "space space-box" +
      (selected ? "" : " space-available") +
      (RendererUtils.isSpaceVertical(item.width, item.height, item.rotation)
        ? " space-box-vertical"
        : "");
    return (
      <div
        key={item.spaceId}
        style={boxStyle}
        className={className}
        role="button"
        tabIndex={0}
        aria-label={item.spaceName}
        aria-pressed={selected}
        onClick={() => this.onMapSpaceSelect(item.spaceId)}
        onKeyDown={(e) => {
          if (e.key === "Enter" || e.key === " ") {
            e.preventDefault();
            this.onMapSpaceSelect(item.spaceId);
          }
        }}
      >
        <div style={innerStyle}>
          <p style={textStyle}>{item.spaceName}</p>
        </div>
      </div>
    );
  };

  openMapModal = () => {
    const selectedSpace = this.getSelectedSpace();
    if (!selectedSpace) {
      return;
    }
    this.setState({ mapModalOpen: true });
    this.loadMap(selectedSpace.locationId);
  };

  onMapSpaceSelect = (spaceId: string) => {
    this.onSpaceChange(spaceId);
    this.setState({ mapModalOpen: false });
  };

  renderMap = () => {
    const mapData = this.state.mapData;
    const selectedSpace = this.getSelectedSpace();
    if (!selectedSpace) {
      return <></>;
    }
    if (this.state.mapLoading) {
      return <Loading />;
    }
    if (!mapData) {
      return <p>{this.props.t("publicBookingNoMap")}</p>;
    }
    const mapWidth = mapData.width * mapData.scale;
    const mapHeight = mapData.height * mapData.scale;
    const fitScale =
      this.state.mapContainerWidth > 0 &&
      this.state.mapContainerHeight > 0 &&
      mapWidth > 0 &&
      mapHeight > 0
        ? Math.min(
            this.state.mapContainerWidth / mapWidth,
            this.state.mapContainerHeight / mapHeight,
          )
        : 1;
    const scaledStyle: React.CSSProperties = {
      width: mapWidth * fitScale + "px",
      height: mapHeight * fitScale + "px",
      flexShrink: 0,
    };
    const floorPlanStyle: React.CSSProperties = {
      position: "relative",
      width: mapWidth + "px",
      height: mapHeight + "px",
      transform: `scale(${fitScale})`,
      transformOrigin: "top left",
      backgroundSize: "contain",
      backgroundRepeat: "no-repeat",
      backgroundImage:
        "url(data:image/" + mapData.mimeType + ";base64," + mapData.data + ")",
    };
    const spaces = this.state.spaces.filter(
      (s) => s.locationId === selectedSpace.locationId,
    );
    return (
      <div
        className="public-booking-map border rounded bg-body-secondary"
        data-testid="public-booking-map"
        ref={this.setMapContainerRef}
      >
        <div style={scaledStyle}>
          <div style={floorPlanStyle}>
            {spaces.map((s) => this.renderMapSpace(s))}
          </div>
        </div>
      </div>
    );
  };

  renderMapModal = () => {
    if (!this.state.showMap) {
      return <></>;
    }
    return (
      <Modal
        show={this.state.mapModalOpen}
        onHide={() => this.setState({ mapModalOpen: false })}
        size="xl"
        dialogClassName="public-booking-map-modal"
      >
        <Modal.Header
          closeButton
          className={
            this.getLocations().length > 1
              ? "public-booking-map-modal-header-tabs"
              : undefined
          }
        >
          {this.getLocations().length > 1 ? (
            <Nav
              variant="tabs"
              aria-label={this.props.t("area")}
              activeKey={this.getSelectedSpace()?.locationId}
              onSelect={(key) => key && this.onLocationChange(key)}
            >
              {this.getLocations().map((l) => (
                <Nav.Item key={l.id}>
                  <Nav.Link eventKey={l.id}>{l.name}</Nav.Link>
                </Nav.Item>
              ))}
            </Nav>
          ) : (
            <Modal.Title>{this.getSelectedSpace()?.locationName}</Modal.Title>
          )}
        </Modal.Header>
        <Modal.Body>{this.renderMap()}</Modal.Body>
        <Modal.Footer>
          {this.state.mapData && !this.state.mapLoading && (
            <span className="text-muted me-auto">
              {this.props.t("publicBookingSelectOnMap")}
            </span>
          )}
          <Button
            variant="secondary"
            onClick={() => this.setState({ mapModalOpen: false })}
          >
            {this.props.t("close")}
          </Button>
        </Modal.Footer>
      </Modal>
    );
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
          <Form.Group className="mb-3" controlId="select-space">
            <Form.Label>{this.props.t("space")}</Form.Label>
            <InputGroup>
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
              {this.state.showMap && (
                <Button variant="outline-secondary" onClick={this.openMapModal}>
                  {this.props.t("publicBookingShowFloorplan")}
                </Button>
              )}
            </InputGroup>
          </Form.Group>
          {this.renderMapModal()}
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
      </div>
    );
  }
}

export default withTranslation(withReadyRouter(PublicBooking as any));
