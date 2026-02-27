import React from "react";
import { TranslationFunc, withTranslation } from "./withTranslation";
import { Button, ListGroup, InputGroup, Form } from "react-bootstrap";
import Passkey, {
  PasskeyListItem,
  prepareCreationOptions,
  serializeAttestationResponse,
} from "@/types/Passkey";
import Formatting from "@/util/Formatting";

interface State {
  passkeys: PasskeyListItem[];
  loading: boolean;
  registering: boolean;
  newName: string;
  error: string;
  passkeyPlatformAvailable: boolean;
}

interface Props {
  t: TranslationFunc;
  hidden?: boolean;
  onPasskeyAdded?: () => void;
  onPasskeyDeleted?: () => void;
}

class PasskeySettings extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      passkeys: [],
      loading: true,
      registering: false,
      newName: "",
      error: "",
      // Sync pre-check; refined by the async call in componentDidMount (Finding #14)
      passkeyPlatformAvailable: Passkey.isSupported(),
    };
  }

  componentDidMount() {
    if (!this.props.hidden) {
      this.loadPasskeys();
    }
    // Refine the platform authenticator availability check asynchronously (Finding #14)
    Passkey.isPlatformAuthAvailable().then((available) =>
      this.setState({ passkeyPlatformAvailable: available }),
    );
  }

  componentDidUpdate(prevProps: Props) {
    if (prevProps.hidden && !this.props.hidden) {
      this.loadPasskeys();
    }
  }

  loadPasskeys = () => {
    if (!this.state.passkeyPlatformAvailable) {
      this.setState({ loading: false });
      return;
    }
    Passkey.list()
      .then((passkeys) => this.setState({ passkeys, loading: false }))
      .catch(() => this.setState({ loading: false }));
  };

  registerPasskey = async () => {
    const { newName } = this.state;
    if (!newName.trim()) {
      this.setState({ error: this.props.t("passkeyNameRequired") });
      return;
    }
    this.setState({ registering: true, error: "" });
    try {
      const beginResponse = await Passkey.beginRegistration();
      const rawOptions =
        beginResponse.challenge?.publicKey ?? beginResponse.challenge;
      const publicKeyOptions = prepareCreationOptions(rawOptions);
      const credential = await navigator.credentials.create({
        publicKey: publicKeyOptions,
      });
      if (!credential) {
        throw new Error("No credential returned");
      }
      const serialized = serializeAttestationResponse(
        credential as PublicKeyCredential,
      );
      const newPasskey = await Passkey.finishRegistration(
        beginResponse.stateId,
        newName.trim(),
        serialized,
      );
      const passkeys = [...this.state.passkeys, newPasskey];
      this.setState({ passkeys, newName: "", registering: false });
      if (this.props.onPasskeyAdded) {
        this.props.onPasskeyAdded();
      }
    } catch (e: any) {
      this.setState({
        registering: false,
        error: e?.message ?? this.props.t("passkeyRegisterError"),
      });
    }
  };

  deletePasskey = (id: string, name: string) => {
    if (!window.confirm(this.props.t("passkeyDeleteConfirm", { name }))) {
      return;
    }
    Passkey.deletePasskey(id)
      .then(() => {
        const passkeys = this.state.passkeys.filter((p) => p.id !== id);
        this.setState({ passkeys });
        if (this.props.onPasskeyDeleted) {
          this.props.onPasskeyDeleted();
        }
      })
      .catch(() => {
        this.setState({ error: this.props.t("passkeyDeleteError") });
      });
  };

  render() {
    if (!this.state.passkeyPlatformAvailable && !this.state.registering) {
      return null;
    }
    const { passkeys, loading, registering, newName, error } = this.state;
    return (
      <div hidden={this.props.hidden}>
        <h5 className="mt-5">{this.props.t("passkeys")}</h5>
        <p>{this.props.t("passkeysHint")}</p>
        {loading ? null : (
          <>
            {passkeys.length > 0 && (
              <ListGroup className="mb-3">
                {passkeys.map((pk) => (
                  <ListGroup.Item
                    key={pk.id}
                    className="d-flex justify-content-between align-items-center"
                  >
                    <span>
                      <strong>{pk.name}</strong>
                      {pk.lastUsedAt ? (
                        <small className="text-muted ms-2">
                          {Formatting.decodeHtmlEntities(
                            this.props.t("passkeyLastUsed", {
                              date: Formatting.getFormatterShort(false).format(
                                new Date(pk.lastUsedAt),
                              ),
                            }),
                          )}
                        </small>
                      ) : null}
                    </span>
                    <Button
                      variant="outline-danger"
                      size="sm"
                      onClick={() => this.deletePasskey(pk.id, pk.name)}
                    >
                      {this.props.t("delete")}
                    </Button>
                  </ListGroup.Item>
                ))}
              </ListGroup>
            )}
            <InputGroup className="mb-2" style={{ maxWidth: 400 }}>
              <Form.Control
                type="text"
                placeholder={this.props.t("passkeyNamePlaceholder")}
                value={newName}
                onChange={(e) => this.setState({ newName: e.target.value })}
                disabled={registering}
                maxLength={255}
              />
              <Button
                variant="primary"
                onClick={this.registerPasskey}
                disabled={registering || !newName.trim()}
              >
                {registering
                  ? this.props.t("passkeyRegistering")
                  : this.props.t("passkeyAdd")}
              </Button>
            </InputGroup>
            {error && <p className="text-danger">{error}</p>}
          </>
        )}
      </div>
    );
  }
}

export default withTranslation(PasskeySettings as any);
