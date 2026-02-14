import React from "react";
import { TranslationFunc, withTranslation } from "./withTranslation";
import { Button, Form, Modal } from "react-bootstrap";
import User from "@/types/User";
import TotpInput from "./TotpInput";
import RuntimeConfig from "./RuntimeConfig";

interface State {
  qrCode: string;
  stateId: string;
  showTotpSetup: boolean;
  code: string;
  submitting: boolean;
  invalid?: boolean;
  totpEnabled?: boolean;
  showSecret: boolean;
  secret: string;
  secretLoading: boolean;
}

interface Props {
  t: TranslationFunc;
  hidden?: boolean | undefined;
}

class TotpSettings extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      qrCode: "",
      stateId: "",
      showTotpSetup: false,
      code: "",
      submitting: false,
      invalid: false,
      totpEnabled: RuntimeConfig.INFOS.totpEnabled,
      showSecret: false,
      secret: "",
      secretLoading: false,
    };
  }

  setupTotp = () => {
    User.generateTotp().then((result) => {
      this.setState({
        qrCode: result.qrCode,
        stateId: result.stateId,
        showTotpSetup: true,
        showSecret: false,
        secret: "",
      });
    });
  };

  loadSecret = () => {
    this.setState({ secretLoading: true });
    User.getTotpSecret(this.state.stateId)
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
    User.validateTotp(this.state.stateId, this.state.code)
      .then(() => {
        RuntimeConfig.INFOS.totpEnabled = true;
        this.setState({
          showTotpSetup: false,
          code: "",
          submitting: false,
          totpEnabled: true,
        });
      })
      .catch(() => {
        this.setState({ invalid: true, submitting: false });
      });
  };

  disableTotp = () => {
    if (window.confirm(this.props.t("disableTotpConfirm"))) {
      User.disableTotp().then(() => {
        RuntimeConfig.INFOS.totpEnabled = false;
        this.setState({ totpEnabled: false });
      });
    }
  };

  render() {
    return (
      <div hidden={this.props.hidden}>
        <Modal
          show={this.state.showTotpSetup}
          onHide={() => this.setState({ showTotpSetup: false })}
        >
          <Modal.Header closeButton>
            <Modal.Title>{this.props.t("enableTotp")}</Modal.Title>
          </Modal.Header>
          <Modal.Body>
            <p>{this.props.t("enableTotpHint")}</p>
            <p>
              <img src={"data:image/png;base64," + this.state.qrCode} alt="" />
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
        </Modal>
        <h5 className="mt-5">{this.props.t("totp")}</h5>
        <p>{this.props.t("totpHint")}</p>
        <Button
          variant="primary"
          onClick={() => this.setupTotp()}
          hidden={this.state.totpEnabled}
        >
          {this.props.t("enableTotp")}
        </Button>
        <Button
          variant="danger"
          onClick={() => this.disableTotp()}
          hidden={!this.state.totpEnabled}
        >
          {this.props.t("disableTotp")}
        </Button>
      </div>
    );
  }
}

export default withTranslation(TotpSettings as any);
