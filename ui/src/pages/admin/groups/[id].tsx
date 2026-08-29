import React from "react";
import {
  Form,
  Col,
  Row,
  Button,
  Alert,
  Table,
  InputGroup,
} from "react-bootstrap";
import {
  ChevronLeft as IconBack,
  Save as IconSave,
  Trash2 as IconDelete,
} from "react-feather";
import { NextRouter } from "next/router";
import FullLayout from "@/components/FullLayout";
import Link from "next/link";
import Loading from "@/components/Loading";
import withReadyRouter from "@/components/withReadyRouter";
import withPermission from "@/components/withPermission";
import ProfilePicture from "@/components/ProfilePicture";
import UserSearchTypeahead from "@/components/UserSearchTypeahead";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import User from "@/types/User";
import Group from "@/types/Group";
import ConfirmModal from "@/components/ConfirmModal";

import RendererUtils from "@/util/RendererUtils";
import AjaxError from "@/util/AjaxError";
import ErrorText from "@/types/ErrorText";
import RuntimeConfig from "@/components/RuntimeConfig";
import { Permission, PermissionLevel } from "@/types/Permission";

interface State {
  loading: boolean;
  submitting: boolean;
  saved: boolean;
  error: boolean;
  errorText: string;
  goBack: boolean;
  name: string;
  addUserIds: string[];
  members: User[];
  removeUserIds: string[];
  showDeleteConfirm: boolean;
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
}

class EditUser extends React.Component<Props, State> {
  entity: Group = new Group();
  typeahead: any = null;

  constructor(props: any) {
    super(props);
    this.state = {
      loading: true,
      submitting: false,
      saved: false,
      error: false,
      errorText: "",
      goBack: false,
      name: "",
      addUserIds: [],
      members: [],
      removeUserIds: [],
      showDeleteConfirm: false,
    };
  }

  componentDidMount = () => {
    this.loadData();
  };

  loadData = () => {
    let promises: Promise<any>[] = [];
    const { id } = this.props.router.query;
    if (id && typeof id === "string" && id !== "add") {
      promises.push(Group.get(id));
    }
    Promise.all(promises).then((values) => {
      if (values.length >= 1) {
        let group = values[0];
        this.entity = group;
        this.loadMembers().then(() => {
          this.setState({
            name: group.name,
          });
        });
      }
      this.setState({
        loading: false,
      });
    });
  };

  loadMembers = () => {
    return this.entity.getMembers().then((members) => {
      this.setState({
        members: members,
      });
    });
  };

  onSubmit = (e: any) => {
    e.preventDefault();
    this.setState({
      error: false,
      saved: false,
    });
    this.entity.name = this.state.name;
    this.entity
      .save()
      .then((e) => {
        this.entity.id = e.id;
        this.props.router.push("/admin/groups/" + this.entity.id);
        this.setState({ saved: true });
      })
      .catch((e) => {
        let code: number = 0;
        if (e instanceof AjaxError) {
          code = e.appErrorCode;
        }
        this.setState({
          error: true,
          errorText: code
            ? ErrorText.getTextForAppCode(code, this.props.t)
            : this.props.t("errorSave"),
        });
      });
  };

  deleteItem = () => {
    this.setState({ showDeleteConfirm: true });
  };

  onSearchSelected = (selected: any) => {
    this.setState({
      addUserIds: selected.map((user: any) => user.id),
    });
  };

  addMembers = () => {
    if (this.typeahead !== null) {
      this.entity
        .addMembers(this.state.addUserIds)
        .then(() => {
          this.typeahead.clear();
          this.setState({ addUserIds: [] });
          this.loadMembers();
        })
        .catch(() => {});
    }
  };

  getMemberRow = (user: User) => {
    const canManageMembers = RuntimeConfig.hasPermission(
      Permission.Groups,
      PermissionLevel.Write,
    );
    const fullname = RendererUtils.fullname(user.firstname, user.lastname);
    return (
      <tr
        key={user.id}
        onClick={() =>
          canManageMembers &&
          this.selectMember(
            user.id,
            !this.state.removeUserIds.includes(user.id),
          )
        }
        style={{ cursor: canManageMembers ? "pointer" : "default" }}
      >
        <td style={{ tableLayout: "fixed", width: "20px" }}>
          {canManageMembers && (
            <Form.Check
              type="checkbox"
              onChange={(e: any) =>
                this.selectMember(user.id, e.target.checked)
              }
              checked={this.state.removeUserIds.includes(user.id)}
              onClick={(e: any) => e.stopPropagation()}
            />
          )}
        </td>
        <td style={{ tableLayout: "fixed", width: "64px" }}>
          <ProfilePicture width={48} height={48} />
        </td>
        <td style={{ tableLayout: "auto" }}>
          <div style={{ marginLeft: "10px" }}>
            {user.email}
            {fullname && (
              <>
                <br />
                {fullname}
              </>
            )}
          </div>
        </td>
      </tr>
    );
  };

