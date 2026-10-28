import React from "react";
import { Alert, Col, Row } from "react-bootstrap";
import MarkdownRenderer from "./MarkdownRenderer";

interface Props {
  text: string;
}

const OrgAnnouncement: React.FC<Props> = ({ text }) => {
  if (!text.trim()) {
    return null;
  }
  return (
    <Row className="mb-4">
      <Col sm="12" xl="8">
        <Alert variant="warning" className="org-announcement">
          <MarkdownRenderer relativeLinksSameTab>{text}</MarkdownRenderer>
        </Alert>
      </Col>
    </Row>
  );
};

export default OrgAnnouncement;
