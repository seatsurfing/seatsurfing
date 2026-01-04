import { Entity } from "./Entity";
import Ajax from "../util/Ajax";
import SpaceAttributeValue from "./SpaceAttributeValue";

export default class Location extends Entity {
  name: string;
  description: string;
  maxConcurrentBookings: number;
  timezone: string;
  enabled: boolean;
  mapWidth: number;
  mapHeight: number;
  mapMimeType: string;
  mapScale: number;

  constructor() {
    super();
    this.name = "";
    this.description = "";
    this.maxConcurrentBookings = 0;
    this.timezone = "";
    this.enabled = true;
    this.mapWidth = 0;
    this.mapHeight = 0;
    this.mapMimeType = "";
    this.mapScale = 1.0;
  }

  serialize(): Object {
    return Object.assign(super.serialize(), {
      name: this.name,
      description: this.description,
      maxConcurrentBookings: this.maxConcurrentBookings,
      timezone: this.timezone,
      enabled: this.enabled,
      mapScale: this.mapScale,
    });
  }

  deserialize(input: any): void {
    super.deserialize(input);
    this.name = input.name;
    this.description = input.description;
    this.maxConcurrentBookings = input.maxConcurrentBookings;
    this.timezone = input.timezone;
    this.enabled = input.enabled;
    this.mapWidth = input.mapWidth;
    this.mapHeight = input.mapHeight;
    this.mapMimeType = input.mapMimeType;
    this.mapScale = input.mapScale;
  }

  getBackendUrl(): string {
    return "/location/";
  }

  getMapUrl(): string {
    return "/location/" + this.id + "/map";
  }

  async save(): Promise<Location> {
    return Ajax.saveEntity(this, this.getBackendUrl()).then(() => this);
  }

  async delete(): Promise<void> {
    return Ajax.delete(this.getBackendUrl() + this.id).then(() => undefined);
  }

  async getMap(): Promise<LocationMap> {
    return Ajax.get(this.getMapUrl()).then((result) => {
      return {
        width: result.json.width,
        height: result.json.height,
        mimeType: result.json.mimeType,
        scale: result.json.scale,
        data: result.json.data,
      } as LocationMap;
    });
  }

  async setMap(file: File): Promise<void> {
    return Ajax.postData(this.getBackendUrl() + this.id + "/map", file).then(
      () => undefined,
    );
  }

  async getAttributes(): Promise<SpaceAttributeValue[]> {
    return Ajax.get(this.getBackendUrl() + this.id + "/attribute").then(
      (result) => {
        let list: SpaceAttributeValue[] = [];
        (result.json as []).forEach((item) => {
          let e: SpaceAttributeValue = new SpaceAttributeValue();
          e.deserialize(item);
          list.push(e);
        });
        return list;
      },
    );
  }

  async setAttribute(attributeId: string, value: string): Promise<void> {
    let payload = {
      value: value,
    };
    return Ajax.postData(
      this.getBackendUrl() + this.id + "/attribute/" + attributeId,
      payload,
    ).then(() => undefined);
  }

  async deleteAttribute(attributeId: string): Promise<void> {
    return Ajax.delete(
      this.getBackendUrl() + this.id + "/attribute/" + attributeId,
    ).then(() => undefined);
  }

  static async get(id: string): Promise<Location> {
    return Ajax.get("/location/" + id).then((result) => {
      let e: Location = new Location();
      e.deserialize(result.json);
      return e;
    });
  }

  static async list(): Promise<Location[]> {
    return Ajax.get("/location/").then((result) => {
      let list: Location[] = [];
      (result.json as []).forEach((item) => {
        let e: Location = new Location();
        e.deserialize(item);
        list.push(e);
      });
      return list;
    });
  }
}

export interface LocationMap {
  width: number;
  height: number;
  scale: number;
  mimeType: string;
  data: string;
}
