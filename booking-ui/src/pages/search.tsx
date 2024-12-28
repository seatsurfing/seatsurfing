import React, { RefObject } from 'react';
import { Form, Col, Row, Modal, Button, ListGroup, Badge, InputGroup } from 'react-bootstrap';
import { Location, Booking, Buddy, User, Ajax, Formatting, Space, AjaxError, UserPreference, SpaceAttributeValue, SpaceAttribute, SearchAttribute } from 'flexspace-commons';
// @ts-ignore
import DateTimePicker from 'react-datetime-picker';
import DatePicker from 'react-date-picker';
import 'react-datetime-picker/dist/DateTimePicker.css';
import 'react-date-picker/dist/DatePicker.css';
import 'react-calendar/dist/Calendar.css';
import 'react-clock/dist/Clock.css';
import Loading from '../components/Loading';
import { FilterOutline as FilterIcon, InformationOutline as InfoIcon, EnterOutline as EnterIcon, ExitOutline as ExitIcon, LocationOutline as LocationIcon, ChevronUpOutline as CollapseIcon, ChevronDownOutline as CollapseIcon2, SettingsOutline as SettingsIcon, MapOutline as MapIcon, CalendarOutline as WeekIcon } from 'react-ionicons'
import ErrorText from '../types/ErrorText';
import { NextRouter } from 'next/router';
import { WithTranslation, withTranslation } from 'next-i18next';
import NavBar from '@/components/NavBar';
import RuntimeConfig from '@/components/RuntimeConfig';
import withReadyRouter from '@/components/withReadyRouter';
import { Tooltip } from 'react-tooltip';

interface State {
  earliestEnterDate: Date;
  enter: Date
  leave: Date
  daySlider: number
  daySliderDisabled: boolean
  locationId: string
  canSearch: boolean
  canSearchHint: string
  showBookingNames: boolean
  selectedSpace: Space | null
  showConfirm: boolean
  showLocationDetails: boolean
  showSearchModal: boolean
  showSuccess: boolean
  showError: boolean
  errorText: string
  loading: boolean
  listView: boolean
  prefEnterTime: number
  prefWorkdayStart: number
  prefWorkdayEnd: number
  prefWorkdays: number[]
  prefLocationId: string
  prefBookedColor: string
  prefNotBookedColor: string
  prefSelfBookedColor: string
  prefPartiallyBookedColor: string
  prefBuddyBookedColor: string
  attributeValues: SpaceAttributeValue[]
  searchAttributes: SearchAttribute[]
}

interface Props extends WithTranslation {
  router: NextRouter
}

class Search extends React.Component<Props, State> {
  static PreferenceEnterTimeNow: number = 1;
  static PreferenceEnterTimeNextDay: number = 2;
  static PreferenceEnterTimeNextWorkday: number = 3;

  data: Space[];
  locations: Location[]
  mapData: any;
  curBookingCount: number = 0;
  searchContainerRef: RefObject<any>;
  enterChangeTimer: number | undefined;
  leaveChangeTimer: number | undefined;
  buddies: Buddy[];
  availableAttributes: SpaceAttribute[];

  constructor(props: any) {
    super(props);
    this.data = [];
    this.locations = [];
    this.mapData = null;
    this.buddies = [];
    this.availableAttributes = [];
    this.searchContainerRef = React.createRef();
    this.enterChangeTimer = undefined;
    this.leaveChangeTimer = undefined;
    this.state = {
      earliestEnterDate: new Date(),
      enter: new Date(),
      leave: new Date(),
      locationId: "",
      daySlider: 0,
      daySliderDisabled: false,
      canSearch: false,
      canSearchHint: "",
      showBookingNames: false,
      selectedSpace: null,
      showConfirm: false,
      showLocationDetails: false,
      showSearchModal: false,
      showSuccess: false,
      showError: false,
      errorText: "",
      loading: true,
      listView: false,
      prefEnterTime: 0,
      prefWorkdayStart: 0,
      prefWorkdayEnd: 0,
      prefWorkdays: [],
      prefLocationId: "",
      prefBookedColor: "#ff453a",
      prefNotBookedColor: "#30d158",
      prefSelfBookedColor: "#b825de",
      prefPartiallyBookedColor: "#ff9100",
      prefBuddyBookedColor: "#2415c5",
      attributeValues: [],
      searchAttributes: [],
    };
  }

  componentDidMount = () => {
    console.log(RuntimeConfig.INFOS);
    if (!Ajax.CREDENTIALS.accessToken) {
      this.props.router.push({ pathname: "/login", query: { redir: this.props.router.asPath } });
      return;
    }
    this.loadItems();
  }

  loadItems = () => {
    let promises = [
      this.loadLocations(),
      this.loadPreferences(),
      this.loadBuddies(),
      this.loadAvailableAttributes(),
    ];
    Promise.all(promises).then(() => {
      this.initDates();
      if (this.state.locationId === "" && this.locations.length > 0) {
        let defaultLocationId = this.getPreferredLocationId(this.props.router.query["lid"] as string || '');
        let sidParam = this.props.router.query["sid"] as string || "";
        this.setState({ locationId: defaultLocationId }, () => {
          this.getLocation()?.getAttributes().then((attributes) => {
            this.loadMap(this.state.locationId).then(() => {
              this.setState({
                attributeValues: attributes,
                loading: false
              });
              if (sidParam) {
                let space = this.data.find((item) => item.id == sidParam);
                if (space) this.onSpaceSelect(space);
              }
            });
          });
        });
      } else {
        this.setState({ loading: false });
      }
    });
  }

