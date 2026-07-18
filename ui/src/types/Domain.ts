import Ajax from "../util/Ajax";

export default class Domain {
  organizationId: string;
  domain: string;
  active: boolean;
  verifyToken: string;
  primary: boolean;
  accessible: boolean;
  accessCheck: Date | null;

  constructor() {
    this.organizationId = "";
    this.domain = "";
    this.active = false;
    this.verifyToken = "";
    this.primary = false;
    this.accessible = false;
    this.accessCheck = null;
  }

  deserialize(input: any): void {
    this.domain = input.domain;
    this.active = input.active;
    this.verifyToken = input.verifyToken;
    this.primary = input.primary;
    this.accessible = input.accessible;
    if (input.accessCheck) {
      this.accessCheck = new Date(input.accessCheck);
    }
  }

  async delete(): Promise<void> {
    await Ajax.delete(
      `/organization/${encodeURIComponent(this.organizationId)}/domain/${encodeURIComponent(this.domain)}`,
    );
  }

  async setPrimary(): Promise<void> {
    await Ajax.postData(
      `/organization/${encodeURIComponent(this.organizationId)}/domain/${encodeURIComponent(this.domain)}/primary`,
    );
  }

  async verify(): Promise<void> {
    await Ajax.postData(
      `/organization/${encodeURIComponent(this.organizationId)}/domain/${encodeURIComponent(this.domain)}/verify`,
      null,
      (status) => status === 400,
    );
  }

  static async add(orgId: string, domain: string): Promise<void> {
    await Ajax.postData(
      `/organization/${encodeURIComponent(orgId)}/domain/${encodeURIComponent(domain)}`,
    );
  }

  static async list(orgId: string): Promise<Domain[]> {
    const result = await Ajax.get(
      `/organization/${encodeURIComponent(orgId)}/domain/`,
    );
    const list: Domain[] = [];
    (result.json as []).forEach((item) => {
      const e: Domain = new Domain();
      e.deserialize(item);
      e.organizationId = orgId;
      list.push(e);
    });
    return list;
  }
}
