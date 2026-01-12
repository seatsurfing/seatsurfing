import { Entity } from "./Entity";
import Ajax from "../util/Ajax";

export default class Organization extends Entity {
  name: string;
  contactFirstname: string;
  contactLastname: string;
  contactEmail: string;
  language: string;
  country: string;
  addressLine1: string;
  addressLine2: string;
  postalCode: string;
  city: string;
  vatId: string;
  company: string;

  constructor() {
    super();
    this.name = "";
    this.contactFirstname = "";
    this.contactLastname = "";
    this.contactEmail = "";
    this.language = "";
    this.country = "";
    this.addressLine1 = "";
    this.addressLine2 = "";
    this.postalCode = "";
    this.city = "";
    this.vatId = "";
    this.company = "";
  }

  serialize(): Object {
    const obj: any = {
      name: this.name,
      firstname: this.contactFirstname,
      lastname: this.contactLastname,
      email: this.contactEmail,
      language: this.language,
    };
    if (this.country !== undefined) obj.country = this.country;
    if (this.addressLine1 !== undefined) obj.addressLine1 = this.addressLine1;
    if (this.addressLine2 !== undefined) obj.addressLine2 = this.addressLine2;
    if (this.postalCode !== undefined) obj.postalCode = this.postalCode;
    if (this.city !== undefined) obj.city = this.city;
    if (this.vatId !== undefined) obj.vatId = this.vatId;
    if (this.company !== undefined) obj.company = this.company;
    return Object.assign(super.serialize(), obj);
  }

  deserialize(input: any): void {
    super.deserialize(input);
    this.name = input.name;
    this.contactFirstname = input.firstname;
    this.contactLastname = input.lastname;
    this.contactEmail = input.email;
    this.language = input.language;
    this.country = input.country;
    this.addressLine1 = input.addressLine1;
    this.addressLine2 = input.addressLine2;
    this.postalCode = input.postalCode;
    this.city = input.city;
    this.vatId = input.vatId;
    this.company = input.company;
  }

  getBackendUrl(): string {
    return "/organization/";
  }

  async save(): Promise<Organization> {
    return Ajax.saveEntity(this, this.getBackendUrl()).then(() => this);
  }

  async delete(): Promise<number> {
    return Ajax.delete(this.getBackendUrl() + this.id).then(
      (result) => result.json.code as number,
    );
  }

  static async get(id: string): Promise<Organization> {
    return Ajax.get("/organization/" + id).then((result) => {
      let e: Organization = new Organization();
      e.deserialize(result.json);
      return e;
    });
  }

  static async list(): Promise<Organization[]> {
    return Ajax.get("/organization/").then((result) => {
      let list: Organization[] = [];
      (result.json as []).forEach((item) => {
        let e: Organization = new Organization();
        e.deserialize(item);
        list.push(e);
      });
      return list;
    });
  }

  static async getOrgForDomain(domain: string): Promise<Organization> {
    return Ajax.get("/organization/domain/" + domain).then((result) => {
      let e: Organization = new Organization();
      e.deserialize(result.json);
      return e;
    });
  }
}
