import Ajax from "@/util/Ajax";
import Navigation from "@/util/Navigation";

export default class UpdateChecker {
  static async check(): Promise<{
    version: string;
    updateAvailable: boolean;
  }> {
    try {
      const res = await Ajax.get(`${Navigation.PATH_API_UC}/`);
      return res.json;
    } catch {
      console.warn("Could not check for updates.");
      return { version: "", updateAvailable: false };
    }
  }
}
