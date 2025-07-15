import React from "react";
import { Alert, Col, Row, Table } from "react-bootstrap";
import { Plus as IconPlus, Download as IconDownload } from "react-feather";
import { Ajax, Group } from "seatsurfing-commons";
import FullLayout from "@/components/FullLayout";
import Loading from "@/components/Loading";
import Link from "next/link";
import { NextRouter } from "next/router";
import withReadyRouter from "@/components/withReadyRouter";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import RuntimeConfig from "@/components/RuntimeConfig";
import CloudFeatureHint from "@/components/CloudFeatureHint";

interface State {
  selectedItem: string;
  loading: boolean;
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
}

class Groups extends React.Component<Props, State> {
  data: Group[] = [];
  ExcellentExport: any;

  constructor(props: any) {
    super(props);
    this.state = {
      selectedItem: "",
      loading: true,
    };
  }

  componentDidMount = () => {
    if (!Ajax.CREDENTIALS.accessToken) {
      this.props.router.push("/login");
      return;
    }
    import("excellentexport").then(
      (imp) => (this.ExcellentExport = imp.default)
    );
    this.loadItems();
  };

  loadItems = () => {
    Group.list().then((list) => {
      this.data = list;
      this.setState({ loading: false });
    });
  };

  onItemSelect = (group: Group) => {
    this.setState({ selectedItem: group.id });
  };

  renderItem = (group: Group) => {
    return (
      <tr key={group.id} onClick={() => this.onItemSelect(group)}>
        <td>{group.name}</td>
      </tr>
    );
  };

  exportTable = (e: any) => {
    return this.ExcellentExport.convert(
      { anchor: e.target, filename: "seatsurfing-groups", format: "xlsx" },
      [{ name: "Seatsurfing Groups", from: { table: "datatable" } }]
    );
  };

  render() {
    if (
      RuntimeConfig.INFOS.cloudHosted &&
      !RuntimeConfig.INFOS.subscriptionActive
    ) {
      return (
        <FullLayout headline={this.props.t("groups")}>
          <CloudFeatureHint />
        </FullLayout>
      );
    }

    if (this.state.selectedItem) {
      this.props.router.push(`/groups/${this.state.selectedItem}`);
      return <></>;
    }
    // eslint-disable-next-line
    let downloadButton = (
      <a
        download="seatsurfing-groups.xlsx"
        href="#"
        className="btn btn-sm btn-outline-secondary"
        onClick={this.exportTable}
      >
        <IconDownload className="feather" /> {this.props.t("download")}
      </a>
    );
    let buttons = (
      <>
        {this.data && this.data.length > 0 ? downloadButton : <></>}
        <Link href="/groups/add" className="btn btn-sm btn-outline-secondary">
          <IconPlus className="feather" /> {this.props.t("add")}
        </Link>
      </>
    );

    if (this.state.loading) {
      return (
        <FullLayout headline={this.props.t("groups")} buttons={buttons}>
          <Loading />
        </FullLayout>
      );
    }

    let rows = this.data.map((item) => this.renderItem(item));
    if (rows.length === 0) {
      return (
        <FullLayout headline={this.props.t("groups")} buttons={buttons}>
          <p>{this.props.t("noRecords")}</p>
        </FullLayout>
      );
    }
    return (
      <FullLayout headline={this.props.t("groups")} buttons={buttons}>
        <Table
          striped={true}
          hover={true}
          className="clickable-table"
          id="datatable"
        >
          <thead>
            <tr>
              <th>{this.props.t("name")}</th>
            </tr>
          </thead>
          <tbody>{rows}</tbody>
        </Table>
      </FullLayout>
    );
  }
}

export default withTranslation(withReadyRouter(Groups as any));
