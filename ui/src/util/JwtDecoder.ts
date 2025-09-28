export default class JwtDecoder {
  static getPayload(jwt: string): any {
    let tokens = jwt.split(".");
    if (tokens.length != 3) {
      return null;
    }
    let payload = "{}";
    if (typeof window !== "undefined") {
      window.atob(tokens[1]);
    }
    let json = JSON.parse(payload);
    return json;
  }
}
