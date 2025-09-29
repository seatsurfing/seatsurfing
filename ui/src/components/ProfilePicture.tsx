import React from "react";
import { User as IconUser } from "react-feather";
import styles from "./ProfilePicture.module.css";

interface State {}

interface Props {
  width: number;
  height: number;
}

class ProfilePicture extends React.Component<Props, State> {
  render() {
    return (
      <div
        className={styles.profilePicWrapper}
        style={{
          width: this.props.width + "px",
          height: this.props.height + "px",
        }}
      >
        <IconUser className={styles.ProfilePic} />
      </div>
    );
  }
}

export default ProfilePicture;
