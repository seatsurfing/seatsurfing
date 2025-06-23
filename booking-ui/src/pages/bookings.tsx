import React from "react";
import {
  Ajax,
  AjaxError,
  Booking,
  Formatting,
  RecurringBooking,
} from "seatsurfing-commons";
import Loading from "../components/Loading";
import { Button, Form, ListGroup, Modal } from "react-bootstrap";
import {
  LogIn as IconEnter,
  LogOut as IconLeave,
  MapPin as IconLocation,
  Clock as IconPending,
  RefreshCw as IconRecurring,
} from "react-feather";
import { NextRouter } from "next/router";
import NavBar from "@/components/NavBar";
import withReadyRouter from "@/components/withReadyRouter";
import RuntimeConfig from "@/components/RuntimeConfig";
import ErrorText from "@/types/ErrorText";
import { Loader as IconLoad, Calendar as IconCalendar } from "react-feather";
import { getIcal } from "@/components/Ical";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";

interface State {
  loading: boolean;
  deletingItem: boolean;
  selectedItem: Booking | null;
  cancelSeries: boolean;
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
}

class Bookings extends React.Component<Props, State> {
  data: Booking[];

  constructor(props: any) {
    super(props);
    this.data = [];
    this.state = {
      loading: true,
      deletingItem: false,
      selectedItem: null,
      cancelSeries: false,
    };
  }

  componentDidMount = () => {
    if (!Ajax.CREDENTIALS.accessToken) {
      this.props.router.push("/login");
      return;
    }
    this.loadData();
  };

  loadData = () => {
    Booking.list().then((list) => {
      this.data = list;
      this.setState({ loading: false });
    });
  };

  onItemPress = (item: Booking) => {
    this.setState({
      selectedItem: item,
      cancelSeries: false,
    });
  };

  cancelBooking = async () => {
    if (!this.state.selectedItem) {
      return;
    }
    this.setState({
      deletingItem: true,
    });
    let item: any;
    item = this.state.selectedItem;
    if (this.state.cancelSeries && item.isRecurring()) {
      item = await RecurringBooking.get(item.recurringId);
    }
    item.delete().then(
      () => {
        this.setState(
          {
            selectedItem: null,
            deletingItem: false,
            loading: true,
          },
          this.loadData
        );
      },
      (reason: any) => {
        if (reason instanceof AjaxError && reason.httpStatusCode === 403) {
          window.alert(
            ErrorText.getTextForAppCode(reason.appErrorCode, this.props.t)
          );
        } else {
          window.alert(this.props.t("errorDeleteBooking"));
        }
        this.setState(
          {
            selectedItem: null,
            deletingItem: false,
            loading: true,
          },
          this.loadData
        );
      }
    );
  };

  renderItem = (item: Booking) => {
    let formatter = Formatting.getFormatter();
    if (RuntimeConfig.INFOS.dailyBasisBooking) {
      formatter = Formatting.getFormatterNoTime();
    }
    let pending = <></>;
    if (item.approved === false) {
      pending = (
        <>
          <IconPending className="feather" />
          &nbsp;{this.props.t("approval")}: {this.props.t("pending")}
          <br />
        </>
      );
    }
    let recurringIcon = <></>;
    if (item.isRecurring()) {
      recurringIcon = (
        <IconRecurring className="feather recurring-booking-icon" />
      );
    }
    return (
      <ListGroup.Item
        key={item.id}
        action={true}
        onClick={(e) => {
          e.preventDefault();
          this.onItemPress(item);
        }}
      >
        <h5>{Formatting.getDateOffsetText(item.enter, item.leave)}</h5>
        {recurringIcon}
        <h6 hidden={!item.subject}>{item.subject}</h6>
        <p>
          {pending}
          <IconLocation className="feather" />
          &nbsp;{item.space.location.name}, {item.space.name}
          <br />
          <IconEnter className="feather" />
          &nbsp;{formatter.format(item.enter)}
          <br />
          <IconLeave className="feather" />
          &nbsp;{formatter.format(item.leave)}
        </p>
      </ListGroup.Item>
    );
  };

  render() {
    if (this.state.loading) {
      return <Loading />;
    }
    if (this.data.length === 0) {
      return (
        <>
          <NavBar />
          <div className="container-signin">
            <Form className="form-signin">
              <p>{this.props.t("noBookings")}</p>
            </Form>
          </div>
        </>
      );
    }
    let formatter = Formatting.getFormatter();
    if (RuntimeConfig.INFOS.dailyBasisBooking) {
      formatter = Formatting.getFormatterNoTime();
    }
    return (
      <>
        <NavBar />
        <div className="container-signin">
          <Form className="form-signin">
            <ListGroup>
              {this.data.map((item) => this.renderItem(item))}
            </ListGroup>
          </Form>
        </div>
        <Modal
          show={this.state.selectedItem != null}
          onHide={() => this.setState({ selectedItem: null })}
        >
          <Modal.Header closeButton>
            <Modal.Title>{this.props.t("cancelBooking")}</Modal.Title>
          </Modal.Header>
          <Modal.Body>
            <h6 hidden={!this.state.selectedItem?.subject}>
              {this.state.selectedItem?.subject}
            </h6>
            <p dangerouslySetInnerHTML={{__html: this.props.t("confirmCancelBooking", {
                enter: formatter.format(this.state.selectedItem?.enter),
              })}}>
            </p>
            <p hidden={!this.state.selectedItem?.isRecurring()}>
              <Form.Check
                type="checkbox"
                id="cancelAllUpcomingBookings"
                onChange={(e) =>
                  this.setState({ cancelSeries: e.target.checked })
                }
                checked={this.state.cancelSeries}
                label={this.props.t("cancelAllUpcomingBookings")}
              />
            </p>
          </Modal.Body>
          <Modal.Footer>
            <Button
              variant="secondary"
              onClick={() => this.setState({ selectedItem: null })}
              disabled={this.state.deletingItem}
            >
              {this.props.t("back")}
            </Button>
            <Button
              variant="secondary"
              onClick={() =>
                getIcal(
                  this.state.selectedItem ? this.state.selectedItem.id : ""
                )
              }
            >
              <IconCalendar
                className="feather"
                style={{ marginRight: "5px" }}
              />{" "}
              Event
            </Button>
            <Button
              variant="danger"
              onClick={() => this.cancelBooking()}
              disabled={this.state.deletingItem}
            >
              {this.props.t("cancelBooking")}
              {this.state.deletingItem ? (
                <IconLoad
                  className="feather loader"
                  style={{ marginLeft: "5px" }}
                />
              ) : (
                <></>
              )}
            </Button>
          </Modal.Footer>
        </Modal>
      </>
    );
  }
}

export default withTranslation(withReadyRouter(Bookings as any));
