"use client"

import Link from "next/link"
import { LeafIcon, BookOpenIcon, RefreshCwIcon, WalletIcon, FileTextIcon, FileSpreadsheetIcon, ArrowRightIcon } from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Button } from "@/components/ui/button"

const cards = [
  { title: "EL Policy", desc: "Configure accrual rate, max balance, encashment basis per company", href: "/earned-leave/policy", icon: BookOpenIcon, color: "bg-emerald-500" },
  { title: "EL Accrual", desc: "Bulk accrual by period (YYYY-MM) — idempotent, batch optimized", href: "/earned-leave/accrual", icon: RefreshCwIcon, color: "bg-blue-500" },
  { title: "EL Balance", desc: "Real-time closing balance per employee (ledger-driven)", href: "/earned-leave/balance", icon: WalletIcon, color: "bg-indigo-500" },
  { title: "EL Ledger", desc: "Audit trail: ACCRUAL, LEAVE_USED, ADJUSTMENT, ENCASHMENT, REVERSAL", href: "/earned-leave/ledger", icon: FileTextIcon, color: "bg-amber-500" },
  { title: "EL Salary Sheet", desc: "Generate, approve, finalize, reverse — bulk, transactional", href: "/earned-leave/salary-sheet", icon: FileSpreadsheetIcon, color: "bg-rose-500" },
]

export default function EarnedLeaveDashboard() {
  return (
    <div className="flex flex-col gap-6 py-6">
      <div className="px-4 lg:px-6">
        <div className="rounded-xl bg-gradient-to-br from-emerald-600 via-teal-600 to-cyan-600 p-6 text-white">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-white/15">
              <LeafIcon className="h-5 w-5" />
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight md:text-3xl">Earned Leave</h1>
              <p className="text-sm text-white/80">Policy → Accrual → Balance → Ledger → Encashment → Salary Sheet</p>
            </div>
          </div>
          <p className="mt-3 max-w-3xl text-sm text-white/85">
            Date-based, company-isolated, ledger-driven. Accrual is idempotent (company+employee+period+type unique), salary sheet is batch-optimized (bulk load → maps → in-memory calc → CreateInBatches), and all operations are transactional with concurrency locks.
          </p>
        </div>
      </div>

      <div className="px-4 lg:px-6 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {cards.map((c) => (
          <Card key={c.href} className="hover:shadow-md transition-shadow">
            <CardHeader className="pb-3">
              <div className={`flex h-9 w-9 items-center justify-center rounded-lg ${c.color} text-white`}>
                <c.icon className="h-4 w-4" />
              </div>
              <CardTitle className="mt-3 text-base">{c.title}</CardTitle>
              <CardDescription className="text-xs leading-relaxed">{c.desc}</CardDescription>
            </CardHeader>
            <CardContent>
              <Link href={c.href}>
                <Button variant="outline" size="sm" className="w-full">
                  Open <ArrowRightIcon className="ml-2 h-3.5 w-3.5" />
                </Button>
              </Link>
            </CardContent>
          </Card>
        ))}
      </div>

      <div className="px-4 lg:px-6">
        <Card>
          <CardHeader><CardTitle className="text-base">Quick Flow</CardTitle></CardHeader>
          <CardContent className="text-sm text-muted-foreground leading-relaxed">
            <ol className="list-decimal pl-5 space-y-1">
              <li>Create <b>EL Policy</b> per company (accrual rate, max 40/60, carry, encashment basis).</li>
              <li>Run <b>Accrual Process</b> for period <code className="bg-muted px-1 rounded">YYYY-MM</code> — bulk loads employees + separations + existing accrual, builds maps, calculates in memory, batch inserts.</li>
              <li>Check <b>Balance</b> (opening + accrued - used - encashed) and <b>Ledger</b> audit.</li>
              <li>Adjust / Encash via dedicated APIs — each creates a ledger transaction.</li>
              <li>Generate <b>EL Salary Sheet</b> — bulk calc <code>days × (basis/divisor)</code>, stored immutably, then Approve → Finalize (or Reverse).</li>
            </ol>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
