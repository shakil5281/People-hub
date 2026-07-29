"use client"

import * as React from "react"
import { MoonIcon, PlusIcon, FilterIcon, XIcon } from "lucide-react"
import { useRouter } from "next/navigation"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { nightBillApi, companyApi, departmentApi, sectionApi, designationApi } from "@/lib/api"
import { FilterBar } from "@/components/filter-bar"
import type { FilterDef } from "@/components/filter-bar"
import { ButtonGroup } from "@/components/ui/button-group"
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger, SheetClose } from "@/components/ui/sheet"

interface NightBill {
  id: string
  company_id: string
  employee_id: string
  date: string
  bill_type: string
  night_hours: number
  rate: number
  amount: number
  status: string
  employee?: {
    name_en: string
    name_bn: string
  }
}

const statusBadge = (status: string) => {
  const map: Record<string, "default" | "secondary" | "outline" | "destructive"> = {
    pending: "outline",
    approved: "default",
    paid: "secondary",
  }
  return <Badge variant={map[status] || "outline"} className="capitalize">{status}</Badge>
}

export default function NightBillPage() {
  const router = useRouter()
  const [data, setData] = React.useState<NightBill[]>([])
  const [loading, setLoading] = React.useState(true)
  const [companies, setCompanies] = React.useState<Array<{ id: string; company_name_en: string }>>([])
  const [departments, setDepartments] = React.useState<Array<{ id: string; name: string }>>([])
  const [sections, setSections] = React.useState<Array<{ id: string; name: string }>>([])
  const [designations, setDesignations] = React.useState<Array<{ id: string; name: string }>>([])
  const [filters, setFilters] = React.useState<Record<string, string>>({})
  const [page, setPage] = React.useState(1)
  const [limit, setLimit] = React.useState(20)
  const [total, setTotal] = React.useState(0)
  const [totalPages, setTotalPages] = React.useState(0)
  const [mobileFilterOpen, setMobileFilterOpen] = React.useState(false)

  const fetchDepartments = React.useCallback(async (companyId: string) => {
    try {
      const { data: res } = await departmentApi.list({ company_id: companyId, limit: "200" })
      const list = res?.data || res?.departments || []
      setDepartments(Array.isArray(list) ? list : [])
    } catch { setDepartments([]) }
  }, [])

  const fetchOrgData = React.useCallback(async (companyId: string) => {
    try {
      const [sRes, desRes] = await Promise.all([
        sectionApi.list(undefined, { company_id: companyId, limit: "200" }),
        designationApi.list(undefined, { company_id: companyId, limit: "200" }),
      ])
      const extract = (res: any) => {
        const d = res?.data || res?.sections || res?.designations || []
        return Array.isArray(d) ? d : []
      }
      setSections(extract(sRes))
      setDesignations(extract(desRes))
    } catch { }
  }, [])

  const filterDefs: FilterDef[] = React.useMemo(() => [
    {
      key: "company_id", label: "Company", type: "select",
      options: companies.map((c) => ({ value: c.id, label: c.company_name_en })),
    },
    {
      key: "date_from", label: "Date From", type: "datepicker",
    },
    {
      key: "date_to", label: "Date To", type: "datepicker",
    },
    {
      key: "bill_type", label: "Bill Type", type: "select",
      options: [
        { value: "fixed", label: "Fixed" },
        { value: "hourly", label: "Hourly" },
      ],
    },
    {
      key: "employee_id", label: "Employee ID", type: "text", placeholder: "EMP-001",
    },
  ], [companies])

  const summaryColumns: ColumnDef<NightBill>[] = [
    { accessorKey: "employee_id", header: "Employee ID" },
    {
      header: "Name",
      accessorFn: (r) => r.employee?.name_en || "-",
    },
    { accessorKey: "date", header: "Date" },
    {
      accessorKey: "bill_type", header: "Type",
      cell: ({ row }) => <span className="capitalize">{row.original.bill_type}</span>,
    },
    { accessorKey: "night_hours", header: "Hours",
      cell: ({ row }) => row.original.night_hours > 0 ? row.original.night_hours.toFixed(2) : "—",
    },
    {
      accessorKey: "rate", header: "Rate",
      cell: ({ row }) => `৳${row.original.rate.toFixed(2)}`,
    },
    {
      accessorKey: "amount", header: "Amount",
      cell: ({ row }) => `৳${row.original.amount.toFixed(2)}`,
    },
    {
      accessorKey: "status", header: "Status",
      cell: ({ row }) => statusBadge(row.original.status),
    },
  ]

  const fetchData = React.useCallback(async (f: Record<string, string>, p?: number, l?: number) => {
    setLoading(true)
    try {
      const params: Record<string, string> = { page: String(p ?? page), limit: String(l ?? limit) }
      ;["company_id", "employee_id", "bill_type", "date_from", "date_to"].forEach((k) => {
        if (f[k]) params[k] = f[k]
      })
      const { data: res } = await nightBillApi.list(params)
      setData(Array.isArray(res.data) ? res.data : [])
      setTotal(res.total ?? 0)
      setTotalPages(res.total_pages ?? 0)
    } catch {
      setData([])
      toast.error("Failed to load night bills")
    } finally {
      setLoading(false)
    }
  }, [page, limit])

  React.useEffect(() => {
    companyApi.list({ limit: "100" }).then((res) => {
      const list = res?.data?.data || res?.data || []
      setCompanies(Array.isArray(list) ? list : [])
    }).catch(() => { })
    fetchData(filters)
  }, [])

  React.useEffect(() => { fetchData(filters) }, [page, limit])

  const handleDelete = async (row: NightBill) => {
    try {
      await nightBillApi.delete(row.id)
      toast.success("Night bill deleted")
      fetchData(filters, page)
    } catch { toast.error("Failed to delete night bill") }
  }

  const handleChange = (key: string, value: string) => {
    setFilters((prev) => ({ ...prev, [key]: value }))
    if (key === "company_id" && value) {
      fetchDepartments(value)
      fetchOrgData(value)
    }
  }

  const handleApply = () => {
    setPage(1)
    const active: Record<string, string> = {}
    for (const [k, v] of Object.entries(filters)) {
      if (v) active[k] = v
    }
    fetchData(active, 1)
  }

  const handleReset = () => {
    setPage(1)
    setLimit(20)
    setFilters({})
    fetchData({}, 1, 20)
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <MoonIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Night Bill</h1>
            <p className="text-muted-foreground mt-1">Manage employee night shift bills</p>
          </div>
        </div>
        <div className="hidden md:flex gap-2">
          <Button onClick={() => router.push("/payroll/night-bill/create")}>
            <PlusIcon className="mr-2 h-4 w-4" />
            Add Night Bill
          </Button>
        </div>
      </div>

      <div className="md:hidden px-4 lg:px-6">
        <ButtonGroup className="w-full">
          <Sheet open={mobileFilterOpen} onOpenChange={setMobileFilterOpen}>
            <SheetTrigger asChild>
              <Button variant="outline" className="flex-1">
                <FilterIcon className="mr-2 h-4 w-4" />
                Filters
              </Button>
            </SheetTrigger>
            <SheetContent side="right" className="w-full sm:max-w-md p-0 flex flex-col" showCloseButton={false}>
              <SheetHeader className="px-4 py-3 border-b flex flex-row items-center justify-between">
                <SheetTitle className="text-base">Filters</SheetTitle>
                <SheetClose asChild>
                  <Button variant="ghost" size="icon-sm">
                    <XIcon className="h-4 w-4" />
                  </Button>
                </SheetClose>
              </SheetHeader>
              <div className="flex-1 overflow-y-auto px-4 py-4">
                <FilterBar filters={filterDefs} values={filters} onChange={handleChange} onApply={() => { handleApply(); setMobileFilterOpen(false) }} onReset={() => { handleReset(); setMobileFilterOpen(false) }} submitting={loading} singleColumn noBorder />
              </div>
            </SheetContent>
          </Sheet>
          <Button onClick={() => router.push("/payroll/night-bill/create")} className="flex-1">
            <PlusIcon className="mr-2 h-4 w-4" />
            Add
          </Button>
        </ButtonGroup>
      </div>

      <div className="px-4 lg:px-6 hidden md:block">
        <FilterBar filters={filterDefs} values={filters} onChange={handleChange} onApply={handleApply} onReset={handleReset} submitting={loading} />
      </div>

      <DataTable
        data={data}
        columns={summaryColumns}
        onDelete={handleDelete}
        serverSide={true}
        page={page}
        pageSize={limit}
        pageCount={totalPages}
        total={total}
        onPageChange={setPage}
        onPageSizeChange={(size) => { setLimit(size); setPage(1) }}
        loading={loading}
      />
    </div>
  )
}
