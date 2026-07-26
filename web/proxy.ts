import { NextResponse } from "next/server"
import type { NextRequest } from "next/server"

const publicPaths = ["/login", "/register"]

export function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl
  const basePath = request.nextUrl.basePath || ""
  const token = request.cookies.get("auth_token")?.value

  const isPublic = publicPaths.some((p) => pathname === p || pathname.startsWith(p + "/"))

  if (isPublic) {
    if (token) {
      return NextResponse.redirect(new URL(`${basePath}/`, request.url))
    }
    return NextResponse.next()
  }

  if (!token) {
    const loginUrl = new URL(`${basePath}/login`, request.url)
    loginUrl.searchParams.set("redirect", pathname)
    return NextResponse.redirect(loginUrl)
  }

  return NextResponse.next()
}

export const config = {
  matcher: ["/((?!api|uploads|_next/static|_next/image|favicon.ico|placeholder.svg).*)"],
}
