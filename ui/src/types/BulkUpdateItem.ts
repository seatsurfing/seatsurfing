export default class BulkUpdateItem {
  id: string;
  success: boolean;

  constructor() {
    this.id = "";
    this.success = false;
  }

  deserialize(input: any): BulkUpdateItem {
    this.id = input.id;
    this.success = input.success;
    return this;
  }
}
