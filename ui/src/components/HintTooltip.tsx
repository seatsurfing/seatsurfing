import React from "react";
import { OverlayTrigger, Tooltip } from "react-bootstrap";
import { Info as IconHelp } from "react-feather";

interface Props {
  hint: string;
}

const HintTooltip: React.FC<Props> = ({ hint }) => (
  <OverlayTrigger placement="right" overlay={<Tooltip>{hint}</Tooltip>}>
    <IconHelp
      size={16}
      style={{ marginLeft: "6px", cursor: "pointer", color: "#6c757d" }}
    />
  </OverlayTrigger>
);

export default HintTooltip;
