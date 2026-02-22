// GET /draft?page=1&size=20 需登录，响应 data.drafts, data.total, data.page, data.size
'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import api from '@/lib/api'
import type { GetDraftResponse, PostListItem } from '@/lib/types'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'

export default function DraftsPage() {
  const router = useRouter()
  const [list, setList] = useState<PostListItem[]>([])
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const size = 20

  useEffect(() => {
    const token = typeof window !== 'undefined' ? localStorage.getItem('token') : null
    if (!token) {
      router.push('/login')
      return
    }

    const fetchList = async () => {
      setLoading(true)
      try {
        const res = await api.get<GetDraftResponse>('/draft', {
          params: { page, size },
        })
        const data = res.data.data ?? res.data
        const drafts = Array.isArray(data?.drafts) ? data.drafts : []
        setList(drafts)
        setTotal(Number(data?.total) ?? 0)
      } catch (e) {
        setList([])
        setTotal(0)
      } finally {
        setLoading(false)
      }
    }
    fetchList()
  }, [page])

  return (
    <div className="min-h-screen bg-slate-50 py-10 px-4">
      <div className="max-w-2xl mx-auto">
        <Card>
          <CardHeader>
            <CardTitle>我的草稿</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {loading && <p className="text-center text-gray-500">加载中...</p>}
            {!loading && list.length === 0 && <p className="text-center text-gray-500">暂无草稿</p>}
            {list.map((item) => (
              <div
                key={item.ID}
                className="p-4 rounded-lg border bg-white cursor-pointer hover:bg-gray-50 flex justify-between items-center"
                onClick={() => router.push(`/posts/${item.ID}/edit`)}
              >
                <p className="font-medium">{item.Title || '无标题'}</p>
                <span className="text-sm text-gray-500">
                  {formatDistanceToNow(new Date(item.CreateAt), { addSuffix: true, locale: zhCN })}
                </span>
              </div>
            ))}
            {total > size && (
              <div className="flex justify-center gap-2 pt-4">
                <Button variant="outline" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
                  上一页
                </Button>
                <span className="flex items-center px-4 text-sm text-gray-600">
                  {page} / {Math.ceil(total / size)}
                </span>
                <Button
                  variant="outline"
                  disabled={page >= Math.ceil(total / size)}
                  onClick={() => setPage((p) => p + 1)}
                >
                  下一页
                </Button>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
