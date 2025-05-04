import { Entity } from "./Entity";
import Ajax from "../util/Ajax";
import User from "./User";

export default class Group extends Entity {
    id: string;
    organizationId: string;
    name: string;

    constructor() {
        super();
        this.id = "";
        this.organizationId = "";
        this.name = "";
    }

    serialize(): Object {
        return Object.assign(super.serialize(), {
            "organizationId": this.organizationId,
            "name": this.name
        });
    }

    deserialize(input: any): void {
        super.deserialize(input);
        this.organizationId = input.organizationId;
        this.name = input.name;
    }

    getBackendUrl(): string {
        return "/group/";
    }

    async save(): Promise<Group> {
        return Ajax.saveEntity(this, this.getBackendUrl()).then(() => this);
    }

    async delete(): Promise<void> {
        return Ajax.delete(this.getBackendUrl() + this.id).then(() => undefined);
    }

    async getMembers(): Promise<User[]> {
        return Ajax.get(this.getBackendUrl() + this.id + "/member").then(result => {
            let list: User[] = [];
            (result.json as []).forEach(item => {
                let e: User = new User();
                e.deserialize(item);
                list.push(e);
            });
            return list;
        });
    }

    async addMembers(userIds: string[]): Promise<void> {
        return Ajax.putData(this.getBackendUrl() + this.id + "/member", userIds).then(() => undefined);
    }

    async removeMembers(userIds: string[]): Promise<void> {
        return Ajax.postData(this.getBackendUrl() + this.id + "/member/remove", userIds).then(() => undefined);
    }

    static async get(id: string): Promise<Group> {
        return Ajax.get("/group/" + id).then(result => {
            let e: Group = new Group();
            e.deserialize(result.json);
            return e;
        });
    }

    static async list(): Promise<Group[]> {
        return Ajax.get("/group/").then(result => {
            let list: Group[] = [];
            (result.json as []).forEach(item => {
                let e: Group = new Group();
                e.deserialize(item);
                list.push(e);
            });
            return list;
        });
    }
}
