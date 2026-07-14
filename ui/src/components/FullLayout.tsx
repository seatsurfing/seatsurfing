import React, { JSX, ReactNode } from "react";
import AdminNavBar from "./AdminNavBar";
import SideBar from "./SideBar";
import RuntimeConfig from "./RuntimeConfig";
import FeedbackButton from "./FeedbackButton";
import PluginEmbed from "./PluginEmbed";

interface Props {
  headline: string;
  buttons?: ReactNode;
  children: string | JSX.Element | JSX.Element[];
}

interface State {}

export default class FullLayout extends React.Component<Props, State> {
  skipWelcomeScreen = () => {
    window.sessionStorage.setItem("skipWelcomeScreen", "true");
    window.location.reload();
  };

  render() {
    if (
      RuntimeConfig.INFOS.pluginWelcomeScreens &&
      RuntimeConfig.INFOS.pluginWelcomeScreens.length > 0 &&
      window.sessionStorage.getItem("skipWelcomeScreen") !== "true"
    ) {
      const screen = RuntimeConfig.INFOS.pluginWelcomeScreens[0];
      return (
        <div style={{ minHeight: "100vh", width: "100%", overflowY: "auto" }}>
          <PluginEmbed
            id="welcome-screen-embed"
            src={screen.src}
            tagName={screen.tagName}
            style={{ width: "100%", borderWidth: 0 }}
            onSkipWelcomeScreen={this.skipWelcomeScreen}
          />
        </div>
      );
    }

    return (
      <div>
        <AdminNavBar />
        <div className="container-fluid">
          <div className="row">
            <SideBar />
            <main
              role="main"
              className="col-11 col-md-9 ms-auto col-lg-10 px-md-4"
            >
              <div className="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
                <h1 className="h2">{this.props.headline}</h1>
                <div className="btn-toolbar mb-2 mb-md-0">
                  <div className="btn-group me-2">{this.props.buttons}</div>
                </div>
              </div>
              {this.props.children}
              <FeedbackButton />
            </main>
          </div>
        </div>
      </div>
    );
  }
}