  loadPreferences = async (): Promise<void> => {
    let self = this;
    return new Promise<void>(function (resolve, reject) {
      UserPreference.list().then(list => {
        let state: any = {};
        list.forEach(s => {
          if (typeof window !== 'undefined') {
            if (s.name === "enter_time") state.prefEnterTime = window.parseInt(s.value);
            if (s.name === "workday_start") state.prefWorkdayStart = window.parseInt(s.value);
            if (s.name === "workday_end") state.prefWorkdayEnd = window.parseInt(s.value);
            if (s.name === "workdays") state.prefWorkdays = s.value.split(",").map(val => window.parseInt(val));
          }
          if (s.name === "location_id") state.prefLocationId = s.value;
          if (s.name === "booked_color") state.prefBookedColor = s.value;
          if (s.name === "not_booked_color") state.prefNotBookedColor = s.value;
          if (s.name === "self_booked_color") state.prefSelfBookedColor = s.value;
          if (s.name === "partially_booked_color") state.prefPartiallyBookedColor = s.value;
          if (s.name === "buddy_booked_color") state.prefBuddyBookedColor = s.value;
        });
        if (RuntimeConfig.INFOS.dailyBasisBooking) {
          state.prefWorkdayStart = 0;
          state.prefWorkdayEnd = 23;
        }
        self.setState({
          ...state
        }, () => resolve());
      }).catch(e => reject(e));
    });
  }

  initCurrentBookingCount = () => {
    Booking.list().then(list => {
      this.curBookingCount = list.length;
      this.updateCanSearch();
    });
  }

  getPreferredLocationId = (previousLocationId?: string) => {
    if (previousLocationId !== undefined) {
      if (this.locations.find((e) => (e.id === previousLocationId) && e.enabled) !== undefined) {
        return previousLocationId
      }
    }
    if (this.state.prefLocationId && this.locations.find((e) => (e.id === this.state.prefLocationId) && e.enabled) !== undefined) {
      return this.state.prefLocationId;
    }
    for (let location of this.locations) {
      if (location.enabled) {
        return location.id;
      }
    }
    return '';
  }

  initDates = () => {
    let enter = new Date();
    if (this.state.prefEnterTime === Search.PreferenceEnterTimeNow) {
      enter.setHours(enter.getHours() + 1, 0, 0);
      if (enter.getHours() < this.state.prefWorkdayStart) {
        enter.setHours(this.state.prefWorkdayStart, 0, 0, 0);
      }
      if (enter.getHours() >= this.state.prefWorkdayEnd) {
        enter.setDate(enter.getDate() + 1);
        enter.setHours(this.state.prefWorkdayStart, 0, 0, 0);
      }
    } else if (this.state.prefEnterTime === Search.PreferenceEnterTimeNextDay) {
      enter.setDate(enter.getDate() + 1);
      enter.setHours(this.state.prefWorkdayStart, 0, 0, 0);
    } else if (this.state.prefEnterTime === Search.PreferenceEnterTimeNextWorkday) {
      enter.setDate(enter.getDate() + 1);
      let add = 0;
      let nextDayFound = false;
      let lookFor = enter.getDay();
      while (!nextDayFound) {
        if (this.state.prefWorkdays.includes(lookFor) || add > 7) {
          nextDayFound = true;
        } else {
          add++;
          lookFor++;
          if (lookFor > 6) {
            lookFor = 0;
          }
        }
      }
      enter.setDate(enter.getDate() + add);
      enter.setHours(this.state.prefWorkdayStart, 0, 0, 0);
    }

    let leave = new Date(enter);
    leave.setHours(this.state.prefWorkdayEnd, 0, 0);

    if (RuntimeConfig.INFOS.dailyBasisBooking) {
      enter.setHours(0, 0, 0, 0);
      leave.setHours(23, 59, 59, 0);
    }

    this.setState({
      earliestEnterDate: enter,
      enter: enter,
      leave: leave
    });
  }

  loadLocations = async (): Promise<void> => {
    return Location.list().then(list => {
      this.locations = list;
    });
  }

  loadAvailableAttributes = async (): Promise<void> => {
    return SpaceAttribute.list().then(attributes => {
      this.availableAttributes = attributes;
    });
  }

  loadBuddies = async (): Promise<void> => {
    return Buddy.list().then(list => {
      this.buddies = list;
    });
  }

  loadMap = async (locationId: string) => {
    this.setState({ loading: true });
    return Location.get(locationId).then(location => {
      return this.loadSpaces(location.id).then(() => {
        return Ajax.get(location.getMapUrl()).then(mapData => {
          this.mapData = mapData.json;
          this.centerMapView();
        });
      });
    })
  }

  centerMapView = () => {
    if (typeof window !== 'undefined') {
      let timer: number | undefined = undefined;
      let cb = () => {
        const el = document.querySelector('.mapScrollContainer');
        if (el) {
          window.clearInterval(timer);
          el.scrollLeft = (this.mapData ? this.mapData.width : 0) / 2 - (window.innerWidth / 2);
          el.scrollTop = (this.mapData ? this.mapData.height : 0) / 2 - (window.innerHeight / 2);
        }
      };
      timer = window.setInterval(cb, 10);
    }
  }

