import { Entity } from "./Entity";
import Ajax from "../util/Ajax";

export default class AuthProvider extends Entity {
  name: string;
  providerType: number;
  authUrl: string;
  tokenUrl: string;
  authStyle: number;
  scopes: string;
  userInfoUrl: string;
  userInfoEmailField: string;
  userInfoFirstnameField: string;
  userInfoLastnameField: string;
  /** Claim carrying the user's groups, used to assign roles and groups. */
  userInfoGroupsField: string;
  clientId: string;
  clientSecret: string;
  logoutUrl: string;
  profilePageUrl: string;
  readOnly: boolean;

  constructor() {
    super();
    this.name = "";
    this.providerType = 0;
    this.authUrl = "";
    this.tokenUrl = "";
    this.authStyle = 0;
    this.scopes = "";
    this.userInfoUrl = "";
    this.userInfoEmailField = "";
    this.userInfoFirstnameField = "";
    this.userInfoLastnameField = "";
    this.userInfoGroupsField = "";
    this.clientId = "";
    this.clientSecret = "";
    this.logoutUrl = "";
    this.profilePageUrl = "";
    this.readOnly = false;
  }

  serialize(): Object {
    return Object.assign(super.serialize(), {
      name: this.name,
      providerType: this.providerType,
      authUrl: this.authUrl,
      tokenUrl: this.tokenUrl,
      authStyle: this.authStyle,
      scopes: this.scopes,
      userInfoUrl: this.userInfoUrl,
      userInfoEmailField: this.userInfoEmailField,
      userInfoFirstnameField: this.userInfoFirstnameField,
      userInfoLastnameField: this.userInfoLastnameField,
      userInfoGroupsField: this.userInfoGroupsField,
      clientId: this.clientId,
      clientSecret: this.clientSecret,
      logoutUrl: this.logoutUrl,
      profilePageUrl: this.profilePageUrl,
    });
  }

  deserialize(input: any): void {
    super.deserialize(input);
    this.name = input.name;
    this.providerType = input.providerType;
    this.authUrl = input.authUrl;
    this.tokenUrl = input.tokenUrl;
    this.authStyle = input.authStyle;
    this.scopes = input.scopes;
    this.userInfoUrl = input.userInfoUrl;
    this.userInfoEmailField = input.userInfoEmailField;
    this.userInfoFirstnameField = input.userInfoFirstnameField;
    this.userInfoLastnameField = input.userInfoLastnameField;
    this.userInfoGroupsField = input.userInfoGroupsField ?? "";
    this.clientId = input.clientId;
    this.clientSecret = input.clientSecret;
    this.logoutUrl = input.logoutUrl;
    this.profilePageUrl = input.profilePageUrl;
    this.readOnly = input.readOnly;
  }

  getBackendUrl(): string {
    return "/auth-provider/";
  }

  async save(): Promise<AuthProvider> {
    return Ajax.saveEntity(this, this.getBackendUrl()).then(() => this);
  }

  async delete(): Promise<void> {
    return Ajax.delete(this.getBackendUrl() + this.id).then(() => undefined);
  }

  static async get(id: string): Promise<AuthProvider> {
    return Ajax.get("/auth-provider/" + id).then((result) => {
      let e: AuthProvider = new AuthProvider();
      e.deserialize(result.json);
      return e;
    });
  }

  static async list(): Promise<AuthProvider[]> {
    return Ajax.get("/auth-provider/").then((result) => {
      let list: AuthProvider[] = [];
      (result.json as []).forEach((item) => {
        let e: AuthProvider = new AuthProvider();
        e.deserialize(item);
        list.push(e);
      });
      return list;
    });
  }

  static async listPublicForOrg(id: string): Promise<AuthProvider[]> {
    return Ajax.get("/auth-provider/org/" + id).then((result) => {
      let list: AuthProvider[] = [];
      (result.json as []).forEach((item) => {
        let e: AuthProvider = new AuthProvider();
        e.deserialize(item);
        list.push(e);
      });
      return list;
    });
  }
}

/** A claim value mapped onto a role or a group. */
export class AuthProviderMapping {
  claimValue: string;
  targetType: "role" | "group";
  targetId: string;

  constructor(
    claimValue: string = "",
    targetType: "role" | "group" = "role",
    targetId: string = "",
  ) {
    this.claimValue = claimValue;
    this.targetType = targetType;
    this.targetId = targetId;
  }

  static async list(authProviderId: string): Promise<AuthProviderMapping[]> {
    return Ajax.get("/auth-provider/" + authProviderId + "/mapping").then(
      (result) =>
        (result.json as any[]).map(
          (item) =>
            new AuthProviderMapping(
              item.claimValue,
              item.targetType,
              item.targetId,
            ),
        ),
    );
  }

  static async save(
    authProviderId: string,
    mappings: AuthProviderMapping[],
  ): Promise<void> {
    return Ajax.putData("/auth-provider/" + authProviderId + "/mapping", {
      mappings: mappings,
    }).then(() => undefined);
  }
}
