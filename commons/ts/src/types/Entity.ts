export abstract class Entity {
  id: string = "";

  constructor(id?: string) {
    if (id) {
      this.id = id;
    }
  }

  serialize(): Object {
    return {};
  }

  deserialize(input: any): void {
    this.id = input.id;
  }

  abstract getBackendUrl(): string;
}
