"use client"

import * as React from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import {
  DatabaseIcon,
  DownloadIcon,
  UploadIcon,
  RotateCcwIcon,
  AlertTriangleIcon,
  Loader2,
  CheckCircleIcon,
  FileTextIcon,
  Trash2Icon,
} from "lucide-react"
import { databaseApi } from "@/lib/api"
import { toast } from "sonner"
import { isSuperAdmin } from "@/lib/auth"
import {
  Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger,
} from "@/components/ui/dialog"

interface BackupFile {
  name: string
  size_kb: number
  modified: string
}

export default function DatabasePage() {
  const [backups, setBackups] = React.useState<BackupFile[]>([])
  const [loading, setLoading] = React.useState(false)
  const [backingUp, setBackingUp] = React.useState(false)
  const [importing, setImporting] = React.useState(false)
  const [resetting, setResetting] = React.useState(false)
  const [importFile, setImportFile] = React.useState<File | null>(null)
  const [importResult, setImportResult] = React.useState<string | null>(null)
  const [resetDialogOpen, setResetDialogOpen] = React.useState(false)

  const fetchBackups = async () => {
    setLoading(true)
    try {
      const { data } = await databaseApi.listBackups()
      setBackups(Array.isArray(data) ? data : [])
    } catch {
      toast.error("Failed to load backups")
    } finally {
      setLoading(false)
    }
  }

  React.useEffect(() => { fetchBackups() }, [])

  const handleBackup = async () => {
    setBackingUp(true)
    try {
      const { data } = await databaseApi.backup()
      toast.success(data.message || "Backup created")
      fetchBackups()
    } catch {
      toast.error("Backup failed")
    } finally {
      setBackingUp(false)
    }
  }

  const handleExport = async (filename: string) => {
    try {
      const res = await databaseApi.export(filename)
      const blob = new Blob([res.data], { type: "application/octet-stream" })
      const url = URL.createObjectURL(blob)
      const a = document.createElement("a")
      a.href = url
      a.download = filename
      a.click()
      URL.revokeObjectURL(url)
    } catch {
      toast.error("Export failed")
    }
  }

  const handleDeleteBackup = async (filename: string) => {
    if (!window.confirm(`Delete backup "${filename}"? This cannot be undone.`)) return
    try {
      const { data } = await databaseApi.deleteBackup(filename)
      toast.success(data.message || "Backup deleted")
      fetchBackups()
    } catch {
      toast.error("Failed to delete backup")
    }
  }

  const handleImport = async () => {
    if (!importFile) { toast.error("Select a file first"); return }
    setImporting(true)
    setImportResult(null)
    try {
      const { data } = await databaseApi.importSql(importFile)
      setImportResult(data.output || "Import completed")
      toast.success(data.message || "Import complete")
      fetchBackups()
    } catch (err: unknown) {
      const axiosErr = err as { response?: { data?: { error?: string; output?: string } } }
      setImportResult(axiosErr?.response?.data?.output || axiosErr?.response?.data?.error || "Import failed")
      toast.error("Import failed")
    } finally {
      setImporting(false)
    }
  }

  const handleReset = async () => {
    setResetting(true)
    setResetDialogOpen(false)
    try {
      const { data } = await databaseApi.reset()
      toast.success(data.message || "Database reset completed")
    } catch {
      toast.error("Reset failed")
    } finally {
      setResetting(false)
    }
  }

  const downloadBlob = (data: BlobPart, filename: string, mime: string) => {
    const blob = new Blob([data], { type: mime })
    const url = URL.createObjectURL(blob)
    const a = document.createElement("a")
    a.href = url
    a.download = filename
    a.click()
    URL.revokeObjectURL(url)
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center gap-2">
        <DatabaseIcon className="h-6 w-6 text-muted-foreground" />
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Database Management</h1>
          <p className="text-muted-foreground mt-1">Backup, import, export, and reset the database</p>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2 px-4 lg:px-6">
        <Card>
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <DownloadIcon className="h-4 w-4" />
              Backup &amp; Export
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {isSuperAdmin() && (
              <Button onClick={handleBackup} disabled={backingUp} className="w-full">
                {backingUp ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <DatabaseIcon className="mr-2 h-4 w-4" />}
                {backingUp ? "Creating Backup..." : "Create Backup"}
              </Button>
            )}

            <div className="border rounded-lg">
              <div className="p-3 border-b bg-muted/50 text-sm font-medium">Recent Backups</div>
              {loading ? (
                <div className="flex justify-center py-6">
                  <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
                </div>
              ) : backups.length === 0 ? (
                <div className="p-6 text-center text-sm text-muted-foreground">No backups yet</div>
              ) : (
                <div className="divide-y max-h-60 overflow-y-auto">
                  {backups.map((b) => (
                    <div key={b.name} className="flex items-center justify-between p-3 text-sm hover:bg-muted/30">
                      <div className="flex items-center gap-2 min-w-0">
                        <FileTextIcon className="h-4 w-4 text-muted-foreground shrink-0" />
                        <span className="truncate">{b.name}</span>
                      </div>
                      <div className="flex items-center gap-2 shrink-0">
                        <span className="text-xs text-muted-foreground">{b.size_kb} KB</span>
                        <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => handleExport(b.name)} title="Download">
                          <DownloadIcon className="h-3.5 w-3.5" />
                        </Button>
                        {isSuperAdmin() && (
                          <Button variant="ghost" size="icon" className="h-7 w-7 text-red-500 hover:text-red-700 hover:bg-red-50" onClick={() => handleDeleteBackup(b.name)} title="Delete">
                            <Trash2Icon className="h-3.5 w-3.5" />
                          </Button>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <UploadIcon className="h-4 w-4" />
              Import SQL
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="border-2 border-dashed rounded-lg p-6 text-center">
              <input
                type="file"
                accept=".sql"
                onChange={(e) => { setImportFile(e.target.files?.[0] ?? null); setImportResult(null) }}
                className="hidden"
                id="sql-upload"
              />
              <label htmlFor="sql-upload" className="cursor-pointer">
                <UploadIcon className="h-8 w-8 mx-auto text-muted-foreground" />
                <p className="mt-2 text-sm font-medium">
                  {importFile ? importFile.name : "Click to select .sql file"}
                </p>
              </label>
            </div>
            {isSuperAdmin() && (
              <Button onClick={handleImport} disabled={importing || !importFile} className="w-full">
                {importing ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <UploadIcon className="mr-2 h-4 w-4" />}
                {importing ? "Importing..." : "Import SQL"}
              </Button>
            )}
            {importResult && (
              <div className="rounded-md bg-emerald-50 dark:bg-emerald-950/20 p-3">
                <div className="flex items-center gap-2">
                  <CheckCircleIcon className="h-4 w-4 text-emerald-600" />
                  <p className="text-sm text-emerald-700 dark:text-emerald-400">Import result</p>
                </div>
                <pre className="mt-1 text-xs text-muted-foreground whitespace-pre-wrap max-h-32 overflow-y-auto">{importResult}</pre>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      <div className="px-4 lg:px-6">
        <Card className="border-red-200 dark:border-red-900">
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2 text-red-600">
              <AlertTriangleIcon className="h-4 w-4" />
              Danger Zone
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground mb-4">
              This will drop all tables and re-run auto-migration. All data will be permanently lost.
            </p>
            {isSuperAdmin() && (
              <Dialog open={resetDialogOpen} onOpenChange={setResetDialogOpen}>
                <DialogTrigger asChild>
                  <Button variant="destructive" disabled={resetting}>
                    {resetting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <RotateCcwIcon className="mr-2 h-4 w-4" />}
                    {resetting ? "Resetting..." : "Reset Database"}
                  </Button>
                </DialogTrigger>
                <DialogContent>
                  <DialogHeader>
                    <DialogTitle className="text-red-600">Confirm Database Reset</DialogTitle>
                    <DialogDescription>
                      This action is irreversible. All data across all tables will be permanently deleted
                      and the database schema will be re-created from scratch.
                    </DialogDescription>
                  </DialogHeader>
                  <div className="flex justify-end gap-3">
                    <Button variant="outline" onClick={() => setResetDialogOpen(false)}>Cancel</Button>
                    <Button variant="destructive" onClick={handleReset} disabled={resetting}>
                      {resetting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Trash2Icon className="mr-2 h-4 w-4" />}
                      Yes, Reset Everything
                    </Button>
                  </div>
                </DialogContent>
              </Dialog>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
