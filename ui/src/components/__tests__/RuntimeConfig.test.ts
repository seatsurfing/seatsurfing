import { describe, it, expect } from "vitest";
import RuntimeConfig from "../RuntimeConfig";

describe("RuntimeConfig.parseArray", () => {
  it("returns parsed arrays", () => {
    expect(RuntimeConfig.parseArray('[{"id":"a"}]')).toEqual([{ id: "a" }]);
  });

  it("returns an empty array for non-arrays and invalid JSON", () => {
    for (const v of ["", "true", "1", '{"id":"a"}', "null", "[{"]) {
      expect(RuntimeConfig.parseArray(v)).toEqual([]);
    }
  });
});
