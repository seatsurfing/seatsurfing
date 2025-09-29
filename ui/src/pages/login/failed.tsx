import React from "react";
import { Form, Alert } from "react-bootstrap";
import Link from "next/link";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import RuntimeConfig from "@/components/RuntimeConfig";

interface State {}

interface Props {
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

    return (
      <div className="container-signin">
        <Form className="form-signin">
          <Alert variant="danger">{this.props.t("errorLoginFailed")}</Alert>
          <p>{this.props.t("loginFailedDescription")}</p>
          {backButton}
        </Form>
      </div>
    );
  }
}

export default withTranslation(LoginFailed as any);
