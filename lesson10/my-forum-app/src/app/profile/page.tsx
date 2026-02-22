// src/app/profile/page.tsx — GET /user/:id 响应 data 与 UserPublicInfo 对齐
'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import api from '@/lib/api'
import { staticUrl } from '@/lib/api'
import type { UserPublicInfo } from '@/lib/types'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { toast } from "sonner"
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'

export default function ProfilePage() {
  const router = useRouter()
  const [profile, setProfile] = useState<UserPublicInfo | null>(null)
  const [loading, setLoading] = useState(true)
  const [errorMsg, setErrorMsg] = useState<string | null>(null)

  useEffect(() => {
    const fetchProfile = async () => {
      const userId = localStorage.getItem('user_id')
      const token = localStorage.getItem('token')

      if (!userId || !token) {
        setErrorMsg("请先登录才能查看个人信息")
        setLoading(false)
        return
      }

      try {
        setLoading(true)
        // GET /user/:id 响应 { message, data: UserPublicInfo }
        const res = await api.get<{ message: string; data: UserPublicInfo }>(`/user/${userId}`)
        setProfile(res.data.data ?? null)
      } catch (err: any) {
        const msg = err.response?.data?.message || '加载个人信息失败'
        setErrorMsg(msg)
        if (err.response?.status === 401) {
          localStorage.clear()
        }
      } finally {
        setLoading(false)
      }
    }

    fetchProfile()
  }, [router])

  const handleLogout = () => {
    localStorage.clear()
    toast.success("已退出登录")
    router.push('/login')
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

  if (errorMsg || !profile) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center bg-slate-50">
        <p className="text-xl text-gray-600 mb-6">{errorMsg || '加载失败'}</p>
        <Button onClick={() => router.push('/login')}>
          去登录
        </Button>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-slate-50 py-10 px-4">
      <div className="max-w-4xl mx-auto">
        <Card>
          <CardHeader className="flex flex-row items-center gap-6">
            <Avatar className="h-24 w-24">
              <AvatarImage src={profile.avatar_url ? staticUrl(profile.avatar_url) : undefined} alt={profile.username} />
              <AvatarFallback>{profile.username?.[0]?.toUpperCase() || 'U'}</AvatarFallback>
            </Avatar>
            <div>
              <CardTitle className="text-3xl">{profile.username}</CardTitle>
              <p className="text-sm text-gray-600">ID: {profile.id}</p>
              <CardDescription className="mt-2 text-lg">
                {profile.profile || '暂无个人简介'}
              </CardDescription>
              <div className="mt-2 flex gap-4 text-sm text-gray-600">
                <span>角色：{profile.role === 0 ? '普通用户' : profile.role === 1 ? 'VIP' : '管理员'}</span>
                <span>•</span>
                <span>VIP：{profile.is_vip ? '是' : '否'}</span>
              </div>
            </div>
          </CardHeader>

          <CardContent className="space-y-6">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <Card>
                <CardHeader>
                  <CardTitle>发帖数</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-3xl font-bold">{profile.post_total || 0}</p>
                </CardContent>
              </Card>

              {/* 点击关注跳转到 /users/:id/following */}
              <Card 
                className="cursor-pointer hover:bg-gray-50 transition-colors"
                onClick={() => router.push(`/users/${profile.id}/following`)}
              >
                <CardHeader>
                  <CardTitle>关注</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-3xl font-bold">{profile.following_count || 0}</p>
                </CardContent>
              </Card>

              {/* 点击粉丝跳转到 /users/:id/followers */}
              <Card 
                className="cursor-pointer hover:bg-gray-50 transition-colors"
                onClick={() => router.push(`/users/${profile.id}/followers`)}
              >
                <CardHeader>
                  <CardTitle>粉丝</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-3xl font-bold">{profile.followers_count || 0}</p>
                </CardContent>
              </Card>
            </div>

            {/* 最近发帖 */}
            <Card>
              <CardHeader>
                <CardTitle>最近发帖</CardTitle>
              </CardHeader>
              <CardContent>
                {profile.posts.length > 0 ? (
                  <ul className="space-y-3">
                    {profile.posts.map(post => (
                      <li 
                        key={post.id} 
                        className="flex justify-between items-center cursor-pointer hover:bg-gray-50 p-2 rounded transition-colors"
                        onClick={() => router.push(`/posts/${post.id}`)}
                      >
                        <span className="text-gray-800 font-medium">{post.title}</span>
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

            <div className="flex gap-4">
              <Button onClick={() => router.push('/profile/edit')}>
                编辑资料
              </Button>
              <Button variant="outline" onClick={() => router.push('/posts/create')}>
                发布新帖
              </Button>
              <Button variant="destructive" onClick={() => {
                localStorage.clear()
                toast.success("已退出登录")
                router.push('/login')
              }}>
                退出登录
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
