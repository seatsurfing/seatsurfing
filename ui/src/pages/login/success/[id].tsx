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

interface State {}

interface Props {
  router: NextRouter;
}

// The verification token is single-use, so concurrent mounts (e.g. React
// StrictMode mounting the component twice) must share one request per ID
// instead of firing a second one that fails with 404.
const verifyRequests: Map<string, ReturnType<typeof Ajax.get>> = new Map();

class LoginSuccess extends React.Component<Props, State> {
  componentDidMount = () => {
    this.loadData();
  };

  loadData = async () => {
    const id = this.props.router.query["id"] as string;
    if (!id) {
      return;
    }
    let request = verifyRequests.get(id);
    if (!request) {
      request = Ajax.get("/auth/verify/" + id, () => true);
      verifyRequests.set(id, request);
    }
    try {
      const res = await request;
      if (res.json && res.json.accessToken) {
        const credentials: AjaxCredentials = {
          accessToken: res.json.accessToken,
          accessTokenExpiry: JwtDecoder.getExpiryDate(res.json.accessToken),
          logoutUrl: res.json.logoutUrl,
          profilePageUrl: res.json.profilePageUrl,
        };
        Ajax.PERSISTER.persistRefreshTokenInLocalStorage(res.json.refreshToken);
        Ajax.PERSISTER.updateCredentialsLocalStorage(credentials);
        await RuntimeConfig.loadUserAndSettings();
        const redirParam = this.props.router.query["redir"] as string;
        const redirect =
          redirParam && Validation.isRelativeUrl(redirParam)
            ? redirParam
            : "/search";
        this.props.router.push(redirect);
      } else {
        verifyRequests.delete(id);
        this.props.router.push("/login/failed/");
      }
    } catch {
      verifyRequests.delete(id);
      this.props.router.push("/login/failed/");
    }
  };

  render() {
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
