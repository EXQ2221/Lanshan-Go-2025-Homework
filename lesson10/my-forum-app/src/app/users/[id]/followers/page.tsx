// src/app/users/[id]/followers/page.tsx
'use client'

import { useState, useEffect, useRef, useCallback } from 'react'
import { useParams, useRouter } from 'next/navigation'
import api from '@/lib/api'
import { staticUrl } from '@/lib/api'
import type { FollowListResponse, FollowUserInfo } from '@/lib/types'
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Button } from "@/components/ui/button"
import { toast } from "sonner"

export default function FollowersPage() {
  const router = useRouter()
  const params = useParams()
  const userId = params.id ? parseInt(params.id as string) : null

  const [users, setUsers] = useState<FollowUserInfo[]>([])
  const [page, setPage] = useState(1)
  const [hasMore, setHasMore] = useState(true)
  const [loading, setLoading] = useState(false)
  const observerRef = useRef<HTMLDivElement>(null)

  // GET /users/followers/:id?page=1&size=20 响应 data.users, data.total, data.page, data.size
  const fetchFollowers = useCallback(async (pageNum: number) => {
    if (!userId || !hasMore || loading) return

    setLoading(true)
    try {
      const res = await api.get<FollowListResponse>(`/users/followers/${userId}`, {
        params: { page: pageNum, size: 20 },
      })
      const newUsers = res.data.data?.users ?? []
      setUsers(prev => (pageNum === 1 ? newUsers : [...prev, ...newUsers]))

      if (newUsers.length < 20) setHasMore(false)
    } catch (err) {
      toast.error("加载粉丝列表失败")
    } finally {
      setLoading(false)
    }
  }, [userId, hasMore, loading])

  useEffect(() => {
    if (userId) fetchFollowers(1)
  }, [userId, fetchFollowers])

  useEffect(() => {
    if (!observerRef.current || !hasMore || loading) return

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting) {
          setPage(prev => prev + 1)
          fetchFollowers(page + 1)
        }
      },
      { threshold: 0.1 }
    )

    observer.observe(observerRef.current)
    return () => observer.disconnect()
  }, [hasMore, loading, fetchFollowers, page])

  const toggleFollow = async (targetId: number, isFollowed: boolean, index: number) => {
    try {
      if (isFollowed) {
        await api.delete(`/follow/${targetId}`)
        toast.success("已取消关注")
      } else {
        await api.post(`/follow/${targetId}`)
        toast.success("关注成功")
      }
      setUsers(prev => {
        const newUsers = [...prev]
        newUsers[index].is_followed = !isFollowed
        return newUsers
      })
    } catch (err) {
      toast.error("操作失败")
    }
  }

  if (!userId) {
    return <div className="min-h-screen flex items-center justify-center bg-slate-50">无效的用户ID</div>
  }

  return (
    <div className="min-h-screen bg-slate-50 py-10 px-4">
      <div className="max-w-4xl mx-auto">
        <Card>
          <CardHeader>
            <CardTitle className="text-2xl">我的粉丝</CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            {users.length === 0 && !loading && <p className="text-center text-gray-500">暂无粉丝</p>}

            {users.map((user, index) => (
              <div key={user.id} className="flex items-center gap-4 py-3 border-b last:border-b-0">
                <Avatar className="h-12 w-12 cursor-pointer" onClick={() => router.push(`/users/${user.id}`)}>
                  <AvatarImage src={user.avatar_url ? staticUrl(user.avatar_url) : undefined} alt={user.username} />
                  <AvatarFallback>{user.username?.[0]?.toUpperCase() || 'U'}</AvatarFallback>
                </Avatar>
                <div className="flex-1">
                  <p className="font-medium cursor-pointer" onClick={() => router.push(`/users/${user.id}`)}>
                    {user.username}
                  </p>
                  <p className="text-sm text-gray-600">{user.profile || '暂无简介'}</p>
                </div>
                <Button
                  variant={user.is_followed ? "outline" : "default"}
                  size="sm"
                  onClick={() => toggleFollow(user.id, user.is_followed, index)}
                >
                  {user.is_followed ? '取消关注' : '关注'}
                </Button>
              </div>
            ))}

            {hasMore && (
              <div ref={observerRef} className="py-4 text-center">
                {loading ? (
                  <div className="flex justify-center">
                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900" />
                  </div>
                ) : (
                  <p className="text-gray-500">继续向下滚动加载更多...</p>
                )}
              </div>
            )}

            {!hasMore && users.length > 0 && (
              <p className="text-center text-gray-500 py-4">已加载全部粉丝</p>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}