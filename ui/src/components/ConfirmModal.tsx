import React from "react";
import { Button, Modal } from "react-bootstrap";
import { TranslationFunc, withTranslation } from "./withTranslation";

interface Props {
  t: TranslationFunc;
  show: boolean;
  title?: string;
  message: React.ReactNode;
  confirmLabel?: string;
  cancelLabel?: string;
  confirmVariant?: string;
  onConfirm: () => void;
  onCancel: () => void;
}

class ConfirmModal extends React.Component<Props> {
  private confirmClicked = false;

  componentDidUpdate(prevProps: Props) {
    // Reset the guard each time the modal is opened.
    if (!prevProps.show && this.props.show) {
      this.confirmClicked = false;
    }
  }

  onConfirmClicked = () => {
    if (this.confirmClicked) {
      return;
    }
    this.confirmClicked = true;
    this.props.onConfirm();
  };

  render() {
    return (
      <Modal show={this.props.show} onHide={this.props.onCancel}>
        {this.props.title ? (
          <Modal.Header closeButton>
            <Modal.Title>{this.props.title}</Modal.Title>
          </Modal.Header>
        ) : null}
        <Modal.Body>{this.props.message}</Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={this.props.onCancel}>
            {this.props.cancelLabel || this.props.t("cancel")}
          </Button>
          <Button
            variant={this.props.confirmVariant || "danger"}
            onClick={this.onConfirmClicked}
          >
            {this.props.confirmLabel || this.props.t("ok")}
          </Button>
        </Modal.Footer>
      </Modal>
    );
  }
}

export default withTranslation(ConfirmModal as any);
