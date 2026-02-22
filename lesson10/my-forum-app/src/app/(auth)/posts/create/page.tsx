// src/app/posts/create/page.tsx
'use client'

import { useState, useCallback, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useForm, SubmitHandler } from 'react-hook-form'
import api from '@/lib/api'
import { staticUrl } from '@/lib/api'
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Checkbox } from "@/components/ui/checkbox"
import { toast } from "sonner"

// Tiptap 编辑器相关
import { EditorContent, useEditor } from '@tiptap/react'
import StarterKit from '@tiptap/starter-kit'
import Placeholder from '@tiptap/extension-placeholder'
import Image from '@tiptap/extension-image'
import Dropcursor from '@tiptap/extension-dropcursor'
import Gapcursor from '@tiptap/extension-gapcursor'

// 表单类型
type CreatePostForm = {
  type: '1' | '2'
  title: string
  content: string
  status: boolean
}

export default function CreatePostPage() {
  const router = useRouter()
  const [loading, setLoading] = useState(false)
  const [uploading, setUploading] = useState(false)

  const { register, handleSubmit, setValue, watch, formState: { errors } } = useForm<CreatePostForm>({
    defaultValues: {
      type: '1',
      title: '',
      content: '',
      status: false,
    },
  })

  // Tiptap 编辑器
  const editor = useEditor({
    extensions: [
      StarterKit,
      Placeholder.configure({
        placeholder: '写下你的想法... 支持拖拽/粘贴图片上传',
      }),
      Image.configure({
        inline: true,
        allowBase64: false,
      }),
      Dropcursor,
      Gapcursor,
    ],
    content: '',
    immediatelyRender: false, // 解决 SSR hydration 问题
    onUpdate: ({ editor }) => {
      setValue('content', editor.getHTML(), { shouldValidate: true })
    },
  })

  // 上传图片
// 上传图片函数（只改这一部分）
const uploadImage = async (file: File): Promise<string | null> => {
  if (!file.type.startsWith('image/')) {
    toast.error("只能上传图片文件")
    return null
  }

  setUploading(true)
  const formData = new FormData()
  formData.append('image', file)  // 后端用的是 "image"，不是 "file"

  try {
    console.log('开始上传图片：', file.name, file.type, file.size)  // 调试

    const res = await api.post('/upload/article-image', formData)

    console.log('上传成功响应：', res.data)  // 调试

    // 后端返回的是 "image_url"（如 /static/uploads/images/xxx），插入时用完整 URL 以便展示
    const imageUrl = res.data.image_url
    if (!imageUrl) throw new Error('未返回 image_url')
    const fullUrl = staticUrl(imageUrl)

    editor?.chain().focus().setImage({ src: fullUrl, alt: file.name }).run()

    toast.success("图片上传成功，已插入编辑器")
    return imageUrl
  } catch (err: any) {
    console.error('上传失败：', err)  // 调试
    const msg = err.response?.data?.message || '图片上传失败，请检查后端日志'
    toast.error(msg)
    return null
  } finally {
    setUploading(false)
  }
}

// 处理拖拽/粘贴（加强阻止默认行为）
const handleImageInput = useCallback((e: DragEvent | ClipboardEvent) => {
  e.preventDefault()
  e.stopPropagation()
  e.stopImmediatePropagation()  // 加强阻止，防止浏览器默认打开窗口

  const files = 'dataTransfer' in e 
    ? e.dataTransfer?.files 
    : (e as ClipboardEvent).clipboardData?.files

  console.log('检测到文件输入：', files?.length || 0, '个文件')  // 调试

  if (files && files.length > 0) {
    Array.from(files).forEach(file => {
      if (file.type.startsWith('image/')) {
        uploadImage(file)
      }
    })
  }
}, [uploadImage])

// 手动选择文件（加强调试）
const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
  const files = e.target.files
  console.log('手动选择文件：', files?.length || 0, '个文件')  // 调试

  if (files && files.length > 0) {
    Array.from(files).forEach(file => {
      if (file.type.startsWith('image/')) {
        uploadImage(file)
      }
    })
    e.target.value = ''
  }
}
  // 绑定事件
  useEffect(() => {
    const view = editor?.view
    if (view) {
      const dom = view.dom
      dom.addEventListener('drop', handleImageInput as any)
      dom.addEventListener('paste', handleImageInput as any)
      return () => {
        dom.removeEventListener('drop', handleImageInput as any)
        dom.removeEventListener('paste', handleImageInput as any)
      }
    }
  }, [editor, handleImageInput])

  // 提交函数
  const onSubmit: SubmitHandler<CreatePostForm> = async (data) => {
    setLoading(true)

    try {
      const payload = {
        type: parseInt(data.type),
        title: data.title,
        content: data.content,
        status: data.status ? 1 : 0,
      }

      const res = await api.post<{ ok: boolean; post: { ID: number } }>('/posts', payload)

      toast.success("发布成功！")

      const postId = res.data?.post?.ID
      if (postId) {
        router.push(`/posts/${postId}`)
      } else {
        router.push('/')
      }
    } catch (err: any) {
      const msg = err.response?.data?.error || err.response?.data?.message || '发布失败，请稍后重试'
      toast.error(msg)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-slate-50 py-10 px-4">
      <div className="max-w-4xl mx-auto">
        <Card>
          <CardHeader>
            <CardTitle className="text-2xl">发布新帖</CardTitle>
            <CardDescription>
              支持拖拽、粘贴或选择文件上传图片，内容使用 Markdown 格式
            </CardDescription>
          </CardHeader>

          <CardContent>
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
              {/* 类型 */}
              <div className="space-y-2">
                <Label>类型</Label>
                <Select
                  defaultValue="1"
                  onValueChange={(v) => setValue('type', v as '1' | '2')}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="请选择类型" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="1">文章</SelectItem>
                    <SelectItem value="2">问题</SelectItem>
                  </SelectContent>
                </Select>
                {errors.type && <p className="text-red-500 text-sm">{errors.type.message}</p>}
              </div>

              {/* 标题 */}
              <div className="space-y-2">
                <Label htmlFor="title">标题</Label>
                <Input
                  id="title"
                  placeholder="请输入标题（最多200字符）"
                  {...register('title', { required: '标题不能为空', maxLength: { value: 200, message: '标题最多200字符' } })}
                />
                {errors.title && <p className="text-red-500 text-sm">{errors.title.message}</p>}
              </div>

              {/* 正文 */}
              <div className="space-y-2">
                <Label>正文</Label>
                {uploading && <p className="text-blue-600 text-sm">图片上传中...</p>}
                <div className="border rounded-md min-h-[400px] bg-white overflow-hidden">
                  <EditorContent editor={editor} className="p-4" />
                </div>
                {errors.content && <p className="text-red-500 text-sm">{errors.content.message}</p>}
              </div>

              {/* 手动上传按钮 */}
              <div className="space-y-2">
                <Button type="button" variant="secondary" onClick={() => document.getElementById('image-upload')?.click()}>
                  选择图片上传
                </Button>
                <input
                  id="image-upload"
                  type="file"
                  accept="image/*"
                  multiple
                  hidden
                  onChange={handleFileChange}
                />
              </div>

              {/* 草稿 */}
              <div className="flex items-center space-x-2">
                <Checkbox
                  id="draft"
                  checked={watch('status')}
                  onCheckedChange={(checked) => setValue('status', !!checked)}
                />
                <Label htmlFor="draft">保存为草稿（不公开）</Label>
              </div>

              <div className="flex justify-end">
                <Button type="submit" disabled={loading || uploading}>
                  {loading ? '发布中...' : uploading ? '上传中...' : '发布'}
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
