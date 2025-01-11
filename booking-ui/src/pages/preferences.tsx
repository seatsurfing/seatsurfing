import React from 'react';
import { Ajax, Location, User, UserPreference } from 'flexspace-commons';
import Loading from '../components/Loading';
import { Alert, Button, ButtonGroup, Col, Form, Nav, Row } from 'react-bootstrap';
import { WithTranslation, withTranslation } from 'next-i18next';
import { NextRouter } from 'next/router';
import NavBar from '@/components/NavBar';
import withReadyRouter from '@/components/withReadyRouter';
import RuntimeConfig from '@/components/RuntimeConfig';

interface State {
  loading: boolean
  submitting: boolean
  saved: boolean
  error: boolean
  enterTime: number
  workdayStart: number
  workdayEnd: number
  workdays: boolean[]
  booked: string
  notBooked: string
  selfBooked: string
  partiallyBooked: string
  buddyBooked: string
  locationId: string
  changePassword: boolean
  password: string
  activeTab: string
  caldavUrl: string
  caldavUser: string
  caldavPass: string
  caldavCalendar: string
  caldavCalendars: any[]
  caldavCalendarsLoaded: boolean
  caldavError: boolean
}

interface Props extends WithTranslation {
  router: NextRouter
}

class Preferences extends React.Component<Props, State> {
  locations: Location[];

  constructor(props: any) {
    super(props);
    this.locations = [];
    this.state = {
      loading: true,
      submitting: false,
      saved: false,
      error: false,
      enterTime: 0,
      workdayStart: 0,
      workdayEnd: 0,
      workdays: [],
      booked: "#ff453a",
      notBooked: "#30d158",
      selfBooked: "#b825de",
      partiallyBooked: "#ff9100",
      buddyBooked: "#2415c5",
      locationId: "",
      changePassword: false,
      password: "",
      activeTab: "tab-bookings",
      caldavUrl: "",
      caldavUser: "",
      caldavPass: "",
      caldavCalendar: "",
      caldavCalendars: [],
      caldavCalendarsLoaded: false,
      caldavError: false,
    };
  }

  componentDidMount = () => {
    if (!Ajax.CREDENTIALS.accessToken) {
      this.props.router.push("/login");
      return;
    }
    let promises = [
      this.loadPreferences(),
      this.loadLocations(),
    ];
    Promise.all(promises).then(() => {
      this.setState({ loading: false });
    });
  }

  loadPreferences = async (): Promise<void> => {
    let self = this;
    return new Promise<void>(function (resolve, reject) {
      UserPreference.list().then(list => {
        let state: any = {};
        list.forEach(s => {
          if (typeof window !== 'undefined') {
            if (s.name === "enter_time") state.enterTime = window.parseInt(s.value);
            if (s.name === "workday_start") state.workdayStart = window.parseInt(s.value);
            if (s.name === "workday_end") state.workdayEnd = window.parseInt(s.value);
          }
          if (s.name === "workdays") {
            state.workdays = [];
            for (let i = 0; i <= 6; i++) {
              state.workdays[i] = false;
            }
            s.value.split(",").forEach(val => state.workdays[val] = true)
          }
          if (s.name === "booked_color") state.booked = s.value;
          if (s.name === "not_booked_color") state.notBooked = s.value;
          if (s.name === "self_booked_color") state.selfBooked = s.value;
          if (s.name === "partially_booked_color") state.partiallyBooked = s.value;
          if (s.name === "buddy_booked_color") state.buddyBooked = s.value;
          if (s.name === "location_id") state.locationId = s.value;
          if (s.name === "caldav_url") state.caldavUrl = s.value;
          if (s.name === "caldav_user") state.caldavUser = s.value;
          if (s.name === "caldav_pass") state.caldavPass = s.value;
          if (s.name === "caldav_path") state.caldavCalendar = s.value;
        });
        self.setState({
          ...self.state,
          ...state
        }, () => resolve());
      }).catch(e => reject(e));
    });
  }

  loadLocations = async (): Promise<void> => {
    let self = this;
    return new Promise<void>(function (resolve, reject) {
      Location.list().then(list => {
        self.locations = list;
        resolve();
      }).catch(e => reject(e));
    });
  }

