import React from 'react';
import Loading from '../../../components/Loading';
import { Form } from 'react-bootstrap';
import { Ajax, JwtDecoder, User } from 'flexspace-commons';
import { NextRouter } from 'next/router';
import { WithTranslation, withTranslation } from 'next-i18next';
import withReadyRouter from '@/components/withReadyRouter';

interface State {
  redirect: string | null
}

interface Props extends WithTranslation {
  router: NextRouter
}

class LoginSuccess extends React.Component<Props, State> {
  constructor(props: any) {
    super(props);
    this.state = {
      redirect: null
    };
  }

  componentDidMount = () => {
    this.loadData();
  }

  loadData = () => {
    const { id } = this.props.router.query;
    if (id) {
      return Ajax.get("/auth/verify/" + id).then(res => {
        if (res.json && res.json.accessToken) {
          let jwtPayload = JwtDecoder.getPayload(res.json.accessToken);
          if (jwtPayload.role < User.UserRoleSpaceAdmin) {
              this.setState({
              redirect: "/login/failed"
            });
            return;
          }
          Ajax.CREDENTIALS = {
            accessToken: res.json.accessToken,
            refreshToken: res.json.refreshToken,
            accessTokenExpiry: new Date(new Date().getTime() + Ajax.ACCESS_TOKEN_EXPIRY_OFFSET),
            logoutUrl: res.json.logoutUrl,
          };
          if (res.json.longLived) {
            Ajax.PERSISTER.persistRefreshTokenInLocalStorage(Ajax.CREDENTIALS);
          }
          Ajax.PERSISTER.updateCredentialsSessionStorage(Ajax.CREDENTIALS).then(() => {
            this.setState({
              redirect: "/dashboard"
            });
          });
        } else {
          this.setState({
            redirect: "/login/failed"
          });
        }
      }).catch(() => {
        this.setState({
          redirect: "/login/failed"
        });
      });
    }
  }

  render() {
    if (this.state.redirect != null) {
      this.props.router.push(this.state.redirect);
      return <></>
    }

    return (
      <div className="container-signin">
        <Form className="form-signin">
          <Loading />
        </Form>
      </div>
    );
  }
}

export default withTranslation(['admin'])(withReadyRouter(LoginSuccess as any));
