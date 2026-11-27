import React from "react";
import { Alert, Button, Modal } from "react-bootstrap";
import { TranslationFunc, withTranslation } from "./withTranslation";

interface Props {
  t: TranslationFunc;
  show: boolean;
  url: string;
  onHide: () => void;
}

interface State {
  loadedUrl: string;
  errorUrl: string;
  width: number;
  height: number;
}

class LogoPreviewModal extends React.Component<Props, State> {
  constructor(props: any) {
    super(props);
    this.state = { loadedUrl: "", errorUrl: "", width: 0, height: 0 };
  }

  onLoad = (e: React.SyntheticEvent<HTMLImageElement>) => {
    const img = e.currentTarget;
    this.setState({
      loadedUrl: this.props.url,
      width: img.naturalWidth,
      height: img.naturalHeight,
    });
  };

  renderBody = () => {
    if (this.state.errorUrl === this.props.url) {
      return (
        <Alert variant="danger" className="mb-0">
          {this.props.t("customLogoPreviewError")}
        </Alert>
      );
    }
    return (
      <>
        <div className="custom-logo-preview bg-body-tertiary border rounded">
          <img
            key={this.props.url}
            src={this.props.url}
            alt={this.props.t("preview")}
            onLoad={this.onLoad}
            onError={() => this.setState({ errorUrl: this.props.url })}
          />
        </div>
        <p className="text-muted small text-center mt-2 mb-0">
          {this.state.loadedUrl === this.props.url
            ? this.props.t("imageDimensions", {
                width: this.state.width,
                height: this.state.height,
              })
            : this.props.t("loading")}
        </p>
      </>
    );
  };

  render() {
    return (
      <Modal show={this.props.show} onHide={this.props.onHide}>
        <Modal.Body>{this.props.show ? this.renderBody() : null}</Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={this.props.onHide}>
            {this.props.t("close")}
          </Button>
        </Modal.Footer>
      </Modal>
    );
  }
}

export default withTranslation(LogoPreviewModal as any);
