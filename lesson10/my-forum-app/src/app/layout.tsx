// src/app/layout.tsx
import type { Metadata } from "next"
import { Inter } from "next/font/google"
import "./globals.css"
import QueryProvider from "@/components/QueryProvider"
import Nav from "@/components/Nav"
import { Toaster } from "sonner"

const inter = Inter({ subsets: ["latin"] })

export const metadata: Metadata = {
  title: "我的论坛",
  description: "分享想法、提问解惑的小社区",
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang="zh-CN">
      <body className={inter.className}>
        <QueryProvider>
          <Nav />
          <main>{children}</main>
          <Toaster position="top-center" />
        </QueryProvider>
      </body>
    </html>
  )
}