  loadSpaces = async (locationId: string) => {
    this.setState({ loading: true });
    let leave = new Date(this.state.leave);
    if (!RuntimeConfig.INFOS.dailyBasisBooking) {
      leave.setSeconds(leave.getSeconds() - 1);
    }
    return Space.listAvailability(locationId, this.state.enter, leave).then(list => {
      this.data = list;
    });
  }

  updateCanSearch = async () => {
    let res = true;
    let hint = "";
    let isAdmin = RuntimeConfig.INFOS.noAdminRestrictions && User.UserRoleSpaceAdmin;
    if (this.curBookingCount >= RuntimeConfig.INFOS.maxBookingsPerUser && !isAdmin) {
      res = false;
      hint = this.props.t("errorBookingLimit", { "num": RuntimeConfig.INFOS.maxBookingsPerUser });
    }
    if (!this.state.locationId) {
      res = false;
      hint = this.props.t("errorPickArea");
    }
    let now = new Date();
    let enterTime = new Date(this.state.enter);
    if (RuntimeConfig.INFOS.dailyBasisBooking) {
      enterTime.setHours(23, 59, 59);
    }
    if (enterTime.getTime() <= now.getTime()) {
      res = false;
      hint = this.props.t("errorEnterFuture");
    }
    if (this.state.leave.getTime() <= this.state.enter.getTime()) {
      res = false;
      hint = this.props.t("errorLeaveAfterEnter");
    }
    const MS_PER_MINUTE = 1000 * 60;
    const MS_PER_HOUR = MS_PER_MINUTE * 60;
    const MS_PER_DAY = MS_PER_HOUR * 24;
    let bookingAdvanceDays = Math.floor((this.state.enter.getTime() - new Date().getTime()) / MS_PER_DAY);
    if (bookingAdvanceDays > RuntimeConfig.INFOS.maxDaysInAdvance && !isAdmin) {
      res = false;
      hint = this.props.t("errorDaysAdvance", { "num": RuntimeConfig.INFOS.maxDaysInAdvance });
    }
    let bookingDurationHours = Math.floor((this.state.leave.getTime() - this.state.enter.getTime()) / MS_PER_MINUTE) / 60;
    if (bookingDurationHours > RuntimeConfig.INFOS.maxBookingDurationHours && !isAdmin) {
      res = false;
      hint = this.props.t("errorMaxBookingDuration", { "num": RuntimeConfig.INFOS.maxBookingDurationHours });
    }
    if (bookingDurationHours < RuntimeConfig.INFOS.minBookingDurationHours && !isAdmin) {
      res = false;
      hint = this.props.t("errorMinBookingDuration", { "num": RuntimeConfig.INFOS.minBookingDurationHours });
    }
    let self = this;
    return new Promise<void>(function (resolve, reject) {
      self.setState({
        canSearch: res,
        canSearchHint: hint
      }, () => resolve());
    });
  }

  renderLocations = () => {
    return this.locations.map(location => {
      return <option value={location.id} key={location.id} disabled={!location.enabled}>{location.name}</option>;
    });
  }

  changeEnterDay = (value: number) => {
    let enter = new Date(this.state.earliestEnterDate.valueOf());
    enter.setDate(enter.getDate() + value);
    if (Formatting.getDayValue(enter) > Formatting.getDayValue(this.state.earliestEnterDate)) {
      enter.setHours(this.state.prefWorkdayStart, 0, 0, 0);
    }
    let leave = new Date(enter.valueOf());
    leave.setHours(this.state.prefWorkdayEnd, 0, 0, 0);
    this.setEnterDate(enter);
    this.setLeaveDate(leave);
    this.setState({ daySlider: value });
  }

  setEnterDate = (value: Date | [Date | null, Date | null]) => {
    console.log("New enter date: " + value);
    let dateChangedCb = () => {
      this.updateCanSearch().then(() => {
        if (!this.state.canSearch) {
          this.setState({ loading: false });
        } else {
          let promises = [
            this.initCurrentBookingCount(),
            this.loadSpaces(this.state.locationId),
          ];
          Promise.all(promises).then(() => {
            this.setState({ loading: false });
          });
        }
      });
    };
    let performChange = () => {
      let diff = this.state.leave.getTime() - this.state.enter.getTime();
      let date = (value instanceof Date) ? value : value[0];
      if (date == null) {
        return;
      }
      if (RuntimeConfig.INFOS.dailyBasisBooking) {
        date.setHours(0, 0, 0);
      }
      let leave = new Date();
      leave.setTime(date.getTime() + diff);
      const daySlider = Formatting.getDayDiff(date, this.state.earliestEnterDate);
      const daySliderDisabled = (daySlider > RuntimeConfig.INFOS.maxDaysInAdvance) || (daySlider < 0);
      this.setState({
        enter: date,
        leave: leave,
        daySlider: daySlider,
        daySliderDisabled: daySliderDisabled
      }, () => dateChangedCb());
    };
    if (typeof window !== 'undefined') {
      window.clearTimeout(this.enterChangeTimer);
      this.enterChangeTimer = window.setTimeout(performChange, 1000);
    }
  }

