import React from 'react';
import { WithTranslation, withTranslation } from 'next-i18next';
import FullLayout from '@/components/FullLayout';
import Loading from '@/components/Loading';
import withReadyRouter from '@/components/withReadyRouter';

interface State {
  iFrameLoaded: boolean
}

interface Props extends WithTranslation {
}

class UpgradeCloud extends React.Component<Props, State> {
  constructor(props: any) {
    super(props);
    this.state = {
      iFrameLoaded: false
    };
  }

  componentDidMount = () => {
    this.checkiFrameHeight();
  }

  checkiFrameHeight(): void {
    window.setTimeout(() => {
      if (!window.location.pathname.endsWith('/upgrade-cloud')) return;
      this.checkiFrameHeight();
      let iFrame = document.getElementById("payment-iframe") as HTMLIFrameElement;
      if (!iFrame || !iFrame.contentWindow || !iFrame.contentWindow.document || !iFrame.contentWindow.document.body) return;
      let height = iFrame.contentWindow.document.body.scrollHeight;
      iFrame.style.height = height + 'px';
      if (height > 0) {
        this.setState({ iFrameLoaded: true });
      }
    }, 2000);
  }

  render() {
    return (
      <FullLayout headline={this.props.t("upgradeCloud")}>
        { this.state.iFrameLoaded ? <></> : <Loading /> }
        <iframe src="https://app.seatsurfing.io/cloud/" style={{ width: '100%', height: '0', borderWidth: 0 }} id="payment-iframe">
        </iframe>
      </FullLayout>
    );
  }
}

export default withTranslation(['admin'])(withReadyRouter(UpgradeCloud as any));
