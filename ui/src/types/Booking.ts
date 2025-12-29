import Location from "./Location";
import Space from "./Space";
import { Entity } from "./Entity";
import Ajax from "../util/Ajax";
import User from "./User";
import Formatting from "../util/Formatting";

export default class Booking extends Entity {
  enter: Date;
  leave: Date;
  location: Location;
  space: Space;
  user: User;
  approved: boolean;
  subject: string;
  recurringId: string;

  constructor() {
    super();
    this.enter = new Date();
    this.leave = new Date();
    this.location = new Location();
    this.space = new Space();
    this.user = new User();
    this.approved = false;
    this.subject = "";
    this.recurringId = "";
  }

  serialize(): Object {
    // Convert the local dates to UTC dates without changing the date/time ("fake" UTC)
    let enter = Formatting.convertToFakeUTCDate(this.enter);
    let leave = Formatting.convertToFakeUTCDate(this.leave);

    if (this.user) {
      return Object.assign(super.serialize(), {
        enter: enter.toISOString(),
        leave: leave.toISOString(),
        spaceId: this.space.id,
        subject: this.subject,
        userEmail: this.user.email,
      });
    } else {
      return Object.assign(super.serialize(), {
        enter: enter.toISOString(),
        leave: leave.toISOString(),
        spaceId: this.space.id,
        subject: this.subject,
      });
    }
  }

  deserialize(input: any): void {
    super.deserialize(input);
    // Discard time zone information from date
    input.enter = Formatting.stripTimezoneDetails(input.enter);
    input.leave = Formatting.stripTimezoneDetails(input.leave);
    this.enter = new Date(input.enter);
    this.leave = new Date(input.leave);
    if (input.space) {
      this.space.deserialize(input.space);
    }
    if (input.userId) {
      this.user.id = input.userId;
    }
    if (input.userEmail) {
      this.user.email = input.userEmail;
    }
    if (input.approved !== undefined) {
      this.approved = input.approved;
    }
    if (input.subject) {
      this.subject = input.subject;
    }
    if (input.recurringId) {
      this.recurringId = input.recurringId;
    }
  }

  getBackendUrl(): string {
    return "/booking/";
  }

  isRecurring(): boolean {
    return this.recurringId !== undefined && this.recurringId !== "";
  }

  async save(): Promise<Booking> {
    return Ajax.saveEntity(this, this.getBackendUrl()).then(() => this);
  }

  async delete(): Promise<void> {
    return Ajax.delete(this.getBackendUrl() + this.id).then(() => undefined);
  }

  async approve(approved: boolean): Promise<void> {
    let payload = {
      approved: approved,
    };
    return Ajax.postData(
      this.getBackendUrl() + this.id + "/approve",
      payload,
    ).then(() => undefined);
  }

  static async get(id: string): Promise<Booking> {
    return Ajax.get("/booking/" + id).then((result) => {
      let e: Booking = new Booking();
      e.deserialize(result.json);
      return e;
    });
  }

  static async list(): Promise<Booking[]> {
    return Ajax.get("/booking/").then((result) => {
      let list: Booking[] = [];
      (result.json as []).forEach((item) => {
        let e: Booking = new Booking();
        e.deserialize(item);
        list.push(e);
      });
      return list;
    });
  }

  static async listPendingApprovals(): Promise<Booking[]> {
    return Ajax.get("/booking/pendingapprovals/").then((result) => {
      let list: Booking[] = [];
      (result.json as []).forEach((item) => {
        let e: Booking = new Booking();
        e.deserialize(item);
        list.push(e);
      });
      return list;
    });
  }

  static async getPendingApprovalsCount(): Promise<number> {
    return Ajax.get("/booking/pendingapprovals/count").then((result) => {
      return result.json.count as number;
    });
  }

  static async listFiltered(start: Date, end: Date): Promise<Booking[]> {
    let params =
      "start=" +
      encodeURIComponent(Formatting.convertToFakeUTCDate(start).toISOString());
    params +=
      "&end=" +
      encodeURIComponent(Formatting.convertToFakeUTCDate(end).toISOString());
    return Ajax.get("/booking/filter/?" + params).then((result) => {
      let list: Booking[] = [];
      (result.json as []).forEach((item) => {
        let e: Booking = new Booking();
        e.deserialize(item);
        list.push(e);
      });
      return list;
    });
  }

  static async listCurrent(): Promise<Booking[]> {
    return Ajax.get("/booking/current/").then((result) => {
      let list: Booking[] = [];
      (result.json as []).forEach((item) => {
        let e: Booking = new Booking();
        e.deserialize(item);
        list.push(e);
      });
      return list;
    });
  }

  static createFromRawArray(arr: any[]): Booking[] {
    return arr.map((booking) => {
      let res = new Booking();
      res.deserialize(booking);
      return res;
    });
  }
}