  onSubmit = (e: any) => {
    e.preventDefault();
    this.setState({
      submitting: true,
      saved: false,
      error: false,
      caldavError: false
    });
    let workdays: string[] = [];
    this.state.workdays.forEach((val, day) => {
      if (val) {
        workdays.push(day.toString());
      }
    });
    let payload = [
      new UserPreference("enter_time", this.state.enterTime.toString()),
      new UserPreference("workday_start", this.state.workdayStart.toString()),
      new UserPreference("workday_end", this.state.workdayEnd.toString()),
      new UserPreference("workdays", workdays.join(",")),
      new UserPreference("location_id", this.state.locationId),
    ];
    UserPreference.setAll(payload).then(() => {
      this.setState({
        submitting: false,
        saved: true
      });
    }).catch(() => {
      this.setState({
        submitting: false,
        error: true
      });
    });
  }

  onSubmitSecurity = (e: any) => {
    e.preventDefault();
    if (!this.state.changePassword) {
      return;
    }
    this.setState({
      submitting: true,
      saved: false,
      error: false,
      caldavError: false
    });
    let payload = {
      password: this.state.password
    };
    Ajax.putData("/user/me/password", payload).then(() => {
      this.setState({
        submitting: false,
        saved: true
      });
    });
  }

  onSubmitColors = (e: any) => {
    e.preventDefault();
    this.setState({
      submitting: true,
      saved: false,
      error: false,
      caldavError: false
    });
    let workdays: string[] = [];
    this.state.workdays.forEach((val, day) => {
      if (val) {
        workdays.push(day.toString());
      }
    });
    let payload = [
      new UserPreference("booked_color", this.state.booked),
      new UserPreference("not_booked_color", this.state.notBooked),
      new UserPreference("self_booked_color", this.state.selfBooked),
      new UserPreference("partially_booked_color", this.state.partiallyBooked),
      new UserPreference("buddy_booked_color", this.state.buddyBooked)
    ];
    UserPreference.setAll(payload).then(() => {
      this.setState({
        submitting: false,
        saved: true
      });
    }).catch(() => {
      this.setState({
        submitting: false,
        error: true
      });
    });
  }

  resetColors = () => {
    this.setState({
      booked: "#ff453a",
      notBooked: "#30d158",
      selfBooked: "#b825de",
      partiallyBooked: "#ff9100",
      buddyBooked: "#2415c5",
    })
  }

  onWorkdayCheck = (day: number, checked: boolean) => {
    let workdays = this.state.workdays.map((val, i) => (i === day) ? checked : val);
    this.setState({
      workdays: workdays
    });
  }

  connectCalDav = () => {
    this.setState({
      submitting: true,
      saved: false,
      error: false,
      caldavError: false,
      caldavCalendarsLoaded: false,
    });
    let payload = {
      url: this.state.caldavUrl,
      username: this.state.caldavUser,
      password: this.state.caldavPass,
    };
    Ajax.postData("/preference/caldav/listCalendars", payload).then(res => {
      this.setState({
        caldavCalendarsLoaded: true,
        caldavCalendars: res.json,
        caldavCalendar: res.json && res.json.length > 0 ? res.json[0].path : '',
        submitting: false,
      });
    }).catch(() => {
      this.setState({
        submitting: false,
        caldavError: true,
      });
    });
  }

  disconnectCalDav = () => {
    this.setState({
      submitting: true,
      saved: false,
      error: false,
      caldavError: false,
      caldavCalendarsLoaded: false
    });
    let payload = [
      new UserPreference("caldav_url", ""),
      new UserPreference("caldav_user", ""),
      new UserPreference("caldav_pass", ""),
      new UserPreference("caldav_path", ""),
    ];
    UserPreference.setAll(payload).then(() => {
      this.setState({
        submitting: false,
        saved: true,
        caldavUrl: '',
        caldavUser: '',
        caldavPass: '',
        caldavCalendar: '',
        caldavCalendars: []
      });
    }).catch(() => {
      this.setState({
        submitting: false,
        error: true
      });
    });
  }

  saveCaldavSettings = (e: any) => {
    e.preventDefault();
    this.setState({
      submitting: true,
      saved: false,
      error: false,
      caldavError: false
    });
    let payload = [
      new UserPreference("caldav_url", this.state.caldavUrl),
      new UserPreference("caldav_user", this.state.caldavUser),
      new UserPreference("caldav_pass", this.state.caldavPass),
      new UserPreference("caldav_path", this.state.caldavCalendar),
    ];
    UserPreference.setAll(payload).then(() => {
      this.setState({
        submitting: false,
        saved: true
      });
    }).catch(() => {
      this.setState({
        submitting: false,
        error: true
      });
    });
  }

