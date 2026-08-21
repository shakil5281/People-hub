"use client"

import Image from "next/image"
import { LoginForm } from "@/components/form/login-form"
import { GalleryVerticalEndIcon } from "lucide-react"

export default function LoginPage() {
  return (
    <div className="grid min-h-svh lg:grid-cols-2">
      <div className="flex flex-col gap-4 p-6 md:p-10">
        <div className="flex justify-center gap-2 md:justify-start">
          <a href="#" className="flex items-center gap-2 font-medium">
            <div className="flex size-8 items-center justify-center rounded-md bg-gradient-to-br from-emerald-500 to-purple-600 text-white">
              <svg viewBox="0 0 32 32" className="size-5" fill="none">
                <circle cx="10" cy="10" r="2.5" fill="currentColor" />
                <circle cx="22" cy="10" r="2.5" fill="currentColor" />
                <circle cx="16" cy="22" r="2.5" fill="currentColor" />
                <path d="M10 13v2.5a2 2 0 0 0 2 2h1.5" className="stroke-[2]" stroke="currentColor" strokeLinecap="round" />
                <path d="M22 13v2.5a2 2 0 0 1-2 2h-1.5" className="stroke-[2]" stroke="currentColor" strokeLinecap="round" />
                <path d="M13.5 18.5L16 22l2.5-3.5" className="stroke-[2]" stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" />
              </svg>
            </div>
            <span className="bg-gradient-to-r from-emerald-500 to-purple-600 bg-clip-text text-transparent font-semibold">People Hub</span>
          </a>
        </div>
        <div className="flex flex-1 items-center justify-center">
          <div className="w-full max-w-xs">
            <LoginForm />
          </div>
        </div>
      </div>
      <div className="relative hidden bg-muted lg:block">
        <Image
          src="https://images.unsplash.com/photo-1522071820081-009f0129c71c?auto=format&fit=crop&w=1200&q=80"
          alt="Office team working"
          fill
          className="object-cover dark:brightness-[0.2] dark:grayscale"
        />
      </div>
    </div>
  )
}
