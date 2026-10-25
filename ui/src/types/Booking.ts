import Location from "./Location";
import Space from "./Space";
import { Entity } from "./Entity";
import Ajax from "../util/Ajax";
import User from "./User";
import Formatting from "../util/Formatting";
import DateUtil from "../util/DateUtil";
import { BookingAttributeValue } from "./BookingAttribute";

export default class Booking extends Entity {
  enter: Date;
  leave: Date;
  location: Location;
  space: Space;
  user: User;
  approved: boolean;
  subject: string;
  recurringId: string;
  public: boolean;
  attributes: BookingAttributeValue[];

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
    this.public = false;
    this.attributes = [];
  }

  serialize(): Object {
    // Convert the local dates to UTC dates without changing the date/time ("fake" UTC)
    let enter = DateUtil.convertToFakeUTCDate(this.enter);
    let leave = DateUtil.convertToFakeUTCDate(this.leave);

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
    if (input.userFirstname) {
      this.user.firstname = input.userFirstname;
    }
    if (input.userLastname) {
      this.user.lastname = input.userLastname;
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
    if (input.public !== undefined) {
      this.public = input.public;
    }
    if (Array.isArray(input.attributes)) {
      this.attributes = input.attributes;
    }
  }

  getBackendUrl(): string {
    return "/booking/";
  }

  isRecurring(): boolean {
    return this.recurringId !== undefined && this.recurringId !== "";
  }

  async save(): Promise<Booking> {
    await Ajax.saveEntity(this, this.getBackendUrl());
    return this;
  }

  async delete(): Promise<void> {
    await Ajax.delete(this.getBackendUrl() + this.id);
  }

  async approve(approved: boolean): Promise<void> {
    const payload = {
      approved,
    };
    await Ajax.postData(this.getBackendUrl() + this.id + "/approve", payload);
  }

  static async get(id: string): Promise<Booking> {
    const result = await Ajax.get("/booking/" + id);
    const e: Booking = new Booking();
    e.deserialize(result.json);
    return e;
  }

  static async list(): Promise<Booking[]> {
    const result = await Ajax.get("/booking/");
    return Booking.createFromRawArray(result.json);
  }

  static async listPendingApprovals(): Promise<Booking[]> {
    const result = await Ajax.get("/booking/pendingapprovals/");
    return Booking.createFromRawArray(result.json);
  }

  static async getPendingApprovalsCount(): Promise<number> {
    const result = await Ajax.get("/booking/pendingapprovals/count");
    return result.json.count as number;
  }

  static async listFiltered(
    start: Date,
    end: Date,
    users: string[],
    location: string,
    subject: string = "",
  ): Promise<Booking[]> {
    const queryParams = new URLSearchParams();
    queryParams.set(
      "start",
      DateUtil.convertToFakeUTCDate(start).toISOString(),
    );
    queryParams.set("end", DateUtil.convertToFakeUTCDate(end).toISOString());
    users.forEach((user) => queryParams.append("user", user));
    if (location) queryParams.set("location", location);
    if (subject) queryParams.set("subject", subject);
    const result = await Ajax.get(`/booking/filter/?${queryParams.toString()}`);
    return Booking.createFromRawArray(result.json);
  }

  static async listCurrent(
    users: string[],
    location: string,
    subject: string = "",
  ): Promise<Booking[]> {
    const queryParams = new URLSearchParams();
    users.forEach((user) => queryParams.append("user", user));
    if (location) queryParams.set("location", location);
    if (subject) queryParams.set("subject", subject);
    const params = queryParams.toString() ? `?${queryParams.toString()}` : "";

    const result = await Ajax.get(`/booking/current/${params}`);
    return Booking.createFromRawArray(result.json);
  }

  static createFromRawArray(arr: any[]): Booking[] {
    return arr.map((booking) => {
      const res = new Booking();
      res.deserialize(booking);
      return res;
    });
  }
}
