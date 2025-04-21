import 'bootstrap/dist/css/bootstrap.min.css'
import '@/styles/App.css'
import '@/styles/CenterContent.css'
import '@/styles/Dashboard.css'
import '@/styles/EditLocation.css'
import '@/styles/Login.css'
import '@/styles/NavBar.css'
import '@/styles/Settings.css'
import '@/styles/SideBar.css'
import 'react-calendar/dist/Calendar.css';
import '@/styles/Booking.css'
import type { AppProps } from 'next/app'
import nextI18nConfig from '../../next-i18next.config'
import { WithTranslation, appWithTranslation, withTranslation } from 'next-i18next'
import { Ajax, Formatting } from 'seatsurfing-commons'
import React from 'react'
import Loading from '@/components/Loading'
import Head from 'next/head'
import RuntimeConfig from '@/components/RuntimeConfig'


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
    }
    setTimeout(() => {
      RuntimeConfig.verifyToken(() => {
        this.setState({isLoading: false});
      });
    }, 10);
  }

  render() {
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
          <link rel="icon" href="/admin/favicon.ico" />
          <meta name="viewport" content="width=device-width, initial-scale=1" />
          <meta name="theme-color" content="#343a40" />
          <link rel="manifest" href="/admin/manifest.json" />
          <meta name="apple-mobile-web-app-capable" content="yes" />
          <meta name="apple-mobile-web-app-status-bar-style" content="default" />
          <link rel="shortcut icon" href="/admin/favicon-192.png" />
          <link rel="apple-touch-icon" href="/admin/favicon-192.png" />
          <link rel="apple-touch-startup-image" href="/admin/favicon-1024.png" />
          <title>Seatsurfing</title>
        </Head>
        <Component {...pageProps} />
      </>
    );
  }
}

export default appWithTranslation(withTranslation(['admin'])(App), nextI18nConfig);
