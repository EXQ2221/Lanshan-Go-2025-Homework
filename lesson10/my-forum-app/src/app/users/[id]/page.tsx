// 查看任意用户 — GET /user/:id?page=1 响应 { message, data: UserPublicInfo }
'use client'

import { useEffect, useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import api, { staticUrl } from '@/lib/api'
import type { UserPublicInfo } from '@/lib/types'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { toast } from 'sonner'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'

export default function UserPage() {
  const params = useParams()
  const router = useRouter()
  const userId = params.id as string
  const currentId = typeof window !== 'undefined' ? localStorage.getItem('user_id') : null
  const isSelf = currentId === userId

  const [user, setUser] = useState<UserPublicInfo | null>(null)
  const [loading, setLoading] = useState(true)
  const [following, setFollowing] = useState(false)

  useEffect(() => {
    const fetchUser = async () => {
      try {
        setLoading(true)
        const res = await api.get<{ message: string; data: UserPublicInfo }>(`/user/${userId}`, {
          params: { page: 1 },
        })
        const data = res.data.data ?? null
        setUser(data)
        setFollowing(Boolean(data?.is_followed))
      } catch (err: unknown) {
        const msg = (err as { response?: { data?: { message?: string } } })?.response?.data?.message ?? '加载失败'
        toast.error(msg)
      } finally {
        setLoading(false)
      }
    }
    fetchUser()
  }, [userId])

  const handleFollow = async () => {
    try {
      await api.post(`/follow/${userId}`)
      setFollowing(true)
      toast.success('关注成功')
      if (user) setUser({ ...user, followers_count: user.followers_count + 1, is_followed: true })
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { message?: string } } })?.response?.data?.message ?? '操作失败'
      toast.error(msg)
    }
  }

  const handleUnfollow = async () => {
    try {
      await api.delete(`/follow/${userId}`)
      setFollowing(false)
      toast.success('已取消关注')
      if (user) setUser({ ...user, followers_count: Math.max(0, user.followers_count - 1), is_followed: false })
    } catch {
      toast.error('操作失败')
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-slate-50 py-10 px-4">
        <div className="max-w-4xl mx-auto">
          <Skeleton className="h-32 w-full rounded-lg" />
          <Skeleton className="h-64 w-full mt-6 rounded-lg" />
        </div>
      </div>
    )
  }

  if (!user) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center bg-slate-50">
        <p className="text-xl text-gray-600 mb-6">用户不存在或加载失败</p>
        <Button onClick={() => router.push('/')}>返回首页</Button>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-slate-50 py-10 px-4">
      <div className="max-w-4xl mx-auto">
        <Card>
          <CardHeader className="flex flex-row items-center gap-6">
            <Avatar className="h-24 w-24">
              <AvatarImage src={user.avatar_url ? staticUrl(user.avatar_url) : undefined} alt={user.username} />
              <AvatarFallback>{user.username?.[0]?.toUpperCase() || 'U'}</AvatarFallback>
            </Avatar>
            <div className="flex-1">
              <CardTitle className="text-3xl">{user.username}</CardTitle>
              <p className="text-sm text-gray-600">ID: {user.id}</p>
              <CardDescription className="mt-2 text-lg">{user.profile || '暂无个人简介'}</CardDescription>
              <div className="mt-2 flex gap-4 text-sm text-gray-600">
                <span>角色：{user.role === 0 ? '普通用户' : user.role === 1 ? 'VIP' : '管理员'}</span>
                <span>·</span>
                <span>VIP：{user.is_vip ? '是' : '否'}</span>
              </div>
              {!isSelf && (
                <div className="mt-4">
                  <Button variant="outline" size="sm" onClick={() => router.push(`/users/${userId}/followers`)}>
                    粉丝 {user.followers_count}
                  </Button>
                  <Button variant="outline" size="sm" className="ml-2" onClick={() => router.push(`/users/${userId}/following`)}>
                    关注 {user.following_count}
                  </Button>
                  <Button
                    className="ml-2"
                    variant={following ? 'outline' : 'default'}
                    size="sm"
                    onClick={following ? handleUnfollow : handleFollow}
                  >
                    {following ? '取消关注' : '关注'}
                  </Button>
                </div>
              )}
            </div>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <Card className="cursor-pointer hover:bg-gray-50" onClick={() => {}}>
                <CardHeader><CardTitle>发帖数</CardTitle></CardHeader>
                <CardContent><p className="text-3xl font-bold">{user.post_total ?? 0}</p></CardContent>
              </Card>
              <Card className="cursor-pointer hover:bg-gray-50" onClick={() => router.push(`/users/${userId}/following`)}>
                <CardHeader><CardTitle>关注</CardTitle></CardHeader>
                <CardContent><p className="text-3xl font-bold">{user.following_count ?? 0}</p></CardContent>
              </Card>
              <Card className="cursor-pointer hover:bg-gray-50" onClick={() => router.push(`/users/${userId}/followers`)}>
                <CardHeader><CardTitle>粉丝</CardTitle></CardHeader>
                <CardContent><p className="text-3xl font-bold">{user.followers_count ?? 0}</p></CardContent>
              </Card>
            </div>
            <Card>
              <CardHeader><CardTitle>最近发帖</CardTitle></CardHeader>
              <CardContent>
                {user.posts?.length ? (
                  <ul className="space-y-3">
                    {user.posts.map((post) => (
                      <li
                        key={post.id}
                        className="flex justify-between items-center cursor-pointer hover:bg-gray-50 p-2 rounded"
                        onClick={() => router.push(`/posts/${post.id}`)}
                      >
                        <span className="font-medium">{post.title}</span>
                        <span className="text-sm text-gray-500">
                          {formatDistanceToNow(new Date(post.created_at), { addSuffix: true, locale: zhCN })}
                        </span>
                      </li>
                    ))}
                  </ul>
                ) : (
                  <p className="text-gray-500">暂无发帖记录</p>
                )}
              </CardContent>
            </Card>
            {isSelf && (
              <div className="flex gap-4">
                <Button onClick={() => router.push('/profile')}>我的主页</Button>
                <Button variant="outline" onClick={() => router.push('/posts/create')}>发布新帖</Button>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}