import React from "react";
import { Form, Alert } from "react-bootstrap";
import Link from "next/link";
import { NextRouter } from "next/router";
import withReadyRouter from "@/components/withReadyRouter";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import RuntimeConfig from "@/components/RuntimeConfig";

interface State {}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
}

class LoginFailed extends React.Component<Props, State> {
  render() {
    let backButton = <></>;
    if (!RuntimeConfig.EMBEDDED) {
      backButton = (
        <Link className="btn btn-primary" href="/login">
          {this.props.t("back")}
        </Link>
      );
    }

    const reason = this.props.router.query["reason"];
    const error = this.props.router.query["error"];
    const errorDescription = this.props.router.query["error_description"];

    return (
      <div className="container-signin">
        <Form className="form-signin">
          <Alert variant="danger">{this.props.t("errorLoginFailed")}</Alert>
          <p>{this.props.t("loginFailedDescription")}</p>
          <p>
            {reason && `${this.props.t("code")}: ${reason}`}
            {error && ` (${error})`}
            {errorDescription && ` ${errorDescription}`}
          </p>
          {backButton}
        </Form>
      </div>
    );
  }
}

export default withTranslation(withReadyRouter(LoginFailed as any) as any);
