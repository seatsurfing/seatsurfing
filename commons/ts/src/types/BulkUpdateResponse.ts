import BulkUpdateItem from "./BulkUpdateItem";

export default class BulkUpdateResponse {
  creates: BulkUpdateItem[];
  updates: BulkUpdateItem[];
  deletes: BulkUpdateItem[];

  constructor() {
    this.creates = [];
    this.updates = [];
    this.deletes = [];
  }

  deserialize(input: any): void {
    if (input.creates) {
      this.creates = input.creates.map((e: any) =>
        new BulkUpdateItem().deserialize(e),
      );
    }
    if (input.updates) {
      this.updates = input.updates.map((e: any) =>
        new BulkUpdateItem().deserialize(e),
      );
    }
    if (input.deletes) {
      this.deletes = input.deletes.map((e: any) =>
        new BulkUpdateItem().deserialize(e),
      );
    }
  }
}
