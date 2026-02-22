'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { usePathname } from 'next/navigation'

export default function Nav() {
  const pathname = usePathname()
  const [username, setUsername] = useState<string | null>(null)

  useEffect(() => {
    setUsername(typeof window !== 'undefined' ? localStorage.getItem('username') : null)
  }, [pathname])

  const link = (href: string, label: string) => (
    <Link
      href={href}
      className={`px-3 py-2 rounded-md text-sm font-medium ${
        pathname === href ? 'bg-slate-200 text-slate-900' : 'text-slate-600 hover:bg-slate-100'
      }`}
    >
      {label}
    </Link>
  )

  return (
    <header className="border-b bg-white/80 backdrop-blur sticky top-0 z-10">
      <div className="max-w-5xl mx-auto px-4 h-14 flex items-center justify-between">
        <nav className="flex items-center gap-1">
          {link('/', '首页')}
          {username && (
            <>
              {link('/profile', '个人中心')}
              {link('/notifications', '消息')}
              {link('/favorites', '收藏')}
              {link('/drafts', '草稿')}
            </>
          )}
        </nav>
        <div className="flex items-center gap-2">
          {username ? (
            <>
              <span className="text-sm text-slate-600">欢迎，{username}</span>
              <Link
                href="/login"
                onClick={() => {
                  localStorage.removeItem('token')
                  localStorage.removeItem('user_id')
                  localStorage.removeItem('username')
                }}
                className="text-sm text-slate-600 hover:underline"
              >
                退出
              </Link>
            </>
          ) : (
            <>
              <Link href="/login" className="text-sm font-medium text-slate-700 hover:underline">
                登录
              </Link>
              <Link href="/register" className="text-sm font-medium text-primary hover:underline">
                注册
              </Link>
            </>
          )}
        </div>
      </div>
    </header>
  )
}
