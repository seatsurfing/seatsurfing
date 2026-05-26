import { Entity } from "./Entity";
import Ajax from "../util/Ajax";
import Location from "./Location";
import DateUtil from "../util/DateUtil";
import BulkUpdateResponse from "./BulkUpdateResponse";
import SpaceAttributeValue from "./SpaceAttributeValue";
import SearchAttribute from "./SearchAttribute";
import Group from "./Group";

export default class Space extends Entity {
  name: string;
  x: number;
  y: number;
  width: number;
  height: number;
  rotation: number;
  requireSubject: boolean;
  enabled: boolean;
  kioskEnabled: boolean;
  shape: string;
  attributes: SpaceAttributeValue[];
  approverGroupIds: string[];
  allowedBookerGroupIds: string[];
  available: boolean;
  locationId: string;
  location: Location;
  rawBookings: any[];
  allowed: boolean;
  approvalRequired: boolean;

  constructor() {
    super();
    this.name = "";
    this.x = 0;
    this.y = 0;
    this.width = 0;
    this.height = 0;
    this.rotation = 0;
    this.requireSubject = false;
    this.enabled = true;
    this.kioskEnabled = false;
    this.shape = "";
    this.attributes = [];
    this.approverGroupIds = [];
    this.allowedBookerGroupIds = [];
    this.available = false;
    this.locationId = "";
    this.location = new Location();
    this.rawBookings = [];
    this.allowed = true;
    this.approvalRequired = false;
  }

  normalizePosition(): {
    x: number;
    y: number;
    width: number;
    height: number;
    rotation: number;
  } {
    const { x, y, width: w, height: h, rotation } = this;
    const r = ((rotation % 360) + 360) % 360;

    // Representation A: current (x, y), size (w, h), rotation r
    // Representation B: same visual center, swapped dimensions, rotation r+90
    // Both are visually identical (same center, same shape, just described differently).
    const cx = x + w / 2;
    const cy = y + h / 2;
    const xB = cx - h / 2;
    const yB = cy - w / 2;
    const rB = (r + 90) % 360;

    const aOk = x >= 0 && y >= 0;
    const bOk = xB >= 0 && yB >= 0;

    if (aOk)
      return {
        x: Math.round(x),
        y: Math.round(y),
        width: w,
        height: h,
        rotation: r,
      };
    if (bOk)
      return {
        x: Math.round(xB),
        y: Math.round(yB),
        width: h,
        height: w,
        rotation: rB,
      };

    // For arbitrary angles where neither representation gives non-negative coords,
    // clamp as last resort — the space may shift slightly on screen.
    const useB = xB + yB > x + y;
    return useB
      ? {
          x: Math.max(0, Math.round(xB)),
          y: Math.max(0, Math.round(yB)),
          width: h,
          height: w,
          rotation: rB,
        }
      : {
          x: Math.max(0, Math.round(x)),
          y: Math.max(0, Math.round(y)),
          width: w,
          height: h,
          rotation: r,
        };
  }

  serialize(): Object {
    const { x, y, width, height, rotation } = this.normalizePosition();
    return Object.assign(super.serialize(), {
      id: this.id,
      name: this.name,
      x,
      y,
      width,
      height,
      rotation,
      requireSubject: this.requireSubject,
      enabled: this.enabled,
      kioskEnabled: this.kioskEnabled,
      shape: this.shape,
      attributes: this.attributes.map((a) => a.serialize()),
      approverGroupIds: this.approverGroupIds,
      allowedBookerGroupIds: this.allowedBookerGroupIds,
    });
  }

  deserialize(input: any): void {
    super.deserialize(input);
    this.name = input.name;
    this.locationId = input.locationId;
    this.x = input.x;
    this.y = input.y;
    this.width = input.width;
    this.height = input.height;
    this.rotation = input.rotation;
    this.requireSubject = input.requireSubject;
    this.enabled = input.enabled;
    this.kioskEnabled = input.kioskEnabled ?? false;
    this.shape = input.shape ?? "";
    if (input.allowed !== undefined) {
      this.allowed = input.allowed;
    }
    if (input.approvalRequired !== undefined) {
      this.approvalRequired = input.approvalRequired;
    }
    if (input.available) {
      this.available = input.available;
    }
    if (input.location) {
      this.location.deserialize(input.location);
    }
    if (input.bookings && Array.isArray(input.bookings)) {
      this.rawBookings = input.bookings;
    }
    if (input.attributes) {
      this.attributes = input.attributes.map((a: any) => {
        const e = new SpaceAttributeValue();
        e.deserialize(a);
        return e;
      });
    }
    if (input.approverGroupIds) {
      this.approverGroupIds = input.approverGroupIds;
    }
    if (input.allowedBookerGroupIds) {
      this.allowedBookerGroupIds = input.allowedBookerGroupIds;
    }
  }

