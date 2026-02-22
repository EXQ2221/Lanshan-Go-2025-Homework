// src/app/page.tsx
'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useQuery } from '@tanstack/react-query'
import api from '@/lib/api'
import type { ListPostsResponse, PostListItem } from '@/lib/types'
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'

const PAGE_SIZE = 20

export default function Home() {
  const router = useRouter()
  const [username, setUsername] = useState<string | null>(null)
  const [page, setPage] = useState(1)
  const [keyword, setKeyword] = useState('')
  const [searchInput, setSearchInput] = useState('')

  useEffect(() => {
    const stored = localStorage.getItem('username')
    if (stored) setUsername(stored)
  }, [])

  const handleLogout = () => {
    localStorage.clear()
    router.push('/login')
  }

  // GET /posts?page=1&size=20&type=0&keyword= 公开接口，无需登录
  const { data, isLoading, error } = useQuery<ListPostsResponse>({
    queryKey: ['posts', page, keyword],
    queryFn: async () => {
      const res = await api.get<ListPostsResponse>('/posts', {
        params: { page, size: PAGE_SIZE, keyword: keyword || undefined },
      })
      return res.data
    },
  })

  const list: PostListItem[] = data?.list ?? []
  const total = data?.total ?? 0
  const totalPages = Math.ceil(total / PAGE_SIZE) || 1

  const handleSearch = () => {
    setKeyword(searchInput.trim())
    setPage(1)
  }

  return (
    <div className="min-h-screen bg-slate-50 p-6">
      <div className="max-w-5xl mx-auto">
        <div className="flex justify-between items-center mb-10">
          <h1 className="text-4xl font-bold text-gray-900">我的论坛</h1>
          {username ? (
            <div className="flex items-center gap-6">
              <span
                className="text-lg text-gray-700 cursor-pointer hover:underline"
                onClick={() => router.push('/profile')}
              >
                欢迎，{username}
              </span>
              <Button variant="outline" onClick={handleLogout}>
                退出
              </Button>
            </div>
          ) : (
            <Button onClick={() => router.push('/login')}>
              登录 / 注册
            </Button>
          )}
        </div>

        {/* 搜索：keyword 与后端 form keyword 对齐 */}
        <div className="mb-6 flex gap-2">
          <Input
            placeholder="搜索标题或内容"
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
            className="max-w-sm"
          />
          <Button variant="secondary" onClick={handleSearch}>搜索</Button>
        </div>

        {username && (
          <div className="mb-8">
            <Button size="lg" onClick={() => router.push('/posts/create')}>
              发布新帖
            </Button>
          </div>
        )}

        <div className="space-y-6">
          {isLoading && <p className="text-center text-gray-500">加载中...</p>}
          {error && (
            <p className="text-center text-red-500">
              加载失败：{(error as Error).message}
            </p>
          )}

          {!isLoading && list.length === 0 && (
            <Card>
              <CardHeader>
                <CardTitle>暂无帖子</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-gray-600">
                  {username ? '快来发布第一篇吧！' : '登录后即可发帖、参与讨论'}
                </p>
              </CardContent>
            </Card>
          )}

          {list.length > 0 && list.map((post) => (
            <Card
              key={post.ID}
              className="hover:shadow-lg transition-shadow cursor-pointer"
              onClick={() => router.push(`/posts/${post.ID}`)}
            >
              <CardHeader className="pb-2">
                <div className="flex justify-between items-start">
                  <div>
                    <span className="inline-block px-3 py-1 bg-blue-100 text-blue-800 rounded-full text-sm mr-3">
                      {post.Type === 1 ? '文章' : '问题'}
                    </span>
                    <CardTitle className="text-xl inline">{post.Title || '无标题'}</CardTitle>
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                <div className="flex items-center gap-4 text-sm text-gray-600">
                  <span>作者：{post.AuthorName || '匿名'}</span>
                  <span>•</span>
                  <span>
                    {post.CreateAt && !isNaN(new Date(post.CreateAt).getTime())
                      ? formatDistanceToNow(new Date(post.CreateAt), { addSuffix: true, locale: zhCN })
                      : '时间未知'}
                  </span>
                </div>
              </CardContent>
            </Card>
          ))}

          {totalPages > 1 && (
            <div className="flex justify-center gap-2 pt-4">
              <Button
                variant="outline"
                disabled={page <= 1}
                onClick={() => setPage((p) => Math.max(1, p - 1))}
              >
                上一页
              </Button>
              <span className="flex items-center px-4 text-sm text-gray-600">
                {page} / {totalPages}（共 {total} 条）
              </span>
              <Button
                variant="outline"
                disabled={page >= totalPages}
                onClick={() => setPage((p) => p + 1)}
              >
                下一页
              </Button>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
