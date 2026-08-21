"use client"

import * as React from "react"
import { MapPinIcon, PlusIcon, Loader2 } from "lucide-react"
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
import { Division, District, Upazila, Union } from "@/components/data/address-data"
import { divisionApi, districtApi, upazilaApi, unionApi } from "@/lib/api"

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

const divisionColumns: ColumnDef<Division>[] = [
  { accessorKey: "name", header: "Name (EN)" },
  { accessorKey: "name_bn", header: "Name (BN)", cell: bnCell },
]

const districtColumns: ColumnDef<District>[] = [
  { accessorKey: "name", header: "Name (EN)" },
  { accessorKey: "name_bn", header: "Name (BN)", cell: bnCell },
]

const upazilaColumns: ColumnDef<Upazila>[] = [
  { accessorKey: "name", header: "Name (EN)" },
  { accessorKey: "name_bn", header: "Name (BN)", cell: bnCell },
]

const unionColumns: ColumnDef<Union>[] = [
  { accessorKey: "name", header: "Name (EN)" },
  { accessorKey: "name_bn", header: "Name (BN)", cell: bnCell },
]

type FormItem = Division | District | Upazila | Union

export default function AddressPage() {
  const [divisions, setDivisions] = React.useState<Division[]>([])
  const [districts, setDistricts] = React.useState<District[]>([])
  const [upazilas, setUpazilas] = React.useState<Upazila[]>([])
  const [unions, setUnions] = React.useState<Union[]>([])

  const [loading, setLoading] = React.useState({ div: true, dist: false, upa: false, uni: false })

  const [selectedDivision, setSelectedDivision] = React.useState("")
  const [selectedDistrict, setSelectedDistrict] = React.useState("")
  const [selectedUpazila, setSelectedUpazila] = React.useState("")

  const [form, setForm] = React.useState<{ open: boolean; edit?: FormItem | null; type: string }>({ open: false, type: "" })
  const [tab, setTab] = React.useState("division")

  const [page, setPage] = React.useState(1)
  const [limit, setLimit] = React.useState(20)
  const [total, setTotal] = React.useState(0)
  const [totalPages, setTotalPages] = React.useState(0)

  const fetchDivisions = async (p?: number, l?: number) => {
    setLoading(prev => ({ ...prev, div: true }))
    try { 
      const { data } = await divisionApi.list({ page: String(p ?? page), limit: String(l ?? limit) })
      setDivisions(Array.isArray(data.data) ? data.data : [])
      setTotal(data.total ?? 0)
      setTotalPages(data.total_pages ?? 0)
    } catch { toast.error("Failed to load divisions") }
    finally { setLoading(prev => ({ ...prev, div: false })) }
  }
  const fetchDistricts = async (divId: string, p?: number, l?: number) => {
    if (!divId) return setDistricts([])
    setLoading(prev => ({ ...prev, dist: true }))
    try { 
      const { data } = await districtApi.list(divId, { page: String(p ?? page), limit: String(l ?? limit) })
      setDistricts(Array.isArray(data.data) ? data.data : [])
      setTotal(data.total ?? 0)
      setTotalPages(data.total_pages ?? 0)
    } catch { toast.error("Failed to load districts") }
    finally { setLoading(prev => ({ ...prev, dist: false })) }
  }
  const fetchUpazilas = async (distId: string, p?: number, l?: number) => {
    if (!distId) return setUpazilas([])
    setLoading(prev => ({ ...prev, upa: true }))
    try { 
      const { data } = await upazilaApi.list(distId, { page: String(p ?? page), limit: String(l ?? limit) })
      setUpazilas(Array.isArray(data.data) ? data.data : [])
      setTotal(data.total ?? 0)
      setTotalPages(data.total_pages ?? 0)
    } catch { toast.error("Failed to load upazilas") }
    finally { setLoading(prev => ({ ...prev, upa: false })) }
  }
  const fetchUnions = async (upaId: string, p?: number, l?: number) => {
    if (!upaId) return setUnions([])
    setLoading(prev => ({ ...prev, uni: true }))
    try { 
      const { data } = await unionApi.list(upaId, { page: String(p ?? page), limit: String(l ?? limit) })
      setUnions(Array.isArray(data.data) ? data.data : [])
      setTotal(data.total ?? 0)
      setTotalPages(data.total_pages ?? 0)
    } catch { toast.error("Failed to load unions") }
    finally { setLoading(prev => ({ ...prev, uni: false })) }
  }

  React.useEffect(() => { fetchDivisions(1, 20) }, [])

  React.useEffect(() => {
    setSelectedDistrict("")
    setSelectedUpazila("")
    setUpazilas([])
    setUnions([])
    setPage(1)
    fetchDistricts(selectedDivision, 1, limit)
  }, [selectedDivision])

  React.useEffect(() => {
    setSelectedUpazila("")
    setUnions([])
    setPage(1)
    fetchUpazilas(selectedDistrict, 1, limit)
  }, [selectedDistrict])

  React.useEffect(() => {
    setPage(1)
    fetchUnions(selectedUpazila, 1, limit)
  }, [selectedUpazila])

  React.useEffect(() => {
    if (tab === "division") fetchDivisions()
    else if (tab === "district" && selectedDivision) fetchDistricts(selectedDivision)
    else if (tab === "upazila" && selectedDistrict) fetchUpazilas(selectedDistrict)
    else if (tab === "union" && selectedUpazila) fetchUnions(selectedUpazila)
  }, [page, limit])

  const openForm = (type: string, edit?: FormItem | null) => setForm({ open: true, edit, type })
  const closeForm = () => setForm({ open: false, edit: null, type: "" })

  const handleSave = async (name: string, nameBn: string) => {
    const { edit, type } = form
    try {
      if (type === "division") {
        if (edit) await divisionApi.update(edit.id, { name, name_bn: nameBn })
        else await divisionApi.create({ name, name_bn: nameBn })
        toast.success(edit ? "Updated" : "Created")
        fetchDivisions(1, limit)
      } else if (type === "district") {
        const payload = { name, name_bn: nameBn, division_id: selectedDivision }
        if (edit) await districtApi.update(edit.id, payload)
        else await districtApi.create(payload)
        toast.success(edit ? "Updated" : "Created")
        fetchDistricts(selectedDivision, 1, limit)
      } else if (type === "upazila") {
        const payload = { name, name_bn: nameBn, district_id: selectedDistrict }
        if (edit) await upazilaApi.update(edit.id, payload)
        else await upazilaApi.create(payload)
        toast.success(edit ? "Updated" : "Created")
        fetchUpazilas(selectedDistrict, 1, limit)
      } else if (type === "union") {
        const payload = { name, name_bn: nameBn, upazila_id: selectedUpazila }
        if (edit) await unionApi.update(edit.id, payload)
        else await unionApi.create(payload)
        toast.success(edit ? "Updated" : "Created")
        fetchUnions(selectedUpazila, 1, limit)
      }
    } catch { }
  }

  const handleDelete = async (type: string, item: FormItem) => {
    try {
      if (type === "division") {
        await divisionApi.delete(item.id)
        toast.success("Deleted")
        fetchDivisions(1, limit)
        if (selectedDivision === item.id) setSelectedDivision("")
      } else if (type === "district") {
        await districtApi.delete(item.id)
        toast.success("Deleted")
        fetchDistricts(selectedDivision, 1, limit)
        if (selectedDistrict === item.id) setSelectedDistrict("")
      } else if (type === "upazila") {
        await upazilaApi.delete(item.id)
        toast.success("Deleted")
        fetchUpazilas(selectedDistrict, 1, limit)
      } else {
        await unionApi.delete(item.id)
        toast.success("Deleted")
        fetchUnions(selectedUpazila, 1, limit)
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
      <div className="px-4 lg:px-6 flex items-center gap-2">
        <MapPinIcon className="h-6 w-6 text-muted-foreground" />
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Address Management</h1>
          <p className="text-muted-foreground mt-1">Manage divisions, districts, upazilas and unions</p>
        </div>
      </div>

      <div className="px-4 lg:px-6">
        <Tabs value={tab} onValueChange={setTab}>
          <TabsList>
            <TabsTrigger value="division">Divisions</TabsTrigger>
            <TabsTrigger value="district">Districts</TabsTrigger>
            <TabsTrigger value="upazila">Upazilas</TabsTrigger>
            <TabsTrigger value="union">Unions</TabsTrigger>
          </TabsList>

          <TabsContent value="division" className="mt-4">
            <div className="flex items-center justify-between mb-4">
              <div />
              <Button onClick={() => openForm("division")}><PlusIcon className="mr-2 h-4 w-4" /> Add Division</Button>
            </div>
            <DataTable
              data={divisions}
              columns={divisionColumns}
              onEdit={(row) => openForm("division", row)}
              onDelete={(row) => handleDelete("division", row)}
              {...tableProps}
              loading={loading.div}
            />
          </TabsContent>

          <TabsContent value="district" className="mt-4">
            <div className="flex items-center justify-between mb-4">
              <select
                className="flex h-10 w-full max-w-xs rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                value={selectedDivision}
                onChange={e => setSelectedDivision(e.target.value)}
              >
                <option value="">-- Select Division --</option>
                {divisions.map(d => <option key={d.id} value={d.id}>{d.name}</option>)}
              </select>
              <Button disabled={!selectedDivision} onClick={() => openForm("district")}><PlusIcon className="mr-2 h-4 w-4" /> Add District</Button>
            </div>
            <DataTable
              data={selectedDivision ? districts : []}
              columns={districtColumns}
              onEdit={(row) => openForm("district", row)}
              onDelete={(row) => handleDelete("district", row)}
              {...tableProps}
              loading={loading.dist}
            />
          </TabsContent>

          <TabsContent value="upazila" className="mt-4">
            <div className="flex items-center justify-between mb-4">
              <select
                className="flex h-10 w-full max-w-xs rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                value={selectedDistrict}
                onChange={e => setSelectedDistrict(e.target.value)}
              >
                <option value="">-- Select District --</option>
                {districts.map(d => <option key={d.id} value={d.id}>{d.name}</option>)}
              </select>
              <Button disabled={!selectedDistrict} onClick={() => openForm("upazila")}><PlusIcon className="mr-2 h-4 w-4" /> Add Upazila</Button>
            </div>
            <DataTable
              data={selectedDistrict ? upazilas : []}
              columns={upazilaColumns}
              onEdit={(row) => openForm("upazila", row)}
              onDelete={(row) => handleDelete("upazila", row)}
              {...tableProps}
              loading={loading.upa}
            />
          </TabsContent>

          <TabsContent value="union" className="mt-4">
            <div className="flex items-center justify-between mb-4">
              <select
                className="flex h-10 w-full max-w-xs rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                value={selectedUpazila}
                onChange={e => setSelectedUpazila(e.target.value)}
              >
                <option value="">-- Select Upazila --</option>
                {upazilas.map(u => <option key={u.id} value={u.id}>{u.name}</option>)}
              </select>
              <Button disabled={!selectedUpazila} onClick={() => openForm("union")}><PlusIcon className="mr-2 h-4 w-4" /> Add Union</Button>
            </div>
            <DataTable
              data={selectedUpazila ? unions : []}
              columns={unionColumns}
              onEdit={(row) => openForm("union", row)}
              onDelete={(row) => handleDelete("union", row)}
              {...tableProps}
              loading={loading.uni}
            />
          </TabsContent>
        </Tabs>
      </div>

      <ItemForm title={formTitle} open={form.open} onOpenChange={closeForm} initialData={form.edit || undefined} onSave={handleSave} />
    </div>
  )
}
