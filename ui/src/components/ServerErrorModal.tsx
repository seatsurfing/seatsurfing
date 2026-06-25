import React from "react";
import { TranslationFunc, withTranslation } from "./withTranslation";
import { Button, Modal } from "react-bootstrap";

interface Props {
  t: TranslationFunc;
  show: boolean;
}

class ServerErrorModal extends React.Component<Props> {
  render() {
    return (
      <Modal show={this.props.show} backdrop="static" keyboard={false}>
        <Modal.Header>
          <Modal.Title>{this.props.t("serverErrorTitle")}</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <p>{this.props.t("serverErrorBody")}</p>
        </Modal.Body>
        <Modal.Footer>
          <Button
            variant="secondary"
            onClick={() => {
              window.location.href = "/ui/";
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

export default withTranslation(ServerErrorModal as any);