  getBackendUrl(): string {
    return `/location/${encodeURIComponent(this.locationId)}/space/`;
  }

  async save(): Promise<Space> {
    await Ajax.saveEntity(this, this.getBackendUrl());
    return this;
  }

  async delete(): Promise<void> {
    await Ajax.delete(`${this.getBackendUrl()}${encodeURIComponent(this.id)}`);
  }

  async getApprovers(): Promise<Group[]> {
    const result = await Ajax.get(
      `${this.getBackendUrl()}${encodeURIComponent(this.id)}/approver`,
    );
    const list: Group[] = [];
    (result.json as []).forEach((item) => {
      const e: Group = new Group();
      e.deserialize(item);
      list.push(e);
    });
    return list;
  }

  async addApprovers(groupIds: string[]): Promise<void> {
    await Ajax.putData(
      `${this.getBackendUrl()}${encodeURIComponent(this.id)}/approver`,
      groupIds,
    );
  }

  async removeApprovers(groupIds: string[]): Promise<void> {
    await Ajax.postData(
      `${this.getBackendUrl()}${encodeURIComponent(this.id)}/approver/remove`,
      groupIds,
    );
  }

  async getAllowedBookers(): Promise<Group[]> {
    const result = await Ajax.get(
      `${this.getBackendUrl()}${encodeURIComponent(this.id)}/allowedbooker`,
    );
    const list: Group[] = [];
    (result.json as []).forEach((item) => {
      const e: Group = new Group();
      e.deserialize(item);
      list.push(e);
    });
    return list;
  }

  async addAllowedBookers(groupIds: string[]): Promise<void> {
    await Ajax.putData(
      `${this.getBackendUrl()}${encodeURIComponent(this.id)}/allowedbooker`,
      groupIds,
    );
  }

  async removeAllowedBookers(groupIds: string[]): Promise<void> {
    await Ajax.postData(
      `${this.getBackendUrl()}${encodeURIComponent(this.id)}/allowedbooker/remove`,
      groupIds,
    );
  }

  static async get(locationId: string, id: string): Promise<Space> {
    const result = await Ajax.get(
      `/location/${encodeURIComponent(locationId)}/space/${encodeURIComponent(id)}`,
    );
    const e = new Space();
    e.deserialize(result.json);
    return e;
  }

  private static deserializeList(json: []): Space[] {
    return json.map((item) => {
      const e = new Space();
      e.deserialize(item);
      return e;
    });
  }

  static async list(locationId: string): Promise<Space[]> {
    const result = await Ajax.get(
      `/location/${encodeURIComponent(locationId)}/space/`,
    );
    return Space.deserializeList(result.json as []);
  }

  static async listAvailability(
    locationId: string,
    enter: Date,
    leave: Date,
    attributes?: SearchAttribute[],
  ): Promise<Space[]> {
    let params = `enter=${encodeURIComponent(DateUtil.convertToFakeUTCDate(enter).toISOString())}`;
    params += `&leave=${encodeURIComponent(DateUtil.convertToFakeUTCDate(leave).toISOString())}`;
    if (attributes && attributes.length > 0) {
      params += `&attributes=${encodeURIComponent(JSON.stringify(attributes.map((a) => a.serialize())))}`;
    }
    const result = await Ajax.get(
      `/location/${encodeURIComponent(locationId)}/space/availability?${params}`,
    );
    return Space.deserializeList(result.json as []);
  }

  static async listSingleAvailability(
    locationId: string,
    spaceId: string,
    enter: Date,
    leave: Date,
  ): Promise<Space[]> {
    const params = `enter=${encodeURIComponent(DateUtil.convertToFakeUTCDate(enter).toISOString())}&leave=${encodeURIComponent(DateUtil.convertToFakeUTCDate(leave).toISOString())}`;
    const result = await Ajax.get(
      `/location/${encodeURIComponent(locationId)}/space/${encodeURIComponent(spaceId)}/availability?${params}`,
    );
    return Space.deserializeList(result.json as []);
  }

  static async bulkUpdate(
    locationId: string,
    creates: Space[],
    updates: Space[],
    deleteIds: string[],
  ): Promise<BulkUpdateResponse> {
    const payload = {
      creates: creates.map((s) => s.serialize()),
      updates: updates.map((s) => s.serialize()),
      deleteIds: deleteIds,
    };
    const result = await Ajax.postData(
      `/location/${encodeURIComponent(locationId)}/space/bulk`,
      payload,
    );
    const e = new BulkUpdateResponse();
    e.deserialize(result);
    return e;
  }
}
