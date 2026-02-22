// 编辑资料：PUT /profile, PUT /change_pass, POST /avatar (FormData field: avatar)，头像先圆形裁剪再上传
'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import api from '@/lib/api'
import { staticUrl } from '@/lib/api'
import type { UpdateProfileRequest } from '@/lib/types'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import AvatarCropModal from '@/components/AvatarCropModal'
import { toast } from 'sonner'

export default function ProfileEditPage() {
  const router = useRouter()
  const [profile, setProfile] = useState('')
  const [avatarUrl, setAvatarUrl] = useState<string | null>(null)
  const [oldPass, setOldPass] = useState('')
  const [newPass, setNewPass] = useState('')
  const [loadingProfile, setLoadingProfile] = useState(false)
  const [loadingPass, setLoadingPass] = useState(false)
  const [loadingAvatar, setLoadingAvatar] = useState(false)
  const [cropFile, setCropFile] = useState<File | null>(null)

  useEffect(() => {
    const token = typeof window !== 'undefined' ? localStorage.getItem('token') : null
    const userId = typeof window !== 'undefined' ? localStorage.getItem('user_id') : null
    if (!token || !userId) {
      router.push('/login')
      return
    }
    api.get<{ message: string; data: { profile?: string; avatar_url?: string } }>(`/user/${userId}`)
      .then((res) => {
        const d = res.data.data
        if (d) {
          setProfile(d.profile ?? '')
          setAvatarUrl(d.avatar_url ?? null)
        }
      })
      .catch(() => {})
  }, [router])

  const handleSaveProfile = async () => {
    setLoadingProfile(true)
    try {
      await api.put<{ ok: boolean }>('/profile', { profile: profile || undefined } as UpdateProfileRequest)
      toast.success('资料已保存')
    } catch {
      toast.error('保存失败')
    } finally {
      setLoadingProfile(false)
    }
  }

  const handleChangePass = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!oldPass || !newPass) {
      toast.error('请填写旧密码和新密码')
      return
    }
    setLoadingPass(true)
    try {
      await api.put('/change_pass', { old_pass: oldPass, new_pass: newPass })
      toast.success('密码已修改，请重新登录')
      setOldPass('')
      setNewPass('')
      setTimeout(() => {
        localStorage.clear()
        router.push('/login')
      }, 1500)
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error ?? '修改失败'
      toast.error(msg)
    } finally {
      setLoadingPass(false)
    }
  }

  const handleAvatarFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    if (!file.type.startsWith('image/')) {
      toast.error('请选择图片文件（jpg/png/webp）')
      e.target.value = ''
      return
    }
    setCropFile(file)
    e.target.value = ''
  }

  const handleCropConfirm = async (blob: Blob) => {
    setCropFile(null)
    setLoadingAvatar(true)
    const formData = new FormData()
    formData.append('avatar', blob, 'avatar.png')
    try {
      const res = await api.post<{ avatar_url: string }>('/avatar', formData)
      setAvatarUrl(res.data.avatar_url)
      toast.success('头像已更新')
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error ?? '上传失败'
      toast.error(msg)
    } finally {
      setLoadingAvatar(false)
    }
  }

  return (
    <div className="min-h-screen bg-slate-50 py-10 px-4">
      <div className="max-w-xl mx-auto space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>编辑资料</CardTitle>
            <CardDescription>修改个人简介与头像</CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="flex items-center gap-4">
              <Avatar className="h-20 w-20">
                <AvatarImage src={avatarUrl ? staticUrl(avatarUrl) : undefined} />
                <AvatarFallback>U</AvatarFallback>
              </Avatar>
              <div>
                <Label htmlFor="avatar" className="cursor-pointer">
                  <span className="inline-block px-4 py-2 bg-primary text-primary-foreground rounded-md hover:opacity-90">
                    {loadingAvatar ? '上传中...' : '更换头像'}
                  </span>
                </Label>
                <input
                  id="avatar"
                  type="file"
                  accept="image/jpeg,image/png,image/webp"
                  className="hidden"
                  onChange={handleAvatarFileSelect}
                  disabled={loadingAvatar}
                />
              </div>
            </div>
            <div className="space-y-2">
              <Label htmlFor="profile">个人简介</Label>
              <Input
                id="profile"
                placeholder="一句话介绍自己"
                value={profile}
                onChange={(e) => setProfile(e.target.value)}
                maxLength={255}
              />
              <Button onClick={handleSaveProfile} disabled={loadingProfile}>
                {loadingProfile ? '保存中...' : '保存简介'}
              </Button>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>修改密码</CardTitle>
            <CardDescription>修改后需重新登录</CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleChangePass} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="old_pass">旧密码</Label>
                <Input
                  id="old_pass"
                  type="password"
                  value={oldPass}
                  onChange={(e) => setOldPass(e.target.value)}
                  placeholder="请输入旧密码"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="new_pass">新密码</Label>
                <Input
                  id="new_pass"
                  type="password"
                  value={newPass}
                  onChange={(e) => setNewPass(e.target.value)}
                  placeholder="请输入新密码"
                />
              </div>
              <Button type="submit" disabled={loadingPass}>
                {loadingPass ? '提交中...' : '修改密码'}
              </Button>
            </form>
          </CardContent>
        </Card>

        <Button variant="outline" onClick={() => router.push('/profile')}>
          返回个人主页
        </Button>
      </div>

      {cropFile && (
        <AvatarCropModal
          imageFile={cropFile}
          onConfirm={handleCropConfirm}
          onCancel={() => setCropFile(null)}
        />
      )}
    </div>
  )
}
