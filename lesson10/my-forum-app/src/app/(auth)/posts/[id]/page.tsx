// src/app/posts/[id]/page.tsx — URL/字段与后端严格对齐
'use client'

import { useState, useEffect, useRef, useCallback } from 'react'
import { useParams, useRouter } from 'next/navigation'
import api from '@/lib/api'
import { staticUrl } from '@/lib/api'
import type {
  PostDetailResponse,
  GetCommentsResponse,
  GetRepliesResponse,
  CommentItem,
  PostCommentRequest,
} from '@/lib/types'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { toast } from 'sonner'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'

const COMMENT_PAGE_SIZE = 20

/** 将正文中的 /static/ 图片路径转为完整 URL 以便跨域显示 */
function rewriteStaticImgSrc(html: string): string {
  if (!html) return html
  return html.replace(/src="(\/static\/[^"]*)"/g, (_, path) => `src="${staticUrl(path)}"`)
}

type CommentWithReplies = CommentItem & {
  replies?: CommentItem[]
  isExpanded?: boolean
  loadingReplies?: boolean
}

export default function PostDetailPage() {
  const params = useParams()
  const postId = params.id as string
  const router = useRouter()
  const [post, setPost] = useState<PostDetailResponse | null>(null)
  const [comments, setComments] = useState<CommentWithReplies[]>([])
  const [hasMoreComments, setHasMoreComments] = useState(true)
  const [loadingPost, setLoadingPost] = useState(true)
  const [loadingComments, setLoadingComments] = useState(false)
  const [commentInput, setCommentInput] = useState('')
  const [commentLoading, setCommentLoading] = useState(false)
  const [liked, setLiked] = useState(false)
  const [likeCount, setLikeCount] = useState(0)
  const [isFavorited, setIsFavorited] = useState(false)
  const [postStatus, setPostStatus] = useState<number>(0)
  const observerRef = useRef<HTMLDivElement>(null)
  const nextCommentPageRef = useRef(2)

  // GET /posts/:id — 可选登录，响应为 PostDetailResponse 直接返回
  useEffect(() => {
    const fetchPost = async () => {
      try {
        setLoadingPost(true)
        const res = await api.get<PostDetailResponse>(`/posts/${postId}`)
        setPost(res.data)
        setLikeCount(res.data.LikeCount ?? 0)
        setPostStatus(res.data.Status ?? 0)
      } catch {
        toast.error('加载帖子失败')
        setPost(null)
      } finally {
        setLoadingPost(false)
      }
    }
    fetchPost()
  }, [postId])

  // GET /posts/comments?target_type=1&target_id=:id&page=1&size=20
  const fetchComments = useCallback(
    async (pageNum: number) => {
      if (!postId) return
      setLoadingComments(true)
      try {
        const res = await api.get<GetCommentsResponse>('/posts/comments', {
          params: {
            target_type: 1,
            target_id: parseInt(postId, 10),
            page: pageNum,
            size: COMMENT_PAGE_SIZE,
          },
        })
        const newComments = (res.data.data?.comments ?? []).map((c) => ({
          ...c,
          isExpanded: false,
          replies: [] as CommentItem[],
          loadingReplies: false,
        }))
        setComments((prev) => (pageNum === 1 ? newComments : [...prev, ...newComments]))
        if (newComments.length < COMMENT_PAGE_SIZE) setHasMoreComments(false)
      } catch {
        toast.error('加载评论失败')
      } finally {
        setLoadingComments(false)
      }
    },
    [postId]
  )

  useEffect(() => {
    if (!postId) return
    setHasMoreComments(true)
    nextCommentPageRef.current = 2
    fetchComments(1)
  }, [postId, fetchComments])

  useEffect(() => {
    if (!observerRef.current || !hasMoreComments || loadingComments) return
    const observer = new IntersectionObserver(
      (entries) => {
        if (!entries[0].isIntersecting || !hasMoreComments || loadingComments) return
        const pageToFetch = nextCommentPageRef.current
        nextCommentPageRef.current = pageToFetch + 1
        fetchComments(pageToFetch)
      },
      { threshold: 0.1 }
    )
    observer.observe(observerRef.current)
    return () => observer.disconnect()
  }, [hasMoreComments, loadingComments, fetchComments])

  // GET /comments/:parent_id/replies — 无查询参数，响应 data.replies（含二级及以上全部回复）
  const loadReplies = async (parentId: number, commentIndex: number) => {
    setComments((prev) => {
      const next = [...prev]
      if (next[commentIndex]) next[commentIndex] = { ...next[commentIndex], loadingReplies: true }
      return next
    })
    try {
      const res = await api.get<GetRepliesResponse>(`/comments/${parentId}/replies`)
      const replies = res.data.data?.replies ?? res.data.replies ?? []
      setComments((prev) => {
        const next = [...prev]
        next[commentIndex] = {
          ...next[commentIndex],
          replies: Array.isArray(replies) ? replies : [],
          isExpanded: true,
          loadingReplies: false,
        }
        return next
      })
    } catch {
      toast.error('加载回复失败')
      setComments((prev) => {
        const next = [...prev]
        if (next[commentIndex]) next[commentIndex] = { ...next[commentIndex], loadingReplies: false }
        return next
      })
    }
  }

  const toggleReplies = (index: number, parentId: number) => {
    setComments((prev) => {
      const next = [...prev]
      if (next[index].isExpanded) {
        next[index] = { ...next[index], isExpanded: false }
        return next
      }
      const needLoad = !next[index].replies?.length && !next[index].loadingReplies
      next[index] = {
        ...next[index],
        isExpanded: true,
        loadingReplies: needLoad,
      }
      if (needLoad) loadReplies(parentId, index)
      return next
    })
  }

  // POST /comments — body: target_type, target_id, content（无 parent_id）
  const postComment = async () => {
    if (!commentInput.trim()) {
      toast.error('评论内容不能为空')
      return
    }
    setCommentLoading(true)
    try {
      const payload: PostCommentRequest = {
        target_type: 1,
        target_id: parseInt(postId, 10),
        content: commentInput.trim(),
      }
      await api.post('/comments', payload)
      toast.success('评论发布成功')
      setCommentInput('')
      setComments([])
      setHasMoreComments(true)
      nextCommentPageRef.current = 2
      fetchComments(1)
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { message?: string } } })?.response?.data?.message ?? '评论发布失败'
      toast.error(msg)
    } finally {
      setCommentLoading(false)
    }
  }

  // POST /reactions — 点赞
  const toggleLike = async () => {
    try {
      const res = await api.post<{ message: string; status: boolean }>('/reactions', {
        target_type: 1,
        target_id: parseInt(postId, 10),
      })
      setLiked(res.data.status)
      setLikeCount((c) => (res.data.status ? c + 1 : Math.max(0, c - 1)))
    } catch {
      toast.error('操作失败')
    }
  }

  const handleDelete = async () => {
    if (!confirm('确定要删除这篇帖子吗？')) return
    try {
      await api.delete(`/posts/${postId}`)
      toast.success('已删除')
      router.push('/')
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { message?: string } } })?.response?.data?.message ?? '删除失败'
      toast.error(msg)
    }
  }

  // PUT /posts/:id 仅更新状态：设为草稿(1) 或 设为发布(0)
  const setStatus = async (status: 0 | 1) => {
    try {
      await api.put(`/posts/${postId}`, { status })
      setPostStatus(status)
      toast.success(status === 1 ? '已设为草稿' : '已设为发布')
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { message?: string } } })?.response?.data?.message ?? '操作失败'
      toast.error(msg)
    }
  }

  // POST /favorites — 收藏/取消收藏 target_type: 1=文章
  const toggleFavorite = async () => {
    try {
      const res = await api.post<{ message: string; data: { is_favorited: boolean } }>('/favorites', {
        target_type: 1,
        target_id: parseInt(postId, 10),
      })
      setIsFavorited(res.data.data?.is_favorited ?? false)
      toast.success(res.data.data?.is_favorited ? '已收藏' : '已取消收藏')
    } catch {
      toast.error('操作失败')
    }
  }

  if (loadingPost) {
    return (
      <div className="min-h-screen bg-slate-50 py-10 px-4">
        <div className="max-w-4xl mx-auto space-y-6">
          <Skeleton className="h-12 w-3/4" />
          <Skeleton className="h-64 w-full" />
          <Skeleton className="h-32 w-full" />
        </div>
      </div>
    )
  }

  if (!post) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-slate-50">
        <p className="text-xl text-gray-600">帖子不存在或加载失败</p>
        <Button variant="link" onClick={() => router.push('/')}>
          返回首页
        </Button>
      </div>
    )
  }

  const currentUserId = typeof window !== 'undefined' ? localStorage.getItem('user_id') : null
  const isAuthor = currentUserId && String(post.AuthorID) === currentUserId

  return (
    <div className="min-h-screen bg-slate-50 py-10 px-4">
      <div className="max-w-4xl mx-auto space-y-8">
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-4">
                <Avatar>
                  <AvatarImage src="" alt={post.AuthorName} />
                  <AvatarFallback>{post.AuthorName?.[0] || 'A'}</AvatarFallback>
                </Avatar>
                <div>
                  <p className="font-medium">{post.AuthorName || '匿名'}</p>
                  <p className="text-sm text-gray-500">
                    {formatDistanceToNow(new Date(post.CreatedAt), { addSuffix: true, locale: zhCN })}
                  </p>
                </div>
              </div>
              <div className="flex items-center gap-2">
                <Button variant="outline" size="sm" onClick={toggleLike}>
                  {liked ? '已赞' : '赞'} {likeCount}
                </Button>
                <Button variant="outline" size="sm" onClick={toggleFavorite}>
                  {isFavorited ? '已收藏' : '收藏'}
                </Button>
                {isAuthor && (
                  <>
                    <Button variant="outline" size="sm" onClick={() => router.push(`/posts/${postId}/edit`)}>
                      编辑
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setStatus(postStatus === 1 ? 0 : 1)}
                    >
                      {postStatus === 1 ? '设为发布' : '设为草稿'}
                    </Button>
                    <Button variant="destructive" size="sm" onClick={handleDelete}>
                      删除
                    </Button>
                  </>
                )}
              </div>
            </div>
            <CardTitle className="mt-4 text-2xl">{post.Title}</CardTitle>
          </CardHeader>
          <CardContent>
            <div
              className="prose max-w-none"
              dangerouslySetInnerHTML={{
                __html: rewriteStaticImgSrc(post.content || '') || '<p>暂无内容</p>',
              }}
            />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>评论</CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            {comments.length === 0 && !loadingComments && (
              <p className="text-center text-gray-500">暂无评论，快来抢沙发！</p>
            )}

            {comments.map((comment, index) => (
              <div key={comment.id}>
                <div className="flex gap-4">
                  <Avatar className="h-10 w-10">
                    <AvatarFallback>{comment.author_name?.[0] || 'A'}</AvatarFallback>
                  </Avatar>
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <span className="font-medium">{comment.author_name || '匿名'}</span>
                      <span className="text-sm text-gray-500">
                        {formatDistanceToNow(new Date(comment.created_at), {
                          addSuffix: true,
                          locale: zhCN,
                        })}
                      </span>
                    </div>
                    <p className="mt-1 text-gray-800">{comment.content}</p>
                    <Button
                      variant="link"
                      className="mt-2 p-0 h-auto text-blue-600"
                      onClick={() => toggleReplies(index, comment.id)}
                    >
                      {comment.loadingReplies
                        ? '加载中...'
                        : comment.isExpanded
                          ? '收起回复'
                          : '查看回复'}
                    </Button>
                  </div>
                </div>
                {comment.isExpanded && (
                  <div className="ml-14 mt-4 space-y-4 border-l-2 border-gray-200 pl-4">
                    {comment.loadingReplies && (
                      <p className="text-sm text-gray-500">加载回复中...</p>
                    )}
                    {!comment.loadingReplies && comment.replies && comment.replies.length > 0 && (
                      comment.replies.map((reply) => (
                        <div key={reply.id} className="flex gap-4">
                          <Avatar className="h-8 w-8">
                            <AvatarFallback>{reply.author_name?.[0] || 'A'}</AvatarFallback>
                          </Avatar>
                          <div className="flex-1">
                            <div className="flex items-center gap-2">
                              <span className="font-medium">{reply.author_name || '匿名'}</span>
                              <span className="text-xs text-gray-500">
                                {formatDistanceToNow(new Date(reply.created_at), {
                                  addSuffix: true,
                                  locale: zhCN,
                                })}
                              </span>
                            </div>
                            <p className="mt-1 text-gray-800">{reply.content}</p>
                          </div>
                        </div>
                      ))
                    )}
                    {!comment.loadingReplies && comment.replies && comment.replies.length === 0 && (
                      <p className="text-sm text-gray-500">暂无回复</p>
                    )}
                  </div>
                )}
                <Separator className="my-6" />
              </div>
            ))}

            {hasMoreComments && (
              <div ref={observerRef} className="py-4 text-center">
                {loadingComments ? (
                  <div className="flex justify-center">
                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900" />
                  </div>
                ) : (
                  <p className="text-gray-500">继续向下滚动加载更多评论...</p>
                )}
              </div>
            )}

            <div className="mt-8">
              <Label htmlFor="comment">添加评论</Label>
              <Input
                id="comment"
                placeholder="输入你的评论..."
                value={commentInput}
                onChange={(e) => setCommentInput(e.target.value)}
                className="mt-2"
              />
              <Button
                className="mt-4"
                onClick={postComment}
                disabled={commentLoading || !commentInput.trim()}
              >
                {commentLoading ? '提交中...' : '提交评论'}
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
