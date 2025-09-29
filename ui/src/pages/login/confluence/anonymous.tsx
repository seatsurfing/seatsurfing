import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import React from "react";
import { Form, Alert } from "react-bootstrap";

interface State {}

interface Props {
  t: TranslationFunc;
}

class ConfluenceAnonymous extends React.Component<Props, State> {
  render() {
    return (
      <div className="container-signin">
        <Form className="form-signin">
          <Alert variant="danger">
            {this.props.t("errorConfluenceAnonymous")}
          </Alert>
          <p>{this.props.t("confluenceAnonymousHint")}</p>
        </Form>
      </div>
    );
  }
}

export default withTranslation(ConfluenceAnonymous as any);
