import { NextRequest, NextResponse } from 'next/server';

const PUBLIC_FILE = /\.(.*)$/;

export async function middleware(req: NextRequest) {
  if (
    req.nextUrl.pathname.startsWith('/_next') ||
    req.nextUrl.pathname.startsWith('/ui/_next') ||
    req.nextUrl.pathname.includes('/api/') ||
    req.nextUrl.pathname.includes('/ui/api/') ||
    PUBLIC_FILE.test(req.nextUrl.pathname)
  ) {
    return;
  }

  if (req.nextUrl.locale === 'default') {
    const locale = req.cookies.get('NEXT_LOCALE')?.value || 'en';
    let scheme = (req.headers.get('X-Forwarded-Proto') || req.nextUrl.protocol).toLowerCase();
    if ((scheme !== 'http') && (scheme !== 'https')) {
      scheme = 'https';
    }
    let port = req.headers.get('X-Forwarded-Port') || req.nextUrl.port;
    if ((port === '80') && (scheme === 'http')) {
      port = '';
    } else if ((port === '443') && (scheme === 'https')) {
      port = '';
    }
    const host = req.headers.get('X-Forwarded-Host') || req.nextUrl.host;
    if ((port !== '') && (host.includes(':'))) {
      port = '';
    }
    const reqUrl = scheme + "://" + host + (port ? ':' + port : '');
    const url = new URL(
      `/ui/${locale}${req.nextUrl.pathname}${req.nextUrl.search}`,
      reqUrl,
    );
    return NextResponse.redirect(url);
  }
}
