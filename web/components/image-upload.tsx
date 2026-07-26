"use client"

import * as React from "react"
import { UploadIcon, XIcon, Loader2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import { uploadApi } from "@/lib/api"
import { getUploadBaseUrl } from "@/lib/utils"

interface ImageUploadProps {
  value?: string
  onChange: (url: string) => void
  label?: string
  accept?: string
}

export function ImageUpload({ value, onChange, label = "Upload Image", accept = "image/*" }: ImageUploadProps) {
  const [uploading, setUploading] = React.useState(false)
  const [error, setError] = React.useState("")
  const inputRef = React.useRef<HTMLInputElement>(null)

  const handleUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    setUploading(true)
    setError("")
    try {
      const { data } = await uploadApi.file(file)
      onChange(data.url)
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "Upload failed"
      if (typeof err === "object" && err !== null && "response" in err) {
        const axiosErr = err as { response?: { data?: { error?: string } } }
        setError(axiosErr.response?.data?.error || message)
      } else {
        setError(message)
      }
    } finally {
      setUploading(false)
      if (inputRef.current) inputRef.current.value = ""
    }
  }

  const handleRemove = () => {
    onChange("")
    setError("")
  }

  const baseUrl = getUploadBaseUrl()
  const imageUrl = value?.startsWith("/uploads") ? `${baseUrl}${value}` : value

  return (
    <div className="space-y-2">
      <input
        ref={inputRef}
        type="file"
        accept={accept}
        onChange={handleUpload}
        className="hidden"
        id={`upload-${label}`}
      />

      {value ? (
        <div className="relative inline-block">
          <img
            src={imageUrl}
            alt={label}
            className="h-32 w-32 rounded-md border object-cover"
            onError={(e) => {
              const target = e.target as HTMLImageElement
              target.style.display = "none"
            }}
          />
          <Button
            type="button"
            variant="destructive"
            size="icon"
            className="absolute -top-2 -right-2 h-6 w-6 rounded-full"
            onClick={handleRemove}
          >
            <XIcon className="h-3 w-3" />
          </Button>
        </div>
      ) : (
        <label
          htmlFor={`upload-${label}`}
          className="flex h-32 w-44 cursor-pointer flex-col items-center justify-center rounded-md border-2 border-dashed text-muted-foreground hover:border-primary/50 hover:text-primary transition-colors"
        >
          {uploading ? (
            <Loader2 className="h-6 w-6 animate-spin" />
          ) : (
            <>
              <UploadIcon className="h-6 w-6 mb-1" />
              <span className="text-xs text-center">{label}</span>
            </>
          )}
        </label>
      )}

      {error && <p className="text-sm text-destructive">{error}</p>}
    </div>
  )
}
