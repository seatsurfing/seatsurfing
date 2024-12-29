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
	clientId: string;
	clientSecret: string;
    logoutUrl: string;

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
	    this.clientId = "";
	    this.clientSecret = "";
        this.logoutUrl = "";
    }

    serialize(): Object {
        return Object.assign(super.serialize(), {
            "name": this.name,
            "providerType": this.providerType,
            "authUrl": this.authUrl,
            "tokenUrl": this.tokenUrl,
            "authStyle": this.authStyle,
            "scopes": this.scopes,
            "userInfoUrl": this.userInfoUrl,
            "userInfoEmailField": this.userInfoEmailField,
            "clientId": this.clientId,
            "clientSecret": this.clientSecret,
            "logoutUrl": this.logoutUrl,
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
        this.clientId = input.clientId;
        this.clientSecret = input.clientSecret;
        this.logoutUrl = input.logoutUrl;
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
        return Ajax.get("/auth-provider/" + id).then(result => {
            let e: AuthProvider = new AuthProvider();
            e.deserialize(result.json);
            return e;
        });
    }

    static async list(): Promise<AuthProvider[]> {
        return Ajax.get("/auth-provider/").then(result => {
            let list: AuthProvider[] = [];
            (result.json as []).forEach(item => {
                let e: AuthProvider = new AuthProvider();
                e.deserialize(item);
                list.push(e);
            });
            return list;
        });
    }

    static async listPublicForOrg(id: string): Promise<AuthProvider[]> {
        return Ajax.get("/auth-provider/org/" + id).then(result => {
            let list: AuthProvider[] = [];
            (result.json as []).forEach(item => {
                let e: AuthProvider = new AuthProvider();
                e.deserialize(item);
                list.push(e);
            });
            return list;
        });
    }
}
