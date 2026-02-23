// GET /notifications?page=1&size=20&unread_only=0|1 需登录
'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import api from '@/lib/api'
import type { GetNotificationsResponse, NotificationItem } from '@/lib/types'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'

export default function NotificationsPage() {
  const router = useRouter()
  const [list, setList] = useState<NotificationItem[]>([])
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [unreadOnly, setUnreadOnly] = useState(false)
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
        await api.post('/notifications/read-all')
        const res = await api.get<GetNotificationsResponse>('/notifications', {
          params: { page, size, unread_only: unreadOnly ? '1' : '0' },
        })
        const data = res.data.data
        setList(data?.notifications ?? [])
        setTotal(data?.total ?? 0)
      } catch {
        setList([])
      } finally {
        setLoading(false)
      }
    }
    fetchList()
  }, [page, unreadOnly, router])

  return (
    <div className="min-h-screen bg-slate-50 py-10 px-4">
      <div className="max-w-2xl mx-auto">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle>消息通知</CardTitle>
            <Button variant={unreadOnly ? 'default' : 'outline'} size="sm" onClick={() => setUnreadOnly((v) => !v)}>
              {unreadOnly ? '全部' : '仅未读'}
            </Button>
          </CardHeader>
          <CardContent className="space-y-4">
            {loading && <p className="text-center text-gray-500">加载中...</p>}
            {!loading && list.length === 0 && <p className="text-center text-gray-500">暂无通知</p>}
            {list.map((n) => (
              <div
                key={n.id}
                className={`p-4 rounded-lg border ${n.is_read ? 'bg-white' : 'bg-blue-50/50'}`}
              >
                <p className="text-sm text-gray-500">
                  {n.actor_name && <span className="font-medium">{n.actor_name}</span>}
                  {' · '}
                  {formatDistanceToNow(new Date(n.created_at), { addSuffix: true, locale: zhCN })}
                </p>
                <p className="mt-1">{n.content}</p>
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
