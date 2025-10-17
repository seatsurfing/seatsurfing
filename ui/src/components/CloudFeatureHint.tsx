import React from "react";
import { TranslationFunc, withTranslation } from "./withTranslation";
import { Alert, Col, Row } from "react-bootstrap";
import Link from "next/link";
import PremiumFeatureIcon from "./PremiumFeatureIcon";

interface State {}

interface Props {
  t: TranslationFunc;
}

class CloudFeatureHint extends React.Component<Props, State> {
  render() {
    return (
      <Row className="mb-4">
        <Col sm="8">
          <Alert variant="info">
            <p style={{ fontWeight: "bold" }}>
              <PremiumFeatureIcon
                style={{ marginLeft: "0px", marginRight: "5px" }}
              />
              {this.props.t("cloudFeatureHeadline")}
            </p>
            <p>
              <Link href="/admin/plugin/subscription/">
                {this.props.t("cloudUpgradeHintText")}
              </Link>{" "}
              🚀
            </p>
          </Alert>
        </Col>
      </Row>
    );
  }
}

export default withTranslation(CloudFeatureHint as any);
