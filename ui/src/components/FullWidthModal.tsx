import React from "react";
import { Modal, ModalDialog } from "react-bootstrap";
import type { ModalProps } from "react-bootstrap";

interface Props extends Omit<ModalProps, "size" | "fullscreen" | "dialogAs"> {
  maxWidth?: number | string;
  children?: React.ReactNode;
}

const FullWidthModal: React.FC<Props> = ({
  maxWidth = 1200,
  children,
  ...modalProps
}) => {
  const maxWidthValue =
    typeof maxWidth === "number" ? `${maxWidth}px` : maxWidth;

  const DialogComponent = React.useMemo(
    () =>
      function CustomDialog({ style, children: dc, ...dialogProps }: any) {
        return (
          <ModalDialog
            {...dialogProps}
            style={{
              ...style,
              width: "calc(100% - 1rem)",
              maxWidth: maxWidthValue,
              margin: "0.5rem auto",
            }}
          >
            {dc}
          </ModalDialog>
        );
      },
    [maxWidthValue],
  );

  return (
    <Modal {...modalProps} dialogAs={DialogComponent}>
      {children}
    </Modal>
  );
};

export default FullWidthModal;
