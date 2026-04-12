import React from "react";
import { Form, Col, Row, Button, Alert } from "react-bootstrap";
import {
  ChevronLeft as IconBack,
  Save as IconSave,
  Trash2 as IconDelete,
} from "react-feather";
import { AsyncTypeahead } from "react-bootstrap-typeahead";
import { NextRouter } from "next/router";
import Link from "next/link";
import Loading from "@/components/Loading";
import OrgSettings from "@/types/Settings";
import withReadyRouter from "@/components/withReadyRouter";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import Space from "@/types/Space";
import User from "@/types/User";
import Location from "@/types/Location";
import Booking from "@/types/Booking";
import Ajax from "@/util/Ajax";
import UserPreference from "@/types/UserPreference";
import Formatting from "@/util/Formatting";
import FullLayout from "@/components/FullLayout";
import RedirectUtil from "@/util/RedirectUtil";
import DateUtil from "@/util/DateUtil";
import DateTimePicker from "@/components/DateTimePicker";
import RuntimeConfig from "@/components/RuntimeConfig";
import RendererUtils from "@/util/RendererUtils";
import ProfilePicture from "@/components/ProfilePicture";
import Search, { SearchOptions } from "@/types/Search";

interface State {
  loading: boolean;
  saved: boolean;
  error: boolean;
  wasCreated: boolean;
  goBack: boolean;
  enter: Date;
  leave: Date;
  location: Location;
  space: Space;
  user: User;
  selectedUserEmail: string;
  selectedLocationId: string;
  selectedSpaceId: string;
  users: User[];
  locations: Location[];
  spaces: Space[];
  isDisabledLocation: boolean;
  isDisabledSpace: boolean;
  canSearch: boolean;
  canSearchHint: string;
  canSave: boolean;
  canEdit: boolean;
  prefEnterTime: number;
  prefWorkdayStart: number;
  prefWorkdayEnd: number;
  prefWorkdays: number[];
  prefLocationId: string;
  selfEmail: string;
  subject: string;
  typeaheadOptions: any[];
  typeaheadLoading: boolean;
  typeaheadSelected: [{ email: string }];
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
}

class EditBooking extends React.Component<Props, State> {
  static PreferenceEnterTimeNow: number = 1;
  static PreferenceEnterTimeNextDay: number = 2;
  static PreferenceEnterTimeNextWorkday: number = 3;
  entity: Booking = new Booking();
  authProviders: { [key: string]: string } = {};
  dailyBasisBooking: boolean;
  noAdminRestrictions: boolean;
  maxBookingsPerUser: number;
  maxDaysInAdvance: number;
  maxBookingDurationHours: number;
  minBookingDurationHours: number;
  isNewBooking: boolean;
  enterChangeTimer: number | undefined;
  leaveChangeTimer: number | undefined;
  curBookingCount: number = 0;
  typeahead: any = null;

  constructor(props: any) {
    super(props);
    this.dailyBasisBooking = false;
    this.noAdminRestrictions = false;
    this.maxBookingsPerUser = 0;
    this.maxBookingDurationHours = 0;
    this.minBookingDurationHours = 0;
    this.maxDaysInAdvance = 0;
    this.isNewBooking = false;
    this.enterChangeTimer = undefined;
    this.leaveChangeTimer = undefined;
    this.state = {
      loading: true,
      saved: false,
      error: false,
      wasCreated: false,
      goBack: false,
      enter: new Date(),
      leave: new Date(),
      location: new Location(),
      space: new Space(),
      user: new User(),
      selectedUserEmail: "",
      selectedLocationId: "",
      selectedSpaceId: "",
      users: [],
      locations: [],
      spaces: [],
      isDisabledLocation: true,
      isDisabledSpace: true,
      canSearch: false,
      canSearchHint: "",
      canSave: false,
      canEdit: false,
      prefEnterTime: 0,
      prefWorkdayStart: 0,
      prefWorkdayEnd: 0,
      prefWorkdays: [],
      prefLocationId: "",
      selfEmail: "",
      subject: "",
      typeaheadOptions: [],
      typeaheadLoading: false,
      typeaheadSelected: [{ email: "" }],
    };
  }

