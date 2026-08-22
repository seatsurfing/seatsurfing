import { Entity } from "./Entity";
import Ajax from "../util/Ajax";
import User from "./User";

export class BuddySearchResult {
  id: string;
  email: string;
  firstname: string;
  lastname: string;

  constructor() {
    this.id = "";
    this.email = "";
    this.firstname = "";
    this.lastname = "";
  }

  deserialize(input: any): void {
    this.id = input.id;
    this.email = input.email;
    this.firstname = input.firstname;
    this.lastname = input.lastname;
  }

  static from(input: any): BuddySearchResult {
    const e = new BuddySearchResult();
    e.deserialize(input);
    return e;
  }
}

export default class Buddy extends Entity {
  buddy: User;

  constructor() {
    super();
    this.buddy = new User();
  }

  serialize(): Object {
    return Object.assign(super.serialize(), {
      buddyId: this.buddy.id,
      buddyEmail: this.buddy.email,
    });
  }

  deserialize(input: any): void {
    super.deserialize(input);
    if (input.buddyId) {
      this.buddy.id = input.buddyId;
    }
    if (input.buddyEmail) {
      this.buddy.email = input.buddyEmail;
    }
    if (input.buddyFirstBooking) {
      this.buddy.firstBooking = input.buddyFirstBooking as BuddyBooking;
    }
  }

  getBackendUrl(): string {
    return "/buddy/";
  }

  async save(): Promise<Buddy> {
    return Ajax.putData(this.getBackendUrl(), this.serialize()).then(
      (result) => {
        this.id = result.objectId;
        return this;
      },
    );
  }

  async delete(): Promise<void> {
    return Ajax.delete(this.getBackendUrl() + this.id).then(() => undefined);
  }

  static async list(): Promise<Buddy[]> {
    return Ajax.get("/buddy/").then((result) => {
      const list: Buddy[] = [];
      (result.json as []).forEach((item) => {
        const e: Buddy = new Buddy();
        e.deserialize(item);
        list.push(e);
      });
      return list;
    });
  }

  static async search(query: string): Promise<BuddySearchResult[]> {
    return Ajax.get("/buddy/search?q=" + encodeURIComponent(query)).then(
      (result) => (result.json as any[]).map(BuddySearchResult.from),
    );
  }

  static createFromRawArray(arr: any[]): Buddy[] {
    return arr.map((buddy) => {
      const res = new Buddy();
      res.deserialize(buddy);
      return res;
    });
  }
}

export interface BuddyBooking {
  enter: Date;
  leave: Date;
  room: string;
  desk: string;
}