  setLeaveDate = (value: Date | [Date | null, Date | null]) => {
    let dateChangedCb = () => {
      this.updateCanSearch().then(() => {
        if (!this.state.canSearch) {
          this.setState({ loading: false });
        } else {
          let promises = [
            this.initCurrentBookingCount(),
            this.loadSpaces(this.state.locationId),
          ];
          Promise.all(promises).then(() => {
            this.setState({ loading: false });
          });
        }
      });
    };
    let performChange = () => {
      let date = (value instanceof Date) ? value : value[0];
      if (date == null) {
        return;
      }
      if (RuntimeConfig.INFOS.dailyBasisBooking) {
        date.setHours(23, 59, 59);
      }
      this.setState({
        leave: date
      }, () => dateChangedCb());
    };
    if (typeof window !== 'undefined') {
      window.clearTimeout(this.leaveChangeTimer);
      this.leaveChangeTimer = window.setTimeout(performChange, 1000);
    }
  }

  changeLocation = (id: string) => {
    this.setState({
      locationId: id,
      loading: true,
    }, () => {
      this.getLocation()?.getAttributes().then((attributes) => {
        this.loadMap(id).then(() => {
          this.setState({
            attributeValues: attributes,
            loading: false
          });
        });
      });
    });
  }

  onSpaceSelect = (item: Space) => {
    if (item.available) {
      this.setState({
        showConfirm: true,
        selectedSpace: item
      });
    } else {
      let bookings = Booking.createFromRawArray(item.rawBookings);
      if (!item.available && bookings && bookings.length > 0) {
        this.setState({
          showBookingNames: true,
          selectedSpace: item
        });
      }
    }
  }

  getAvailibilityStyle = (item: Space, bookings: Booking[]) => {
    const mydesk = (bookings.find(b => b.user.email === RuntimeConfig.INFOS.username));
    const buddiesEmails = this.buddies.map(i => i.buddy.email);
    const myBuddyDesk = (bookings.find(b => buddiesEmails.includes(b.user.email)));

    if (myBuddyDesk) {
      return this.state.prefBuddyBookedColor;
    }

    if (mydesk) {
      return this.state.prefSelfBookedColor;
    }

    if (RuntimeConfig.INFOS.maxHoursPartiallyBookedEnabled && bookings.length > 0) {
      let prefWorkdayStartDate = new Date(this.state.enter);
      prefWorkdayStartDate.setHours(this.state.prefWorkdayStart, 0, 0);
      prefWorkdayStartDate = Formatting.convertToFakeUTCDate(prefWorkdayStartDate);
      let prefWorkdayEndDate = new Date(this.state.leave);
      prefWorkdayEndDate.setHours(this.state.prefWorkdayEnd, 0, 0);
      prefWorkdayEndDate = Formatting.convertToFakeUTCDate(prefWorkdayEndDate);

      let leastEnter = bookings.reduce((a, b) => a.enter < b.enter ? a : b).enter;
      if (leastEnter < prefWorkdayStartDate) {
        leastEnter = prefWorkdayStartDate;
      }

      let maxLeave = bookings.reduce((a, b) => a.leave > b.leave ? a : b).leave;
      if (maxLeave > prefWorkdayEndDate) {
        maxLeave = prefWorkdayEndDate;
      }
      const hours = (maxLeave.getTime() - leastEnter.getTime()) / 1000 / 60 / 60;

      if (hours < RuntimeConfig.INFOS.maxHoursPartiallyBooked) {
        return this.state.prefPartiallyBookedColor;
      }
    }

    return (item.available ? this.state.prefNotBookedColor : this.state.prefBookedColor);
  }

  getBookersList = (bookings: Booking[]) => {
    if (!bookings.length) return "";
    let str = "";
    bookings.forEach(b => {
      str += (str ? ", " : "") + b.user.email
    });
    return str;
  }

  renderItem = (item: Space) => {
    let bookings = Booking.createFromRawArray(item.rawBookings);
    const boxStyle: React.CSSProperties = {
      position: "absolute",
      left: item.x,
      top: item.y,
      width: item.width,
      height: item.height,
      transform: "rotate: " + item.rotation + "deg",
      cursor: (item.available || (bookings && bookings.length > 0)) ? "pointer" : "default",
      backgroundColor: this.getAvailibilityStyle(item, bookings)
    };
    const textStyle: React.CSSProperties = {
      textAlign: "center"
    };
    const className = "space space-box"
      + ((item.width < item.height) ? " space-box-vertical" : "");
    return (

      <div key={item.id} style={boxStyle} className={className} data-tooltip-id="my-tooltip" data-tooltip-content={item.rawBookings[0] ? item.rawBookings[0].userEmail : "Free"}
        onClick={() => this.onSpaceSelect(item)}
        title={this.getBookersList(bookings)}>
        <Tooltip id="my-tooltip" />
        <p style={textStyle}>{item.name}</p>
      </div>
    );
  }

  renderListItem = (item: Space) => {
    let bookings: Booking[] = [];
    bookings = Booking.createFromRawArray(item.rawBookings);
    const bgColor = this.getAvailibilityStyle(item, bookings);
    let bookerCount = 0;
    if (bgColor === this.state.prefSelfBookedColor) {
      bookerCount = 1;
    } else if (bgColor === this.state.prefBookedColor || bgColor === this.state.prefBuddyBookedColor) {
      bookerCount = (bookings.length > 0 ? bookings.length : 1);
    }
    return (
      <ListGroup.Item key={item.id} action={true} onClick={(e) => { e.preventDefault(); this.onSpaceSelect(item); }} className="d-flex justify-content-between align-items-start space-list-item">
        <div className="ms-2 me-auto">
          <div className="fw-bold space-list-item-content">{item.name}</div>
          {bookings.map((booking) => (
            <div key={booking.user.id} className="space-list-item-content">
              {booking.user.email}
            </div>
          ))}
        </div>
        <span className='badge badge-pill' style={{ backgroundColor: bgColor }}>
          {bookerCount}
        </span>
      </ListGroup.Item>
    );
  }

