"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import { UploadIcon, DownloadIcon, Loader2, CheckCircleIcon, AlertCircleIcon, ArrowLeftIcon } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { organizationApi } from "@/lib/api"
import { toast } from "sonner"

export default function ImportOrganizationPage() {
  const router = useRouter()
  const [file, setFile] = React.useState<File | null>(null)
  const [uploading, setUploading] = React.useState(false)
  const [result, setResult] = React.useState<{
    message: string
    created: number
    updated: number
    total: number
    errors?: string[]
    results?: { type: string; name: string; action: string }[]
  } | null>(null)

  const downloadBlob = (data: BlobPart, filename: string, mime: string) => {
    const blob = new Blob([data], { type: mime })
    const url = URL.createObjectURL(blob)
    const a = document.createElement("a")
    a.href = url
    a.download = filename
    a.click()
    URL.revokeObjectURL(url)
  }

  const handleDownloadTemplate = async () => {
    try {
      const res = await organizationApi.downloadTemplate()
      downloadBlob(res.data, "organization_import_template.xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
      toast.success("Template downloaded")
    } catch {
      toast.error("Failed to download template")
    }
  }

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const f = e.target.files?.[0]
    if (f) {
      if (!f.name.endsWith(".xlsx")) {
        toast.error("Only .xlsx files are supported")
        return
      }
      setFile(f)
      setResult(null)
    }
  }

  const handleUpload = async () => {
    if (!file) { toast.error("Select a file first"); return }
    setUploading(true)
    setResult(null)
    try {
      const { data } = await organizationApi.importExcel(file)
      setResult(data)
      toast.success(data.message || "Import complete")
    } catch (err: unknown) {
      const axiosErr = err as { response?: { data?: { error?: string; errors?: string[]; results?: { type: string; name: string; action: string }[] } } }
      const errData = axiosErr?.response?.data
      if (errData?.errors) {
        setResult({ message: errData.error || "Import failed", created: 0, updated: 0, total: 0, errors: errData.errors, results: errData.results })
      } else {
        toast.error(errData?.error || "Import failed")
      }
    } finally {
      setUploading(false)
    }
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <UploadIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Import Organization</h1>
            <p className="text-muted-foreground mt-1">Bulk import departments, sections, designations, lines, groups &amp; floors</p>
          </div>
        </div>
        <Button variant="outline" onClick={() => router.push("/information/organization")}>
          <ArrowLeftIcon className="mr-2 h-4 w-4" />
          Back to Organization
        </Button>
      </div>

      <div className="px-4 lg:px-6">
        <Card className="cursor-pointer hover:border-primary transition-colors" onClick={handleDownloadTemplate}>
          <CardHeader className="pb-2">
            <CardTitle className="text-base flex items-center gap-2">
              <DownloadIcon className="h-4 w-4" /> Download Template
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              Get the Excel template with Format and Demo sheets
            </p>
          </CardContent>
        </Card>
      </div>

      <div className="px-4 lg:px-6">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Upload Excel File</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="border-2 border-dashed rounded-lg p-8 text-center">
              <input
                type="file"
                accept=".xlsx"
                onChange={handleFileChange}
                className="hidden"
                id="file-upload"
              />
              <label htmlFor="file-upload" className="cursor-pointer">
                <UploadIcon className="h-10 w-10 mx-auto text-muted-foreground" />
                <p className="mt-2 text-sm font-medium">
                  {file ? file.name : "Click to select .xlsx file"}
                </p>
                <p className="text-xs text-muted-foreground mt-1">
                  {file ? `${(file.size / 1024).toFixed(1)} KB` : "Supports .xlsx format"}
                </p>
              </label>
            </div>

            {file && (
              <Button onClick={handleUpload} disabled={uploading} className="w-full">
                {uploading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <UploadIcon className="mr-2 h-4 w-4" />}
                {uploading ? "Importing..." : "Import Organization"}
              </Button>
            )}

            {result && (
              <div className={`rounded-md p-4 ${result.errors?.length ? "bg-destructive/15" : "bg-emerald-50 dark:bg-emerald-950/20"}`}>
                <div className="flex items-center gap-2">
                  {result.errors?.length ? <AlertCircleIcon className="h-5 w-5 text-destructive" /> : <CheckCircleIcon className="h-5 w-5 text-emerald-600" />}
                  <p className={`font-medium ${result.errors?.length ? "text-destructive" : "text-emerald-700 dark:text-emerald-400"}`}>
                    {result.message}
                  </p>
                </div>
                {result.total > 0 && (
                  <div className="mt-2 grid grid-cols-3 gap-2 text-sm">
                    <div className="rounded bg-background/50 p-2 text-center">
                      <p className="font-bold text-lg">{result.total}</p>
                      <p className="text-muted-foreground text-xs">Total</p>
                    </div>
                    <div className="rounded bg-background/50 p-2 text-center">
                      <p className="font-bold text-lg text-emerald-600">{result.created}</p>
                      <p className="text-muted-foreground text-xs">Created</p>
                    </div>
                    <div className="rounded bg-background/50 p-2 text-center">
                      <p className="font-bold text-lg text-amber-600">{result.updated}</p>
                      <p className="text-muted-foreground text-xs">Updated</p>
                    </div>
                  </div>
                )}
                {result.results && result.results.length > 0 && !result.errors?.length && (
                  <div className="mt-2 max-h-40 overflow-y-auto text-xs text-muted-foreground space-y-1 border rounded p-2">
                    {result.results.map((r, i) => (
                      <p key={i}>
                        <span className="font-medium">{r.type}:</span> {r.name}
                        <span className={`ml-1 ${r.action === "created" ? "text-emerald-600" : "text-amber-600"}`}>
                          ({r.action})
                        </span>
                      </p>
                    ))}
                  </div>
                )}
                {result.errors && result.errors.length > 0 && (
                  <div className="mt-2 max-h-32 overflow-y-auto text-xs text-destructive space-y-1">
                    {result.errors.map((e, i) => <p key={i}>{e}</p>)}
                  </div>
                )}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
