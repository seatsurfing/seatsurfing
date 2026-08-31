import React from "react";
import { Table } from "react-bootstrap";
import { Plus as IconPlus } from "react-feather";
import FullLayout from "@/components/FullLayout";
import Loading from "@/components/Loading";
import Link from "next/link";
import { NextRouter } from "next/router";
import withReadyRouter from "@/components/withReadyRouter";
import withPermission from "@/components/withPermission";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import RuntimeConfig from "@/components/RuntimeConfig";
import Role from "@/types/Role";
import { Permission, PermissionLevel } from "@/types/Permission";

interface State {
  selectedItem: string;
  loading: boolean;
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
}

class Roles extends React.Component<Props, State> {
  data: Role[] = [];

  constructor(props: any) {
    super(props);
    this.state = {
      selectedItem: "",
      loading: true,
    };
  }

  componentDidMount = () => {
    this.loadItems();
  };

  loadItems = () => {
    Role.list().then((list) => {
      this.data = list;
      this.setState({ loading: false });
    });
  };

  onItemSelect = (role: Role) => {
    this.setState({ selectedItem: role.id });
  };

  renderItem = (role: Role) => {
    // Only the permissions actually granted are stored, so the count is the
    // number of functionalities the role opens up.
    const grantedCount = Object.values(role.permissions).filter(
      (level) => level > PermissionLevel.None,
    ).length;
    return (
      <tr key={role.id} onClick={() => this.onItemSelect(role)}>
        <td>
          {role.name}
          {role.system && (
            <span className="text-muted"> ({this.props.t("systemRole")})</span>
          )}
        </td>
        <td>{role.description}</td>
        <td>{grantedCount}</td>
      </tr>
    );
  };

  render() {
    if (this.state.selectedItem) {
      this.props.router.push(`/admin/roles/${this.state.selectedItem}`);
      return <></>;
    }
    const buttons = RuntimeConfig.hasPermission(
      Permission.Roles,
      PermissionLevel.Admin,
    ) ? (
      <Link
        href="/admin/roles/add"
        className="btn btn-sm btn-outline-secondary"
      >
        <IconPlus className="feather" /> {this.props.t("add")}
      </Link>
    ) : (
      <></>
    );

    if (this.state.loading) {
      return (
        <FullLayout headline={this.props.t("roles")} buttons={buttons}>
          <Loading />
        </FullLayout>
      );
    }

    return (
      <FullLayout headline={this.props.t("roles")} buttons={buttons}>
        <Table striped={true} hover={true} className="clickable-table">
          <thead>
            <tr>
              <th>{this.props.t("name")}</th>
              <th>{this.props.t("description")}</th>
              <th>{this.props.t("grantedPermissions")}</th>
            </tr>
          </thead>
          <tbody>{this.data.map((item) => this.renderItem(item))}</tbody>
        </Table>
      </FullLayout>
    );
  }
}

export default withTranslation(
  withReadyRouter(
    withPermission(Roles as any, Permission.Roles, PermissionLevel.Read) as any,
  ),
);