  renderBookingNameRow = (booking: Booking) => {
    const buddiesEmails = this.buddies.map(i => i.buddy.email);

    return (
      <p key={booking.id}>
        {booking.user.email}<br />
        {Formatting.getFormatterShort().format(new Date(booking.enter))}
        &nbsp;&mdash;&nbsp;
        {Formatting.getFormatterShort().format(new Date(booking.leave))}
        {RuntimeConfig.INFOS.showNames && !RuntimeConfig.INFOS.disableBuddies && booking.user.email !== RuntimeConfig.INFOS.username && !buddiesEmails.includes(booking.user.email) && (
          <Button variant="primary" onClick={(e) => { e.preventDefault(); this.onAddBuddy(booking.user); }} style={{ marginLeft: '10px' }}>
            {this.props.t("addBuddy")}
          </Button>
        )}
      </p>
    );
  }

  onConfirmBooking = () => {
    if (this.state.selectedSpace == null) {
      return;
    }
    this.setState({
      showConfirm: false,
      loading: true
    });
    let booking: Booking = new Booking();
    booking.enter = new Date(this.state.enter);
    booking.leave = new Date(this.state.leave);
    if (!RuntimeConfig.INFOS.dailyBasisBooking) {
      booking.leave.setSeconds(booking.leave.getSeconds() - 1);
    }
    booking.space = this.state.selectedSpace;
    booking.save().then(() => {
      this.setState({
        loading: false,
        showSuccess: true
      });
    }).catch(e => {
      let code: number = 0;
      if (e instanceof AjaxError) {
        code = e.appErrorCode;
      }
      this.setState({
        loading: false,
        showError: true,
        errorText: ErrorText.getTextForAppCode(code, this.props.t)
      });
    });
  }

  onAddBuddy = (buddyUser: User) => {
    if (this.state.selectedSpace == null) {
      return;
    }
    this.setState({
      showBookingNames: false,
      loading: true
    });
    let buddy: Buddy = new Buddy();
    buddy.buddy = buddyUser;
    buddy.save().then(() => {
      this.loadBuddies().then(() => {
        this.setState({ loading: false });
      });
    }).catch(e => {
      let code: number = 0;
      if (e instanceof AjaxError) {
        code = e.appErrorCode;
      }
      this.setState({
        loading: false,
        showError: true,
        errorText: ErrorText.getTextForAppCode(code, this.props.t),
      });
    });
  }

  getLocation = (): Location | undefined => {
    return this.locations.find(e => e.id === this.state.locationId);
  }

  getLocationName = (): string => {
    let name: string = this.props.t("none");
    let location = this.getLocation();
    if (location) {
      name = location.name;
    }
    return name;
  }

  toggleSearchContainer = () => {
    const ref = this.searchContainerRef.current;
    ref.classList.toggle("minimized");

    const map = document.querySelector('.container-map');
    if (map) map.classList.toggle("maximized");
    const list = document.querySelector('.space-list');
    if (list) list.classList.toggle("maximized");
  }

  toggleListView = () => {
    this.setState({ listView: !this.state.listView }, () => {
      if (!this.state.listView) {
        this.centerMapView();
      }
    });
  }

  getLocationAttributeRows = () => {
    let location = this.getLocation();
    if (!location) {
      return <></>;
    }
    return this.state.attributeValues.map((attributeValue) => {
      let attribute = this.availableAttributes.find((attr) => attr.id === attributeValue.attributeId);
      if (!attribute) {
        return <></>;
      }
      return (
        <Form.Group as={Row} key={attribute.id}>
          <Form.Label column sm="4">
            {attribute.label}:
          </Form.Label>
          <Col sm="8">
            <Form.Control plaintext={true} readOnly={true} defaultValue={attribute.type === 2 ? (attributeValue.value === '1' ? this.props.t("yes") : '') : attributeValue.value} />
          </Col>
        </Form.Group>
      );
    });
  }

  getSearchFormComparator = (attribute: SpaceAttribute) => {
    let items = [];
    items.push(<option value=''></option>);
    items.push(<option value='eq'>=</option>);
    items.push(<option value='neq'>&lt;&gt;</option>);
    if (attribute.type === 1) {
      items.push(<option value='gt'>&gt;</option>);
      items.push(<option value='lt'>&lt;</option>);
    }
    if (attribute.type === 3) {
      items.push(<option value='contains'>~</option>);
    }
    return items;
  }

