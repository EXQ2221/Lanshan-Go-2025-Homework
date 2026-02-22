'use client'

import { useState, useRef, useCallback, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'

const PREVIEW_SIZE = 320
const CROP_SIZE = 256
const CIRCLE_R = CROP_SIZE / 2

export type AvatarCropModalProps = {
  imageFile: File
  onConfirm: (blob: Blob) => void
  onCancel: () => void
}

export default function AvatarCropModal({ imageFile, onConfirm, onCancel }: AvatarCropModalProps) {
  const [imageUrl, setImageUrl] = useState<string>('')
  const [scale, setScale] = useState(1)
  const [position, setPosition] = useState({ x: 0, y: 0 })
  const [imageSize, setImageSize] = useState({ w: 0, h: 0 })
  const [dragging, setDragging] = useState(false)
  const dragStart = useRef({ x: 0, y: 0, pos: { x: 0, y: 0 } })
  const imgRef = useRef<HTMLImageElement>(null)

  useEffect(() => {
    const url = URL.createObjectURL(imageFile)
    setImageUrl(url)
    return () => URL.revokeObjectURL(url)
  }, [imageFile])

  const onImageLoad = useCallback(() => {
    const img = imgRef.current
    if (!img) return
    const w = img.naturalWidth
    const h = img.naturalHeight
    setImageSize({ w, h })
    const minScale = Math.max(PREVIEW_SIZE / w, PREVIEW_SIZE / h)
    setScale(minScale * 1.2)
    setPosition({ x: PREVIEW_SIZE / 2 - (w * minScale) / 2, y: PREVIEW_SIZE / 2 - (h * minScale) / 2 })
  }, [])

  const handlePointerDown = useCallback(
    (e: React.PointerEvent) => {
      e.preventDefault()
      setDragging(true)
      dragStart.current = { x: e.clientX, y: e.clientY, pos: { ...position } }
    },
    [position]
  )

  const handlePointerMove = useCallback(
    (e: React.PointerEvent) => {
      if (!dragging) return
      setPosition({
        x: dragStart.current.pos.x + (e.clientX - dragStart.current.x),
        y: dragStart.current.pos.y + (e.clientY - dragStart.current.y),
      })
    },
    [dragging]
  )

  const handlePointerUp = useCallback(() => {
    setDragging(false)
  }, [])

  useEffect(() => {
    if (!dragging) return
    const up = () => setDragging(false)
    window.addEventListener('pointerup', up)
    return () => window.removeEventListener('pointerup', up)
  }, [dragging])

  const getCroppedBlob = useCallback((): Promise<Blob> => {
    return new Promise((resolve, reject) => {
      const img = imgRef.current
      if (!img || !imageSize.w) {
        reject(new Error('Image not ready'))
        return
      }
      const canvas = document.createElement('canvas')
      canvas.width = CROP_SIZE
      canvas.height = CROP_SIZE
      const ctx = canvas.getContext('2d')
      if (!ctx) {
        reject(new Error('No canvas context'))
        return
      }
      const centerX = PREVIEW_SIZE / 2
      const centerY = PREVIEW_SIZE / 2
      const srcCx = (centerX - position.x) / scale
      const srcCy = (centerY - position.y) / scale
      const srcR = CIRCLE_R / scale
      let sx = srcCx - srcR
      let sy = srcCy - srcR
      const sSize = Math.min(2 * srcR, imageSize.w, imageSize.h)
      sx = Math.max(0, Math.min(sx, imageSize.w - sSize))
      sy = Math.max(0, Math.min(sy, imageSize.h - sSize))

      ctx.beginPath()
      ctx.arc(CIRCLE_R, CIRCLE_R, CIRCLE_R, 0, Math.PI * 2)
      ctx.closePath()
      ctx.clip()
      ctx.drawImage(img, sx, sy, sSize, sSize, 0, 0, CROP_SIZE, CROP_SIZE)
      canvas.toBlob(
        (blob) => {
          if (blob) resolve(blob)
          else reject(new Error('Export failed'))
        },
        'image/png',
        0.92
      )
    })
  }, [position, scale, imageSize])

  const handleConfirm = useCallback(async () => {
    try {
      const blob = await getCroppedBlob()
      onConfirm(blob)
    } catch {
      onCancel()
    }
  }, [getCroppedBlob, onConfirm, onCancel])

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-4">
      <div
        className="bg-background rounded-lg shadow-xl max-w-full flex flex-col"
        onPointerMove={handlePointerMove}
        onPointerLeave={handlePointerUp}
      >
        <div className="p-4 border-b">
          <h3 className="font-semibold text-lg">裁剪头像</h3>
          <p className="text-sm text-muted-foreground mt-1">
            拖动图片调整位置，缩放后点击确定，将截取圆形区域作为头像
          </p>
        </div>

        <div
          className="relative overflow-hidden bg-muted flex items-center justify-center"
          style={{ width: PREVIEW_SIZE, height: PREVIEW_SIZE }}
        >
          <div
            className="absolute cursor-move select-none"
            style={{
              width: imageSize.w * scale,
              height: imageSize.h * scale,
              left: position.x,
              top: position.y,
            }}
            onPointerDown={handlePointerDown}
          >
            <img
              ref={imgRef}
              src={imageUrl}
              alt="裁剪预览"
              draggable={false}
              onLoad={onImageLoad}
              className="pointer-events-none w-full h-full object-cover"
              style={{ width: imageSize.w * scale, height: imageSize.h * scale }}
            />
          </div>
          <div
            className="absolute inset-0 pointer-events-none rounded-full border-2 border-white shadow-inner"
            style={{
              width: CROP_SIZE,
              height: CROP_SIZE,
              left: (PREVIEW_SIZE - CROP_SIZE) / 2,
              top: (PREVIEW_SIZE - CROP_SIZE) / 2,
              boxShadow: '0 0 0 9999px rgba(0,0,0,0.5)',
            }}
          />
        </div>

        <div className="p-4 space-y-4 border-t">
          <div className="space-y-2">
            <Label>缩放</Label>
            <input
              type="range"
              min={0.5}
              max={3}
              step={0.05}
              value={scale}
              onChange={(e) => setScale(Number(e.target.value))}
              className="w-full h-2 rounded-lg appearance-none cursor-pointer bg-muted"
            />
          </div>
          <div className="flex justify-end gap-2">
            <Button variant="outline" onClick={onCancel}>
              取消
            </Button>
            <Button onClick={handleConfirm}>确定</Button>
          </div>
        </div>
      </div>
    </div>
  )
}
