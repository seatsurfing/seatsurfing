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
    const scheme = req.headers.get('X-Forwarded-Proto') || req.nextUrl.protocol;
    const host = req.headers.get('X-Forwarded-Host') || req.nextUrl.host;
    const reqUrl = scheme + "://" + host;
    const url = new URL(
      `/ui/${locale}${req.nextUrl.pathname}${req.nextUrl.search}`,
      reqUrl,
    );
    return NextResponse.redirect(url);
  }
}
