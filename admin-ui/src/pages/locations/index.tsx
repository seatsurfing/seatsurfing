import React from 'react';
import { Table } from 'react-bootstrap';
import { Plus as IconPlus, Download as IconDownload, Tag as IconTag } from 'react-feather';
import { Ajax, Location } from 'flexspace-commons';
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

class Locations extends React.Component<Props, State> {
  data: Location[] = [];
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
    Location.list().then(list => {
      this.data = list;
      this.setState({ loading: false });
    });
  }

  onItemSelect = (location: Location) => {
    this.setState({ selectedItem: location.id });
  }

  renderItem = (location: Location) => {
    return (
      <tr key={location.id} onClick={() => this.onItemSelect(location)}>
        <td>{location.name}</td>
        <td>{location.enabled ? this.props.t('yes') : ''}</td>
        <td>{location.mapWidth}x{location.mapHeight}</td>
        <td>{window.location.origin}/ui/search?lid={location.id}</td>
      </tr>
    );
  }

  exportTable = (e: any) => {
    return this.ExcellentExport.convert(
      { anchor: e.target, filename: "seatsurfing-areas", format: "xlsx" },
      [{ name: "Seatsurfing Areas", from: { table: "datatable" } }]
    );
  }

  render() {
    if (this.state.selectedItem) {
      this.props.router.push(`/locations/${this.state.selectedItem}`);
      return <></>
    }

    // eslint-disable-next-line
    let downloadButton = <a download="seatsurfing-areas.xlsx" href="#" className="btn btn-sm btn-outline-secondary" onClick={this.exportTable}><IconDownload className="feather" /> {this.props.t("download")}</a>;
    let buttons = (
      <>
        {this.data && this.data.length > 0 ? downloadButton : <></>}
        <Link href="/attributes" className="btn btn-sm btn-outline-secondary"><IconTag className="feather" /> {this.props.t("attributes")}</Link>
        <Link href="/locations/add" className="btn btn-sm btn-outline-secondary"><IconPlus className="feather" /> {this.props.t("add")}</Link>
      </>
    );

    if (this.state.loading) {
      return (
        <FullLayout headline={this.props.t("areas")} buttons={buttons}>
          <Loading />
        </FullLayout>
      );
    }

    let rows = this.data.map(item => this.renderItem(item));
    if (rows.length === 0) {
      return (
        <FullLayout headline={this.props.t("areas")} buttons={buttons}>
          <p>{this.props.t("noRecords")}</p>
        </FullLayout>
      );
    }
    return (
      <FullLayout headline={this.props.t("areas")} buttons={buttons}>
        <Table striped={true} hover={true} className="clickable-table" id="datatable">
          <thead>
            <tr>
              <th>{this.props.t("name")}</th>
              <th>{this.props.t("enabled")}</th>
              <th>{this.props.t("map")}</th>
              <th>{this.props.t("bookingLink")}</th>
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

export default withTranslation(['admin'])(withReadyRouter(Locations as any));
