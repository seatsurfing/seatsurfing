import React from 'react';
import { WithTranslation, withTranslation } from 'next-i18next';
import FullLayout from '@/components/FullLayout';
import Loading from '@/components/Loading';
import withReadyRouter from '@/components/withReadyRouter';
import { NextRouter } from 'next/router';
import { Ajax } from 'flexspace-commons';

interface State {
  iFrameLoaded: boolean
  pluginMenuItem: any
}

interface Props extends WithTranslation {
  router: NextRouter
}

class PluginPage extends React.Component<Props, State> {
  constructor(props: any) {
    super(props);
    this.state = {
      iFrameLoaded: false,
      pluginMenuItem: {},
    };
  }

  componentDidMount = () => {
    this.loadData();
  }

  loadData = () => {
    const { id } = this.props.router.query;
    Ajax.get('/plugin/admin-menu-items/').then(res => {
      for (let item of res.json) {
        if (item.id === id) {
          this.setState({
            pluginMenuItem: item
          });
          this.checkiFrameHeight();
          return;
        }
      }
    });
  }

  checkiFrameHeight(): void {
    window.setTimeout(() => {
      if (!window.location.pathname.startsWith("/admin/plugin/")) return;
      this.checkiFrameHeight();
      let iFrame = document.getElementById("plugin-iframe") as HTMLIFrameElement;
      if (!iFrame || !iFrame.contentWindow || !iFrame.contentWindow.document || !iFrame.contentWindow.document.body) return;
      let height = iFrame.contentWindow.document.body.scrollHeight;
      iFrame.style.height = height + 'px';
      if (height > 0) {
        this.setState({ iFrameLoaded: true });
      }
    }, 2000);
  }

  render() {
    if (this.state.pluginMenuItem === undefined || this.state.pluginMenuItem === null || !this.state.pluginMenuItem.src) {
      return (
        <FullLayout headline={''}>
          <Loading />
        </FullLayout>
      );
    }
    let url = this.state.pluginMenuItem.src;
    if (!(url.startsWith('http://') || url.startsWith('https://'))) {
      url = Ajax.getBackendUrl() + url;
    }
    return (
      <FullLayout headline={this.state.pluginMenuItem ? this.state.pluginMenuItem.title : ''}>
        {this.state.iFrameLoaded ? <></> : <Loading />}
        <iframe src={url} style={{ width: '100%', height: '0', borderWidth: 0 }} id="plugin-iframe">
        </iframe>
      </FullLayout>
    );
  }
}

export default withTranslation(['admin'])(withReadyRouter(PluginPage as any));
