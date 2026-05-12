import React from "react";
import { Form } from "react-bootstrap";

type Props = Omit<React.ComponentProps<typeof Form.Control>, "type" | "pattern">;

const UrlInput = (props: Props) => (
  <Form.Control type="url" pattern="https?://.+" placeholder="https://..." {...props} />
);

export default UrlInput;
