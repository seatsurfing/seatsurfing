import React from 'react';
import { Loader as IconLoad } from 'react-feather';
import { Ajax } from 'seatsurfing-commons';
import { WithTranslation, withTranslation } from 'next-i18next';
import { NextRouter } from 'next/router';
import Link from 'next/link';
import withReadyRouter from '@/components/withReadyRouter';

interface State {
  loading: boolean
  success: boolean
  domain: string
}

interface Props extends WithTranslation {
  router: NextRouter
}

class ConfirmSignup extends React.Component<Props, State> {
  constructor(props: any) {
    super(props);
    this.state = {
      loading: true,
      success: false,
      domain: ''
    };
  }

  componentDidMount = () => {
    this.loadData();
  }

  loadData = () => {
    const { id } = this.props.router.query;
    if (id) {
      Ajax.postData("/signup/confirm/" + id, null).then((res) => {
        if (res.status >= 200 && res.status <= 299) {
          this.setState({ loading: false, success: true, domain: res.json.domain });
        } else {
          this.setState({ loading: false, success: false });
        }
      }).catch((e) => {
        this.setState({ loading: false, success: false });
      });
    } else {
      this.setState({ loading: false, success: false });
    }
  }

  render() {
    let loading = <></>;
    let result = <></>;
    if (this.state.loading) {
      loading = <div><IconLoad className="feather loader" /> {this.props.t("loadingHint")}</div>;
    } else {
      if (this.state.success) {
        result = (
          <div>
            <p>{this.props.t("orgSignupSuccess")}</p>
            <Link href={"https://"+this.state.domain+"/admin/login"} className="btn btn-primary">{this.props.t("orgSignupGoToLogin")}</Link>
          </div>
        );
      } else {
        result = (
          <div>
            <p>{this.props.t("orgSignupFailed")}</p>
          </div>
        );
      }
    }

    return (
      <div className="container-center">
        <div className="container-center-inner">
          <img src="/admin/seatsurfing.svg" alt="Seatsurfing" className="logo" />
          {loading}
          {result}
        </div>
      </div>
    );
  }
}

export default withTranslation(['admin'])(withReadyRouter(ConfirmSignup as any));
