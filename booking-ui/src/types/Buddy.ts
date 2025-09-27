import { Entity } from "./Entity";
import Ajax from "../util/Ajax";
import User from "./User";

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
    return Ajax.saveEntity(this, this.getBackendUrl()).then(() => this);
  }

  async delete(): Promise<void> {
    return Ajax.delete(this.getBackendUrl() + this.id).then(() => undefined);
  }

  static async list(): Promise<Buddy[]> {
    return Ajax.get("/buddy/").then((result) => {
      let list: Buddy[] = [];
      (result.json as []).forEach((item) => {
        let e: Buddy = new Buddy();
        e.deserialize(item);
        list.push(e);
      });
      return list;
    });
  }

  static createFromRawArray(arr: any[]): Buddy[] {
    return arr.map((buddy) => {
      let res = new Buddy();
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
