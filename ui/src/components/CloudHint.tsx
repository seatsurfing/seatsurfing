import React from "react";
import { TranslationFunc, withTranslation } from "./withTranslation";
import RuntimeConfig from "./RuntimeConfig";
import { Alert, Col, Row } from "react-bootstrap";

interface State {
  current: number;
}

interface Props {
  t: TranslationFunc;
  lang: string;
}

interface Hint {
  icon: string;
  textKey: string;
  linkKey: string;
  href: string;
}

const ROTATION_INTERVAL_MS = 10000;
const FADE_DURATION_MS = 400;

class CloudHint extends React.Component<Props, State> {
  private rotationTimer?: ReturnType<typeof setInterval>;

  constructor(props: any) {
    super(props);
    this.state = {
      current: 0,
    };
  }

  componentDidMount = () => {
    if (RuntimeConfig.INFOS.cloudHosted) {
      return;
    }
    this.startRotation();
  };

  componentWillUnmount = () => {
    this.stopRotation();
  };

  startRotation = () => {
    this.stopRotation();
    this.rotationTimer = setInterval(() => {
      this.setState((state) => ({
        current: (state.current + 1) % this.getHints().length,
      }));
    }, ROTATION_INTERVAL_MS);
  };

  stopRotation = () => {
    if (this.rotationTimer) {
      clearInterval(this.rotationTimer);
      this.rotationTimer = undefined;
    }
  };

  onIndicatorClick = (index: number) => {
    this.setState({ current: index });
    // Restart the timer so the manually selected hint stays visible for the full interval.
    this.startRotation();
  };

  getHints = (): Hint[] => {
    const langPrefix = this.props.lang === "de" ? "/de" : "";
    return [
      {
        icon: "🚀",
        textKey: "cloudHintCloudText",
        linkKey: "cloudHintCloudLink",
        href: `https://seatsurfing.io${langPrefix}/sign-up?paid`,
      },
      {
        icon: "🧩",
        textKey: "cloudHintPlusText",
        linkKey: "cloudHintPlusLink",
        href: `https://seatsurfing.io${langPrefix}/docs/self-hosted/plus-plugin`,
      },
      {
        icon: "❤️",
        textKey: "cloudHintSponsorText",
        linkKey: "cloudHintSponsorLink",
        href: "https://github.com/sponsors/seatsurfing",
      },
    ];
  };

  render() {
    if (RuntimeConfig.INFOS.cloudHosted) {
      return <></>;
    }
    const hints = this.getHints();
    return (
      <Row className="mb-4">
        <Col sm="12" xl="8">
          <Alert variant="info">
            <div className="d-flex" style={{ gap: "0.75rem" }}>
              {/* All hints share the same grid cell, so the container keeps the
                  height of the tallest one and does not jump while rotating. */}
              <div style={{ display: "grid", flex: 1, minWidth: 0 }}>
                {hints.map((hint, i) => {
                  const active = i === this.state.current;
                  return (
                    <div
                      key={hint.linkKey}
                      aria-hidden={!active}
                      style={{
                        gridArea: "1 / 1",
                        display: "flex",
                        flexDirection: "column",
                        opacity: active ? 1 : 0,
                        visibility: active ? "visible" : "hidden",
                        pointerEvents: active ? "auto" : "none",
                        transition: `opacity ${FADE_DURATION_MS}ms ease-in-out, visibility ${FADE_DURATION_MS}ms`,
                      }}
                    >
                      <p>{this.props.t(hint.textKey)}</p>
                      <p className="mb-0 mt-auto">
                        {hint.icon}{" "}
                        <a
                          href={hint.href}
                          target="_blank"
                          rel="noopener noreferrer"
                        >
                          {this.props.t(hint.linkKey)}
                        </a>
                      </p>
                    </div>
                  );
                })}
              </div>
              <div
                className="d-flex flex-column justify-content-center"
                style={{ gap: "0.375rem" }}
              >
                {hints.map((h, i) => (
                  <button
                    key={h.linkKey}
                    type="button"
                    aria-label={this.props.t(h.linkKey)}
                    aria-current={i === this.state.current}
                    onClick={() => this.onIndicatorClick(i)}
                    style={{
                      width: "9px",
                      height: "9px",
                      padding: 0,
                      borderRadius: "50%",
                      border: "1px solid currentColor",
                      backgroundColor:
                        i === this.state.current
                          ? "currentColor"
                          : "transparent",
                      cursor: "pointer",
                      opacity: 0.75,
                    }}
                  />
                ))}
              </div>
            </div>
          </Alert>
        </Col>
      </Row>
    );
  }
}

export default withTranslation(CloudHint as any);
