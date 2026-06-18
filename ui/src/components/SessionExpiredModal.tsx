import React from "react";
import { TranslationFunc, withTranslation } from "./withTranslation";
import { Button, Modal } from "react-bootstrap";
import Router, { NextRouter } from "next/router";

interface Props {
  t: TranslationFunc;
  show: boolean;
  onHide: () => void;
}

class SessionExpiredModal extends React.Component<Props> {
  toLogin(router: NextRouter) {
    const currentPath = router.asPath;
    router.push(`/login?redir=${encodeURIComponent(currentPath)}`);
  }

  render() {
    return (
      <Modal show={this.props.show} backdrop="static" keyboard={false}>
        <Modal.Header>
          <Modal.Title>{this.props.t("sessionExpiredTitle")}</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <p>{this.props.t("sessionExpiredBody")}</p>
        </Modal.Body>
        <Modal.Footer>
          <Button
            variant="primary"
            onClick={() => {
              this.props.onHide();
              this.toLogin(Router);
            }}
          >
            {this.props.t("sessionExpiredLogin")}
          </Button>
        </Modal.Footer>
      </Modal>
    );
  }
}

export default withTranslation(SessionExpiredModal as any);
