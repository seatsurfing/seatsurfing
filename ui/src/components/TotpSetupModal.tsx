import React from "react";
import { TranslationFunc, withTranslation } from "./withTranslation";
import { Button, Form, Modal } from "react-bootstrap";
import User from "@/types/User";
import TotpInput from "./TotpInput";
import RuntimeConfig from "./RuntimeConfig";

interface State {
  code: string;
  submitting: boolean;
  invalid?: boolean;
  showSecret: boolean;
  secret: string;
  secretLoading: boolean;
}

interface Props {
  t: TranslationFunc;
  show: boolean;
  qrCode: string;
  stateId: string;
  onHide: () => void;
  onSuccess: () => void;
  canClose?: boolean;
}

class TotpSetupModal extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      code: "",
      submitting: false,
      invalid: false,
      showSecret: false,
      secret: "",
      secretLoading: false,
    };
  }

  loadSecret = () => {
    this.setState({ secretLoading: true });
    User.getTotpSecret(this.props.stateId)
      .then((secret) => {
        this.setState({
          secret: secret,
          showSecret: true,
          secretLoading: false,
        });
      })
      .catch(() => {
        this.setState({ secretLoading: false });
      });
  };

  copySecret = () => {
    if (this.state.secret) {
      navigator.clipboard.writeText(this.state.secret);
    }
  };

  formatSecret = (secret: string): string => {
    // Format secret in groups of 4 for readability
    return secret.match(/.{1,4}/g)?.join(" ") || secret;
  };

  onSubmit = (e: any) => {
    e.preventDefault();
    this.setState({ submitting: true });
    User.validateTotp(this.props.stateId, this.state.code)
      .then(() => {
        this.setState({
          code: "",
          submitting: false,
          invalid: false,
        });
        this.props.onSuccess();
      })
      .catch(() => {
        this.setState({ invalid: true, submitting: false });
      });
  };

  render() {
    const canClose = this.props.canClose !== false;

    return (
      <Modal
        show={this.props.show}
        onHide={canClose ? this.props.onHide : undefined}
        backdrop={canClose ? true : "static"}
        keyboard={canClose}
      >
        <Modal.Header closeButton={canClose}>
          <Modal.Title>{this.props.t("enableTotp")}</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <p hidden={this.props.canClose}>
            {this.props.t("totpEnforcementMessage")}
          </p>
          <p>{this.props.t("enableTotpHint")}</p>
          <p>
            <img src={"data:image/png;base64," + this.props.qrCode} alt="" />
          </p>

          {/* Manual entry section */}
          <div className="mt-3">
            {!this.state.showSecret ? (
              <Button
                variant="link"
                size="sm"
                onClick={this.loadSecret}
                disabled={this.state.secretLoading}
              >
                {this.state.secretLoading
                  ? this.props.t("loading")
                  : this.props.t("cantScanQrCode")}
              </Button>
            ) : (
              <div className="alert alert-warning" role="alert">
                <small>
                  <strong>⚠️ {this.props.t("securityWarning")}</strong>
                  <br />
                  {this.props.t("totpSecretWarning")}
                </small>
                <div className="mt-2 mb-2">
                  <code
                    style={{
                      fontSize: "1.1rem",
                      letterSpacing: "0.1em",
                      userSelect: "all",
                      display: "block",
                      padding: "8px",
                      backgroundColor: "#f8f9fa",
                      border: "1px solid #dee2e6",
                      borderRadius: "4px",
                    }}
                  >
                    {this.formatSecret(this.state.secret)}
                  </code>
                </div>
                <Button
                  variant="outline-secondary"
                  size="sm"
                  onClick={this.copySecret}
                >
                  📋 {this.props.t("copyToClipboard")}
                </Button>
              </div>
            )}
          </div>

          <Form onSubmit={this.onSubmit}>
            <Form.Group>
              <TotpInput
                value={this.state.code}
                onChange={(value: string) =>
                  this.setState({ code: value, invalid: false })
                }
                onComplete={(value: string) => {
                  this.setState({ code: value, invalid: false }, () => {
                    this.onSubmit(new Event("submit") as any);
                  });
                }}
                required={true}
                invalid={this.state.invalid}
                disabled={this.state.submitting}
              />
            </Form.Group>
          </Form>
        </Modal.Body>
        <Modal.Footer
          hidden={canClose}
          style={{ justifyContent: "flex-start" }}
        >
          <Button variant="secondary" onClick={() => RuntimeConfig.logOut()}>
            {this.props.t("logout")}
          </Button>
        </Modal.Footer>
      </Modal>
    );
  }
}

export default withTranslation(TotpSetupModal as any);
