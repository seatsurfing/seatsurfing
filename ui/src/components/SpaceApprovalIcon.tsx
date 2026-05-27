import React from "react";
import { UserCheck as IconUserCheck } from "react-feather";
import styles from "./SpaceApprovalIcon.module.css";

const SpaceApprovalIcon: React.FC = () => (
  <IconUserCheck
    size={16}
    className={`position-absolute top-0 start-50 translate-middle-x ${styles["space-approval-icon"]}`}
  />
);

export default SpaceApprovalIcon;
