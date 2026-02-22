// GET /favorites?page=1&size=20 需登录
'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import api from '@/lib/api'
import type { GetFavoritesResponse, FavoriteItem } from '@/lib/types'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'

export default function FavoritesPage() {
  const router = useRouter()
  const [list, setList] = useState<FavoriteItem[]>([])
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
        const res = await api.get<GetFavoritesResponse>('/favorites', {
          params: { page, size },
        })
        const data = res.data.data
        setList(data?.favorites ?? [])
        setTotal(data?.total ?? 0)
      } catch {
        setList([])
      } finally {
        setLoading(false)
      }
    }
    fetchList()
  }, [page, router])

  return (
    <div className="min-h-screen bg-slate-50 py-10 px-4">
      <div className="max-w-2xl mx-auto">
        <Card>
          <CardHeader>
            <CardTitle>我的收藏</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {loading && <p className="text-center text-gray-500">加载中...</p>}
            {!loading && list.length === 0 && <p className="text-center text-gray-500">暂无收藏</p>}
            {list.map((item) => (
              <div
                key={item.id}
                className="p-4 rounded-lg border bg-white cursor-pointer hover:bg-gray-50"
                onClick={() => router.push(`/posts/${item.id}`)}
              >
                <p className="font-medium">{item.title || '无标题'}</p>
                <p className="text-sm text-gray-500">
                  {formatDistanceToNow(new Date(item.created_at), { addSuffix: true, locale: zhCN })}
                </p>
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
