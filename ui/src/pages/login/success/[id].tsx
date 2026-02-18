import React from "react";
import { Form } from "react-bootstrap";
import RuntimeConfig from "@/components/RuntimeConfig";
import Loading from "@/components/Loading";
import { NextRouter } from "next/router";
import withReadyRouter from "@/components/withReadyRouter";
import Ajax from "@/util/Ajax";
import AjaxCredentials from "@/util/AjaxCredentials";

interface State {
  redirect: string | null;
}

interface Props {
  router: NextRouter;
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
            const credentials: AjaxCredentials = {
              accessToken: res.json.accessToken,
              accessTokenExpiry: new Date(
                new Date().getTime() + Ajax.ACCESS_TOKEN_EXPIRY_OFFSET,
              ),
              logoutUrl: res.json.logoutUrl,
              profilePageUrl: res.json.profilePageUrl,
            };
            Ajax.PERSISTER.persistRefreshTokenInLocalStorage(
              res.json.refreshToken,
            );
            Ajax.PERSISTER.updateCredentialsLocalStorage(credentials);
            RuntimeConfig.loadUserAndSettings().then(() => {
              let redirect =
                (this.props.router.query["redir"] as string) || "/search";
              this.setState({
                redirect,
              });
            });
          } else {
            this.setState({
              redirect: "/login/failed/",
            });
          }
        })
        .catch(() => {
          this.setState({
            redirect: "/login/failed/",
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

export default withReadyRouter(LoginSuccess as any);
