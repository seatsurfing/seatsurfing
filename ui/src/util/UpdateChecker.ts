import Ajax from "@/util/Ajax";

export default class UpdateChecker {
  static async check(): Promise<{
    version: string;
    updateAvailable: boolean;
  }> {
    try {
      const res = await Ajax.get("/uc/");
      return res.json;
    } catch {
      console.warn("Could not check for updates.");
      return { version: "", updateAvailable: false };
    }
  }
}
