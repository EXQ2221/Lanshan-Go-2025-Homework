// 编辑帖子：PUT /posts/:id body: title?, content?
'use client'

import { useState, useEffect } from 'react'
import { useParams, useRouter } from 'next/navigation'
import api from '@/lib/api'
import type { PostDetailResponse, UpdatePostRequest } from '@/lib/types'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { toast } from 'sonner'

export default function EditPostPage() {
  const params = useParams()
  const router = useRouter()
  const postId = params.id as string
  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [status, setStatus] = useState<0 | 1>(0)
  const [loading, setLoading] = useState(false)
  const [fetching, setFetching] = useState(true)

  useEffect(() => {
    api
      .get<PostDetailResponse>(`/posts/${postId}`)
      .then((res) => {
        setTitle(res.data.Title ?? '')
        setContent(res.data.content ?? '')
        setStatus((res.data.Status === 1 ? 1 : 0) as 0 | 1)
      })
      .catch(() => toast.error('加载帖子失败'))
      .finally(() => setFetching(false))
  }, [postId])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!title.trim()) {
      toast.error('标题不能为空')
      return
    }
    setLoading(true)
    try {
      const payload: UpdatePostRequest = { title: title.trim(), content, status }
      await api.put(`/posts/${postId}`, payload)
      toast.success('更新成功')
      router.push(`/posts/${postId}`)
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { message?: string } } })?.response?.data?.message ?? '更新失败'
      toast.error(msg)
    } finally {
      setLoading(false)
    }
  }

  if (fetching) {
    return (
      <div className="min-h-screen bg-slate-50 flex items-center justify-center">
        <p className="text-gray-500">加载中...</p>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-slate-50 py-10 px-4">
      <div className="max-w-2xl mx-auto">
        <Card>
          <CardHeader>
            <CardTitle>编辑帖子</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="title">标题</Label>
                <Input
                  id="title"
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  placeholder="标题"
                  maxLength={200}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="content">内容</Label>
                <textarea
                  id="content"
                  className="w-full min-h-[300px] rounded-md border border-input bg-background px-3 py-2"
                  value={content}
                  onChange={(e) => setContent(e.target.value)}
                  placeholder="正文（支持 HTML）"
                />
              </div>
              <div className="space-y-2">
                <Label>状态</Label>
                <Select
                  value={String(status)}
                  onValueChange={(v) => setStatus(Number(v) as 0 | 1)}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="选择状态" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="0">发布</SelectItem>
                    <SelectItem value="1">草稿</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="flex gap-2">
                <Button type="submit" disabled={loading}>
                  {loading ? '保存中...' : '保存'}
                </Button>
                <Button type="button" variant="outline" onClick={() => router.push(`/posts/${postId}`)}>
                  取消
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