  componentDidMount = () => {
    if (!Ajax.hasAccessToken()) {
      RedirectUtil.toLogin(this.props.router);
      return;
    }
    const promises = [
      this.loadData(),
      this.loadSettings(),
      this.loadLocations(),
      this.loadPreferences() /* currently same as me */,
      this.loadSelf(),
    ];
    Promise.all(promises).then(() => {
      this.setState({ loading: false });
      this.initDates();
    });
  };

  loadData = () => {
    const { id } = this.props.router.query;
    if (id && typeof id === "string") {
      if (id !== "add") {
        return Booking.get(id).then((booking) => {
          this.entity = booking;
          const canSave = !DateUtil.isInPast(this.entity.leave);
          this.setState({
            enter: DateUtil.convertToUTC(this.entity.enter),
            leave: DateUtil.convertToUTC(this.entity.leave),
            selectedLocationId: this.entity.space.locationId,
            selectedSpaceId: this.entity.space.id,
            selectedUserEmail: this.entity.user.email,
            isDisabledLocation: false,
            isDisabledSpace: false,
            canSave: canSave,
            canEdit: canSave,
            subject: this.entity.subject,
            // loading: false,
            typeaheadSelected: this.entity.user.email
              ? [{ email: this.entity.user.email }]
              : [{ email: "" }],
          });
          this.loadSpaces(
            this.entity.space.locationId,
            this.entity.enter,
            this.entity.leave,
          );
          this.isNewBooking = false;
        });
      } else {
        // add
        this.isNewBooking = true;
        this.setState({
          isDisabledLocation: false,
          isDisabledSpace: false,
          enter: new Date(),
          canSave: true,
          canEdit: true,
          // loading: false,
        });
      }
    } else {
      // no id
    }
  };

