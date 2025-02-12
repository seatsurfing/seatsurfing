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
}

interface Props extends WithTranslation {
}

class UpgradeCloud extends React.Component<Props, State> {
  constructor(props: any) {
    super(props);
    this.state = {
    };
  }

  componentDidMount = () => {
  }

  render() {
    return (
      <FullLayout headline={this.props.t("upgradeCloud")}>
        <iframe src="https://app.seatsurfing.io/cloud/" style={{ width: '100%', height: '100vh', borderWidth: 0 }}>
        </iframe>
      </FullLayout>
    );
  }
}

export default withTranslation(['admin'])(withReadyRouter(UpgradeCloud as any));
