import React from 'react';
import { Ajax, Search } from 'seatsurfing-commons';
import { Card, ListGroup, Col, Row } from 'react-bootstrap';
import { WithTranslation, withTranslation } from 'next-i18next';
import { NextRouter } from 'next/router';
import FullLayout from '@/components/FullLayout';
import Loading from '@/components/Loading';
import Link from 'next/link';
import withReadyRouter from '@/components/withReadyRouter';

interface State {
  loading: boolean
}

interface Props extends WithTranslation {
  router: NextRouter
}

class SearchResult extends React.Component<Props, State> {
  data: Search;

  constructor(props: any) {
    super(props);
    this.data = new Search();
    this.state = {
      loading: true
    };
  }

  componentDidMount = () => {
    if (!Ajax.CREDENTIALS.accessToken) {
      this.props.router.push("/login");
      return;
    }
    this.loadItems();
  }

  componentDidUpdate = (prevProps: Props) => {
    const { keyword } = this.props.router.query;
    if (keyword !== prevProps.router.query['keyword']) {
      this.loadItems();
    }
  }

  loadItems = () => {
    const { keyword } = this.props.router.query;
    if (typeof keyword === 'string') {
      Search.search(keyword ? keyword : "").then(res => {
        this.data = res;
        this.setState({ loading: false });
      });
    } else {
      this.setState({ loading: false });
    }
  }

  escapeHTML = (s: string): string => {
    return s;
  }

  renderUserResults = () => {
    let items = this.data.users.map(user => {
      let link = "/users/" + user.id;
      return (
        <ListGroup.Item key={user.id}><Link href={link}>{user.email}</Link></ListGroup.Item>
      );
    });
    if (items.length === 0) {
      items.push(<ListGroup.Item key="users-no-results">{this.props.t("noResults")}</ListGroup.Item>);
    }
    return (
      <Col sm="4" className="mb-4">
        <Card>
          <Card.Header>{this.props.t("users")}</Card.Header>
          <ListGroup variant="flush">
            {items}
          </ListGroup>
        </Card>
      </Col>
    );
  }

  renderLocationResults = () => {
    let items = this.data.locations.map(location => {
      let link = "/locations/" + location.id;
      return (
        <ListGroup.Item key={location.id}><Link href={link}>{location.name}</Link></ListGroup.Item>
      );
    });
    if (items.length === 0) {
      items.push(<ListGroup.Item key="locations-no-results">{this.props.t("noResults")}</ListGroup.Item>);
    }
    return (
      <Col sm="4" className="mb-4">
        <Card>
          <Card.Header>{this.props.t("areas")}</Card.Header>
          <ListGroup variant="flush">
            {items}
          </ListGroup>
        </Card>
      </Col>
    );
  }

  renderSpaceResults = () => {
    let items = this.data.spaces.map(space => {
      let link = "/locations/" + space.locationId;
      return (
        <ListGroup.Item key={space.id}><Link href={link}>{space.name}</Link></ListGroup.Item>
      );
    });
    if (items.length === 0) {
      items.push(<ListGroup.Item key="spaces-no-results">{this.props.t("noResults")}</ListGroup.Item>);
    }
    return (
      <Col sm="4" className="mb-4">
        <Card>
          <Card.Header>{this.props.t("spaces")}</Card.Header>
          <ListGroup variant="flush">
            {items}
          </ListGroup>
        </Card>
      </Col>
    );
  }

  render() {
    const { keyword } = this.props.router.query;
    let headline = '';
    if (typeof keyword === 'string') {
      headline = this.props.t("searchForX", {keyword: this.escapeHTML(keyword ? keyword : "")});
    } else {
      headline = this.props.t("searchForX", {keyword: ""});
    }

    if (this.state.loading) {
      return (
        <FullLayout headline={headline}>
          <Loading />
        </FullLayout>
      );
    }

    return (
      <FullLayout headline={headline}>
        <Row>
          {this.renderUserResults()}
          {this.renderLocationResults()}
          {this.renderSpaceResults()}
        </Row>
      </FullLayout>
    );
  }
}

export default withTranslation(['admin'])(withReadyRouter(SearchResult as any));
