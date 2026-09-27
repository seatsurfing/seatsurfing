import React from "react";
import { Alert, Form } from "react-bootstrap";
import { TranslationFunc, withTranslation } from "./withTranslation";

interface Props {
  t: TranslationFunc;
}

class FrameBlocked extends React.Component<Props> {
  render() {
    const href = typeof window !== "undefined" ? window.location.href : "/ui/";
    return (
      <div className="container-signin">
        <Form className="form-signin">
          <Alert variant="warning">{this.props.t("frameBlockedTitle")}</Alert>
          <p>{this.props.t("frameBlockedBody")}</p>
          <a
            className="btn btn-primary"
            href={href}
            target="_blank"
            rel="noopener noreferrer"
          >
            {this.props.t("openInNewWindow")}
          </a>
        </Form>
      </div>
    );
  }
}

export default withTranslation(FrameBlocked as any);
