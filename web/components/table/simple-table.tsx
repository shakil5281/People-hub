"use client"

import * as React from "react"
import { Search, ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"

interface Column<T> {
  header: string
  accessor: keyof T | ((item: T) => string | number)
  className?: string
  cell?: (item: T) => React.ReactNode
}

interface SimpleTableProps<T> {
  title: string
  description?: string
  data: T[]
  columns: Column<T>[]
  searchKey?: keyof T
  pageSize?: number
}

export function SimpleTable<T extends Record<string, unknown>>({
  title,
  description,
  data,
  columns,
  searchKey,
  pageSize = 10,
}: SimpleTableProps<T>) {
  const [search, setSearch] = React.useState("")
  const [currentPage, setCurrentPage] = React.useState(1)
  const [sortConfig, setSortConfig] = React.useState<{ key: string; direction: "asc" | "desc" } | null>(null)

  const filtered = React.useMemo(() => {
    let result = data
    if (search && searchKey) {
      const lower = search.toLowerCase()
      result = data.filter((item) => {
        const val = item[searchKey]
        return val != null && String(val).toLowerCase().includes(lower)
      })
    }
    if (sortConfig) {
      result = [...result].sort((a, b) => {
        const aVal = a[sortConfig.key]
        const bVal = b[sortConfig.key]
        if (aVal == null) return 1
        if (bVal == null) return -1
        if (aVal < bVal) return sortConfig.direction === "asc" ? -1 : 1
        if (aVal > bVal) return sortConfig.direction === "asc" ? 1 : -1
        return 0
      })
    }
    return result
  }, [data, search, searchKey, sortConfig])

  const totalPages = Math.ceil(filtered.length / pageSize)
  const paginated = filtered.slice((currentPage - 1) * pageSize, currentPage * pageSize)

  const handleSort = (key: string) => {
    setSortConfig((prev) => ({
      key,
      direction: prev?.key === key && prev.direction === "asc" ? "desc" : "asc",
    }))
  }

  const getCellValue = (item: T, column: Column<T>): string | number => {
    if (typeof column.accessor === "function") {
      return column.accessor(item)
    }
    const val = item[column.accessor]
    return val != null ? String(val) : ""
  }

  return (
    <Card>
      <CardHeader className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-3">
        <div>
          <CardTitle>{title}</CardTitle>
          {description && <p className="text-sm text-muted-foreground mt-1">{description}</p>}
        </div>
        {searchKey && (
          <div className="relative max-w-xs w-full sm:w-auto">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search..."
              value={search}
              onChange={(e) => {
                setSearch(e.target.value)
                setCurrentPage(1)
              }}
              className="pl-10"
            />
          </div>
        )}
      </CardHeader>
      <CardContent className="p-0">
        <div className="overflow-x-auto">
          <Table>
            <TableHeader className="bg-muted/50">
              <TableRow>
                {columns.map((col) => (
                  <TableHead
                    key={col.header}
                    className={`cursor-pointer select-none ${col.className || ""}`}
                    onClick={() => handleSort(typeof col.accessor === "function" ? col.header : (col.accessor as string))}
                  >
                    <div className="flex items-center gap-1">
                      {col.header}
                      {sortConfig?.key === (typeof col.accessor === "function" ? col.header : col.accessor) && (
                        <span className="text-xs">{sortConfig.direction === "asc" ? "▲" : "▼"}</span>
                      )}
                    </div>
                  </TableHead>
                ))}
              </TableRow>
            </TableHeader>
            <TableBody>
              {paginated.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={columns.length} className="text-center py-12 text-muted-foreground">
                    No data found.
                  </TableCell>
                </TableRow>
              ) : (
                paginated.map((item, idx) => (
                  <TableRow key={idx} className="hover:bg-muted/50 transition-colors">
                    {columns.map((col) => (
                      <TableCell key={col.header} className={col.className || ""}>
                        {col.cell ? col.cell(item) : getCellValue(item, col)}
                      </TableCell>
                    ))}
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </div>

        {totalPages > 1 && (
          <div className="flex items-center justify-between px-4 py-3 border-t">
            <p className="text-sm text-muted-foreground">
              Showing {(currentPage - 1) * pageSize + 1} to {Math.min(currentPage * pageSize, filtered.length)} of {filtered.length}
            </p>
            <div className="flex items-center gap-2">
              <Button variant="outline" size="icon" onClick={() => setCurrentPage(1)} disabled={currentPage === 1}>
                <ChevronsLeft className="h-4 w-4" />
              </Button>
              <Button variant="outline" size="icon" onClick={() => setCurrentPage((p) => Math.max(1, p - 1))} disabled={currentPage === 1}>
                <ChevronLeft className="h-4 w-4" />
              </Button>
              <span className="text-sm font-medium w-20 text-center">Page {currentPage} of {totalPages}</span>
              <Button variant="outline" size="icon" onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))} disabled={currentPage === totalPages}>
                <ChevronRight className="h-4 w-4" />
              </Button>
              <Button variant="outline" size="icon" onClick={() => setCurrentPage(totalPages)} disabled={currentPage === totalPages}>
                <ChevronsRight className="h-4 w-4" />
              </Button>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
