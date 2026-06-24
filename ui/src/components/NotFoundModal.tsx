import React from "react";
import { TranslationFunc, withTranslation } from "./withTranslation";
import { Button, Modal } from "react-bootstrap";

interface Props {
  t: TranslationFunc;
  show: boolean;
  onHide: () => void;
}

class NotFoundModal extends React.Component<Props> {
  render() {
    return (
      <Modal
        show={this.props.show}
        onHide={this.props.onHide}
        backdrop="static"
        keyboard={false}
      >
        <Modal.Header>
          <Modal.Title>{this.props.t("notFoundTitle")}</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <p>{this.props.t("notFoundBody")}</p>
        </Modal.Body>
        <Modal.Footer>
          <Button
            variant="secondary"
            onClick={() => {
              window.location.href = "/";
            }}
          >
            {this.props.t("home")}
          </Button>
          <Button variant="primary" onClick={() => window.location.reload()}>
            {this.props.t("reload")}
          </Button>
        </Modal.Footer>
      </Modal>
    );
  }
}

export default withTranslation(NotFoundModal as any);
