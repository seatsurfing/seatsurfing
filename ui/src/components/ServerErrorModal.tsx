import React from "react";
import { TranslationFunc, withTranslation } from "./withTranslation";
import { Button, Modal } from "react-bootstrap";

interface Props {
  t: TranslationFunc;
  show: boolean;
  onHide: () => void;
}

class ServerErrorModal extends React.Component<Props> {
  render() {
    return (
      <Modal
        show={this.props.show}
        onHide={this.props.onHide}
        backdrop={true}
        keyboard={true}
      >
        <Modal.Header closeButton={true}>
          <Modal.Title>{this.props.t("serverErrorTitle")}</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <p>{this.props.t("serverErrorBody")}</p>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="primary" onClick={this.props.onHide}>
            {this.props.t("serverErrorClose")}
          </Button>
        </Modal.Footer>
      </Modal>
    );
  }
}

export default withTranslation(ServerErrorModal as any);
