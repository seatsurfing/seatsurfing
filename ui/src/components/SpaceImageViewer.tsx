import React from "react";
import { Loader as IconLoad } from "react-feather";
import Space from "@/types/Space";

interface Props {
  space: Space;
}

interface State {
  loading: boolean;
  imageData: string;
  imageSpaceId: string;
}

class SpaceImageViewer extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      loading: false,
      imageData: "",
      imageSpaceId: "",
    };
  }

  componentDidMount = () => {
    this.loadImageIfRequired();
  };

  componentDidUpdate = () => {
    this.loadImageIfRequired();
  };

  loadImageIfRequired = async () => {
    const spaceId = this.props.space.id;
    if (
      !this.props.space.hasImage ||
      this.state.loading ||
      this.state.imageSpaceId === spaceId
    ) {
      return;
    }
    this.setState({ loading: true, imageData: "", imageSpaceId: spaceId });
    try {
      const image = await this.props.space.getImage(true);
      if (spaceId !== this.state.imageSpaceId) {
        return;
      }
      this.setState({
        imageData: `data:${image.mimeType};base64,${image.data}`,
        loading: false,
      });
    } catch {
      this.setState({ loading: false });
    }
  };

  render() {
    if (!this.props.space.hasImage) {
      return null;
    }
    return (
      <div className="space-image-viewer">
        {this.state.loading && <IconLoad className="feather loader" />}
        {!this.state.loading && this.state.imageData && (
          <img
            src={this.state.imageData}
            alt={this.props.space.name}
            className="space-image"
          />
        )}
      </div>
    );
  }
}

export default SpaceImageViewer;
