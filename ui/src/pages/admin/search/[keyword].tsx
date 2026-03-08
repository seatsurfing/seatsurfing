import React from "react";
import { Card, ListGroup, Col, Row } from "react-bootstrap";
import { NextRouter } from "next/router";
import FullLayout from "@/components/FullLayout";
import Loading from "@/components/Loading";
import Link from "next/link";
import withReadyRouter from "@/components/withReadyRouter";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import Search, { SearchOptions } from "@/types/Search";
import Ajax from "@/util/Ajax";
import RedirectUtil from "@/util/RedirectUtil";
import * as Navigation from "@/util/Navigation";

interface State {
  loading: boolean;
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
}

class SearchResult extends React.Component<Props, State> {
  data: Search;

  constructor(props: any) {
    super(props);
    this.data = new Search();
    this.state = {
      loading: true,
    };
  }

  componentDidMount = () => {
    if (!Ajax.hasAccessToken()) {
      RedirectUtil.toLogin(this.props.router);
      return;
    }
    this.loadItems();
  };

  componentDidUpdate = (prevProps: Props) => {
    const { keyword } = this.props.router.query;
    if (keyword !== prevProps.router.query["keyword"]) {
      this.loadItems();
    }
  };

  loadItems = async () => {
    const { keyword } = this.props.router.query;
    if (typeof keyword === "string") {
      const options = new SearchOptions();
      options.includeUsers = true;
      options.includeLocations = true;
      options.includeSpaces = true;
      options.includeGroups = true;
      const result = await Search.search(keyword ? keyword : "", options);
      this.data = result;
      this.setState({ loading: false });
    } else {
      this.setState({ loading: false });
    }
  };

  escapeHTML = (s: string): string => {
    return s;
  };

  renderUserResults = () => {
    const items = this.data.users.map((user) => {
      const link = Navigation.adminUserDetails(user.id);
      return (
        <ListGroup.Item key={user.id}>
          <Link href={link}>{user.email}</Link>
        </ListGroup.Item>
      );
    });
    if (items.length === 0) {
      items.push(
        <ListGroup.Item key="users-no-results">
          {this.props.t("noResults")}
        </ListGroup.Item>,
      );
    }
    return (
      <Col sm="4" className="mb-4">
        <Card>
          <Card.Header>
            {this.props.t("users")} ({this.data.users.length})
          </Card.Header>
          <ListGroup variant="flush">{items}</ListGroup>
        </Card>
      </Col>
    );
  };

  renderLocationResults = () => {
    const items = this.data.locations.map((location) => {
      const link = Navigation.adminLocationDetails(location.id);
      return (
        <ListGroup.Item key={location.id}>
          <Link href={link}>{location.name}</Link>
        </ListGroup.Item>
      );
    });
    if (items.length === 0) {
      items.push(
        <ListGroup.Item key="locations-no-results">
          {this.props.t("noResults")}
        </ListGroup.Item>,
      );
    }
    return (
      <Col sm="4" className="mb-4">
        <Card>
          <Card.Header>
            {this.props.t("areas")} ({this.data.locations.length})
          </Card.Header>
          <ListGroup variant="flush">{items}</ListGroup>
        </Card>
      </Col>
    );
  };

  renderSpaceResults = () => {
    const items = this.data.spaces.map((space) => {
      const link = Navigation.adminLocationDetails(space.locationId);
      return (
        <ListGroup.Item key={space.id}>
          <Link href={link}>{space.name}</Link>
        </ListGroup.Item>
      );
    });
    if (items.length === 0) {
      items.push(
        <ListGroup.Item key="spaces-no-results">
          {this.props.t("noResults")}
        </ListGroup.Item>,
      );
    }
    return (
      <Col sm="4" className="mb-4">
        <Card>
          <Card.Header>
            {this.props.t("spaces")} ({this.data.spaces.length})
          </Card.Header>
          <ListGroup variant="flush">{items}</ListGroup>
        </Card>
      </Col>
    );
  };

  render() {
    const { keyword } = this.props.router.query;
    let headline = "";
    if (typeof keyword === "string") {
      headline = this.props.t("searchForX", {
        keyword: this.escapeHTML(keyword ? keyword : ""),
      });
    } else {
      headline = this.props.t("searchForX", { keyword: "" });
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

export default withTranslation(withReadyRouter(SearchResult as any));
