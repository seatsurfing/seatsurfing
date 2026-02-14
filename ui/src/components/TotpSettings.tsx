import React from "react";
import { TranslationFunc, withTranslation } from "./withTranslation";
import { Button, Form, Modal } from "react-bootstrap";
import User from "@/types/User";
import TotpInput from "./TotpInput";
import RuntimeConfig from "./RuntimeConfig";

interface State {
  secet: string;
  qrCode: string;
  stateId: string;
  showTotpSetup: boolean;
  code: string;
  submitting: boolean;
  invalid?: boolean;
  totpEnabled?: boolean;
}

interface Props {
  t: TranslationFunc;
  hidden?: boolean | undefined;
}

class TotpSettings extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      secet: "",
      qrCode: "",
      stateId: "",
      showTotpSetup: false,
      code: "",
      submitting: false,
      invalid: false,
      totpEnabled: RuntimeConfig.INFOS.totpEnabled,
    };
  }

  setupTotp = () => {
    User.generateTotp().then((result) => {
      this.setState({
        secet: result.secret,
        qrCode: result.qrCode,
        stateId: result.stateId,
        showTotpSetup: true,
      });
    });
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
            <p>{this.props.t("noQrCodeTotpHint")}</p>
            <p>{this.state.secet}</p>
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
