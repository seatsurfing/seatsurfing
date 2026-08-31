import { Entity } from "./Entity";
import Ajax from "../util/Ajax";
import { PermissionMap } from "./Permission";

/** A permission as reported by the catalogue endpoint. */
export class PermissionDefinition {
  key: string;
  allowedLevels: number[];
  pluginId: string;

  constructor() {
    this.key = "";
    this.allowedLevels = [];
    this.pluginId = "";
  }

  deserialize(input: any): void {
    this.key = input.key;
    this.allowedLevels = input.allowedLevels ?? [];
    this.pluginId = input.pluginId ?? "";
  }

  static async list(): Promise<PermissionDefinition[]> {
    return Ajax.get("/role/permissions").then((result) => {
      let list: PermissionDefinition[] = [];
      (result.json as any[]).forEach((item) => {
        let e = new PermissionDefinition();
        e.deserialize(item);
        list.push(e);
      });
      return list;
    });
  }
}

export default class Role extends Entity {
  name: string;
  description: string;
  /** System roles can not be edited or deleted. */
  system: boolean;
  permissions: PermissionMap;

  constructor() {
    super();
    this.name = "";
    this.description = "";
    this.system = false;
    this.permissions = {};
  }

  serialize(): Object {
    return Object.assign(super.serialize(), {
      name: this.name,
      description: this.description,
      permissions: this.permissions,
    });
  }

  deserialize(input: any): void {
    super.deserialize(input);
    this.name = input.name;
    this.description = input.description ?? "";
    this.system = input.system ?? false;
    this.permissions = input.permissions ?? {};
  }

  getBackendUrl(): string {
    return "/role/";
  }

  async save(): Promise<Role> {
    return Ajax.saveEntity(this, this.getBackendUrl()).then(() => this);
  }

  async delete(): Promise<void> {
    return Ajax.delete(this.getBackendUrl() + this.id).then(() => undefined);
  }

  static async get(id: string): Promise<Role> {
    return Ajax.get("/role/" + id).then((result) => {
      let e = new Role();
      e.deserialize(result.json);
      return e;
    });
  }

  static async list(): Promise<Role[]> {
    return Ajax.get("/role/").then((result) => {
      let list: Role[] = [];
      (result.json as any[]).forEach((item) => {
        let e = new Role();
        e.deserialize(item);
        list.push(e);
      });
      list.sort((a, b) => a.name.localeCompare(b.name));
      return list;
    });
  }

  static async getUserIds(roleId: string): Promise<string[]> {
    return Ajax.get("/role/" + roleId + "/users").then(
      (result) => result.json as string[],
    );
  }

  static async getRoleIdsForUser(userId: string): Promise<string[]> {
    return Ajax.get("/user/" + userId + "/roles").then(
      (result) => result.json as string[],
    );
  }

  static async setRoleIdsForUser(
    userId: string,
    roleIds: string[],
  ): Promise<void> {
    return Ajax.putData("/user/" + userId + "/roles", {
      roleIds: roleIds,
    }).then(() => undefined);
  }
}