  getSearchFormInput = (attribute: SpaceAttribute) => {
    if (attribute.type === 1) {
      return <Form.Control type="number" min={0} value={this.state.searchAttributes.find((attr) => attr.attributeId === attribute.id)?.value || ''} onChange={(e: any) => this.setSearchAttributeValue(attribute.id, e.target.value)} disabled={this.state.searchAttributes.find((attr) => attr.attributeId === attribute.id) === undefined} />;
    } else if (attribute.type === 2) {
      return <Form.Check type="checkbox" style={{ paddingTop: '5px' }} label={this.props.t("yes")} checked={this.state.searchAttributes.find((attr) => attr.attributeId === attribute.id)?.value === '1' || false} onChange={(e: any) => this.setSearchAttributeValue(attribute.id, e.target.checked ? '1' : '0')} disabled={this.state.searchAttributes.find((attr) => attr.attributeId === attribute.id) === undefined} />;
    } else if (attribute.type === 3) {
      return <Form.Control type="text" value={this.state.searchAttributes.find((attr) => attr.attributeId === attribute.id)?.value || ''} onChange={(e: any) => this.setSearchAttributeValue(attribute.id, e.target.value)} disabled={this.state.searchAttributes.find((attr) => attr.attributeId === attribute.id) === undefined} />;
    }
  }

  setSearchAttributeComparator = (attributeId: string, comparator: string) => {
    let searchAttributes = this.state.searchAttributes;
    if (comparator === '') {
      searchAttributes = searchAttributes.filter((attr) => attr.attributeId !== attributeId);
      this.setState({ searchAttributes: searchAttributes });
      return;
    }
    let searchAttribute = searchAttributes.find((attr) => attr.attributeId === attributeId);
    if (!searchAttribute) {
      searchAttribute = new SearchAttribute();
      searchAttribute.attributeId = attributeId;
      searchAttributes.push(searchAttribute);
    }
    searchAttribute.comparator = comparator;
    this.setState({ searchAttributes: searchAttributes });
  }

  setSearchAttributeValue = (attributeId: string, value: string) => {
    let searchAttributes = this.state.searchAttributes;
    let searchAttribute = searchAttributes.find((attr) => attr.attributeId === attributeId);
    if (!searchAttribute) {
      searchAttribute = new SearchAttribute();
      searchAttribute.attributeId = attributeId;
      searchAttributes.push(searchAttribute);
    }
    searchAttribute.value = value;
    this.setState({ searchAttributes: searchAttributes });
  }

  getSearchFormRows = () => {
    return this.availableAttributes.map(attribute => {
      if (!attribute.locationApplicable) {
        return <></>;
      }
      return (
        <Form.Group as={Row} key={attribute.id}>
          <Form.Label column sm="4">{attribute.label}</Form.Label>
          <Col sm="3">
            <Form.Select value={this.state.searchAttributes.find((attr) => attr.attributeId === attribute.id)?.comparator || ''} onChange={(e: any) => this.setSearchAttributeComparator(attribute.id, e.target.value)}>
              {this.getSearchFormComparator(attribute)}
            </Form.Select>
          </Col>
          <Col sm="5">
            {this.getSearchFormInput(attribute)}
          </Col>
        </Form.Group>
      );
    });
  }

  resetSearch = () => {
    this.setState({
      searchAttributes: []
    }, () => {
      this.applySearch();
    });
  }

  applySearch = () => {
    this.setState({
      showSearchModal: false,
      loading: true,
    });
    SearchAttribute.search(this.state.searchAttributes).then((locations) => {
      this.locations = locations;
      if ((locations.length === 0) || (this.locations.find((e) => e.enabled) === undefined)) {
        this.setState({
          locationId: "",
          loading: false,
        });
        return;
      }
      let newLocationId = this.getPreferredLocationId(this.state.locationId);
      this.setState({
        locationId: newLocationId,
      }, () => {
        this.loadMap(this.state.locationId).then(() => {
          this.getLocation()?.getAttributes().then((attributes) => {
            this.setState({ loading: false });
          });
        });
      });
    });
  }

