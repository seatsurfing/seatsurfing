import React from "react";
import { Button, Modal } from "react-bootstrap";
import { TranslationFunc, withTranslation } from "./withTranslation";

interface Props {
  t: TranslationFunc;
  show: boolean;
  title?: string;
  message: string;
  confirmLabel?: string;
  onConfirm: () => void;
}

class AlertModal extends React.Component<Props> {
  render() {
    return (
      <Modal show={this.props.show} onHide={this.props.onConfirm}>
        {this.props.title ? (
          <Modal.Header closeButton>
            <Modal.Title>{this.props.title}</Modal.Title>
          </Modal.Header>
        ) : null}
        <Modal.Body>{this.props.message}</Modal.Body>
        <Modal.Footer>
          <Button variant="primary" onClick={this.props.onConfirm}>
            {this.props.confirmLabel || this.props.t("ok")}
          </Button>
        </Modal.Footer>
      </Modal>
    );
  }
}

export default withTranslation(AlertModal as any);
