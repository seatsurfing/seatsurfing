import React from "react";
import Loading from "../components/Loading";
import { Button, Form, InputGroup, ListGroup, Modal } from "react-bootstrap";
import {
  LogIn as IconEnter,
  LogOut as IconLeave,
  MapPin as IconLocation,
} from "react-feather";
import { NextRouter } from "next/router";
import NavBar from "@/components/NavBar";
import withReadyRouter from "@/components/withReadyRouter";
import RuntimeConfig from "@/components/RuntimeConfig";
import AlertModal from "@/components/AlertModal";
import UserSearchTypeahead from "@/components/UserSearchTypeahead";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import Buddy, { BuddySearchResult } from "@/types/Buddy";
import Formatting from "@/util/Formatting";

interface State {
  loading: boolean;
  selectedItem: Buddy | null;
  selectedBuddy: BuddySearchResult | null;
  alertMessage: string | null;
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
}

class Buddies extends React.Component<Props, State> {
  data: Buddy[];
  typeahead: any = null;

  constructor(props: any) {
    super(props);
    this.data = [];
    this.state = {
      loading: true,
      selectedItem: null,
      selectedBuddy: null,
      alertMessage: null,
    };
  }

  componentDidMount = () => {
    if (!RuntimeConfig.INFOS.showNames || RuntimeConfig.INFOS.disableBuddies) {
      this.props.router.push("/search");
      return;
    }
    this.loadData();
  };

  loadData = () => {
    Buddy.list().then((list) => {
      this.data = list;
      this.setState({ loading: false });
    });
  };

  onItemPress = (item: Buddy) => {
    this.setState({ selectedItem: item });
  };

  removeBuddy = (_item: Buddy | null) => {
    this.setState({
      loading: true,
    });
    this.state.selectedItem?.delete().then(() => {
      this.setState(
        {
          selectedItem: null,
        },
        this.loadData,
      );
    });
  };

  searchBuddyCandidates = (query: string): Promise<BuddySearchResult[]> => {
    return Buddy.search(query).then((users) => {
      const existingIds = this.data.map((item) => item.buddy.id);
      return users.filter((user) => !existingIds.includes(user.id));
    });
  };

  onSearchSelected = (selected: BuddySearchResult[]) => {
    this.setState({ selectedBuddy: selected.length > 0 ? selected[0] : null });
  };

  addBuddy = () => {
    const { selectedBuddy } = this.state;

    if (!selectedBuddy) {
      return;
    }

    const buddy = new Buddy();
    buddy.buddy.id = selectedBuddy.id;
    buddy.buddy.email = selectedBuddy.email;
    buddy
      .save()
      .then(() => {
        if (this.typeahead !== null) {
          this.typeahead.clear();
        }
        this.setState({ selectedBuddy: null });
        this.loadData();
      })
      .catch(() => {
        this.setState({ alertMessage: this.props.t("userNotFound") });
      });
  };

  renderAddBuddy() {
    return (
      <ListGroup.Item key={"add-buddy"} style={{ minWidth: "300px" }}>
        <Form.Group className="grid-item">
          <InputGroup>
            <UserSearchTypeahead
              t={this.props.t}
              id="search-buddy"
              multiple={false}
              searchFn={this.searchBuddyCandidates}
              onChange={(selected: any) => this.onSearchSelected(selected)}
              ref={(ref: any) => {
                this.typeahead = ref;
              }}
            />
            <Button
              variant="primary"
              type="submit"
              onClick={(e) => {
                e.preventDefault();
                this.addBuddy();
              }}
              disabled={!this.state.selectedBuddy}
            >
              {this.props.t("addBuddy")}
            </Button>
          </InputGroup>
        </Form.Group>
      </ListGroup.Item>
    );
  }

  renderItem = (item: Buddy) => {
    const {
      id,
      buddy: { email, firstBooking },
    } = item;
    const formatter = Formatting.getBookingDateFormatter();
    return (
      <ListGroup.Item
        key={id}
        style={{
          minWidth: "300px",
          maxWidth: "100%",
          overflowWrap: "break-word",
        }}
      >
        <h5 style={{ overflowWrap: "break-word" }}>{email}</h5>
        {(firstBooking == null && <p>{this.props.t("noBooking")}</p>) || (
          <p>
            <IconLocation className="feather" />
            &nbsp;{firstBooking!.room}, {firstBooking!.desk}
            <br />
            <IconEnter className="feather" />
            &nbsp;{formatter.format(new Date(firstBooking!.enter))}
            <br />
            <IconLeave className="feather" />
            &nbsp;{formatter.format(new Date(firstBooking!.leave))}
          </p>
        )}
        <Button variant="danger" onClick={() => this.onItemPress(item)}>
          {this.props.t("removeBuddy")}
        </Button>
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
              <p>{this.props.t("noBuddies")}</p>
              {this.renderAddBuddy()}
            </Form>
          </div>
          <AlertModal
            show={this.state.alertMessage !== null}
            message={this.state.alertMessage || ""}
            onConfirm={() => this.setState({ alertMessage: null })}
          />
        </>
      );
    }
    return (
      <>
        <NavBar />
        <div className="container-signin">
          <Form className="form-signin">
            <ListGroup>
              {this.data.map((item) => this.renderItem(item))}
              {this.renderAddBuddy()}
            </ListGroup>
          </Form>
        </div>
        <Modal
          show={this.state.selectedItem != null}
          onHide={() => this.setState({ selectedItem: null })}
        >
          <Modal.Header closeButton>
            <Modal.Title>{this.props.t("removeBuddy")}</Modal.Title>
          </Modal.Header>
          <Modal.Body>
            <p>{this.props.t("confirmRemoveBuddy")}</p>
          </Modal.Body>
          <Modal.Footer>
            <Button
              variant="secondary"
              onClick={() => this.setState({ selectedItem: null })}
            >
              {this.props.t("back")}
            </Button>
            <Button
              variant="danger"
              onClick={() => this.removeBuddy(this.state.selectedItem)}
            >
              {this.props.t("removeBuddy")}
            </Button>
          </Modal.Footer>
        </Modal>
        <AlertModal
          show={this.state.alertMessage !== null}
          message={this.state.alertMessage || ""}
          onConfirm={() => this.setState({ alertMessage: null })}
        />
      </>
    );
  }
}

export default withTranslation(withReadyRouter(Buddies as any));
