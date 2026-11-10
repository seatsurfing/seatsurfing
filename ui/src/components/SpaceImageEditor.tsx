import React from "react";
import { Button, Form, InputGroup } from "react-bootstrap";
import { Trash2 as IconDelete, Loader as IconLoad } from "react-feather";
import { TranslationFunc, withTranslation } from "./withTranslation";
import Space from "@/types/Space";

export const SPACE_IMAGE_MAX_BYTES = 5 * 1024 * 1024;
export const SPACE_IMAGE_MIME_TYPES = ["image/png", "image/jpeg"];

interface Props {
  t: TranslationFunc;
  locationId: string;
  spaceId: string;
  hasImage: boolean;
  pendingImage: File | null;
  removeImage: boolean;
  onSelect: (file: File) => void;
  onRemove: () => void;
}

interface State {
  loading: boolean;
  storedImageData: string;
  pendingImageUrl: string;
  error: string;
}

class SpaceImageEditor extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      loading: false,
      storedImageData: "",
      pendingImageUrl: "",
      error: "",
    };
  }

  componentDidMount = () => {
    this.updatePendingImageUrl();
    if (this.props.hasImage && this.props.spaceId) {
      this.loadStoredImage();
    }
  };

  componentDidUpdate = (prevProps: Props) => {
    if (prevProps.pendingImage !== this.props.pendingImage) {
      this.updatePendingImageUrl();
    }
  };

  // Read as data URL, as the Content Security Policy does not allow blob: images.
  updatePendingImageUrl = async () => {
    const file = this.props.pendingImage;
    if (!file) {
      this.setState({ pendingImageUrl: "" });
      return;
    }
    try {
      const url = await new Promise<string>((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = () => resolve(reader.result as string);
        reader.onerror = () => reject(reader.error);
        reader.readAsDataURL(file);
      });
      if (file === this.props.pendingImage) {
        this.setState({ pendingImageUrl: url });
      }
    } catch {
      this.setState({ pendingImageUrl: "" });
    }
  };

  loadStoredImage = async () => {
    const space = new Space();
    space.id = this.props.spaceId;
    space.locationId = this.props.locationId;
    this.setState({ loading: true });
    try {
      const image = await space.getImage();
      this.setState({
        storedImageData: `data:${image.mimeType};base64,${image.data}`,
        loading: false,
      });
    } catch {
      this.setState({ loading: false });
    }
  };

  onFileSelected = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.item(0);
    e.target.value = "";
    if (!file) {
      return;
    }
    if (!SPACE_IMAGE_MIME_TYPES.includes(file.type)) {
      this.setState({ error: this.props.t("errorSpaceImageInvalidFormat") });
      return;
    }
    if (file.size > SPACE_IMAGE_MAX_BYTES) {
      this.setState({ error: this.props.t("errorSpaceImageTooLarge") });
      return;
    }
    this.setState({ error: "" });
    this.props.onSelect(file);
  };

  getPreviewUrl = (): string => {
    if (this.props.pendingImage) {
      return this.state.pendingImageUrl;
    }
    if (this.props.hasImage && !this.props.removeImage) {
      return this.state.storedImageData;
    }
    return "";
  };

  render() {
    const previewUrl = this.getPreviewUrl();
    return (
      <>
        {this.state.loading && (
          <div>
            <IconLoad className="feather loader" />
          </div>
        )}
        {!this.state.loading && previewUrl && (
          <div className="space-image-editor-preview">
            <img
              src={previewUrl}
              alt={this.props.t("spaceImage")}
              className="space-image"
            />
          </div>
        )}
        <InputGroup>
          <Form.Control
            id="space-image"
            type="file"
            accept={SPACE_IMAGE_MIME_TYPES.join(",")}
            onChange={this.onFileSelected}
          />
          {previewUrl && (
            <Button
              variant="outline-danger"
              onClick={() => {
                this.setState({ error: "" });
                this.props.onRemove();
              }}
              aria-label={this.props.t("remove")}
              title={this.props.t("remove")}
            >
              <IconDelete className="feather" />
            </Button>
          )}
        </InputGroup>
        {this.state.error && (
          <div className="invalid-feedback d-block">{this.state.error}</div>
        )}
      </>
    );
  }
}

export default withTranslation(SpaceImageEditor as any);
