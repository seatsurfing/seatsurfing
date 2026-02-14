import React from "react";
import { TranslationFunc, withTranslation } from "./withTranslation";
import { Button } from "react-bootstrap";
import User from "@/types/User";
import TotpSetupModal from "./TotpSetupModal";
import RuntimeConfig from "./RuntimeConfig";

interface State {
  qrCode: string;
  stateId: string;
  showTotpSetup: boolean;
  totpEnabled?: boolean;
}

interface Props {
  t: TranslationFunc;
  hidden?: boolean | undefined;
}

class TotpSettings extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      qrCode: "",
      stateId: "",
      showTotpSetup: false,
      totpEnabled: RuntimeConfig.INFOS.totpEnabled,
    };
  }

  setupTotp = () => {
    User.generateTotp().then((result) => {
      this.setState({
        qrCode: result.qrCode,
        stateId: result.stateId,
        showTotpSetup: true,
      });
    });
  };

  onTotpSuccess = () => {
    RuntimeConfig.INFOS.totpEnabled = true;
    this.setState({
      showTotpSetup: false,
      totpEnabled: true,
    });
  };

  disableTotp = () => {
    if (window.confirm(this.props.t("disableTotpConfirm"))) {
      User.disableTotp().then(() => {
        RuntimeConfig.INFOS.totpEnabled = false;
        this.setState({ totpEnabled: false });
      });
    }
  };

  render() {
    return (
      <div hidden={this.props.hidden}>
        <TotpSetupModal
          show={this.state.showTotpSetup}
          qrCode={this.state.qrCode}
          stateId={this.state.stateId}
          onHide={() => this.setState({ showTotpSetup: false })}
          onSuccess={this.onTotpSuccess}
          canClose={true}
        />
        <h5 className="mt-5">{this.props.t("totp")}</h5>
        <p>{this.props.t("totpHint")}</p>
        <Button
          variant="primary"
          onClick={() => this.setupTotp()}
          hidden={this.state.totpEnabled}
        >
          {this.props.t("enableTotp")}
        </Button>
        <Button
          variant="danger"
          onClick={() => this.disableTotp()}
          hidden={!this.state.totpEnabled}
        >
          {this.props.t("disableTotp")}
        </Button>
      </div>
    );
  }
}

export default withTranslation(TotpSettings as any);
