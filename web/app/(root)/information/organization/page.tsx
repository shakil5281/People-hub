"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import { BuildingIcon, PlusIcon, UploadIcon, Loader2 } from "lucide-react"
import { toast } from "sonner"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import {
  Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle
} from "@/components/ui/dialog"
import { Department, Section, Designation, Line } from "@/components/data/organization-data"
import { departmentApi, sectionApi, designationApi, lineApi } from "@/lib/api"

function ItemForm({ title, open, onOpenChange, initialData, onSave }: {
  title: string; open: boolean; onOpenChange: (v: boolean) => void; initialData?: { name: string; name_bn: string }; onSave: (name: string, nameBn: string) => Promise<void>
}) {
  const [name, setName] = React.useState(initialData?.name || "")
  const [nameBn, setNameBn] = React.useState(initialData?.name_bn || "")
  const [saving, setSaving] = React.useState(false)
  React.useEffect(() => { setName(initialData?.name || ""); setNameBn(initialData?.name_bn || "") }, [initialData, open])
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>Enter the details below.</DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
          <div className="space-y-2">
            <Label>Name (English) <span className="text-red-500">*</span></Label>
            <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="Enter English name" />
          </div>
          <div className="space-y-2">
            <Label>Name (Bangla)</Label>
            <Input value={nameBn} onChange={(e) => setNameBn(e.target.value)} placeholder="Enter Bangla name" className="bangla-input" />
          </div>
          <div className="flex justify-end gap-3">
            <Button variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
            <Button onClick={async () => { setSaving(true); try { await onSave(name, nameBn); onOpenChange(false) } catch { } finally { setSaving(false) } }} disabled={!name.trim() || saving}>
              {saving ? <Loader2 className="h-4 w-4 animate-spin" /> : "Save"}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}

const bnCell = ({ getValue }: { getValue: () => unknown }) => (
  <span className="bangla-input">{getValue() as string}</span>
)

const deptColumns: ColumnDef<Department>[] = [
  { id: "sl", header: "SL", cell: ({ row }) => row.index + 1 },
  { accessorKey: "name", header: "Name (EN)" },
  { accessorKey: "name_bn", header: "Name (BN)", cell: bnCell },
]

const sectionColumns: ColumnDef<Section>[] = [
  { id: "sl", header: "SL", cell: ({ row }) => row.index + 1 },
  { accessorKey: "name", header: "Name (EN)" },
  { accessorKey: "name_bn", header: "Name (BN)", cell: bnCell },
]

const designationColumns: ColumnDef<Designation>[] = [
  { id: "sl", header: "SL", cell: ({ row }) => row.index + 1 },
  { accessorKey: "name", header: "Name (EN)" },
  { accessorKey: "name_bn", header: "Name (BN)", cell: bnCell },
]

const lineColumns: ColumnDef<Line>[] = [
  { id: "sl", header: "SL", cell: ({ row }) => row.index + 1 },
  { accessorKey: "name", header: "Name (EN)" },
  { accessorKey: "name_bn", header: "Name (BN)", cell: bnCell },
]

