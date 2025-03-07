import Ajax from "../util/Ajax";

export default class Domain {
    organizationId: string;
    domain: string;
    active: boolean;
    verifyToken: string;
    primary: boolean;

    constructor() {
        this.organizationId = "";
        this.domain = "";
        this.active = false;
        this.verifyToken = "";
        this.primary = false;
    }

    deserialize(input: any): void {
        this.domain = input.domain;
        this.active = input.active;
        this.verifyToken = input.verifyToken;
        this.primary = input.primary;
    }

    async delete(): Promise<void> {
        return Ajax.delete("/organization/" + this.organizationId + "/domain/" + this.domain).then(() => undefined);
    }

    async setPrimary(): Promise<void> {
        return Ajax.postData("/organization/" + this.organizationId + "/domain/" + this.domain + "/primary").then(() => undefined);
    }

    async verify(): Promise<void> {
        return Ajax.postData("/organization/" + this.organizationId + "/domain/" + this.domain + "/verify").then(() => undefined);
    }

    static async add(orgId: string, domain: string): Promise<void> {
        return Ajax.postData("/organization/" + orgId + "/domain/" + domain).then(() => undefined);
    }

    static async list(orgId: string): Promise<Domain[]> {
        return Ajax.get("/organization/" + orgId + "/domain/").then(result => {
            let list: Domain[] = [];
            (result.json as []).forEach(item => {
                let e: Domain = new Domain();
                e.deserialize(item);
                e.organizationId = orgId;
                list.push(e);
            });
            return list;
        });
    }
}
