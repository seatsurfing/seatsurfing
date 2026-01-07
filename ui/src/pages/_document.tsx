import Document, {
  Html,
  Head,
  Main,
  NextScript,
  DocumentProps,
} from "next/document";
import { randomBytes } from "crypto";
import RuntimeConfig from "@/components/RuntimeConfig";

type Props = DocumentProps & {
  // add custom document props
};

class Doc extends Document<Props> {
  render() {
    const nonce = randomBytes(128).toString("base64");
    let csp = new Map<string, string[]>();
    csp.set("default-src", ["'self'"]);
    csp.set("img-src", ["'self'", "data:", "https:", "'unsafe-eval'", "blob:"]);
    csp.set("style-src", ["'self'", "data:", "'unsafe-inline'"]);
    csp.set("object-src", ["data:", "'unsafe-eval'"]);
    csp.set("base-uri", ["'none'"]);
    csp.set("script-src", [
      "'self'",
      "'nonce-" + nonce + "'",
      "'strict-dynamic'",
    ]);
    csp.set("connect-src", ["'self'", "data:"]);
    if (process.env.NODE_ENV.toLowerCase() === "development") {
      csp.set("connect-src", ["'self'", "http://localhost:8080", "data:"]);
      csp.set(
        "script-src",
        Object.assign(
          [],
          csp.get("script-src")?.concat(["'unsafe-eval'", "'unsafe-inline'"]),
        ),
      );
    }
    let cspString = ``;
    csp.keys().forEach((key) => {
      cspString += `${key} ${csp.get(key)?.join(" ")}; `;
    });
    return (
      <Html lang={RuntimeConfig.getLanguage()}>
        <Head nonce={nonce}>
          <meta name="robots" content="noindex" />
          <meta httpEquiv="Content-Security-Policy" content={cspString} />
        </Head>
        <body>
          <Main />
          <NextScript nonce={nonce} />
        </body>
      </Html>
    );
  }
}

export default Doc;
