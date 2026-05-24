import React from "react";
import { TranslationFunc, withTranslation } from "./withTranslation";
import RuntimeConfig from "./RuntimeConfig";
import { Alert, Col, Row } from "react-bootstrap";

interface State {}

interface Props {
  t: TranslationFunc;
}

class CloudHint extends React.Component<Props, State> {
  constructor(props: any) {
    super(props);
    this.state = {};
  }

  render() {
    if (RuntimeConfig.INFOS.cloudHosted) {
      return <></>;
    }
    return (
      <Row className="mb-4">
        <Col sm="8">
          <Alert variant="info">
            <p>💎 {this.props.t("cloudHint")}</p>
            <Row>
              <Col sm="6">
                🚀{" "}
                <a
                  href="https://seatsurfing.io/sign-up?paid"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  {this.props.t("cloudHintLink")}
                </a>
              </Col>
              <Col sm="6">
                ❤️{" "}
                <a
                  href="https://github.com/sponsors/seatsurfing"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  {this.props.t("sponsorLink")}
                </a>
              </Col>
            </Row>
          </Alert>
        </Col>
      </Row>
    );
  }
}

export default withTranslation(CloudHint as any);
