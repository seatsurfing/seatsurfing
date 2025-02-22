import React from 'react';
import { Ajax, Buddy, User, Formatting } from 'seatsurfing-commons';
import Loading from '../components/Loading';
import { Button, Form, ListGroup, Modal } from 'react-bootstrap';
import { LogIn as IconEnter, LogOut as IconLeave, MapPin as IconLocation } from 'react-feather';
import { WithTranslation, withTranslation } from 'next-i18next';
import { NextRouter } from 'next/router';
import NavBar from '@/components/NavBar';
import withReadyRouter from '@/components/withReadyRouter';
import RuntimeConfig from '@/components/RuntimeConfig';

interface State {
  loading: boolean
  selectedItem: Buddy | null
  email: string
}

interface Props extends WithTranslation {
  router: NextRouter
}

class Buddies extends React.Component<Props, State> {
  data: Buddy[];

  constructor(props: any) {
    super(props);
    this.data = [];
    this.state = {
      loading: true,
      selectedItem: null,
      email: ''
    };
  }

  componentDidMount = () => {
    if (!Ajax.CREDENTIALS.accessToken) {
      this.props.router.push("/login");
      return;
    }
    if (!RuntimeConfig.INFOS.showNames || RuntimeConfig.INFOS.disableBuddies) {
      this.props.router.push("/search");
      return;
    }
    this.loadData();
  }

  loadData = () => {
    Buddy.list().then(list => {
      this.data = list;
      this.setState({ loading: false });
    });
  }

  onItemPress = (item: Buddy) => {
    this.setState({ selectedItem: item });
  }

  removeBuddy = (item: Buddy | null) => {
    this.setState({
      loading: true
    });
    this.state.selectedItem?.delete().then(() => {
      this.setState({
        selectedItem: null,
      }, this.loadData);
    });
  }

  addBuddy = () => {
    const { email } = this.state;

    if (!email) {
      return;
    }

    if (this.data.find(item => item.buddy.email.toLowerCase() === email.toLowerCase())) {
      this.setState({ email: '' });
      return;
    }

    User.getByEmail(email).then((user: User) => {
      const buddy = new Buddy();
      buddy.buddy = user;
      buddy.save().then(() => {
        this.setState({ email: '' });
        this.loadData()
      });
    }).catch(() => {
      alert(this.props.t("userNotFound"));
    })
  };

  renderAddBuddy() {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    const isValidEmail = emailRegex.test(this.state.email);

    return (
      <ListGroup.Item key={'add-buddy'} style={{ minWidth: "300px" }}>
        <Form.Group className='grid-item'>
          <Form.Control
            type="email"
            placeholder="Email..."
            value={this.state.email}
            onChange={(e) => this.setState({ email: e.target.value })}
            style={{ marginBottom: '10px', padding: '10px' }}
            isInvalid={!isValidEmail && this.state.email !== ''}
          />
          <Form.Control.Feedback type='invalid'>
            {this.props.t("validEmailRequired")}
          </Form.Control.Feedback>
          <Button
            variant="primary"
            type='submit'
            onClick={(e) => {
              e.preventDefault();
              if (isValidEmail) {
                this.addBuddy();
              }
            }}
            style={{ backgroundColor: '#007bff', borderColor: '#007bff', color: 'white' }}
            disabled={!isValidEmail}
          >
            {this.props.t("addBuddy")}
          </Button>
        </Form.Group>
      </ListGroup.Item>
    );
  }

  renderItem = (item: Buddy) => {
    const { id, buddy: { email, firstBooking } } = item;
    let formatter = Formatting.getFormatter();
    if (RuntimeConfig.INFOS.dailyBasisBooking) {
      formatter = Formatting.getFormatterNoTime();
    }
    return (
      <ListGroup.Item key={id} style={{ minWidth: "300px" }}>
        <h5>{email}</h5>
        {firstBooking == null && <p>{this.props.t("noBooking")}</p> ||
          <p>
            <IconLocation className="feather" />&nbsp;{firstBooking!.room}, {firstBooking!.desk}<br />
            <IconEnter className="feather" />&nbsp;{formatter.format(new Date(firstBooking!.enter))}<br />
            <IconLeave className="feather" />&nbsp;{formatter.format(new Date(firstBooking!.leave))}
          </p>}
        <Button variant="danger" onClick={() => this.onItemPress(item)}>
          {this.props.t("removeBuddy")}
        </Button>
      </ListGroup.Item>
    );
  }

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
        </>
      );
    }
    return (
      <>
        <NavBar />
        <div className="container-signin">
          <Form className="form-signin">
              <ListGroup>
                {this.data.map(item => this.renderItem(item))}
                {this.renderAddBuddy()}
              </ListGroup>
          </Form>
        </div>
        <Modal show={this.state.selectedItem != null} onHide={() => this.setState({ selectedItem: null })}>
          <Modal.Header closeButton>
            <Modal.Title>{this.props.t("removeBuddy")}</Modal.Title>
          </Modal.Header>
          <Modal.Body>
            <p>{this.props.t("confirmRemoveBuddy", { interpolation: { escapeValue: false } })}</p>
          </Modal.Body>
          <Modal.Footer>
            <Button variant="secondary" onClick={() => this.setState({ selectedItem: null })}>
              {this.props.t("back")}
            </Button>
            <Button variant="danger" onClick={() => this.removeBuddy(this.state.selectedItem)}>
              {this.props.t("removeBuddy")}
            </Button>
          </Modal.Footer>
        </Modal>
      </>
    );
  }
}

export default withTranslation()(withReadyRouter(Buddies as any));
