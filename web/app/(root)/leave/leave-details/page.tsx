"use client"

import * as React from "react"
import { FileTextIcon, FilterIcon, XIcon } from "lucide-react"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { leaveBalanceApi, employeeApi } from "@/lib/api"
import { FilterBar } from "@/components/filter-bar"
import type { FilterDef } from "@/components/filter-bar"
import { Button } from "@/components/ui/button"
import { ButtonGroup } from "@/components/ui/button-group"
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger, SheetClose } from "@/components/ui/sheet"

interface BalanceRecord {
  id: string
  employee_id: string
  leave_type: string
  year: number
  total: number
  used: number
  pending: number
  remaining: number
}

interface Employee { id: string; employee_id: string; name_en: string }

const thisYear = new Date().getFullYear()

const columns: ColumnDef<BalanceRecord>[] = [
  { id: "sl", header: "Sl", cell: ({ row }) => row.index + 1 },
  { accessorKey: "leave_type", header: "Leave Type" },
  { accessorKey: "total", header: "Total" },
  { accessorKey: "used", header: "Used" },
  { accessorKey: "pending", header: "Pending" },
  { accessorKey: "remaining", header: "Remaining" },
]

export default function LeaveDetailsPage() {
  const [data, setData] = React.useState<BalanceRecord[]>([])
  const [loading, setLoading] = React.useState(true)
  const [employees, setEmployees] = React.useState<Employee[]>([])
  const [filters, setFilters] = React.useState<Record<string, string>>({})
  const [mobileFilterOpen, setMobileFilterOpen] = React.useState(false)

  const [page, setPage] = React.useState(1)
  const [limit, setLimit] = React.useState(20)
  const [total, setTotal] = React.useState(0)
  const [totalPages, setTotalPages] = React.useState(0)

  const filterDefs: FilterDef[] = React.useMemo(() => [
    {
      key: "employee_id", label: "Employee", type: "select",
      options: employees.map((e) => ({ value: e.id, label: `${e.employee_id} - ${e.name_en}` })),
    },
  ], [employees])

  const fetchData = React.useCallback(async (params: Record<string, string>, p?: number, l?: number) => {
    setLoading(true)
    try {
      const active: Record<string, string> = { year: String(thisYear) }
      if (params.employee_id) active.employee_id = params.employee_id
      active.page = String(p ?? page)
      active.limit = String(l ?? limit)
      const { data: res } = await leaveBalanceApi.list(active)
      const list = Array.isArray(res.data) ? res.data : []
      setData(list.map((r: any, i: number) => ({ ...r, id: r.id || `${r.employee_id}-${r.leave_type}-${i}` })))
      setTotal(res.total ?? 0)
      setTotalPages(res.total_pages ?? 0)
    } catch {
      setData([])
    } finally {
      setLoading(false)
    }
  }, [page, limit])

  React.useEffect(() => {
    employeeApi.list({ limit: "100" }).then((res) => {
      const list = res.data?.data?.employees || res.data?.data || []
      setEmployees(Array.isArray(list) ? list : [])
    }).catch(() => {})
    fetchData({})
  }, [])

  React.useEffect(() => {
    fetchData(filters)
  }, [page, limit])

  const handleChange = (key: string, value: string) => setFilters((prev) => ({ ...prev, [key]: value }))

  const handleApply = () => {
    setPage(1)
    const active: Record<string, string> = {}
    for (const [k, v] of Object.entries(filters)) if (v) active[k] = v
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
      <div className="px-4 lg:px-6">
        <div className="flex items-center gap-2">
          <FileTextIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Leave Details</h1>
            <p className="text-muted-foreground mt-1">Employee leave balance details</p>
          </div>
        </div>
        <div className="md:hidden mt-3">
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
          </ButtonGroup>
        </div>
      </div>
      <div className="px-4 lg:px-6 hidden md:block">
        <FilterBar filters={filterDefs} values={filters} onChange={handleChange} onApply={handleApply} onReset={handleReset} submitting={loading} />
      </div>
      <DataTable
        data={data}
        columns={columns}
        serverSide={true}
        page={page}
        pageSize={limit}
        pageCount={totalPages}
        total={total}
        onPageChange={setPage}
        onPageSizeChange={(size) => { setLimit(size); setPage(1); }}
        loading={loading}
      />
    </div>
  )
}