  initDates = () => {
    if (!this.isNewBooking) return;
    const enter = new Date();
    if (this.state.prefEnterTime === EditBooking.PreferenceEnterTimeNow) {
      enter.setHours(enter.getHours() + 1, 0, 0);
      if (enter.getHours() < this.state.prefWorkdayStart) {
        enter.setHours(this.state.prefWorkdayStart, 0, 0, 0);
      }
      if (enter.getHours() >= this.state.prefWorkdayEnd) {
        enter.setDate(enter.getDate() + 1);
        enter.setHours(this.state.prefWorkdayStart, 0, 0, 0);
      }
    } else if (
      this.state.prefEnterTime === EditBooking.PreferenceEnterTimeNextDay
    ) {
      enter.setDate(enter.getDate() + 1);
      enter.setHours(this.state.prefWorkdayStart, 0, 0, 0);
    } else if (
      this.state.prefEnterTime === EditBooking.PreferenceEnterTimeNextWorkday
    ) {
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

    const leave = new Date(enter);
    leave.setHours(this.state.prefWorkdayEnd, 0, 0);

    if (this.dailyBasisBooking) {
      enter.setHours(0, 0, 0, 0);
      leave.setHours(23, 59, 59, 0);
    }
    this.setState({
      enter,
      leave,
    });
  };

  loadSpaces = async (
    selectedLocationId: string,
    enter: Date,
    leave: Date,
  ): Promise<void> => {
    return Space.listAvailability(selectedLocationId, enter, leave).then(
      (list) => {
        this.setState({
          spaces: list,
          isDisabledSpace: false,
        });
      },
    );
  };

  loadSelf = async (): Promise<void> => {
    User.getSelf().then((user) => {
      this.setState({
        selfEmail: user.email,
      });
      if (!this.state.typeaheadSelected[0]?.email) {
        this.setState({
          typeaheadSelected: [{ email: user.email }],
        });
      }
    });
  };

  loadSettings = async (): Promise<void> => {
    return OrgSettings.list().then((settings) => {
      settings.forEach((s) => {
        if (s.name === "daily_basis_booking") {
          this.dailyBasisBooking = s.value === "1";
        }
        if (s.name === "no_admin_restrictions") {
          this.noAdminRestrictions = s.value === "1";
        }
        if (s.name === "max_bookings_per_user") {
          this.maxBookingsPerUser = window.parseInt(s.value);
        }
        if (s.name === "max_days_in_advance") {
          this.maxDaysInAdvance = window.parseInt(s.value);
        }
        if (s.name === "max_booking_duration_hours") {
          this.maxBookingDurationHours = window.parseInt(s.value);
        }
        if (s.name === "min_booking_duration_hours") {
          this.minBookingDurationHours = window.parseInt(s.value);
        }
      });
    });
  };

  loadPreferences = async (): Promise<void> => {
    let self = this;
    return new Promise<void>(function (resolve, reject) {
      UserPreference.list()
        .then((list) => {
          let state: any = {};
          list.forEach((s) => {
            if (typeof window !== "undefined") {
              if (s.name === UserPreference.PREF_ENTER_TIME)
                state.prefEnterTime = window.parseInt(s.value);
              if (s.name === UserPreference.PREF_WORKDAY_START)
                state.prefWorkdayStart = window.parseInt(s.value);
              if (s.name === UserPreference.PREF_WORKDAY_END)
                state.prefWorkdayEnd = window.parseInt(s.value);
              if (s.name === UserPreference.PREF_WORKDAYS)
                state.prefWorkdays = s.value
                  .split(",")
                  .map((val) => window.parseInt(val));
            }
            if (s.name === UserPreference.PREF_LOCATION_ID)
              state.prefLocationId = s.value;
            if (s.name === UserPreference.PREF_BOOKED_COLOR)
              state.prefBookedColor = s.value;
            if (s.name === UserPreference.PREF_NOT_BOOKED_COLOR)
              state.prefNotBookedColor = s.value;
            if (s.name === UserPreference.PREF_SELF_BOOKED_COLOR)
              state.prefSelfBookedColor = s.value;
            if (s.name === UserPreference.PREF_BUDDY_BOOKED_COLOR)
              state.prefBuddyBookedColor = s.value;
          });
          if (self.dailyBasisBooking) {
            state.prefWorkdayStart = 0;
            state.prefWorkdayEnd = 23;
          }
          self.setState(
            {
              ...state,
            },
            () => resolve(),
          );
        })
        .catch((e) => reject(e));
    });
  };

  loadLocations = async (): Promise<void> => {
    return Location.list().then((list) => {
      this.setState({ locations: list });
      // this.setState({ loading: false });
    });
  };

  //TODO: modify to init according to selcted user
  // initCurrentBookingCount = () => {
  //     Booking.list().then(list => {
  //         this.curBookingCount = list.length;
  //         this.updateCanSearch();
  //     });
  // }

  onSubmit = (e: any) => {
    e.preventDefault();
    this.setState({
      error: false,
      saved: false,
      canSearchHint: "",
    });

    if (this.dailyBasisBooking) {
      let enter = new Date();
      enter = this.state.enter;
      enter.setHours(0, 0, 0, 0);

      let leave = new Date();
      leave = this.state.leave;
      leave.setHours(23, 59, 59, 0);

      this.setState({
        enter: enter,
        leave: leave,
      });
    } else {
      const enter = this.state.enter;
      const leave = this.state.leave;
      enter.setSeconds(0);
      enter.setMilliseconds(0);
      leave.setSeconds(0);
      leave.setMilliseconds(0);
      this.setState({
        enter,
        leave,
      });
    }

    if (this.isNewBooking) {
      let user = this.state.selectedUserEmail;
      if (!user) {
        user = this.state.selfEmail;
      }
      this.entity.enter = this.state.enter;
      this.entity.leave = this.state.leave;
      this.entity.space.id = this.state.selectedSpaceId;
      this.entity.user.email = user;
      this.entity.subject = this.state.subject;
      this.entity
        .save()
        .then(() => {
          this.isNewBooking = false;
          this.props.router.push("/admin/bookings/" + this.entity.id);
          this.setState({
            saved: true,
            isDisabledLocation: false,
            isDisabledSpace: false,
            wasCreated: true,
            selectedUserEmail: user,
          });
        })
        .catch(() => {
          this.setState({
            error: true,
            saved: false,
            wasCreated: true,
          });
        });
    } else {
      if (!this.state.selectedUserEmail) {
        this.setState({
          canSearchHint: "errorUserRequired",
          saved: false,
        });
        return;
      }
      this.entity.enter = this.state.enter;
      this.entity.leave = this.state.leave;
      this.entity.space.id = this.state.selectedSpaceId;
      this.entity.user.email = this.state.selectedUserEmail;
      this.entity.subject = this.state.subject;
      this.entity
        .save()
        .then(() => {
          this.setState({
            saved: true,
            wasCreated: false,
          });
        })
        .catch(() => {
          this.setState({
            error: true,
            saved: false,
            wasCreated: false,
          });
        });
    }
  };

  deleteItem = () => {
    const formatter = Formatting.getBookingDateFormatter();
    const confirmMessage = this.props.t("confirmCancelBooking", {
      enter: formatter.format(this.entity.enter),
    });
    if (!window.confirm(RendererUtils.decodeHtmlEntities(confirmMessage))) {
      return;
    }
    this.entity.delete().then(() => {
      this.setState({ goBack: true });
    });
  };

  updateCanSearch = async () => {
    let res = true;
    let hint = "";
    if (this.curBookingCount >= this.maxBookingsPerUser) {
      res = false;
      hint = this.props.t("errorBookingLimit", {
        num: this.maxBookingsPerUser,
      });
    }
    const todayMorning = DateUtil.convertToUTC(new Date());
    todayMorning.setHours(0, 0, 0);
    const enterTime = new Date(this.state.enter);
    if (this.dailyBasisBooking) {
      enterTime.setHours(23, 59, 59);
    }
    if (enterTime.getTime() < todayMorning.getTime()) {
      res = false;
      hint = this.props.t("errorEnterFuture");
    }
    if (this.state.leave.getTime() <= this.state.enter.getTime()) {
      res = false;
      hint = this.props.t("errorLeaveAfterEnter");
    }
    if (this.state.leave.getTime() < new Date().getTime()) {
      res = false;
      hint = this.props.t("errorLeavePast");
    }

    const bookingAdvanceDays = Math.floor(
      (this.state.enter.getTime() - new Date().getTime()) / DateUtil.MS_PER_DAY,
    );
    if (
      bookingAdvanceDays > this.maxDaysInAdvance &&
      !this.noAdminRestrictions
    ) {
      res = false;
      hint = this.props.t("errorDaysAdvance", { num: this.maxDaysInAdvance });
    }
    let bookingDurationHours =
      Math.floor(
        (this.state.leave.getTime() - this.state.enter.getTime()) /
          DateUtil.MS_PER_MINUTE,
      ) / 60;
    if (
      bookingDurationHours > this.maxBookingDurationHours &&
      !this.noAdminRestrictions
    ) {
      res = false;
      hint = this.props.t("errorMaxBookingDuration", {
        num: this.maxBookingDurationHours,
      });
    }
    if (
      bookingDurationHours < this.minBookingDurationHours &&
      !this.noAdminRestrictions
    ) {
      res = false;
      hint = this.props.t("errorMinBookingDuration", {
        num: this.minBookingDurationHours,
      });
    }
    const self = this;
    return new Promise<void>(function (resolve, reject) {
      self.setState(
        {
          canSearch: res,
          canSearchHint: hint,
        },
        () => resolve(),
      );
    });
  };

  setEnterDate = (value: Date | [Date | null, Date | null]) => {
    const dateChangedCb = () => {
      this.updateCanSearch().then(() => {
        if (!this.state.canSearch) {
          this.setState({ loading: false });
        } else {
          // let promises = [
          //     this.initCurrentBookingCount(),
          //     this.loadSpaces(this.state.locationId),
          // ];
          // Promise.all(promises).then(() => {
          //     this.setState({ loading: false });
          // });
        }
      });
    };
    const performChange = () => {
      const enter = value instanceof Date ? value : value[0];
      if (enter == null) {
        return;
      }
      const leave = new Date(enter);
      leave.setHours(leave.getHours() + 1);
      if (this.dailyBasisBooking) {
        enter.setHours(0, 0, 0);
        leave.setHours(23, 59, 59);
      }
      this.setState(
        {
          enter: enter,
          leave: leave,
          isDisabledLocation: false,
          isDisabledSpace: true,
        },
        () => dateChangedCb(),
      );

      if (this.state.selectedLocationId) {
        this.loadSpaces(this.state.selectedLocationId, enter, leave);
      }
    };
    window.clearTimeout(this.leaveChangeTimer);
    this.leaveChangeTimer = window.setTimeout(performChange, 1000);
    return true;
  };

  setLeaveDate = (value: Date | [Date | null, Date | null]) => {
    let dateChangedCb = () => {
      //TODO: check for parameters *maxBookingDur ...

      this.updateCanSearch().then(() => {
        if (!this.state.canSearch) {
          this.setState({ loading: false });
        } else {
          // let promises = [
          //     this.initCurrentBookingCount(),
          //     this.loadSpaces(this.state.locationId),
          // ];
          // Promise.all(promises).then(() => {
          //     this.setState({ loading: false });
          // });
        }
      });
    };
    const performChange = () => {
      let date = value instanceof Date ? value : value[0];
      if (date == null) {
        return;
      }
      if (this.dailyBasisBooking) {
        date.setHours(23, 59, 59);
      }
      this.setState(
        {
          leave: date,
          isDisabledLocation: false,
          isDisabledSpace: true,
        },
        () => dateChangedCb(),
      );
      if (this.state.selectedLocationId) {
        this.loadSpaces(this.state.selectedLocationId, this.state.enter, date);
      }
    };
    window.clearTimeout(this.leaveChangeTimer);
    this.leaveChangeTimer = window.setTimeout(performChange, 1000);
  };

  getBookersList = (bookings: Booking[]) => {
    if (!bookings.length) return "";
    let str = "";
    bookings.forEach((b) => {
      str += (str ? ", " : "") + b.user.email;
    });
    return str;
  };

  userOnChange = (val: string) => {
    this.setState({ selectedUserEmail: val });
    /* IMPROVEMENT: LoadPreferences from selected user
        let promises = [
            this.loadPreferences()
          ];
          Promise.all(promises).then(() => {
            this.initDates()
          });
        */
  };

  getSelectedSpace = () => {
    if (this.state.selectedSpaceId) {
      return this.state.spaces.find(
        (space) => space.id === this.state.selectedSpaceId,
      );
    }
    return undefined;
  };

  filterSearch = () => {
    return true;
  };

  onSearchSelected = (selected: any) => {
    this.setState({
      selectedUserEmail: selected[0]?.email,
    });
  };

  handleSearch = (query: string) => {
    this.setState({ typeaheadLoading: true });
    const options = new SearchOptions();
    options.includeUsers = true;
    options.keyword = query ? query : "";
    Search.search(options).then((res) => {
      this.setState({
        typeaheadOptions: res.users,
        typeaheadLoading: false,
      });
    });
  };

  render() {
    if (this.state.goBack) {
      this.props.router.push("/admin/bookings");
      return <></>;
    }

    let hint = <></>;
    if (!this.state.canSearch && this.state.canSearchHint) {
      hint = (
        <Form.Group as={Row} className="margin-top-10">
          <Col xs="2"></Col>
          <Col xs="10">
            <div className="invalid-search-config">
              {this.state.canSearchHint}
            </div>
          </Col>
        </Form.Group>
      );
    }

    const enterDatePicker = (
      <DateTimePicker
        id="booking-enter"
        value={this.state.enter}
        onChange={(value: Date | null) => {
          if (value != null) this.setEnterDate(value);
        }}
        required={true}
        disabled={!this.state.canEdit}
        enableTime={!this.dailyBasisBooking}
      />
    );
    const leaveDatePicker = (
      <DateTimePicker
        id="booking-leave"
        value={this.state.leave}
        onChange={(value: Date | null) => {
          if (value != null) this.setLeaveDate(value);
        }}
        required={true}
        disabled={!this.state.canEdit}
        enableTime={!this.dailyBasisBooking}
      />
    );

    const backButton = (
      <Link href="/admin/bookings" className="btn btn-sm btn-outline-secondary">
        <IconBack className="feather" /> {this.props.t("back")}
      </Link>
    );
    let buttons = backButton;

    if (this.state.loading) {
      return (
        <FullLayout
          headline={this.props.t(
            this.isNewBooking ? "newBooking" : "editBooking",
          )}
          buttons={buttons}
        >
          <Loading />
        </FullLayout>
      );
    }

    if (this.state.saved) {
      hint = (
        <Alert variant="success">
          {this.props.t(
            this.state.wasCreated ? "entryCreated" : "entryUpdated",
          )}
        </Alert>
      );
    } else if (this.state.canSearchHint) {
      hint = (
        <Alert variant="danger">{this.props.t(this.state.canSearchHint)}</Alert>
      );
    } else if (this.state.error) {
      hint = <Alert variant="danger">{this.props.t("errorSave")}</Alert>;
    }

    const buttonDelete = (
      <Button
        className="btn-sm"
        variant="outline-secondary"
        onClick={this.deleteItem}
        disabled={!this.state.canEdit}
      >
        <IconDelete className="feather" /> {this.props.t("delete")}
      </Button>
    );
    const buttonSave = (
      <Button
        disabled={!(this.state.canSave && this.state.canEdit)}
        className="btn-sm"
        variant="outline-secondary"
        type="submit"
        form="form"
      >
        <IconSave className="feather" /> {this.props.t("save")}
      </Button>
    );
    if (this.entity.id) {
      buttons = (
        <>
          {backButton} {buttonDelete} {buttonSave}
        </>
      );
    } else {
      buttons = (
        <>
          {backButton} {buttonSave}
        </>
      );
    }
    let userField = <></>;
    if (this.state.canEdit) {
      userField = (
        <AsyncTypeahead
          selected={this.state.typeaheadSelected}
          filterBy={this.filterSearch}
          isLoading={this.state.typeaheadLoading}
          inputProps={{ id: "booking-user" }}
          labelKey="email"
          multiple={false}
          minLength={3}
          onChange={(selected) => {
            this.setState({
              typeaheadSelected: selected as [{ email: string }],
            });
            this.onSearchSelected(selected);
          }}
          onSearch={this.handleSearch}
          options={this.state.typeaheadOptions}
          placeholder={this.props.t("searchForUser")}
          ref={(ref: any) => {
            this.typeahead = ref;
          }}
          renderMenuItemChildren={(option: any) => (
            <div className="d-flex">
              <ProfilePicture width={24} height={24} />
              <span style={{ marginLeft: "10px" }}>
                {option.email}
                {RendererUtils.preAndSuffixIfDefined(
                  RendererUtils.fullname(option.firstname, option.lastname),
                  " (",
                  ")",
                )}{" "}
              </span>
            </div>
          )}
        />
      );
    } else {
      userField = (
        <Form.Control
          id="booking-user"
          type="text"
          disabled
          value={this.state.selectedUserEmail}
        />
      );
    }

    const getDisabledInfo = (enabled: boolean): string => {
      return enabled ? "" : ` - ${this.props.t("disabled")}`;
    };

    return (
      <FullLayout
        headline={this.props.t(
          this.isNewBooking ? "newBooking" : "editBooking",
        )}
        buttons={buttons}
      >
        <Form onSubmit={this.onSubmit} id="form">
          {hint}

          <Form.Group as={Row}>
            <Form.Label column sm="2" htmlFor="booking-user">
              {this.props.t("user")}
            </Form.Label>
            <Col sm="4">{userField}</Col>
          </Form.Group>

          <Form.Group as={Row}>
            <Form.Label column sm="2" htmlFor="booking-enter">
              {this.props.t("enter")}
            </Form.Label>
            <Col sm="4">{enterDatePicker}</Col>
          </Form.Group>

          <Form.Group as={Row}>
            <Form.Label column sm="2" htmlFor="booking-leave">
              {this.props.t("leave")}
            </Form.Label>
            <Col sm="4">{leaveDatePicker}</Col>
          </Form.Group>

          <Form.Group as={Row}>
            <Form.Label column sm="2" htmlFor="booking-location">
              {this.props.t("area")}
            </Form.Label>
            <Col sm="4">
              <Form.Select
                id="booking-location"
                disabled={this.state.isDisabledLocation || !this.state.canEdit}
                required={true}
                value={this.state.selectedLocationId}
                onChange={(e: any) => {
                  this.setState({
                    selectedLocationId: e.target.value,
                    isDisabledSpace: false,
                    selectedSpaceId: "",
                  });
                  this.loadSpaces(
                    e.target.value,
                    this.state.enter,
                    this.state.leave,
                  );
                }}
              >
                <option disabled={true} value="">
                  -
                </option>
                {this.state.locations.map(
                  (location: {
                    name: string | undefined;
                    id: string | undefined;
                    enabled: boolean;
                  }) => (
                    <option key={location.id} value={location.id}>
                      {location.name}
                      {getDisabledInfo(location.enabled)}
                    </option>
                  ),
                )}
              </Form.Select>
            </Col>
          </Form.Group>

          <Form.Group as={Row}>
            <Form.Label column sm="2" htmlFor="booking-space">
              {this.props.t("space")}
            </Form.Label>
            <Col sm="4">
              <Form.Select
                id="booking-space"
                disabled={this.state.isDisabledSpace || !this.state.canEdit}
                required={true}
                value={this.state.selectedSpaceId}
                onChange={(e: any) =>
                  this.setState({ selectedSpaceId: e.target.value })
                }
              >
                <option disabled={true} value="">
                  -
                </option>
                {this.state.spaces.map(
                  (space: {
                    id: string | undefined;
                    name: string | null | undefined;
                    available: boolean;
                    rawBookings: any[];
                    enabled: boolean;
                  }) => {
                    const bookings = Booking.createFromRawArray(
                      space.rawBookings,
                    );
                    if (space.available) {
                      return (
                        <option key={space.id} value={space.id}>
                          {space.name}
                          {getDisabledInfo(space.enabled)}
                        </option>
                      );
                    } else {
                      const booker = RendererUtils.preAndSuffixIfDefined(
                        this.getBookersList(bookings),
                        " (",
                        ")",
                      );
                      return (
                        <option key={space.id} disabled value={space.id}>
                          {space.name}
                          {booker}
                          {getDisabledInfo(space.enabled)}
                        </option>
                      );
                    }
                  },
                )}
              </Form.Select>
            </Col>
          </Form.Group>

          <Form.Group as={Row}>
            <Form.Label column sm="2" htmlFor="booking-subject">
              {this.props.t("subject")}
            </Form.Label>
            <Col sm="4">
              <Form.Control
                id="booking-subject"
                type="subject"
                value={this.state.subject}
                minLength={this.getSelectedSpace()?.requireSubject ? 3 : 0}
                disabled={!this.state.canEdit}
                onChange={(e: any) =>
                  this.setState({ subject: e.target.value })
                }
                required={this.getSelectedSpace()?.requireSubject}
              />
            </Col>
          </Form.Group>
        </Form>
      </FullLayout>
    );
  }
}

export default withTranslation(withReadyRouter(EditBooking as any));
