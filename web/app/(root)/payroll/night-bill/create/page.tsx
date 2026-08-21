"use client"

import * as React from "react"
import { MoonIcon, ArrowLeftIcon, Loader2, PlusIcon, Trash2Icon, SaveIcon, SearchIcon, UserIcon } from "lucide-react"
import { useRouter } from "next/navigation"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from "@/components/ui/select"
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger,
} from "@/components/ui/dialog"
import { Card, CardContent } from "@/components/ui/card"
import { nightBillApi, employeeApi } from "@/lib/api"

interface EmployeeInfo {
  employee_id: string
  name_en: string
  name_bn: string
  punch_number?: string
  designation_ref?: { name: string; name_bn: string }
  department?: { name: string; name_bn: string }
}

interface TableRow {
  employee_id: string
  name_en: string
  name_bn: string
  designation: string
  date: string
  bill_type: string
  night_hours: number
  rate: number
  amount: number
}

export default function CreateNightBillPage() {
  const router = useRouter()
  const today = new Date().toISOString().slice(0, 10)

  const [rows, setRows] = React.useState<TableRow[]>([])
  const [saving, setSaving] = React.useState(false)
  const [dialogOpen, setDialogOpen] = React.useState(false)

  const [empIdInput, setEmpIdInput] = React.useState("")
  const [searching, setSearching] = React.useState(false)
  const [foundEmp, setFoundEmp] = React.useState<EmployeeInfo | null>(null)
  const [searchError, setSearchError] = React.useState("")

  const [formDate, setFormDate] = React.useState(today)
  const [formBillType, setFormBillType] = React.useState("fixed")
  const [formHours, setFormHours] = React.useState("0")
  const [formRate, setFormRate] = React.useState("80")
  const [formAmount, setFormAmount] = React.useState("80")

  const handleSearch = async () => {
    const q = empIdInput.trim()
    if (!q) { setSearchError("Enter an Employee ID"); return }
    setSearching(true)
    setSearchError("")
    setFoundEmp(null)
    try {
      const { data: res } = await employeeApi.getByCode(q)
      const emp = res?.data || res
      if (!emp || !emp.employee_id) {
        setSearchError("Employee not found")
        return
      }
      setFoundEmp(emp)
    } catch {
      setSearchError("Employee not found")
    } finally {
      setSearching(false)
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") handleSearch()
  }

  React.useEffect(() => {
    if (formBillType === "fixed") {
      setFormAmount(formRate || "0")
      setFormHours("0")
    } else {
      const h = parseFloat(formHours) || 0
      const r = parseFloat(formRate) || 0
      setFormAmount((h * r).toFixed(2))
    }
  }, [formBillType, formHours, formRate])

  const handleAdd = () => {
    if (!foundEmp) { toast.error("Search and select an employee first"); return }
    const empId = foundEmp.employee_id
    const existing = rows.some((r) => r.employee_id === empId && r.date === formDate)
    if (existing) { toast.error("Employee already added for this date"); return }

    const nh = formBillType === "fixed" ? 0 : (parseFloat(formHours) || 0)
    const rt = parseFloat(formRate) || 0
    const amt = parseFloat(formAmount) || 0

    setRows((prev) => [...prev, {
      employee_id: empId,
      name_en: foundEmp.name_en,
      name_bn: foundEmp.name_bn || "",
      designation: foundEmp.designation_ref?.name || "",
      date: formDate,
      bill_type: formBillType,
      night_hours: nh,
      rate: rt,
      amount: amt,
    }])

    setDialogOpen(false)
    setFoundEmp(null)
    setEmpIdInput("")
    setFormBillType("fixed")
    setFormHours("0")
    setFormRate("80")
    setFormAmount("80")
    toast.success("Employee added")
  }

  const removeRow = (idx: number) => {
    setRows((prev) => prev.filter((_, i) => i !== idx))
  }

  const handleSave = async () => {
    if (rows.length === 0) { toast.error("No entries to save"); return }
    setSaving(true)
    try {
      await nightBillApi.bulkCreate({
        company_id: "",
        items: rows.map((r) => ({
          employee_id: r.employee_id,
          date: r.date,
          bill_type: r.bill_type,
          night_hours: r.night_hours,
          rate: r.rate,
          amount: r.amount,
        })),
      })
      toast.success(`${rows.length} night bill(s) saved`)
      router.push("/payroll/night-bill")
    } catch {
      toast.error("Failed to save night bills")
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <MoonIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <div className="flex items-center gap-3">
              <Button variant="ghost" size="icon" onClick={() => router.back()} className="h-8 w-8">
                <ArrowLeftIcon className="h-4 w-4" />
              </Button>
              <h1 className="text-3xl font-bold tracking-tight">Add Night Bill</h1>
            </div>
            <p className="text-muted-foreground mt-1 ml-11">Add employees for night bill calculation</p>
          </div>
        </div>
        {rows.length > 0 && (
          <Button onClick={handleSave} disabled={saving}>
            {saving ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <SaveIcon className="mr-2 h-4 w-4" />}
            Save
          </Button>
        )}
      </div>

      <div className="px-4 lg:px-6">
        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold">Employee List ({rows.length})</h2>
              <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
                <DialogTrigger asChild>
                  <Button>
                    <PlusIcon className="mr-2 h-4 w-4" />
                    Add Employee
                  </Button>
                </DialogTrigger>
                <DialogContent className="sm:max-w-lg">
                  <DialogHeader>
                    <DialogTitle>Add Employee for Night Bill</DialogTitle>
                  </DialogHeader>

                  <div className="space-y-5 py-2">
                    <div className="flex items-end gap-2">
                      <div className="flex-1 space-y-1.5">
                        <Label>Employee ID</Label>
                        <div className="relative">
                          <SearchIcon className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                          <Input
                            value={empIdInput}
                            onChange={(e) => { setEmpIdInput(e.target.value); setSearchError(""); setFoundEmp(null) }}
                            onKeyDown={handleKeyDown}
                            placeholder="Enter Employee ID..."
                            className="pl-8"
                          />
                        </div>
                      </div>
                      <Button onClick={handleSearch} disabled={searching} className="shrink-0">
                        {searching ? <Loader2 className="h-4 w-4 animate-spin" /> : <SearchIcon className="h-4 w-4" />}
                        Search
                      </Button>
                    </div>
                    {searchError && <p className="text-sm text-red-500">{searchError}</p>}

                    {foundEmp && (
                      <div className="rounded-lg border bg-muted/30 p-4 space-y-3">
                        <div className="flex items-center gap-3">
                          <div className="flex h-10 w-10 items-center justify-center rounded-full bg-primary/10">
                            <UserIcon className="h-5 w-5 text-primary" />
                          </div>
                          <div>
                            <p className="font-medium">{foundEmp.name_en}</p>
                            {foundEmp.name_bn && <p className="text-xs text-muted-foreground">{foundEmp.name_bn}</p>}
                          </div>
                        </div>
                        <div className="grid grid-cols-2 gap-2 text-sm">
                          <div>
                            <span className="text-muted-foreground">ID:</span> {foundEmp.employee_id}
                          </div>
                          <div>
                            <span className="text-muted-foreground">Designation:</span> {foundEmp.designation_ref?.name || "-"}
                          </div>
                        </div>
                      </div>
                    )}

                    {foundEmp && (
                      <>
                        <div className="space-y-1.5">
                          <Label>Date</Label>
                          <Input type="date" value={formDate} onChange={(e) => setFormDate(e.target.value)} />
                        </div>

                        <div className="space-y-1.5">
                          <Label>Bill Type</Label>
                          <Select value={formBillType} onValueChange={setFormBillType}>
                            <SelectTrigger>
                              <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectItem value="fixed">Fixed</SelectItem>
                              <SelectItem value="hourly">Hourly Rate</SelectItem>
                            </SelectContent>
                          </Select>
                        </div>

                        {formBillType === "hourly" && (
                          <div className="grid grid-cols-2 gap-4">
                            <div className="space-y-1.5">
                              <Label>Night Hours</Label>
                              <Input
                                type="number" min="0" step="0.5"
                                value={formHours} onChange={(e) => setFormHours(e.target.value)}
                              />
                            </div>
                            <div className="space-y-1.5">
                              <Label>Rate per Hour (৳)</Label>
                              <Input
                                type="number" min="0" step="0.01"
                                value={formRate} onChange={(e) => setFormRate(e.target.value)}
                              />
                            </div>
                          </div>
                        )}

                        <div className="space-y-1.5">
                          <Label>{formBillType === "fixed" ? "Fixed Amount (৳)" : "Total Amount (৳)"}</Label>
                          <Input
                            type="number" min="0" step="0.01"
                            value={formAmount}
                            onChange={(e) => setFormAmount(e.target.value)}
                          />
                        </div>

                        <Button className="w-full" onClick={handleAdd}>
                          <PlusIcon className="mr-2 h-4 w-4" />
                          Add to List
                        </Button>
                      </>
                    )}
                  </div>
                </DialogContent>
              </Dialog>
            </div>

            <div className="overflow-x-auto border rounded-lg">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b bg-muted/50">
                    <th className="p-3 text-left font-medium">Employee ID</th>
                    <th className="p-3 text-left font-medium">Name</th>
                    <th className="p-3 text-left font-medium">Designation</th>
                    <th className="p-3 text-center font-medium">Date</th>
                    <th className="p-3 text-center font-medium">Type</th>
                    <th className="p-3 text-right font-medium">Hours</th>
                    <th className="p-3 text-right font-medium">Rate</th>
                    <th className="p-3 text-right font-medium">Amount</th>
                    <th className="p-3 text-center w-10"></th>
                  </tr>
                </thead>
                <tbody>
                  {rows.length === 0 ? (
                    <tr>
                      <td colSpan={9} className="p-8 text-center text-muted-foreground">
                        No employees added yet. Click "Add Employee" to add.
                      </td>
                    </tr>
                  ) : (
                    rows.map((row, idx) => (
                      <tr key={`${row.employee_id}-${row.date}-${idx}`} className="border-b hover:bg-muted/50">
                        <td className="p-3 font-mono text-xs">{row.employee_id}</td>
                        <td className="p-3 font-medium">{row.name_en}</td>
                        <td className="p-3 text-muted-foreground">{row.designation || "-"}</td>
                        <td className="p-3 text-center font-mono text-xs">{row.date}</td>
                        <td className="p-3 text-center capitalize">{row.bill_type}</td>
                        <td className="p-3 text-right font-mono">{row.night_hours > 0 ? row.night_hours.toFixed(2) : "—"}</td>
                        <td className="p-3 text-right font-mono">৳{row.rate.toFixed(2)}</td>
                        <td className="p-3 text-right font-mono font-medium">৳{row.amount.toFixed(2)}</td>
                        <td className="p-3 text-center">
                          <Button variant="ghost" size="icon" className="h-8 w-8 text-red-500" onClick={() => removeRow(idx)}>
                            <Trash2Icon className="h-4 w-4" />
                          </Button>
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
                {rows.length > 0 && (
                  <tfoot>
                    <tr className="border-t bg-muted/30 font-medium">
                      <td className="p-3" colSpan={5}>Total: {rows.length}</td>
                      <td className="p-3 text-right font-mono">
                        {rows.reduce((s, r) => s + r.night_hours, 0).toFixed(2)}
                      </td>
                      <td className="p-3 text-right">—</td>
                      <td className="p-3 text-right font-mono">
                        ৳{rows.reduce((s, r) => s + r.amount, 0).toFixed(2)}
                      </td>
                      <td></td>
                    </tr>
                  </tfoot>
                )}
              </table>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
