import React from "react";
import Loading from "../../../components/Loading";
import { Form } from "react-bootstrap";
import { Ajax, AjaxCredentials, JwtDecoder, User } from "seatsurfing-commons";
import { NextRouter } from "next/router";
import withReadyRouter from "@/components/withReadyRouter";
import RuntimeConfig from "@/components/RuntimeConfig";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";

interface State {
  redirect: string | null;
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
}

class LoginSuccess extends React.Component<Props, State> {
  constructor(props: any) {
    super(props);
    this.state = {
      redirect: null,
    };
  }

  componentDidMount = () => {
    this.loadData();
  };

  loadData = () => {
    const { id } = this.props.router.query;
    if (id) {
      return Ajax.get("/auth/verify/" + id)
        .then((res) => {
          if (res.json && res.json.accessToken) {
            let jwtPayload = JwtDecoder.getPayload(res.json.accessToken);
            if (jwtPayload.role < User.UserRoleSpaceAdmin) {
              this.setState({
                redirect: "/login/failed",
              });
              return;
            }
            const credentials: AjaxCredentials = {
              accessToken: res.json.accessToken,
              accessTokenExpiry: new Date(
                new Date().getTime() + Ajax.ACCESS_TOKEN_EXPIRY_OFFSET
              ),
              logoutUrl: res.json.logoutUrl,
            };
            Ajax.PERSISTER.persistRefreshTokenInLocalStorage(
              res.json.refreshToken
            );
            Ajax.PERSISTER.updateCredentialsSessionStorage(credentials);
            RuntimeConfig.loadUserAndSettings().then(() => {
              this.setState({
                redirect: "/dashboard",
              });
            });
          } else {
            this.setState({
              redirect: "/login/failed",
            });
          }
        })
        .catch(() => {
          this.setState({
            redirect: "/login/failed",
          });
        });
    }
  };

  render() {
    if (this.state.redirect != null) {
      this.props.router.push(this.state.redirect);
      return <></>;
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

export default withTranslation(withReadyRouter(LoginSuccess as any));
