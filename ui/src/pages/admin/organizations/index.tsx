import React from "react";
import { Table } from "react-bootstrap";
import { Plus as IconPlus } from "react-feather";
import FullLayout from "@/components/FullLayout";
import Loading from "@/components/Loading";
import Link from "next/link";
import { NextRouter } from "next/router";
import withReadyRouter from "@/components/withReadyRouter";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import Organization from "@/types/Organization";
import Ajax from "@/util/Ajax";

interface State {
  selectedItem: string;
  loading: boolean;
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
}

class Organizations extends React.Component<Props, State> {
  data: Organization[] = [];

  constructor(props: any) {
    super(props);
    this.state = {
      selectedItem: "",
      loading: true,
    };
  }

  componentDidMount = () => {
    if (!Ajax.hasAccessToken()) {
      this.props.router.push("/login");
      return;
    }
    this.loadItems();
  };

  loadItems = () => {
    Organization.list().then((list) => {
      this.data = list;
      this.setState({ loading: false });
    });
  };

  onItemSelect = (org: Organization) => {
    this.setState({ selectedItem: org.id });
  };

  renderItem = (org: Organization) => {
    return (
      <tr key={org.id} onClick={() => this.onItemSelect(org)}>
        <td>{org.name}</td>
      </tr>
    );
  };

  render() {
    if (this.state.selectedItem) {
      this.props.router.push(`/admin/organizations/${this.state.selectedItem}`);
      return <></>;
    }

    let buttons = (
      <Link
        href="/admin/organizations/add"
        className="btn btn-sm btn-outline-secondary"
      >
        <IconPlus className="feather" /> {this.props.t("add")}
      </Link>
    );

    if (this.state.loading) {
      return (
        <FullLayout headline={this.props.t("organizations")} buttons={buttons}>
          <Loading />
        </FullLayout>
      );
    }

    let rows = this.data.map((item) => this.renderItem(item));
    if (rows.length === 0) {
      return (
        <FullLayout headline={this.props.t("organizations")} buttons={buttons}>
          <p>{this.props.t("noRecords")}</p>
        </FullLayout>
      );
    }
    return (
      <FullLayout headline={this.props.t("organizations")} buttons={buttons}>
        <Table striped={true} hover={true} className="clickable-table">
          <thead>
            <tr>
              <th>{this.props.t("org")}</th>
            </tr>
          </thead>
          <tbody>{rows}</tbody>
        </Table>
      </FullLayout>
    );
  }
}

export default withTranslation(withReadyRouter(Organizations as any));
