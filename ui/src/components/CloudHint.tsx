import React from "react";
import { TranslationFunc, withTranslation } from "./withTranslation";
import RuntimeConfig from "./RuntimeConfig";
import { Alert, Col, Row } from "react-bootstrap";

interface State {}

interface Props {
  t: TranslationFunc;
  lang: string;
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
    const langPrefix = this.props.lang === "de" ? "/de" : "";
    const links: { icon: string; key: string; href: string }[] = [
      {
        icon: "✨",
        key: "cloudHintFeaturesLink",
        href: `https://seatsurfing.io${langPrefix}/features`,
      },
      {
        icon: "🚀",
        key: "cloudHintCloudLink",
        href: `https://seatsurfing.io${langPrefix}/sign-up?paid`,
      },
      {
        icon: "🧩",
        key: "cloudHintPlusLink",
        href: `https://seatsurfing.io${langPrefix}/docs/self-hosted/plus-plugin`,
      },
      {
        icon: "❤️",
        key: "cloudHintSponsorLink",
        href: "https://github.com/sponsors/seatsurfing",
      },
    ];
    return (
      <Row className="mb-4">
        <Col sm="12" xl="8">
          <Alert variant="info">
            <p>💎 {this.props.t("cloudHint")}</p>
            <ul className="list-unstyled mb-0">
              {links.map((link) => (
                <li key={link.key}>
                  {link.icon}{" "}
                  <a href={link.href} target="_blank" rel="noopener noreferrer">
                    {this.props.t(link.key)}
                  </a>
                </li>
              ))}
            </ul>
          </Alert>
        </Col>
      </Row>
    );
  }
}

export default withTranslation(CloudHint as any);
