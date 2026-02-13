import React from "react";
import { TranslationFunc, withTranslation } from "./withTranslation";
import { Button, Form, Modal } from "react-bootstrap";
import User from "@/types/User";

interface State {
  secet: string;
  qrCode: string;
  stateId: string;
  showTotpSetup: boolean;
  code: string;
  submitting: boolean;
}

interface Props {
  t: TranslationFunc;
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
    User.validateTotp(this.state.stateId, this.state.code).then(() => {
      this.setState({ showTotpSetup: false, code: "", submitting: false });
    });
  }

  render() {
    return (
      <div>
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
            <p>{this.state.secet}</p>
            <Form onSubmit={this.onSubmit}>
              <Form.Group>
                <Form.Label>{this.props.t("totp")}</Form.Label>
            <Form.Control
              type="text"
              value={this.state.code}
              required={true}
              minLength={6}
              maxLength={6}
              onChange={(e) => this.setState({ code: e.target.value })}
            />
            </Form.Group>
            <Button
                            className="margin-top-15"
                            type="submit"
                            disabled={this.state.submitting}
                          >
                            {this.props.t("save")}
                          </Button>
            </Form>
          </Modal.Body>
        </Modal>
        <h5 className="mt-5">{this.props.t("totp")}</h5>
        <Button variant="primary" onClick={() => this.setupTotp()}>
          {this.props.t("enableTotp")}
        </Button>
      </div>
    );
  }
}

export default withTranslation(TotpSettings as any);