export default function OrganizationPage() {
  const [departments, setDepartments] = React.useState<Department[]>([])
  const [sections, setSections] = React.useState<Section[]>([])
  const [designations, setDesignations] = React.useState<Designation[]>([])
  const [lines, setLines] = React.useState<Line[]>([])

  const [loading, setLoading] = React.useState({ dept: true, sec: false, desig: false, line: false })

  const [selectedDept, setSelectedDept] = React.useState("")
  const [selectedSection, setSelectedSection] = React.useState("")

  type FormItem = Department | Section | Designation | Line
  const [form, setForm] = React.useState<{ open: boolean; edit?: FormItem | null; type: string }>({ open: false, type: "" })
  const [tab, setTab] = React.useState("department")
  const router = useRouter()

  const [page, setPage] = React.useState(1)
  const [limit, setLimit] = React.useState(20)
  const [total, setTotal] = React.useState(0)
  const [totalPages, setTotalPages] = React.useState(0)

  const fetchDepartments = async (p?: number, l?: number) => {
    setLoading(p => ({ ...p, dept: true }))
    try { 
      const { data } = await departmentApi.list({ page: String(p ?? page), limit: String(l ?? limit) })
      setDepartments(Array.isArray(data.data) ? data.data : [])
      setTotal(data.total ?? 0)
      setTotalPages(data.total_pages ?? 0)
    } catch { toast.error("Failed to load departments") }
    finally { setLoading(p => ({ ...p, dept: false })) }
  }
  const fetchSections = async (deptId: string, p?: number, l?: number) => {
    if (!deptId) return setSections([])
    setLoading(p => ({ ...p, sec: true }))
    try { 
      const { data } = await sectionApi.list(deptId, { page: String(p ?? page), limit: String(l ?? limit) })
      setSections(Array.isArray(data.data) ? data.data : [])
      setTotal(data.total ?? 0)
      setTotalPages(data.total_pages ?? 0)
    } catch { toast.error("Failed to load sections") }
    finally { setLoading(p => ({ ...p, sec: false })) }
  }
  const fetchDesignations = async (secId: string, p?: number, l?: number) => {
    if (!secId) return setDesignations([])
    setLoading(p => ({ ...p, desig: true }))
    try { 
      const { data } = await designationApi.list(secId, { page: String(p ?? page), limit: String(l ?? limit) })
      setDesignations(Array.isArray(data.data) ? data.data : [])
      setTotal(data.total ?? 0)
      setTotalPages(data.total_pages ?? 0)
    } catch { toast.error("Failed to load designations") }
    finally { setLoading(p => ({ ...p, desig: false })) }
  }
  const fetchLines = async (secId: string, p?: number, l?: number) => {
    if (!secId) return setLines([])
    setLoading(p => ({ ...p, line: true }))
    try { 
      const { data } = await lineApi.list(secId, { page: String(p ?? page), limit: String(l ?? limit) })
      setLines(Array.isArray(data.data) ? data.data : [])
      setTotal(data.total ?? 0)
      setTotalPages(data.total_pages ?? 0)
    } catch { toast.error("Failed to load lines") }
    finally { setLoading(p => ({ ...p, line: false })) }
  }

  React.useEffect(() => { fetchDepartments(1, 20) }, [])

  React.useEffect(() => {
    setSelectedSection("")
    setDesignations([])
    setLines([])
    setPage(1)
    fetchSections(selectedDept, 1, limit)
  }, [selectedDept])

  React.useEffect(() => {
    setPage(1)
    fetchDesignations(selectedSection, 1, limit)
    fetchLines(selectedSection, 1, limit)
  }, [selectedSection])

  React.useEffect(() => {
    if (tab === "department") fetchDepartments()
    else if (tab === "section" && selectedDept) fetchSections(selectedDept)
    else if (tab === "designation" && selectedSection) fetchDesignations(selectedSection)
    else if (tab === "line" && selectedSection) fetchLines(selectedSection)
  }, [page, limit])

  const openForm = (type: string, edit?: FormItem | null) => setForm({ open: true, edit, type })
  const closeForm = () => setForm({ open: false, edit: null, type: "" })

  const handleSave = async (name: string, nameBn: string) => {
    const { edit, type } = form
    try {
      if (type === "department") {
        if (edit) await departmentApi.update(edit.id, { name, name_bn: nameBn })
        else await departmentApi.create({ name, name_bn: nameBn })
        toast.success(edit ? "Updated" : "Created")
        fetchDepartments(1, limit)
      } else if (type === "section") {
        const payload = { name, name_bn: nameBn, department_id: selectedDept }
        if (edit) await sectionApi.update(edit.id, payload)
        else await sectionApi.create(payload)
        toast.success(edit ? "Updated" : "Created")
        fetchSections(selectedDept, 1, limit)
      } else if (type === "designation") {
        const payload = { name, name_bn: nameBn, section_id: selectedSection }
        if (edit) await designationApi.update(edit.id, payload)
        else await designationApi.create(payload)
        toast.success(edit ? "Updated" : "Created")
        fetchDesignations(selectedSection, 1, limit)
      } else if (type === "line") {
        const payload = { name, name_bn: nameBn, section_id: selectedSection }
        if (edit) await lineApi.update(edit.id, payload)
        else await lineApi.create(payload)
        toast.success(edit ? "Updated" : "Created")
        fetchLines(selectedSection, 1, limit)
      }
    } catch { }
  }

  const handleDelete = async (type: string, item: Department | Section | Designation | Line) => {
    try {
      if (type === "department") {
        await departmentApi.delete(item.id)
        toast.success("Deleted")
        fetchDepartments(1, limit)
        if (selectedDept === item.id) setSelectedDept("")
      } else if (type === "section") {
        await sectionApi.delete(item.id)
        toast.success("Deleted")
        fetchSections(selectedDept, 1, limit)
        if (selectedSection === item.id) setSelectedSection("")
      } else {
        if (type === "designation") {
          await designationApi.delete(item.id)
          toast.success("Deleted")
          fetchDesignations(selectedSection, 1, limit)
        } else {
          await lineApi.delete(item.id)
          toast.success("Deleted")
          fetchLines(selectedSection, 1, limit)
        }
      }
    } catch { }
  }

  const formTitle = form.edit ? `Edit ${form.type.charAt(0).toUpperCase() + form.type.slice(1)}` : `Add ${form.type.charAt(0).toUpperCase() + form.type.slice(1)}`

  const tableProps = {
    serverSide: true as const,
    page,
    pageSize: limit,
    pageCount: totalPages,
    total,
    onPageChange: (p: number) => setPage(p),
    onPageSizeChange: (size: number) => { setLimit(size); setPage(1); },
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <BuildingIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Organization</h1>
            <p className="text-muted-foreground mt-1">Manage departments, sections, designations and lines</p>
          </div>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={() => router.push("/information/organization/import")}>
            <UploadIcon className="mr-2 h-4 w-4" />
            Import
          </Button>
        </div>
      </div>

      <div className="px-4 lg:px-6">
        <Tabs value={tab} onValueChange={setTab}>
          <TabsList>
            <TabsTrigger value="department">Departments</TabsTrigger>
            <TabsTrigger value="section">Sections</TabsTrigger>
            <TabsTrigger value="designation">Designations</TabsTrigger>
            <TabsTrigger value="line">Lines</TabsTrigger>
          </TabsList>

          <TabsContent value="department" className="mt-4">
            <div className="flex items-center justify-between mb-4">
              <div />
              <Button onClick={() => openForm("department")}><PlusIcon className="mr-2 h-4 w-4" /> Add Department</Button>
            </div>
            <DataTable
              data={departments}
              columns={deptColumns}
              enableDnd
              onEdit={(row) => openForm("department", row)}
              onDelete={(row) => handleDelete("department", row)}
              {...tableProps}
              loading={loading.dept}
            />
          </TabsContent>

          <TabsContent value="section" className="mt-4">
            <div className="flex items-center justify-between mb-4">
              <select
                className="flex h-10 w-full max-w-xs rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                value={selectedDept}
                onChange={e => setSelectedDept(e.target.value)}
              >
                <option value="">-- Select Department --</option>
                {departments.map(d => <option key={d.id} value={d.id}>{d.name}</option>)}
              </select>
              <Button disabled={!selectedDept} onClick={() => openForm("section")}><PlusIcon className="mr-2 h-4 w-4" /> Add Section</Button>
            </div>
            <DataTable
              data={selectedDept ? sections : []}
              columns={sectionColumns}
              enableDnd
              onEdit={(row) => openForm("section", row)}
              onDelete={(row) => handleDelete("section", row)}
              {...tableProps}
              loading={loading.sec}
            />
          </TabsContent>

          <TabsContent value="designation" className="mt-4">
            <div className="flex items-center justify-between mb-4">
              <select
                className="flex h-10 w-full max-w-xs rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                value={selectedSection}
                onChange={e => setSelectedSection(e.target.value)}
              >
                <option value="">-- Select Section --</option>
                {sections.map(s => <option key={s.id} value={s.id}>{s.name}</option>)}
              </select>
              <Button disabled={!selectedSection} onClick={() => openForm("designation")}><PlusIcon className="mr-2 h-4 w-4" /> Add Designation</Button>
            </div>
            <DataTable
              data={selectedSection ? designations : []}
              columns={designationColumns}
              enableDnd
              onEdit={(row) => openForm("designation", row)}
              onDelete={(row) => handleDelete("designation", row)}
              {...tableProps}
              loading={loading.desig}
            />
          </TabsContent>

          <TabsContent value="line" className="mt-4">
            <div className="flex items-center justify-between mb-4">
              <select
                className="flex h-10 w-full max-w-xs rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                value={selectedSection}
                onChange={e => setSelectedSection(e.target.value)}
              >
                <option value="">-- Select Section --</option>
                {sections.map(s => <option key={s.id} value={s.id}>{s.name}</option>)}
              </select>
              <Button disabled={!selectedSection} onClick={() => openForm("line")}><PlusIcon className="mr-2 h-4 w-4" /> Add Line</Button>
            </div>
            <DataTable
              data={selectedSection ? lines : []}
              columns={lineColumns}
              enableDnd
              onEdit={(row) => openForm("line", row)}
              onDelete={(row) => handleDelete("line", row)}
              {...tableProps}
              loading={loading.line}
            />
          </TabsContent>
        </Tabs>
      </div>

      <ItemForm title={formTitle} open={form.open} onOpenChange={closeForm} initialData={form.edit || undefined} onSave={handleSave} />
    </div>
  )
}
