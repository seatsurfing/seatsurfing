import React from 'react';
import { Table } from 'react-bootstrap';
import { Plus as IconPlus, Download as IconDownload, Tag as IconTag } from 'react-feather';
import { Ajax, SpaceAttribute } from 'seatsurfing-commons';
import { WithTranslation, withTranslation } from 'next-i18next';
import FullLayout from '@/components/FullLayout';
import { NextRouter } from 'next/router';
import Link from 'next/link';
import Loading from '@/components/Loading';
import withReadyRouter from '@/components/withReadyRouter';

interface State {
  selectedItem: string
  loading: boolean
}

interface Props extends WithTranslation {
  router: NextRouter
}

class Attributes extends React.Component<Props, State> {
  data: SpaceAttribute[] = [];
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
    SpaceAttribute.list().then(list => {
      this.data = list;
      this.setState({ loading: false });
    });
  }

  onItemSelect = (e: SpaceAttribute) => {
    this.setState({ selectedItem: e.id });
  }

  getTextForType = (type: Number) => {
    if (type === 1) return this.props.t("number");
    if (type === 2) return this.props.t("boolean");
    if (type === 3) return this.props.t("text");
    return "";
  }

  renderItem = (e: SpaceAttribute) => {
    return (
      <tr key={e.id} onClick={() => this.onItemSelect(e)}>
        <td>{e.label}</td>
        <td>{this.getTextForType(e.type)}</td>
        <td>{e.locationApplicable ? this.props.t("yes") : ""}</td>
        <td>{e.spaceApplicable ? this.props.t("yes") : ""}</td>
      </tr>
    );
  }

  exportTable = (e: any) => {
    return this.ExcellentExport.convert(
      { anchor: e.target, filename: "seatsurfing-attributes", format: "xlsx" },
      [{ name: "Seatsurfing Attributes", from: { table: "datatable" } }]
    );
  }

  render() {
    if (this.state.selectedItem) {
      this.props.router.push(`/attributes/${this.state.selectedItem}`);
      return <></>
    }

    // eslint-disable-next-line
    let downloadButton = <a download="seatsurfing-attributes.xlsx" href="#" className="btn btn-sm btn-outline-secondary" onClick={this.exportTable}><IconDownload className="feather" /> {this.props.t("download")}</a>;
    let buttons = (
      <>
        {this.data && this.data.length > 0 ? downloadButton : <></>}
        <Link href="/attributes/add" className="btn btn-sm btn-outline-secondary"><IconPlus className="feather" /> {this.props.t("add")}</Link>
      </>
    );

    if (this.state.loading) {
      return (
        <FullLayout headline={this.props.t("attributes")} buttons={buttons}>
          <Loading />
        </FullLayout>
      );
    }

    let rows = this.data.map(item => this.renderItem(item));
    if (rows.length === 0) {
      return (
        <FullLayout headline={this.props.t("attributes")} buttons={buttons}>
          <p>{this.props.t("noRecords")}</p>
        </FullLayout>
      );
    }
    return (
      <FullLayout headline={this.props.t("attributes")} buttons={buttons}>
        <Table striped={true} hover={true} className="clickable-table" id="datatable">
          <thead>
            <tr>
              <th>{this.props.t("name")}</th>
              <th>{this.props.t("type")}</th>
              <th>{this.props.t("areas")}</th>
              <th>{this.props.t("spaces")}</th>
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

export default withTranslation(['admin'])(withReadyRouter(Attributes as any));
