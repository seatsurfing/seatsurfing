import React from "react";
import {
  Button,
  Col,
  Form,
  Modal,
  OverlayTrigger,
  Row,
  Table,
  Tooltip,
} from "react-bootstrap";
import {
  Plus as IconPlus,
  Edit2 as IconEdit,
  Trash2 as IconDelete,
  ArrowUp as IconUp,
  ArrowDown as IconDown,
} from "react-feather";
import { TranslationFunc, withTranslation } from "./withTranslation";
import BookingAttribute from "@/types/BookingAttribute";
import HintTooltip from "./HintTooltip";
import ConfirmModal from "./ConfirmModal";
import RendererUtils from "@/util/RendererUtils";

interface Props {
  t: TranslationFunc;
  disabled: boolean;
}

interface State {
  loading: boolean;
  items: BookingAttribute[];
  editItem: BookingAttribute | null;
  label: string;
  type: number;
  required: boolean;
  saving: boolean;
  error: boolean;
  deleteItem: BookingAttribute | null;
}

class BookingAttributesEditor extends React.Component<Props, State> {
  constructor(props: any) {
    super(props);
    this.state = {
      loading: true,
      items: [],
      editItem: null,
      label: "",
      type: BookingAttribute.TYPE_TEXT,
      required: false,
      saving: false,
      error: false,
      deleteItem: null,
    };
  }

  componentDidMount = () => {
    this.loadItems();
  };

  loadItems = async () => {
    try {
      const items = await BookingAttribute.list();
      this.setState({ items: items, loading: false });
    } catch {
      this.setState({ loading: false });
    }
  };

  openEditModal = (item: BookingAttribute) => {
    this.setState({
      editItem: item,
      label: item.label,
      type: item.type,
      required: item.required,
      error: false,
    });
  };

  closeEditModal = () => {
    this.setState({ editItem: null });
  };

  onSave = async (e: React.FormEvent) => {
    // The modal is rendered inside the settings form, which must not submit.
    e.preventDefault();
    e.stopPropagation();
    const editItem = this.state.editItem;
    if (!editItem) {
      return;
    }
    const item = new BookingAttribute(
      editItem.id,
      this.state.label,
      this.state.type,
      this.state.required,
    );
    this.setState({ saving: true, error: false });
    try {
      await item.save();
      this.setState({ saving: false, editItem: null });
      this.loadItems();
    } catch {
      this.setState({ saving: false, error: true });
    }
  };

  onMove = async (item: BookingAttribute, up: boolean) => {
    try {
      await item.move(up);
    } catch {
      // reload below shows the current state
    }
    this.loadItems();
  };

  onDelete = async () => {
    const item = this.state.deleteItem;
    this.setState({ deleteItem: null });
    if (!item) {
      return;
    }
    try {
      await item.delete();
    } catch {
      // reload below shows the current state
    }
    this.loadItems();
  };

  renderItem = (item: BookingAttribute, index: number) => {
    return (
      <tr key={item.id}>
        <td>{item.label}</td>
        <td>
          {this.props.t(BookingAttribute.getTypeTranslationKey(item.type))}
        </td>
        <td>{RendererUtils.state(item.required)}</td>
        <td className="text-end text-nowrap">
          <Button
            variant="link"
            size="sm"
            className="p-0 me-2"
            aria-label={this.props.t("moveUp")}
            disabled={this.props.disabled || index === 0}
            onClick={() => this.onMove(item, true)}
          >
            <IconUp className="feather" />
          </Button>
          <Button
            variant="link"
            size="sm"
            className="p-0 me-2"
            aria-label={this.props.t("moveDown")}
            disabled={
              this.props.disabled || index === this.state.items.length - 1
            }
            onClick={() => this.onMove(item, false)}
          >
            <IconDown className="feather" />
          </Button>
          <Button
            variant="link"
            size="sm"
            className="p-0 me-2"
            aria-label={this.props.t("edit")}
            disabled={this.props.disabled}
            onClick={() => this.openEditModal(item)}
          >
            <IconEdit className="feather" />
          </Button>
          <Button
            variant="link"
            size="sm"
            className="p-0 text-danger"
            aria-label={this.props.t("delete")}
            disabled={this.props.disabled}
            onClick={() => this.setState({ deleteItem: item })}
          >
            <IconDelete className="feather" />
          </Button>
        </td>
      </tr>
    );
  };