  selectMember = (userId: string, checked: boolean) => {
    let removeUserIds = this.state.removeUserIds;
    if (checked && !removeUserIds.includes(userId)) {
      removeUserIds.push(userId);
    } else if (!checked && removeUserIds.includes(userId)) {
      removeUserIds.splice(removeUserIds.indexOf(userId), 1);
    }
    this.setState({
      removeUserIds: removeUserIds,
    });
  };

  persistRemoveMembers = () => {
    this.entity.removeMembers(this.state.removeUserIds).then(() => {
      this.setState({
        removeUserIds: [],
      });
      this.loadMembers();
    });
  };

  render() {
    const canManageGroups = RuntimeConfig.hasPermission(
      Permission.Groups,
      PermissionLevel.Admin,
    );
    const canManageMembers = RuntimeConfig.hasPermission(
      Permission.Groups,
      PermissionLevel.Write,
    );
    if (this.state.goBack) {
      this.props.router.push("/admin/groups");
      return <></>;
    }

    let backButton = (
      <Link href="/admin/groups" className="btn btn-sm btn-outline-secondary">
        <IconBack className="feather" /> {this.props.t("back")}
      </Link>
    );
    let buttons = backButton;

    if (this.state.loading) {
      return (
        <FullLayout headline={this.props.t("editGroup")} buttons={buttons}>
          <Loading />
        </FullLayout>
      );
    }

    let hint = <></>;
    if (this.state.saved) {
      hint = <Alert variant="success">{this.props.t("entryUpdated")}</Alert>;
    } else if (this.state.error) {
      hint = <Alert variant="danger">{this.state.errorText}</Alert>;
    }

    let buttonDelete = (
      <Button
        className="btn-sm"
        variant="outline-secondary"
        onClick={this.deleteItem}
      >
        <IconDelete className="feather" /> {this.props.t("delete")}
      </Button>
    );
    let buttonSave = (
      <Button
        className="btn-sm"
        variant="outline-secondary"
        type="submit"
        form="form"
      >
        <IconSave className="feather" /> {this.props.t("save")}
      </Button>
    );
    // Someone who may only manage membership gets neither: the name, and the
    // group's existence, are not theirs to change.
    if (!canManageGroups) {
      buttons = <>{backButton}</>;
    } else if (this.entity.id) {
      buttons = (
        <>
          {backButton} {buttonDelete} {buttonSave}
        </>
      );
    } else {
      buttons = (
        <>
          {backButton} {buttonSave}
        </>
      );
    }

    let memberTable = <></>;
    if (this.entity.id) {
      memberTable = (
        <>
          <div
            className="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom"
            style={{ marginTop: "50px" }}
          >
            <h4>{this.props.t("members")}</h4>
          </div>
          <Form>
            <Form.Group as={Row}>
              <Col sm="6">
                <InputGroup>
                  <UserSearchTypeahead
                    t={this.props.t}
                    id="search-users"
                    multiple={true}
                    onChange={this.onSearchSelected}
                    ref={(ref: any) => {
                      this.typeahead = ref;
                    }}
                  />
                  <Button
                    onClick={() => {
                      this.addMembers();
                    }}
                    variant="outline-secondary"
                    disabled={!canManageMembers}
                  >
                    {this.props.t("add")}
                  </Button>
                </InputGroup>
              </Col>
            </Form.Group>
          </Form>
          <Table hover>
            <tbody>
              {this.state.members.map((user: User) => this.getMemberRow(user))}
            </tbody>
          </Table>
          <Button
            className="btn-sm"
            variant="outline-secondary"
            hidden={!canManageMembers || this.state.removeUserIds.length === 0}
            onClick={() => {
              this.persistRemoveMembers();
            }}
          >
            {this.props.t("remove")}
          </Button>
        </>
      );
    }

    return (
      <FullLayout headline={this.props.t("editGroup")} buttons={buttons}>
        <Form onSubmit={this.onSubmit} id="form">
          {hint}
          <Form.Group as={Row}>
            <Form.Label column sm="2" htmlFor="name">
              {this.props.t("name")}
            </Form.Label>
            <Col sm="4">
              <Form.Control
                id="name"
                type="text"
                value={this.state.name}
                minLength={3}
                onChange={(e: any) => this.setState({ name: e.target.value })}
                required={true}
                readOnly={!canManageGroups}
              />
            </Col>
          </Form.Group>
        </Form>
        {memberTable}
        <ConfirmModal
          show={this.state.showDeleteConfirm}
          message={this.props.t("confirmDeleteGroup")}
          onCancel={() => this.setState({ showDeleteConfirm: false })}
          onConfirm={() => {
            this.setState({ showDeleteConfirm: false });
            this.entity.delete().then(() => {
              this.setState({ goBack: true });
            });
          }}
        />
      </FullLayout>
    );
  }
}

export default withTranslation(
  withReadyRouter(
    withPermission(
      EditUser as any,
      Permission.Groups,
      PermissionLevel.Read,
    ) as any,
  ),
);
