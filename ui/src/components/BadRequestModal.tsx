import React from "react";
import { TranslationFunc, withTranslation } from "./withTranslation";
import { Button, Modal } from "react-bootstrap";

interface Props {
  t: TranslationFunc;
  show: boolean;
}

class BadRequestModal extends React.Component<Props> {
  render() {
    return (
      <Modal
        show={this.props.show}
        onHide={this.props.onHide}
        backdrop="static"
        keyboard={false}
      >
        <Modal.Header>
          <Modal.Title>{this.props.t("badRequestTitle")}</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <p>{this.props.t("badRequestBody")}</p>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="primary" onClick={() => window.location.reload()}>
            {this.props.t("reload")}
          </Button>
        </Modal.Footer>
      </Modal>
    );
  }
}

export default withTranslation(BadRequestModal as any);
