"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import { Plus, Edit, Trash2, Loader2, Building2, MapPin, Phone, Mail, ChevronLeft, ChevronRight, MoreVertical } from "lucide-react"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { Company, statusOptions } from "../data/company-data"
import { CompanyForm } from "../form/company-form"
import { companyApi } from "@/lib/api"
import { getUploadBaseUrl } from "@/lib/utils"

export function CompanyTable() {
  const router = useRouter()
  const [companies, setCompanies] = React.useState<Company[]>([])
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState("")
  const [deleteId, setDeleteId] = React.useState<string | null>(null)
  const [isDeleting, setIsDeleting] = React.useState(false)
  const [createOpen, setCreateOpen] = React.useState(false)
  const [currentPage, setCurrentPage] = React.useState(1)
  const [sortConfig, setSortConfig] = React.useState<{ key: keyof Company; direction: "asc" | "desc" } | null>(null)
  const [total, setTotal] = React.useState(0)
  const [totalPages, setTotalPages] = React.useState(0)
  const pageSize = 10

  const fetchCompanies = async (p?: number) => {
    setLoading(true)
    setError("")
    try {
      const { data: res } = await companyApi.list({ page: String(p ?? currentPage), limit: String(pageSize) })
      let list = Array.isArray(res.data) ? res.data : []
      if (sortConfig) {
        list = [...list].sort((a: any, b: any) => {
          const aVal = a[sortConfig.key] ?? ""
          const bVal = b[sortConfig.key] ?? ""
          if (aVal < bVal) return sortConfig.direction === "asc" ? -1 : 1
          if (aVal > bVal) return sortConfig.direction === "asc" ? 1 : -1
          return 0
        })
      }
      setCompanies(list)
      setTotal(res.total ?? 0)
      setTotalPages(res.total_pages ?? 0)
    } catch {
      setError("Failed to load companies")
    } finally {
      setLoading(false)
    }
  }

  React.useEffect(() => {
    fetchCompanies()
  }, [])

  React.useEffect(() => {
    fetchCompanies()
  }, [currentPage])

  React.useEffect(() => {
    fetchCompanies(1)
  }, [sortConfig])

  const handleDelete = async (id: string) => {
    setIsDeleting(true)
    try {
      await companyApi.delete(id)
      setCompanies((prev) => prev.filter((c) => c.id !== id))
      setDeleteId(null)
      toast.success("Company deleted successfully")
      fetchCompanies()
    } catch {
      toast.error("Failed to delete company")
    } finally {
      setIsDeleting(false)
    }
  }

  const handleSort = (key: keyof Company) => {
    setSortConfig((prev) => ({
      key,
      direction: prev?.key === key && prev.direction === "asc" ? "desc" : "asc",
    }))
    setCurrentPage(1)
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Companies</h1>
          <p className="text-muted-foreground mt-1">Manage company information and details</p>
        </div>
        <Button onClick={() => setCreateOpen(true)} className="w-full sm:w-auto">
          <Plus className="mr-2 h-4 w-4" />
          Add Company
        </Button>
      </div>

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent className="max-w-6xl min-w-[700px] max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Add New Company</DialogTitle>
            <DialogDescription>Fill in the company details below.</DialogDescription>
          </DialogHeader>
          <CompanyForm
            onSuccess={() => {
              setCreateOpen(false)
              fetchCompanies()
            }}
          />
        </DialogContent>
      </Dialog>

      {error && (
        <div className="rounded-md bg-destructive/15 px-4 py-3 text-sm text-destructive">
          {error}
        </div>
      )}

      <Card>
        <CardHeader className="pb-0">
          <CardTitle className="text-lg">All Companies ({total})</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          <div className="overflow-x-auto">
            <Table>
              <TableHeader className="bg-muted/50">
                <TableRow>
                  <TableHead className="cursor-pointer select-none" onClick={() => handleSort("company_name_en")}>
                    <div className="flex items-center gap-1">
                      Company
                    </div>
                  </TableHead>
                  <TableHead>
                    <div className="flex items-center gap-1">Contact</div>
                  </TableHead>
                  <TableHead>Address</TableHead>
                  <TableHead>Signature</TableHead>
                  <TableHead className="cursor-pointer select-none" onClick={() => handleSort("status")}>
                    <div className="flex items-center gap-1">
                      Status
                    </div>
                  </TableHead>
                  <TableHead className="w-48">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {loading ? (
                  <TableRow>
                    <TableCell colSpan={6} className="text-center py-12">
                      <Loader2 className="mx-auto h-8 w-8 animate-spin text-muted-foreground" />
                      <p className="text-sm text-muted-foreground mt-2">Loading companies...</p>
                    </TableCell>
                  </TableRow>
                ) : companies.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={6} className="text-center py-12 text-muted-foreground">
                      <Building2 className="mx-auto h-12 w-12 text-muted-foreground/50 mb-2" />
                      <p className="font-medium">No companies yet</p>
                      <p className="text-sm">Click &quot;Add Company&quot; to get started</p>
                    </TableCell>
                  </TableRow>
                ) : (
                  companies.map((company) => (
                    <TableRow key={company.id} className="hover:bg-muted/50 transition-colors border-t border-border/50">
                      <TableCell className="font-medium">
                        <div className="flex items-center gap-3">
                          <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary/10 text-primary">
                            <Building2 className="h-5 w-5" />
                          </div>
                          <div>
                            <p>{company.company_name_en}</p>
                            {company.company_name_bn && (
                              <p className="text-xs text-muted-foreground bangla-input">{company.company_name_bn}</p>
                            )}
                          </div>
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="space-y-1 text-sm">
                          <div className="flex items-center gap-2 text-muted-foreground">
                            <Phone className="h-3.5 w-3.5" />
                            <span>{company.phone}</span>
                          </div>
                          {company.email && (
                            <div className="flex items-center gap-2 text-muted-foreground">
                              <Mail className="h-3.5 w-3.5" />
                              <span className="truncate max-w-[200px]">{company.email}</span>
                            </div>
                          )}
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2 text-sm text-muted-foreground">
                          <MapPin className="h-3.5 w-3.5" />
                          <span className="truncate max-w-[250px]">{company.address_en || company.address_bn}</span>
                        </div>
                      </TableCell>
                      <TableCell>
                        {company.signature ? (
                          <img
                            src={company.signature.startsWith("/uploads") ? `${getUploadBaseUrl()}${company.signature}` : company.signature}
                            alt="Signature"
                            className="h-10 w-24 rounded border object-contain"
                          />
                        ) : (
                          <span className="text-xs text-muted-foreground">No signature</span>
                        )}
                      </TableCell>
                      <TableCell>
                        <Badge
                          variant={company.status === "active" ? "default" : "secondary"}
                          className="capitalize"
                        >
                          {statusOptions.find((s) => s.value === company.status)?.label}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="icon" className="h-8 w-8">
                              <MoreVertical className="h-4 w-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem
                              onClick={() => router.push(`/information/company/${company.id}/edit`)}
                            >
                              <Edit className="mr-2 h-4 w-4" />
                              Edit
                            </DropdownMenuItem>
                            <DropdownMenuItem
                              onClick={() => setDeleteId(company.id)}
                              className="text-destructive focus:text-destructive"
                            >
                              <Trash2 className="mr-2 h-4 w-4" />
                              Delete
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>

          {totalPages > 1 && (
            <div className="flex items-center justify-between px-4 py-3 border-t">
              <p className="text-sm text-muted-foreground">
                Showing {(currentPage - 1) * pageSize + 1} to {Math.min(currentPage * pageSize, total)} of {total} companies
              </p>
              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="icon"
                  onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
                  disabled={currentPage === 1}
                >
                  <ChevronLeft className="h-4 w-4" />
                </Button>
                <span className="text-sm font-medium w-20 text-center">
                  Page {currentPage} of {totalPages}
                </span>
                <Button
                  variant="outline"
                  size="icon"
                  onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))}
                  disabled={currentPage === totalPages}
                >
                  <ChevronRight className="h-4 w-4" />
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      <Dialog open={!!deleteId} onOpenChange={(open) => !open && setDeleteId(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Company</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete this company? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteId(null)}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={() => deleteId && handleDelete(deleteId)} disabled={isDeleting}>
              {isDeleting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