  render() {
    let hint = <></>;
    if ((!this.state.canSearch) && (this.state.canSearchHint)) {
      hint = (
        <Form.Group as={Row} className="margin-top-10">
          <Col xs="2"></Col>
          <Col xs="10">
            <div className="invalid-search-config">{this.state.canSearchHint}</div>
          </Col>
        </Form.Group>
      );
    }
    let enterDatePicker = <DateTimePicker disabled={!this.state.locationId} value={this.state.enter} onChange={(value: Date | null) => { if (value != null) this.setEnterDate(value) }} clearIcon={null} required={true} format={this.props.t("datePickerFormat")} />;
    if (RuntimeConfig.INFOS.dailyBasisBooking) {
      enterDatePicker = <DatePicker disabled={!this.state.locationId} value={this.state.enter} onChange={(value: Date | null | [Date | null, Date | null]) => { if (value != null) this.setEnterDate(value) }} clearIcon={null} required={true} format={this.props.t("datePickerFormatDailyBasisBooking")} />;
    }
    let leaveDatePicker = <DateTimePicker disabled={!this.state.locationId} value={this.state.leave} onChange={(value: Date | null) => { if (value != null) this.setLeaveDate(value) }} clearIcon={null} required={true} format={this.props.t("datePickerFormat")} />;
    if (RuntimeConfig.INFOS.dailyBasisBooking) {
      leaveDatePicker = <DatePicker disabled={!this.state.locationId} value={this.state.leave} onChange={(value: Date | null | [Date | null, Date | null]) => { if (value != null) this.setLeaveDate(value) }} clearIcon={null} required={true} format={this.props.t("datePickerFormatDailyBasisBooking")} />;
    }

    let listOrMap = <></>;
    if ((this.locations.length === 0) || (!this.state.locationId)) {
      listOrMap = (
        <div className="container-signin">
          <Form className="form-signin">
            <div style={{paddingBottom: '100px'}} dangerouslySetInnerHTML={{__html: this.props.t("noAreasFounds").replace('.', '.<br />')}}></div>
          </Form>
        </div>
      );
    } else if (this.state.listView) {
      listOrMap = (
        <div className="container-signin">
          <Form className="form-signin">
            <ListGroup className="space-list">
              {this.data.map(item => this.renderListItem(item))}
            </ListGroup>
          </Form>
        </div>
      );
    } else {
      const floorPlanStyle = {
        width: (this.mapData ? this.mapData.width : 0) + "px",
        height: (this.mapData ? this.mapData.height : 0) + "px",
        position: 'relative' as 'relative',
        backgroundImage: (this.mapData ? "url(data:image/" + this.mapData.mapMimeType + ";base64," + this.mapData.data + ")" : "")
      };
      let spaces = this.data.map((item) => {
        return this.renderItem(item);
      });
      listOrMap = (
        <div className="container-map">
          <div className="mapScrollContainer">
            <div style={floorPlanStyle}>
              {spaces}
            </div>
          </div>
        </div>
      );
    }

    let configContainer = (
      <div className="container-search-config" ref={this.searchContainerRef}>
        <div className="collapse-bar" onClick={() => this.toggleSearchContainer()}>
          <CollapseIcon color={'#000'} height="20px" width="20px" cssClasses="collapse-icon collapse-icon-bigscreen" />
          <CollapseIcon2 color={'#000'} height="20px" width="20px" cssClasses="collapse-icon collapse-icon-smallscreen" />
          <SettingsIcon color={'#555'} height="26px" width="26px" cssClasses="expand-icon expand-icon-bigscreen" />
          <CollapseIcon color={'#555'} height="20px" width="20px" cssClasses="expand-icon expand-icon-smallscreen" />
        </div>
        <div className="content-minimized">
          <div className='d-flex'>
            <div className='me-2'><LocationIcon title={this.props.t("area")} color={'#555'} height="20px" width="20px" /></div>
            <div className='ms-2 w-100'>{this.getLocationName()}</div>
          </div>
          <div className='d-flex'>
            <div className='me-2'><EnterIcon title={this.props.t("enter")} color={'#555'} height="20px" width="20px" /></div>
            <div className='ms-2 w-100'>{Formatting.getFormatterShort().format(Formatting.convertToFakeUTCDate(new Date(this.state.enter)))}</div>
          </div>
          <div className='d-flex'>
            <div className='me-2'><ExitIcon title={this.props.t("leave")} color={'#555'} height="20px" width="20px" /></div>
            <div className='ms-2 w-100'>{Formatting.getFormatterShort().format(Formatting.convertToFakeUTCDate(new Date(this.state.leave)))}</div>
          </div>
        </div>
        <div className="content">
          <Form>
            <Form.Group className="d-flex">
              <div className='pt-1 me-2'><LocationIcon title={this.props.t("area")} color={'#555'} height="20px" width="20px" /></div>
              <div className='ms-2 w-100'>
                <InputGroup>
                  <Form.Select required={true} value={this.state.locationId} onChange={(e) => this.changeLocation(e.target.value)} disabled={this.locations.length === 0}>
                    {this.renderLocations()}
                  </Form.Select>
                  <Button variant='outline-secondary' className='addon-button' disabled={(!this.state.locationId) || (this.state.attributeValues.length === 0)} onClick={() => this.setState({ showLocationDetails: true })}><InfoIcon /></Button>
                  <Button variant={this.state.searchAttributes.length === 0 ? 'outline-secondary' : 'primary'} className='addon-button' onClick={() => this.setState({ showSearchModal: true })}><FilterIcon color={this.state.searchAttributes.length === 0 ? undefined : 'white'} /></Button>
                </InputGroup>
              </div>
            </Form.Group>
            <Form.Group className="d-flex margin-top-10">
              <div className='pt-1 me-2'><EnterIcon title={this.props.t("enter")} color={'#555'} height="20px" width="20px" /></div>
              <div className='ms-2 w-100'>
                {enterDatePicker}
              </div>
            </Form.Group>
            <Form.Group className="d-flex margin-top-10">
              <div className='pt-1 me-2'><ExitIcon title={this.props.t("leave")} color={'#555'} height="20px" width="20px" /></div>
              <div className='ms-2 w-100'>
                {leaveDatePicker}
              </div>
            </Form.Group>
            {hint}
            <Form.Group className="d-flex margin-top-10">
              <div className='me-2'><WeekIcon title={this.props.t("week")} color={'#555'} height="20px" width="20px" /></div>
              <div className='ms-2 w-100'>
                <Form.Range disabled={!this.state.locationId || this.state.daySliderDisabled} list="weekDays" min={0} max={RuntimeConfig.INFOS.maxDaysInAdvance} step="1" value={this.state.daySlider} onChange={(event) => this.changeEnterDay(window.parseInt(event.target.value))} />
              </div>
            </Form.Group>
            <Form.Group className="d-flex margin-top-10">
              <div className='me-2'><MapIcon title={this.props.t("map")} color={'#555'} height="20px" width="20px" /></div>
              <div className='ms-2 w-100'>
                <Form.Check disabled={!this.state.locationId} type="switch" checked={!this.state.listView} onChange={() => this.toggleListView()} label={this.state.listView ? this.props.t("showList") : this.props.t("showMap")} />
              </div>
            </Form.Group>
          </Form>
        </div>
      </div>
    );

    let formatter = Formatting.getFormatter();
    if (RuntimeConfig.INFOS.dailyBasisBooking) {
      formatter = Formatting.getFormatterNoTime();
    }
    let locationInfoModal = (
      <Modal show={this.state.showLocationDetails} onHide={() => this.setState({ showLocationDetails: false })}>
        <Modal.Header closeButton>
        </Modal.Header>
        <Modal.Body>
          {this.getLocationAttributeRows()}
        </Modal.Body>
      </Modal>
    );
    let searchModal = (
      <Modal show={this.state.showSearchModal} onHide={() => this.setState({ showSearchModal: false })}>
        <Modal.Header closeButton={true}>
          <Modal.Title>{this.props.t("filter")}</Modal.Title>
        </Modal.Header>
        <Form id='filter-locations-form'>
          <Modal.Body>
            {this.getSearchFormRows()}
          </Modal.Body>
          <Modal.Footer>
            <Button variant="secondary" onClick={() => this.resetSearch()}>{this.props.t("reset")}</Button>
            <Button type='submit' variant="primary" onClick={(e) => { e.preventDefault(); this.applySearch()}}>{this.props.t("apply")}</Button>
          </Modal.Footer>
        </Form>
      </Modal>
    );
    let confirmModal = (
      <Modal show={this.state.showConfirm} onHide={() => this.setState({ showConfirm: false })}>
        <Modal.Header closeButton={true}>
          <Modal.Title>{this.props.t("bookSeat")}</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <p>{this.props.t("space")}: {this.state.selectedSpace?.name}</p>
          <p>{this.props.t("area")}: {this.getLocationName()}</p>
          <p>{this.props.t("enter")}: {formatter.format(Formatting.convertToFakeUTCDate(new Date(this.state.enter)))}</p>
          <p>{this.props.t("leave")}: {formatter.format(Formatting.convertToFakeUTCDate(new Date(this.state.leave)))}</p>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={() => this.setState({ showConfirm: false })}>
            {this.props.t("cancel")}
          </Button>
          <Button variant="primary" onClick={this.onConfirmBooking}>
            {this.props.t("confirmBooking")}
          </Button>
        </Modal.Footer>
      </Modal>
    );
    let bookings: Booking[] = [];
    if (this.state.selectedSpace) {
      bookings = Booking.createFromRawArray(this.state.selectedSpace.rawBookings);
    }
    const myBooking = (bookings.find(b => b.user.email === RuntimeConfig.INFOS.username));
    let gotoBooking;
    if (myBooking) {
      gotoBooking = (
        <Button variant="secondary" onClick={() => {
          this.setState({ showBookingNames: false })
          this.props.router.push("/bookings#" + myBooking.id)
        }}>
          {this.props.t("gotoBooking")}
        </Button>
      )
    }
    let bookingNamesModal = (
      <Modal show={this.state.showBookingNames} onHide={() => this.setState({ showBookingNames: false })}>
        <Modal.Header closeButton>
          <Modal.Title>{this.state.selectedSpace?.name}</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          {bookings.map(item => {
            return <span key={item.user.id}>{this.renderBookingNameRow(item)}</span>
          })}
        </Modal.Body>
        <Modal.Footer>
          <Button variant="primary" onClick={() => this.setState({ showBookingNames: false })}>
            {this.props.t("ok")}
          </Button>
          {gotoBooking}
        </Modal.Footer>
      </Modal>
    );
    let successModal = (
      <Modal show={this.state.showSuccess} onHide={() => this.setState({ showSuccess: false })} backdrop="static" keyboard={false}>
        <Modal.Header closeButton={false}>
          <Modal.Title>{this.props.t("bookSeat")}</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <p>{this.props.t("bookingConfirmed")}</p>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="primary" onClick={() => this.props.router.push("/bookings")}>
            {this.props.t("myBookings").toString()}
          </Button>
          <Button variant="secondary" onClick={() => {
            this.setState({ showSuccess: false });
            this.refreshPage();
          }}>
            {this.props.t("ok").toString()}
          </Button>
        </Modal.Footer>
      </Modal>
    );
    let errorModal = (
      <Modal show={this.state.showError} onHide={() => this.setState({ showError: false })} backdrop="static" keyboard={false}>
        <Modal.Header closeButton={false}>
          <Modal.Title>{this.props.t("bookSeat")}</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <p>{this.state.errorText}</p>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="primary" onClick={() => this.props.router.push("/bookings")}>
            {this.props.t("myBookings").toString()}
          </Button>
        </Modal.Footer>
      </Modal>
    );

    return (
      <>
        <NavBar />
        {locationInfoModal}
        {searchModal}
        {confirmModal}
        {bookingNamesModal}
        {successModal}
        {errorModal}
        {listOrMap}
        <Loading visible={this.state.loading} />
        {configContainer}
      </>
    )
  }

  refreshPage = () => {
    this.setState({
      loading: true,
    });
    this.loadMap(this.state.locationId).then(() => {
      this.setState({ loading: false });
    });
  }
}

export default withTranslation()(withReadyRouter(Search as any));
