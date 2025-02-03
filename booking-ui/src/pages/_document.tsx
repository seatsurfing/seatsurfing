import Document, { Html, Head, Main, NextScript, DocumentProps } from 'next/document'
import i18nextConfig from '../../next-i18next.config'
import { randomBytes } from 'crypto'

type Props = DocumentProps & {
  // add custom document props
}

class Doc extends Document<Props> {
  render() {
    const nonce = randomBytes(128).toString('base64')
    const csp = `object-src 'none'; base-uri 'none'; script-src 'self' 'unsafe-eval' 'unsafe-inline' https: 'nonce-${nonce}' 'strict-dynamic'`
    const currentLocale =
      this.props.__NEXT_DATA__.locale ??
      i18nextConfig.i18n.defaultLocale
    return (
      <Html lang={currentLocale}>
        <Head nonce={nonce}>
          <meta name="robots" content="noindex" />
          <meta httpEquiv="Content-Security-Policy" content={csp} />
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