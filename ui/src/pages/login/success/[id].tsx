import React from "react";
import { Form } from "react-bootstrap";
import RuntimeConfig from "@/components/RuntimeConfig";
import Loading from "@/components/Loading";
import { NextRouter } from "next/router";
import withReadyRouter from "@/components/withReadyRouter";
import Ajax from "@/util/Ajax";
import AjaxCredentials from "@/util/AjaxCredentials";
import JwtDecoder from "@/util/JwtDecoder";
import Validation from "@/util/Validation";

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

  loadData = async () => {
    const { id } = this.props.router.query;
    if (id) {
      try {
        const res = await Ajax.get("/auth/verify/" + id, () => true);
        if (res.json && res.json.accessToken) {
          const credentials: AjaxCredentials = {
            accessToken: res.json.accessToken,
            accessTokenExpiry: JwtDecoder.getExpiryDate(res.json.accessToken),
            logoutUrl: res.json.logoutUrl,
            profilePageUrl: res.json.profilePageUrl,
          };
          Ajax.PERSISTER.persistRefreshTokenInLocalStorage(
            res.json.refreshToken,
          );
          Ajax.PERSISTER.updateCredentialsLocalStorage(credentials);
          await RuntimeConfig.loadUserAndSettings();
          const redirParam = this.props.router.query["redir"] as string;
          const redirect =
            redirParam && Validation.isRelativeUrl(redirParam)
              ? redirParam
              : "/search";
          this.setState({ redirect });
        } else {
          this.setState({ redirect: "/login/failed/" });
        }
      } catch {
        this.setState({ redirect: "/login/failed/" });
      }
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
