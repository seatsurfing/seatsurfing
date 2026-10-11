import { Entity } from "./Entity";
import Ajax from "../util/Ajax";

export interface BookingAttributeValue {
  attributeId: string;
  label: string;
  type: number;
  value: string;
}

export default class BookingAttribute extends Entity {
  static readonly TYPE_TEXT = 1;
  static readonly TYPE_TEXT_MULTILINE = 2;
  static readonly TYPE_NUMBER = 3;
  static readonly MAX_ATTRIBUTES = 3;

  label: string;
  type: number;
  required: boolean;

  constructor(
    id: string = "",
    label: string = "",
    type: number = BookingAttribute.TYPE_TEXT,
    required: boolean = false,
  ) {
    super(id);
    this.label = label;
    this.type = type;
    this.required = required;
  }

  serialize(): Object {
    return Object.assign(super.serialize(), {
      label: this.label,
      type: this.type,
      required: this.required,
    });
  }

  deserialize(input: any): void {
    super.deserialize(input);
    this.label = input.label;
    this.type = input.type;
    this.required = input.required === true;
  }

  getBackendUrl(): string {
    return "/booking-attribute/";
  }

  async save(): Promise<BookingAttribute> {
    await Ajax.saveEntity(this, this.getBackendUrl());
    return this;
  }

  async delete(): Promise<void> {
    await Ajax.delete(this.getBackendUrl() + this.id);
  }

  static async list(): Promise<BookingAttribute[]> {
    const result = await Ajax.get("/booking-attribute/");
    return BookingAttribute.createFromRawArray(result.json || []);
  }

  static createFromRawArray(arr: any[]): BookingAttribute[] {
    return arr.map((item) => {
      const e = new BookingAttribute();
      e.deserialize(item);
      return e;
    });
  }

  static getTypeTranslationKey(type: number): string {
    if (type === BookingAttribute.TYPE_TEXT_MULTILINE) {
      return "bookingAttributeTypeTextMultiline";
    }
    if (type === BookingAttribute.TYPE_NUMBER) {
      return "bookingAttributeTypeNumber";
    }
    return "bookingAttributeTypeText";
  }
}
