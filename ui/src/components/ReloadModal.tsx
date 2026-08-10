import React from "react";
import { TranslationFunc, withTranslation } from "./withTranslation";
import { Button, Modal } from "react-bootstrap";

interface Props {
  t: TranslationFunc;
  show: boolean;
  title: string;
}

class ReloadModal extends React.Component<Props> {
  render() {
    return (
      <Modal
        show={this.props.show}
        onHide={() => window.location.reload()}
        backdrop="static"
        keyboard={false}
      >
        <Modal.Header>
          <Modal.Title>{this.props.title}</Modal.Title>
        </Modal.Header>
        <Modal.Body>{this.props.t("entryUpdatedReloadRequired")}</Modal.Body>
        <Modal.Footer>
          <Button variant="primary" onClick={() => window.location.reload()}>
            {this.props.t("reload")}
          </Button>
        </Modal.Footer>
      </Modal>
    );
  }
}

export default withTranslation(ReloadModal as any);