  renderEditModal = () => {
    const isNew = this.state.editItem !== null && !this.state.editItem.id;
    return (
      <Modal show={this.state.editItem !== null} onHide={this.closeEditModal}>
        <Form onSubmit={this.onSave}>
          <Modal.Header closeButton>
            <Modal.Title>
              {this.props.t(
                isNew ? "addBookingAttribute" : "editBookingAttribute",
              )}
            </Modal.Title>
          </Modal.Header>
          <Modal.Body>
            {this.state.error && (
              <p className="text-danger">
                {this.props.t("bookingAttributeSaveError")}
              </p>
            )}
            <Form.Group className="mb-3" controlId="bookingAttributeLabel">
              <Form.Label>{this.props.t("bookingAttributeLabel")}</Form.Label>
              <Form.Control
                type="text"
                value={this.state.label}
                onChange={(e: any) => this.setState({ label: e.target.value })}
                required={true}
                maxLength={256}
                autoFocus={true}
              />
            </Form.Group>
            <Form.Group className="mb-3" controlId="bookingAttributeType">
              <Form.Label>{this.props.t("type")}</Form.Label>
              <Form.Select
                value={this.state.type}
                onChange={(e: any) =>
                  this.setState({ type: parseInt(e.target.value) })
                }
              >
                {[
                  BookingAttribute.TYPE_TEXT,
                  BookingAttribute.TYPE_TEXT_MULTILINE,
                  BookingAttribute.TYPE_NUMBER,
                  BookingAttribute.TYPE_CHECKBOX,
                ].map((type) => (
                  <option key={type} value={type}>
                    {this.props.t(BookingAttribute.getTypeTranslationKey(type))}
                  </option>
                ))}
              </Form.Select>
            </Form.Group>
            <Form.Check
              type="checkbox"
              id="bookingAttributeRequired"
              label={this.props.t("bookingAttributeRequired")}
              checked={this.state.required}
              onChange={(e: any) =>
                this.setState({ required: e.target.checked })
              }
            />
          </Modal.Body>
          <Modal.Footer>
            <Button variant="secondary" onClick={this.closeEditModal}>
              {this.props.t("cancel")}
            </Button>
            <Button
              variant="primary"
              type="submit"
              disabled={this.state.saving}
            >
              {this.props.t("save")}
            </Button>
          </Modal.Footer>
        </Form>
      </Modal>
    );
  };

  render() {
    if (this.state.loading) {
      return <></>;
    }
    const limitReached =
      this.state.items.length >= BookingAttribute.MAX_ATTRIBUTES;
    const addButton = (
      <Button
        variant="outline-secondary"
        size="sm"
        className="mt-1"
        disabled={this.props.disabled || limitReached}
        style={limitReached ? { pointerEvents: "none" } : undefined}
        onClick={() => this.openEditModal(new BookingAttribute())}
      >
        <IconPlus className="feather" /> {this.props.t("addBookingAttribute")}
      </Button>
    );
    return (
      <Form.Group as={Row}>
        <Form.Label column sm="2">
          {this.props.t("bookingAttributes")}
          <HintTooltip hint={this.props.t("bookingAttributesHint")} />
        </Form.Label>
        <Col sm="4">
          {this.state.items.length > 0 && (
            <Table size="sm" id="bookingAttributesTable">
              <thead>
                <tr>
                  <th>{this.props.t("bookingAttributeLabel")}</th>
                  <th>{this.props.t("type")}</th>
                  <th>{this.props.t("bookingAttributeRequired")}</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {this.state.items.map((item, index) =>
                  this.renderItem(item, index),
                )}
              </tbody>
            </Table>
          )}
          {limitReached ? (
            <OverlayTrigger
              placement="right"
              overlay={
                <Tooltip>
                  {this.props.t("bookingAttributeLimitReached", {
                    max: BookingAttribute.MAX_ATTRIBUTES,
                  })}
                </Tooltip>
              }
            >
              <span className="d-inline-block">{addButton}</span>
            </OverlayTrigger>
          ) : (
            addButton
          )}
        </Col>
        {this.renderEditModal()}
        <ConfirmModal
          show={this.state.deleteItem !== null}
          message={
            this.state.deleteItem
              ? RendererUtils.decodeHtmlEntities(
                  this.props.t("confirmDeleteBookingAttribute", {
                    label: this.state.deleteItem.label,
                  }),
                )
              : ""
          }
          onCancel={() => this.setState({ deleteItem: null })}
          onConfirm={this.onDelete}
        />
      </Form.Group>
    );
  }
}

export default withTranslation(BookingAttributesEditor as any);
