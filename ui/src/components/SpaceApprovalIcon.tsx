import React from "react";
import { UserCheck as IconUserCheck } from "react-feather";
import "./SpaceApprovalIcon.css";

const SpaceApprovalIcon: React.FC = () => (
  <IconUserCheck
    size={16}
    className="position-absolute top-0 end-0 space-approval-icon"
  />
);

export default SpaceApprovalIcon;
