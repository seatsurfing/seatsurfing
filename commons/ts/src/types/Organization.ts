import { Entity } from "./Entity";
import Ajax from "../util/Ajax";

export default class Organization extends Entity {
    name: string;
    contactFirstname: string;
    contactLastname: string;
    contactEmail: string;
    language: string;

    constructor() {
        super();
        this.name = "";
        this.contactFirstname = "";
        this.contactLastname = "";
        this.contactEmail = "";
        this.language = "";
    }

    serialize(): Object {
        return Object.assign(super.serialize(), {
            "name": this.name,
            "firstname": this.contactFirstname,
            "lastname": this.contactLastname,
            "email": this.contactEmail,
            "language": this.language
        });
    }

    deserialize(input: any): void {
        super.deserialize(input);
        this.name = input.name;
        this.contactFirstname = input.firstname;
        this.contactLastname = input.lastname;
        this.contactEmail = input.email;
        this.language = input.language;
    }

    getBackendUrl(): string {
        return "/organization/";
    }

    async save(): Promise<Organization> {
        return Ajax.saveEntity(this, this.getBackendUrl()).then(() => this);
    }

    async delete(): Promise<void> {
        return Ajax.delete(this.getBackendUrl() + this.id).then(() => undefined);
    }

    static async get(id: string): Promise<Organization> {
        return Ajax.get("/organization/" + id).then(result => {
            let e: Organization = new Organization();
            e.deserialize(result.json);
            return e;
        });
    }

    static async list(): Promise<Organization[]> {
        return Ajax.get("/organization/").then(result => {
            let list: Organization[] = [];
            (result.json as []).forEach(item => {
                let e: Organization = new Organization();
                e.deserialize(item);
                list.push(e);
            });
            return list;
        });
    }

    static async getOrgForDomain(domain: string): Promise<Organization> {
        return Ajax.get("/organization/domain/" + domain).then(result => {
            let e: Organization = new Organization();
            e.deserialize(result.json);
            return e;
        });
    }
}
