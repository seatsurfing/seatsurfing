import React from "react";
import FullLayout from "@/components/FullLayout";
import Loading from "@/components/Loading";
import PluginEmbed from "@/components/PluginEmbed";
import withReadyRouter from "@/components/withReadyRouter";
import { NextRouter } from "next/router";
import RuntimeConfig from "@/components/RuntimeConfig";
import { withTranslation } from "@/components/withTranslation";
import Ajax from "@/util/Ajax";

interface State {
  pluginId: string;
  pluginMenuItem: any;
}

interface Props {
  router: NextRouter;
}

class PluginPage extends React.Component<Props, State> {
  constructor(props: any) {
    super(props);
    this.state = {
      pluginId: "",
      pluginMenuItem: {},
    };
  }

  componentDidMount = () => {
    this.loadData();
  };

  componentDidUpdate = () => {
    const { id } = this.props.router.query;
    if (this.state.pluginId !== id) {
      this.setState({ pluginId: id as string });
      this.loadData();
    }
  };

  loadData = () => {
    const { id } = this.props.router.query;
    for (let item of RuntimeConfig.INFOS.pluginMenuItems) {
      if (item.id === id) {
        this.setState({
          pluginId: id as string,
          pluginMenuItem: item,
        });
        return;
      }
    }
  };

  onNavigate = (path: string) => {
    this.props.router.push(path);
  };

  render() {
    if (
      this.state.pluginMenuItem === undefined ||
      this.state.pluginMenuItem === null ||
      !this.state.pluginMenuItem.src
    ) {
      return (
        <FullLayout headline={""}>
          <Loading />
        </FullLayout>
      );
    }
    const src = this.state.pluginMenuItem.src;
    if (
      src.startsWith("http://") ||
      src.startsWith("https://") ||
      src.startsWith("//")
    ) {
      console.error(
        "Plugin URL must be relative, absolute URLs are not allowed:",
        src,
      );
      return <></>;
    }
    const url = `${Ajax.getBackendUrl()}${src}`;
    return (
      <FullLayout
        headline={
          this.state.pluginMenuItem ? this.state.pluginMenuItem.title : ""
        }
      >
        <PluginEmbed
          id="plugin-embed"
          src={url}
          tagName={this.state.pluginMenuItem.tagName}
          style={{ width: "100%", borderWidth: 0 }}
          onNavigate={this.onNavigate}
        />
      </FullLayout>
    );
  }
}

export default withTranslation(withReadyRouter(PluginPage as any));
