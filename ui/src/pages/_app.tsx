import "bootstrap/dist/css/bootstrap.min.css";
import "@/styles/App.css";
import "@/styles/AdminNavBar.css";
import "@/styles/NavBar.css";
import "@/styles/CenterContent.css";
import "@/styles/Dashboard.css";
import "@/styles/EditLocation.css";
import "@/styles/ConfluenceHint.css";
import "@/styles/Login.css";
import "@/styles/Search.css";
import "@/styles/Settings.css";
import "@/styles/SideBar.css";
import "@/styles/FullLayout.css";
import "@/styles/Booking.css";
import type { AppProps } from "next/app";
import RuntimeConfig from "@/components/RuntimeConfig";
import React from "react";
import Loading from "@/components/Loading";
import Head from "next/head";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import Ajax from "@/util/Ajax";
import Formatting from "@/util/Formatting";
import TotpSetupModal from "@/components/TotpSetupModal";
import User from "@/types/User";

interface State {
  isLoading: boolean;
  showTotpEnforcement: boolean;
  totpQrCode: string;
  totpStateId: string;
  enforceTOTP: boolean;
  totpEnabled: boolean;
  idpLogin: boolean;
}

interface Props extends AppProps {
  t: TranslationFunc;
}

class App extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      isLoading: true,
      showTotpEnforcement: false,
      totpQrCode: "",
      totpStateId: "",
      enforceTOTP: false,
      totpEnabled: false,
      idpLogin: false,
    };
    if (typeof window !== "undefined") {
      if (process.env.NODE_ENV.toLowerCase() === "development") {
        Ajax.URL =
          "http://" + window.location.host.split(":").shift() + ":8080";
      }
    }
    setTimeout(() => {
      RuntimeConfig.verifyToken(() => {
        this.setState(
          {
            isLoading: false,
            enforceTOTP: RuntimeConfig.INFOS.enforceTOTP,
            totpEnabled: RuntimeConfig.INFOS.totpEnabled,
            idpLogin: RuntimeConfig.INFOS.idpLogin,
          },
          () => {
            this.checkTotpEnforcement();
          },
        );
      });
    }, 10);
  }

  componentDidUpdate(prevProps: Props, prevState: State) {
    // Sync state with RuntimeConfig.INFOS to detect changes
    const configChanged =
      RuntimeConfig.INFOS.enforceTOTP !== this.state.enforceTOTP ||
      RuntimeConfig.INFOS.totpEnabled !== this.state.totpEnabled ||
      RuntimeConfig.INFOS.idpLogin !== this.state.idpLogin;

    if (configChanged) {
      this.setState({
        enforceTOTP: RuntimeConfig.INFOS.enforceTOTP,
        totpEnabled: RuntimeConfig.INFOS.totpEnabled,
        idpLogin: RuntimeConfig.INFOS.idpLogin,
      });
    }

    // Check enforcement when loading completes
    if (prevState.isLoading && !this.state.isLoading) {
      this.checkTotpEnforcement();
      return;
    }

    // Check when any of the enforcement-related config values change
    if (
      prevState.enforceTOTP !== this.state.enforceTOTP ||
      prevState.totpEnabled !== this.state.totpEnabled ||
      prevState.idpLogin !== this.state.idpLogin
    ) {
      this.checkTotpEnforcement();
      return;
    }

    // Also check when enforcement modal is dismissed (in case user refreshed or navigated away)
    if (
      prevState.showTotpEnforcement &&
      !this.state.showTotpEnforcement &&
      !this.state.totpEnabled
    ) {
      // User dismissed modal somehow without completing setup, check again
      this.checkTotpEnforcement();
    }
  }

  checkTotpEnforcement = () => {
    // Only check on client side
    if (typeof window === "undefined") {
      return;
    }

    // Only check if user is logged in
    if (!Ajax.hasAccessToken()) {
      return;
    }

    const currentPath = window.location.pathname;
    // Don't show on login/auth pages
    if (currentPath.includes("/login") || currentPath.includes("/resetpw")) {
      return;
    }

    // Check if TOTP is enforced and user doesn't have it enabled
    // Skip enforcement for IdP users (those who authenticate via external providers)
    if (
      RuntimeConfig.INFOS.enforceTOTP &&
      !RuntimeConfig.INFOS.totpEnabled &&
      !RuntimeConfig.INFOS.idpLogin
    ) {
      // Generate TOTP setup for the user
      User.generateTotp()
        .then((result) => {
          this.setState({
            showTotpEnforcement: true,
            totpQrCode: result.qrCode,
            totpStateId: result.stateId,
          });
        })
        .catch((err) => {
          console.error("Failed to generate TOTP:", err);
        });
    }
  };

  onTotpEnforcementSuccess = () => {
    RuntimeConfig.INFOS.totpEnabled = true;
    this.setState({
      showTotpEnforcement: false,
      totpEnabled: true,
    });
  };

  render() {
    if (typeof window !== "undefined") {
      if (window !== window.parent) {
        // Add Confluence JS
        if (!document.getElementById("confluence-js")) {
          const script = document.createElement("script");
          script.id = "confluence-js";
          script.src = "https://connect-cdn.atl-paas.net/all.js";
          document.head.appendChild(script);
        }
        RuntimeConfig.EMBEDDED = true;
      }
    }

    if (this.state.isLoading) {
      return <Loading />;
    }

    const { Component, pageProps } = this.props;
    Formatting.Language = RuntimeConfig.getLanguage();
    Formatting.t = this.props.t;
    return (
      <>
        <Head>
          <link rel="icon" href="/ui/favicon.ico" />
          <meta name="viewport" content="width=device-width, initial-scale=1" />
          <meta name="theme-color" content="#343a40" />
          <link rel="manifest" href="/ui/manifest.json" />
          <meta name="mobile-web-app-capable" content="yes" />
          <meta
            name="apple-mobile-web-app-status-bar-style"
            content="default"
          />
          <link rel="shortcut icon" href="/ui/favicon-192.png" />
          <link rel="apple-touch-icon" href="/ui/favicon-192.png" />
          <link rel="apple-touch-startup-image" href="/ui/favicon-1024.png" />
          <title>Seatsurfing</title>
        </Head>
        {this.state.showTotpEnforcement && (
          <TotpSetupModal
            show={true}
            qrCode={this.state.totpQrCode}
            stateId={this.state.totpStateId}
            onHide={() => {}} // Cannot close
            onSuccess={this.onTotpEnforcementSuccess}
            canClose={false}
          />
        )}
        <Component {...pageProps} />
      </>
    );
  }
}

export default withTranslation(App as any);
