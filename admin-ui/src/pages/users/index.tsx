import React from 'react';
import { Table } from 'react-bootstrap';
import { Plus as IconPlus, Download as IconDownload } from 'react-feather';
import { User, AuthProvider, Ajax } from 'seatsurfing-commons';
import { WithTranslation, withTranslation } from 'next-i18next';
import FullLayout from '@/components/FullLayout';
import Loading from '@/components/Loading';
import Link from 'next/link';
import { NextRouter } from 'next/router';
import withReadyRouter from '@/components/withReadyRouter';

interface State {
  selectedItem: string
  loading: boolean
}

interface Props extends WithTranslation {
  router: NextRouter
}

class Users extends React.Component<Props, State> {
  authProviders: { [key: string]: string } = {};
  data: User[] = [];
  ExcellentExport: any;

  constructor(props: any) {
    super(props);
    this.state = {
      selectedItem: "",
      loading: true
    };
  }

  componentDidMount = () => {
    if (!Ajax.CREDENTIALS.accessToken) {
      this.props.router.push("/login");
      return;
    }
    import('excellentexport').then(imp => this.ExcellentExport = imp.default);
    this.loadItems();
  }

  loadItems = () => {
    AuthProvider.list().then(providers => {
      providers.forEach(provider => {
        this.authProviders[provider.id] = provider.name;
      });
      User.list().then(list => {
        this.data = list;
        this.setState({ loading: false });
      });
    });
  }

  onItemSelect = (user: User) => {
    this.setState({ selectedItem: user.id });
  }

  renderItem = (user: User) => {
    let authProvider = "";
    if (user.requirePassword) {
      authProvider = this.props.t("password");
    } else if (this.authProviders[user.authProviderId]) {
      authProvider = this.authProviders[user.authProviderId];
    }
    let role = this.props.t("roleUser");
    if (user.role === User.UserRoleSpaceAdmin) {
      role = this.props.t("roleSpaceAdmin");
    }
    if (user.role === User.UserRoleOrgAdmin) {
      role = this.props.t("roleOrgAdmin");
    }
    if (user.role === User.UserRoleSuperAdmin) {
      role = this.props.t("roleSuperAdmin");
    }
    return (
      <tr key={user.id} onClick={() => this.onItemSelect(user)}>
        <td>{user.email}</td>
        <td>{role}</td>
        <td>{authProvider}</td>
      </tr>
    );
  }

  exportTable = (e: any) => {
    return this.ExcellentExport.convert(
      { anchor: e.target, filename: "seatsurfing-users", format: "xlsx" },
      [{ name: "Seatsurfing Users", from: { table: "datatable" } }]
    );
  }

  render() {
    if (this.state.selectedItem) {
      this.props.router.push(`/users/${this.state.selectedItem}`);
      return <></>
    }
    // eslint-disable-next-line
    let downloadButton = <a download="seatsurfing-users.xlsx" href="#" className="btn btn-sm btn-outline-secondary" onClick={this.exportTable}><IconDownload className="feather" /> {this.props.t("download")}</a>;
    let buttons = (
      <>
        {this.data && this.data.length > 0 ? downloadButton : <></>}
        <Link href="/users/add" className="btn btn-sm btn-outline-secondary"><IconPlus className="feather" /> {this.props.t("add")}</Link>
      </>
    );

    if (this.state.loading) {
      return (
        <FullLayout headline={this.props.t("users")} buttons={buttons}>
          <Loading />
        </FullLayout>
      );
    }

    let rows = this.data.map(item => this.renderItem(item));
    if (rows.length === 0) {
      return (
        <FullLayout headline={this.props.t("users")} buttons={buttons}>
          <p>{this.props.t("noRecords")}</p>
        </FullLayout>
      );
    }
    return (
      <FullLayout headline={this.props.t("users")} buttons={buttons}>
        <Table striped={true} hover={true} className="clickable-table" id="datatable">
          <thead>
            <tr>
              <th>{this.props.t("username")}</th>
              <th>{this.props.t("role")}</th>
              <th>{this.props.t("loginMeans")}</th>
            </tr>
          </thead>
          <tbody>
            {rows}
          </tbody>
        </Table>
      </FullLayout>
    );
  }
}

export default withTranslation(['admin'])(withReadyRouter(Users as any));