  render() {
    if (this.state.loading) {
      return <Loading />;
    }

    let hint = <></>;
    if (this.state.saved) {
      hint = <Alert variant="success" className="margin-top-15">{this.props.t("entryUpdated")}</Alert>
    } else if (this.state.error) {
      hint = <Alert variant="danger" className="margin-top-15">{this.props.t("errorSave")}</Alert>
    } else if (this.state.caldavError) {
      hint = <Alert variant="danger" className="margin-top-15">{this.props.t("errorCaldav")}</Alert>
    }

    return (
      <>
        <NavBar />
        <div className="container-center-top">
          <div className="container-center-inner">
            <Nav variant="underline" defaultActiveKey="tab-bookings" onSelect={(key) => { if (key) this.setState({ activeTab: key }) }}>
              <Nav.Item>
                <Nav.Link eventKey="tab-bookings">{this.props.t('bookings')}</Nav.Link>
              </Nav.Item>
              <Nav.Item>
                <Nav.Link eventKey="tab-security">{this.props.t('security')}</Nav.Link>
              </Nav.Item>
              <Nav.Item>
                <Nav.Link eventKey="tab-integrations">{this.props.t('integrations')}</Nav.Link>
              </Nav.Item>
            </Nav>
            {hint}
            <Form onSubmit={this.onSubmit} hidden={this.state.activeTab !== 'tab-bookings'}>
              <h5 className='margin-top-15'>{this.props.t("preferences")}</h5>
              <Form.Group className="margin-top-15">
                <Form.Label>{this.props.t("notice")}</Form.Label>
                <Form.Select value={this.state.enterTime} onChange={(e: any) => this.setState({ enterTime: e.target.value })}>
                  <option value="1">{this.props.t("earliestPossible")}</option>
                  <option value="2">{this.props.t("nextDay")}</option>
                  <option value="3">{this.props.t("nextWorkday")}</option>
                </Form.Select>
              </Form.Group>
              <Form.Group className="margin-top-15">
                <Form.Label>{this.props.t("workingHours")}</Form.Label>
                <Row>
                  <Col>
                    <Form.Control type="number" value={this.state.workdayStart} onChange={(e: any) => this.setState({ workdayStart: typeof window !== 'undefined' ? window.parseInt(e.target.value) : 0 })} min="0" max="23" />
                  </Col>
                  <Col>
                    <Form.Control plaintext={true} readOnly={true} defaultValue={this.props.t("to").toString()} />
                  </Col>
                  <Col>
                    <Form.Control type="number" value={this.state.workdayEnd} onChange={(e: any) => this.setState({ workdayEnd: e.target.value })} min={this.state.workdayStart + 1} max="23" />
                  </Col>
                </Row>
              </Form.Group>
              <Form.Group className="margin-top-15">
                <Form.Label>{this.props.t("workdays")}</Form.Label>
                <div className="text-left">
                  {[0, 1, 2, 3, 4, 5, 6].map(day => (
                    <Form.Check type="checkbox" key={"workday-" + day} id={"workday-" + day} label={this.props.t("workday-" + day)} checked={this.state.workdays[day]} onChange={(e: any) => this.onWorkdayCheck(day, e.target.checked)} />
                  ))}
                </div>
              </Form.Group>
              <Form.Group className="margin-top-15">
                <Form.Label>{this.props.t("preferredLocation")}</Form.Label>
                <Form.Select value={this.state.locationId} onChange={(e: any) => this.setState({ locationId: e.target.value })}>
                  <option value="">({this.props.t("none")})</option>
                  {this.locations.map(location => <option key={"location-" + location.id} value={location.id}>{location.name}</option>)}
                </Form.Select>
              </Form.Group>
              <Button className="margin-top-15" type="submit" disabled={this.state.submitting}>{this.props.t("save")}</Button>
            </Form>
            <Form onSubmit={this.onSubmitColors} hidden={this.state.activeTab !== 'tab-bookings'}>
              <h5 className='margin-top-50'>{this.props.t("bookingcolors")}</h5>
              <Form.Group className="margin-top-15">
                <Row>
                  <Col>
                    <p>Already booked</p>
                    <Form.Control type="color" key={"booked"} id={"booked"} value={this.state.booked} onChange={(e: any) => this.setState({ booked: e.target.value })} />
                  </Col>
                  <Col>
                    <p>Not booked</p>
                    <Form.Control type="color" key={"notBooked"} id={"notBooked"} value={this.state.notBooked} onChange={(e: any) => this.setState({ notBooked: e.target.value })} />
                  </Col>
                  <Col>
                    <p>Self booked</p>
                    <Form.Control type="color" key={"selfBooked"} id={"selfBooked"} value={this.state.selfBooked} onChange={(e: any) => this.setState({ selfBooked: e.target.value })} />
                  </Col>
                  {RuntimeConfig.INFOS.maxHoursPartiallyBookedEnabled &&
                    <Col>
                      <p>Partially booked</p>
                      <Form.Control type="color" key={"partiallyBooked"} id={"partiallyBooked"} value={this.state.partiallyBooked} onChange={(e: any) => this.setState({ partiallyBooked: e.target.value })} />
                    </Col>
                  }
                  {!RuntimeConfig.INFOS.disableBuddies &&
                    <Col>
                      <p>Buddy booked</p>
                      <Form.Control type="color" key={"buddyBooked"} id={"buddyBooked"} value={this.state.buddyBooked} onChange={(e: any) => this.setState({ buddyBooked: e.target.value })} />
                    </Col>
                  }
                </Row>
              </Form.Group>
              <ButtonGroup className="margin-top-15" >
                <Button type="button" variant='secondary' onClick={() => this.resetColors()}>{this.props.t("reset")}</Button>
                <Button type="submit">{this.props.t("save")}</Button>
              </ButtonGroup>
            </Form>
            <Form onSubmit={this.onSubmitSecurity} hidden={this.state.activeTab !== 'tab-security'}>
              <h5 className='margin-top-15'>{this.props.t("password")}</h5>
              <Form.Group className="margin-top-15">
                <Form.Check type="checkbox" inline={true} id="check-changePassword" label={this.props.t("passwordChange")} checked={this.state.changePassword} onChange={(e: any) => this.setState({ changePassword: e.target.checked })} />
                <Form.Control type="password" value={this.state.password} onChange={(e: any) => this.setState({ password: e.target.value })} required={this.state.changePassword} disabled={!this.state.changePassword} minLength={8} />
              </Form.Group>
              <Button className="margin-top-15" type="submit" disabled={this.state.submitting}>{this.props.t("save")}</Button>
            </Form>
            <Form onSubmit={this.saveCaldavSettings} hidden={this.state.activeTab !== 'tab-integrations'}>
              <h5 className='margin-top-15'>{this.props.t("caldavCalendar")}</h5>
              <Form.Group className="margin-top-15">
                <Form.Label>{this.props.t("caldavUrl")}</Form.Label>
                <Form.Control type='url' value={this.state.caldavUrl} onChange={(e: any) => this.setState({ caldavUrl: e.target.value, caldavCalendarsLoaded: false })} />
              </Form.Group>
              <Form.Group className="margin-top-15">
                <Form.Label>{this.props.t("username")}</Form.Label>
                <Form.Control type='text' value={this.state.caldavUser} onChange={(e: any) => this.setState({ caldavUser: e.target.value, caldavCalendarsLoaded: false })} />
              </Form.Group>
              <Form.Group className="margin-top-15">
                <Form.Label>{this.props.t("password")}</Form.Label>
                <Form.Control type='password' value={this.state.caldavPass} onChange={(e: any) => this.setState({ caldavPass: e.target.value, caldavCalendarsLoaded: false })} />
              </Form.Group>
              <Form.Group className="margin-top-15">
                <Form.Label>{this.props.t("calendar")}</Form.Label>
                <Form.Select value={this.state.caldavCalendar} onChange={(e: any) => this.setState({ caldavCalendar: e.target.value })} disabled={!this.state.caldavCalendarsLoaded}>
                  {this.state.caldavCalendars.map(cal => <option key={cal.path} value={cal.path}>{cal.name}</option>)}
                </Form.Select>
              </Form.Group>
              <ButtonGroup className="margin-top-15" >
                <Button type="button" variant='secondary' disabled={this.state.submitting || this.state.caldavUrl === '' || this.state.caldavUser === '' || this.state.caldavPass === ''} onClick={() => this.connectCalDav()}>{this.props.t("connect")}</Button>
                <Button type="button" variant='secondary' disabled={this.state.submitting || this.state.caldavUrl === '' || this.state.caldavUser === '' || this.state.caldavPass === '' || this.state.caldavCalendar === ''} onClick={() => this.disconnectCalDav()}>{this.props.t("disconnect")}</Button>
                <Button type="submit" disabled={!(this.state.caldavCalendarsLoaded && this.state.caldavCalendar != '') || this.state.submitting}>{this.props.t("save")}</Button>
              </ButtonGroup>
            </Form>
          </div>
        </div>
      </>
    );
  }
}

export default withTranslation()(withReadyRouter(Preferences as any));
