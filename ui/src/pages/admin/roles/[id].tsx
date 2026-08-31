import React from "react";
import { Form, Col, Row, Alert, Button } from "react-bootstrap";
import { ChevronLeft as IconBack, Save as IconSave } from "react-feather";
import FullLayout from "@/components/FullLayout";
import Loading from "@/components/Loading";
import Link from "next/link";
import { NextRouter } from "next/router";
import withReadyRouter from "@/components/withReadyRouter";
import withPermission from "@/components/withPermission";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import RuntimeConfig from "@/components/RuntimeConfig";
import AjaxError from "@/util/AjaxError";
import ErrorText from "@/types/ErrorText";
import Role, { PermissionDefinition } from "@/types/Role";
import {
  Permission,
  PermissionLevel,
  PermissionMap,
  permissionLabelKey,
  permissionLevelLabelKey,
} from "@/types/Permission";

interface State {
  loading: boolean;
  submitting: boolean;
  saved: boolean;
  error: boolean;
  errorText: string;
  goBack: boolean;
  name: string;
  description: string;
  permissions: PermissionMap;
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
}

class EditRole extends React.Component<Props, State> {
  entity: Role = new Role();
  definitions: PermissionDefinition[] = [];

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
      description: "",
      permissions: {},
    };
  }

  componentDidMount = () => {
    this.loadData();
  };

  loadData = async () => {
    this.definitions = await PermissionDefinition.list();
    const { id } = this.props.router.query;
    if (id && typeof id === "string" && id !== "add") {
      const role = await Role.get(id);
      this.entity = role;
      this.setState({
        name: role.name,
        description: role.description,
        permissions: { ...role.permissions },
      });
    }
    this.setState({ loading: false });
  };

  setLevel = (key: string, level: number) => {
    this.setState((state) => {
      const permissions = { ...state.permissions };
      if (level === PermissionLevel.None) {
        delete permissions[key];
      } else {
        permissions[key] = level;
      }
      return { permissions };
    });
  };

  onSubmit = async (e: any) => {
    e.preventDefault();
    this.setState({ error: false, saved: false, submitting: true });
    this.entity.name = this.state.name;
    this.entity.description = this.state.description;
    this.entity.permissions = this.state.permissions;
    try {
      await this.entity.save();
      this.props.router.push(
        `/admin/roles/${encodeURIComponent(this.entity.id)}`,
      );
      this.setState({ saved: true, submitting: false });
    } catch (e) {
      let code: number = 0;
      if (e instanceof AjaxError) {
        code = e.appErrorCode;
      }
      this.setState({
        error: true,
        submitting: false,
        errorText: code
          ? ErrorText.getTextForAppCode(code, this.props.t)
          : this.props.t("errorSave"),
      });
    }
  };

  onDelete = async () => {
    if (!window.confirm(this.props.t("confirmDeleteRole"))) {
      return;
    }
    try {
      await this.entity.delete();
      this.setState({ goBack: true });
    } catch (e) {
      let code: number = 0;
      if (e instanceof AjaxError) {
        code = e.appErrorCode;
      }
      this.setState({
        error: true,
        errorText: code
          ? ErrorText.getTextForAppCode(code, this.props.t)
          : this.props.t("errorDelete"),
      });
    }
  };

  /** Falls back to the raw key for plugin permissions, which have no translation. */
  permissionLabel = (key: string): string => {
    const label = this.props.t(permissionLabelKey(key));
    return label === permissionLabelKey(key) ? key : label;
  };

  renderPermissionRow = (
    definition: PermissionDefinition,
    readOnly: boolean,
  ) => {
    const current =
      this.state.permissions[definition.key] ?? PermissionLevel.None;
    return (
      <Form.Group as={Row} key={definition.key} className="mb-2">
        <Form.Label column sm="4">
          {this.permissionLabel(definition.key)}
        </Form.Label>
        <Col sm="8">
          {/* Only the levels this permission actually declares are offered:
              a uniform ladder would present choices that do nothing. */}
          {definition.allowedLevels.map((level) => (
            <Form.Check
              inline
              type="radio"
              key={definition.key + "-" + level}
              id={definition.key + "-" + level}
              name={definition.key}
              label={this.props.t(permissionLevelLabelKey(level))}
              checked={current === level}
              disabled={readOnly}
              onChange={() => this.setLevel(definition.key, level)}
            />
          ))}
        </Col>
      </Form.Group>
    );
  };

  render() {
    if (this.state.goBack) {
      this.props.router.push("/admin/roles");
      return <></>;
    }
    if (this.state.loading) {
      return (
        <FullLayout headline={this.props.t("roles")}>
          <Loading />
        </FullLayout>
      );
    }

    // System roles are shown but not editable: every organization keeps one
    // role that grants everything, so that it can never be edited away.
    const readOnly =
      this.entity.system ||
      !RuntimeConfig.hasPermission(Permission.Roles, PermissionLevel.Admin);

    const backButton = (
      <Link href="/admin/roles" className="btn btn-sm btn-outline-secondary">
        <IconBack className="feather" /> {this.props.t("back")}
      </Link>
    );
    const buttons = (
      <>
        {backButton}{" "}
        {!readOnly && (
          <>
            {this.entity.id && (
              <Button
                className="btn-sm"
                variant="outline-secondary"
                onClick={this.onDelete}
              >
                {this.props.t("delete")}
              </Button>
            )}{" "}
            <Button
              className="btn-sm"
              variant="outline-secondary"
              type="submit"
              form="form"
              disabled={this.state.submitting}
            >
              <IconSave className="feather" /> {this.props.t("save")}
            </Button>
          </>
        )}
      </>
    );

    const builtIn = this.definitions.filter((d) => !d.pluginId);
    const fromPlugins = this.definitions.filter((d) => d.pluginId);

    return (
      <FullLayout
        headline={this.entity.id ? this.entity.name : this.props.t("addRole")}
        buttons={buttons}
      >
        <Form onSubmit={this.onSubmit} id="form">
          {this.state.saved && (
            <Alert variant="success">{this.props.t("entryUpdated")}</Alert>
          )}
          {this.state.error && (
            <Alert variant="danger">{this.state.errorText}</Alert>
          )}
          {this.entity.system && (
            <Alert variant="info">{this.props.t("systemRoleHint")}</Alert>
          )}
          <Form.Group as={Row}>
            <Form.Label column sm="2" htmlFor="name">
              {this.props.t("name")}
            </Form.Label>
            <Col sm="4">
              <Form.Control
                id="name"
                type="text"
                value={this.state.name}
                onChange={(e: any) => this.setState({ name: e.target.value })}
                required={true}
                maxLength={128}
                readOnly={readOnly}
              />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2" htmlFor="description">
              {this.props.t("description")}
            </Form.Label>
            <Col sm="4">
              <Form.Control
                id="description"
                type="text"
                value={this.state.description}
                onChange={(e: any) =>
                  this.setState({ description: e.target.value })
                }
                maxLength={512}
                readOnly={readOnly}
              />
            </Col>
          </Form.Group>
          <h5 className="mt-4">{this.props.t("permissions")}</h5>
          <p className="text-muted">{this.props.t("permissionsHint")}</p>
          {builtIn.map((d) => this.renderPermissionRow(d, readOnly))}
          {fromPlugins.length > 0 && (
            <>
              <h5 className="mt-4">{this.props.t("pluginPermissions")}</h5>
              {fromPlugins.map((d) => this.renderPermissionRow(d, readOnly))}
            </>
          )}
        </Form>
      </FullLayout>
    );
  }
}

export default withTranslation(
  withReadyRouter(
    withPermission(
      EditRole as any,
      Permission.Roles,
      PermissionLevel.Read,
    ) as any,
  ),
);
