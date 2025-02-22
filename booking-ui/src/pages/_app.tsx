import 'bootstrap/dist/css/bootstrap.min.css'
import '@/styles/App.css'
import '@/styles/NavBar.css'
import '@/styles/CenterContent.css'
import '@/styles/ConfluenceHint.css'
import '@/styles/Login.css'
import '@/styles/Search.css'
import type { AppProps } from 'next/app'
import nextI18nConfig from '../../next-i18next.config'
import { WithTranslation, appWithTranslation, withTranslation } from 'next-i18next'
import { Ajax, Formatting } from 'seatsurfing-commons'
import RuntimeConfig from '@/components/RuntimeConfig'
import React from 'react'
import Loading from '@/components/Loading'
import Head from 'next/head'

interface State {
  isLoading: boolean;
}

interface Props extends WithTranslation, AppProps {
}

class App extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      isLoading: true
    };
    if (typeof window !== 'undefined') {
      if (process.env.NODE_ENV.toLowerCase() === "development") {
        Ajax.URL = "http://" + window.location.host.split(':').shift() + ":8080";
      }
      if (window.location.href.indexOf(".loca.lt/") > -1) {
        Ajax.URL = "https://" + window.location.host.split(':').shift();
      }
    }
    setTimeout(() => {
      RuntimeConfig.verifyToken(() => {
        this.setState({isLoading: false});
      });
    }, 10);
  }

  render() {
    if (typeof window !== 'undefined') {
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

    if ((this.state.isLoading) || (!this.props.tReady)) {
      return <Loading />;
    }

    const { Component, pageProps } = this.props;
    Formatting.Language = this.props.i18n.language;
    // @ts-ignore
    Formatting.t = this.props.t;
    return (
      <>
        <Head>
          <link rel="icon" href="/ui/favicon.ico" />
          <meta name="viewport" content="width=device-width, initial-scale=1" />
          <meta name="theme-color" content="#343a40" />
          <link rel="manifest" href="/ui/manifest.json" />
          <meta name="apple-mobile-web-app-capable" content="yes" />
          <meta name="apple-mobile-web-app-status-bar-style" content="default" />
          <link rel="shortcut icon" href="/ui/favicon-192.png" />
          <link rel="apple-touch-icon" href="/ui/favicon-192.png" />
          <link rel="apple-touch-startup-image" href="/ui/favicon-1024.png" />
          <title>Seatsurfing</title>
        </Head>
        <Component {...pageProps} />
      </>
    );
  }
}

export default appWithTranslation(withTranslation()(App), nextI18nConfig);


