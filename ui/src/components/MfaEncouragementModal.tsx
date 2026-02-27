import React from "react";
import { TranslationFunc, withTranslation } from "./withTranslation";
import { Button, Modal } from "react-bootstrap";

interface Props {
  t: TranslationFunc;
  show: boolean;
  onHide: () => void;
  onSetup: () => void;
}

class MfaEncouragementModal extends React.Component<Props> {
  render() {
    return (
      <Modal
        show={this.props.show}
        onHide={this.props.onHide}
        backdrop={true}
        keyboard={true}
      >
        <Modal.Header closeButton={true}>
          <Modal.Title>{this.props.t("mfaEncouragementTitle")}</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <p>{this.props.t("mfaEncouragementBody")}</p>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="primary" onClick={this.props.onSetup}>
            {this.props.t("mfaEncouragementSetup")}
          </Button>
          <Button variant="link" onClick={this.props.onHide}>
            {this.props.t("mfaEncouragementLater")}
          </Button>
        </Modal.Footer>
      </Modal>
    );
  }
}

export default withTranslation(MfaEncouragementModal as any);
