import React from "react";
import { TranslationFunc, withTranslation } from "./withTranslation";
import { Button, Modal } from "react-bootstrap";
import User from "@/types/User";

interface State {
  secet: string;
  qrCode: string;
  showTotpSetup: boolean;
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
      showTotpSetup: false,
    };
  }

  setupTotp = () => {
    User.generateTotp().then((result) => {
      this.setState({
        secet: result.secret,
        qrCode: result.qrCode,
        showTotpSetup: true,
      });
    });
  };

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
