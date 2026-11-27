import React from "react";
import { Form } from "react-bootstrap";

type Props = Omit<
  React.ComponentProps<typeof Form.Control>,
  "type" | "pattern"
> & { httpsOnly?: boolean };

const UrlInput = ({ httpsOnly, ...props }: Props) => (
  <Form.Control
    type="url"
    pattern={httpsOnly ? "https://.+" : "https?://.+"}
    placeholder="https://..."
    {...props}
  />
);

export default UrlInput;
