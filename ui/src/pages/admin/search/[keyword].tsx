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
import RendererUtils from "@/util/RendererUtils";

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

    if (typeof keyword !== "string") {
      this.setState({ loading: false });
    }

    const options = new SearchOptions();
    options.includeUsers = true;
    options.includeLocations = true;
    options.includeSpaces = true;
    options.includeGroups = true;
    options.expandLocations = true;
    options.keyword = keyword ? (keyword as string) : "";
    const result = await Search.search(options);
    this.data = result;
    this.setState({ loading: false });
  };

  escapeHTML = (s: string): string => {
    return s;
  };

  renderResults = <T extends { id: string }>(
    items: T[],
    titleKey: string,
    getItemData: (item: T) => { text: string; link: string },
  ) => {
    const listItems = items.map((item) => {
      const { text, link } = getItemData(item);
      return (
        <ListGroup.Item
          key={item.id}
          action
          as={Link}
          href={link}
          className="link-primary text-decoration-underline active-transparent"
        >
          {text}
        </ListGroup.Item>
      );
    });

    if (listItems.length === 0) {
      listItems.push(
        <ListGroup.Item key={`${titleKey}-no-results`}>
          {this.props.t("noResults")}
        </ListGroup.Item>,
      );
    }

    return (
      <Col sm="3" className="mb-4">
        <Card>
          <Card.Header>
            {this.props.t(titleKey)} ({items.length})
          </Card.Header>
          <ListGroup variant="flush">{listItems}</ListGroup>
        </Card>
      </Col>
    );
  };

  renderUserResults = () =>
    this.renderResults(this.data.users, "users", (user) => ({
      text: `${user.email} ${RendererUtils.preAndSuffixIfDefined(
        RendererUtils.fullname(user.firstname, user.lastname),
        "(",
        ")",
      )}`.trim(),
      link: Navigation.adminUserDetails(user.id),
    }));

  renderGroupResults = () =>
    this.renderResults(this.data.groups, "groups", (group) => ({
      text: group.name,
      link: Navigation.adminGroupDetails(group.id),
    }));

  renderLocationResults = () =>
    this.renderResults(this.data.locations, "areas", (location) => ({
      text: location.name,
      link: Navigation.adminLocationDetails(location.id),
    }));

  renderSpaceResults = () =>
    this.renderResults(this.data.spaces, "spaces", (space) => ({
      text: `${space.location.name} > ${space.name}`,
      link: Navigation.adminLocationDetails(space.locationId),
    }));

  render() {
    const { keyword } = this.props.router.query;

    const headline = RendererUtils.decodeHtmlEntities(
      this.props.t("searchForX", {
        keyword: this.escapeHTML(
          typeof keyword === "string" && keyword ? keyword : "",
        ),
      }),
    );

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
          {this.renderGroupResults()}
          {this.renderLocationResults()}
          {this.renderSpaceResults()}
        </Row>
      </FullLayout>
    );
  }
}

export default withTranslation(withReadyRouter(SearchResult as any));
