export default class BulkUpdateItem {
  id: string;
  success: boolean;

  constructor() {
    this.id = "";
    this.success = false;
  }

  deserialize(input: any): void {
    this.id = input.id;
    this.success = input.success;
  }
}
