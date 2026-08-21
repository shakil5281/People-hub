"use client"

import * as React from "react"
import {
  CircleHelpIcon,
  MessageSquareTextIcon,
  MailIcon,
  PhoneIcon,
  BookOpenIcon,
  FileTextIcon,
  LifeBuoyIcon,
  ExternalLinkIcon,
  SearchIcon,
  ChevronDownIcon,
} from "lucide-react"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"

const faqs = [
  {
    q: "How do I mark daily attendance?",
    a: "Navigate to Attendance → Daily Attendance from the sidebar. Select the date and employee, then enter check-in and check-out times.",
  },
  {
    q: "How is overtime calculated?",
    a: "OT rate is based on basic salary divided by total days divided by 8 hours. Overtime is calculated from attendance punch data automatically during salary processing.",
  },
  {
    q: "How do I apply for leave?",
    a: "Go to Leave → Leave from the sidebar, click Create Leave, fill in the leave type, dates, and reason. Your manager will be notified for approval.",
  },
  {
    q: "How do I generate salary sheets?",
    a: "Navigate to Payroll → Salary Sheet, select the month/year and company, then click Search. Use the Export button to download as Excel.",
  },
  {
    q: "How do I add a new employee?",
    a: "Go to Human Resource → Employees, click Create Employee. Fill in personal info, office details, and salary information.",
  },
  {
    q: "How do I reset my password?",
    a: "Click your profile icon in the bottom-left, select Settings, then navigate to the Password section to change your password.",
  },
]

const guides = [
  { title: "Getting Started Guide", icon: BookOpenIcon, description: "Learn the basics of PeopleHub" },
  { title: "Attendance Management", icon: FileTextIcon, description: "Daily, manual, and monthly attendance" },
  { title: "Payroll Processing", icon: FileTextIcon, description: "Salary sheets, increments, and payslips" },
  { title: "Employee Management", icon: FileTextIcon, description: "Onboarding, updates, and separations" },
]

export default function HelpPage() {
  const [search, setSearch] = React.useState("")
  const [openIndex, setOpenIndex] = React.useState<number | null>(null)

  const filtered = faqs.filter(
    (f) =>
      f.q.toLowerCase().includes(search.toLowerCase()) ||
      f.a.toLowerCase().includes(search.toLowerCase()),
  )

  return (
    <div className="flex flex-col gap-6 py-4 md:gap-8 md:py-6">
      {/* Hero */}
      <div className="px-4 lg:px-6">
        <div className="relative overflow-hidden rounded-2xl bg-gradient-to-br from-primary/10 via-primary/5 to-background border p-6 md:p-10">
          <div className="absolute top-0 right-0 w-72 h-72 bg-primary/5 rounded-full -translate-y-1/2 translate-x-1/2 blur-3xl" />
          <div className="absolute bottom-0 left-0 w-48 h-48 bg-primary/5 rounded-full translate-y-1/2 -translate-x-1/2 blur-3xl" />
          <div className="relative flex flex-col items-center text-center max-w-2xl mx-auto">
            <div className="p-3 rounded-full bg-primary/10 mb-4">
              <CircleHelpIcon className="h-8 w-8 text-primary" />
            </div>
            <h1 className="text-3xl md:text-4xl font-bold tracking-tight">How can we help you?</h1>
            <p className="text-muted-foreground mt-2 max-w-md">
              Search our help center or browse the topics below.
            </p>
            <div className="relative w-full max-w-lg mt-6">
              <SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder="Search for answers..."
                className="pl-9 h-11 bg-background"
              />
            </div>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4 md:gap-6 px-4 lg:px-6">
        {/* FAQ Section */}
        <div className="lg:col-span-2 space-y-4">
          <div className="flex items-center gap-2">
            <MessageSquareTextIcon className="h-5 w-5 text-muted-foreground" />
            <h2 className="text-lg font-semibold">Frequently Asked Questions</h2>
          </div>

          <div className="space-y-2">
            {filtered.length === 0 ? (
              <Card>
                <CardContent className="py-8 text-center text-muted-foreground text-sm">
                  No results found for &quot;{search}&quot;
                </CardContent>
              </Card>
            ) : (
              filtered.map((faq, i) => (
                <Card
                  key={i}
                  className="cursor-pointer transition-all duration-200 hover:shadow-sm"
                  onClick={() => setOpenIndex(openIndex === i ? null : i)}
                >
                  <CardContent className="p-4">
                    <div className="flex items-center justify-between gap-4">
                      <div className="flex items-center gap-3">
                        <span className="flex items-center justify-center h-7 w-7 rounded-full bg-primary/10 text-primary text-xs font-bold shrink-0">
                          {i + 1}
                        </span>
                        <span className="font-medium text-sm">{faq.q}</span>
                      </div>
                      <ChevronDownIcon
                        className={`h-4 w-4 text-muted-foreground shrink-0 transition-transform duration-200 ${
                          openIndex === i ? "rotate-180" : ""
                        }`}
                      />
                    </div>
                    {openIndex === i && (
                      <div className="mt-3 ml-10 text-sm text-muted-foreground border-t pt-3">
                        {faq.a}
                      </div>
                    )}
                  </CardContent>
                </Card>
              ))
            )}
          </div>
        </div>

        {/* Sidebar */}
        <div className="space-y-4 md:space-y-6">
          {/* Guides */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base flex items-center gap-2">
                <BookOpenIcon className="h-4 w-4 text-muted-foreground" />
                Help Guides
              </CardTitle>
              <CardDescription>Step-by-step tutorials</CardDescription>
            </CardHeader>
            <CardContent className="space-y-2">
              {guides.map((guide) => {
                const Icon = guide.icon
                return (
                  <div
                    key={guide.title}
                    className="flex items-center gap-3 p-2.5 rounded-lg hover:bg-muted/50 transition-colors cursor-pointer"
                  >
                    <div className="p-1.5 rounded-md bg-primary/10 text-primary">
                      <Icon className="h-4 w-4" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium">{guide.title}</p>
                      <p className="text-xs text-muted-foreground truncate">{guide.description}</p>
                    </div>
                    <ExternalLinkIcon className="h-3.5 w-3.5 text-muted-foreground shrink-0" />
                  </div>
                )
              })}
            </CardContent>
          </Card>

          {/* Contact Support */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base flex items-center gap-2">
                <LifeBuoyIcon className="h-4 w-4 text-muted-foreground" />
                Contact Support
              </CardTitle>
              <CardDescription>We&apos;re here to help</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="flex items-center gap-3 text-sm">
                <div className="p-1.5 rounded-md bg-blue-100 text-blue-600 dark:bg-blue-950/40">
                  <MailIcon className="h-4 w-4" />
                </div>
                <div>
                  <p className="font-medium">Email</p>
                  <p className="text-muted-foreground text-xs">support@peoplehub.com</p>
                </div>
              </div>
              <div className="flex items-center gap-3 text-sm">
                <div className="p-1.5 rounded-md bg-green-100 text-green-600 dark:bg-green-950/40">
                  <PhoneIcon className="h-4 w-4" />
                </div>
                <div>
                  <p className="font-medium">Phone</p>
                  <p className="text-muted-foreground text-xs">+880 1234-567890</p>
                </div>
              </div>
              <Button className="w-full mt-2" variant="outline">
                <MessageSquareTextIcon className="mr-2 h-4 w-4" />
                Start a Chat
              </Button>
            </CardContent>
          </Card>

          {/* System Status */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base flex items-center gap-2">
                <div className="h-2 w-2 rounded-full bg-emerald-500 animate-pulse" />
                System Status
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex items-center justify-between text-sm">
                <span className="text-muted-foreground">All systems operational</span>
                <Badge variant="outline" className="text-emerald-600 border-emerald-200 bg-emerald-50 dark:bg-emerald-950/30">
                  Healthy
                </Badge>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
      <div className="px-4 lg:px-6 pb-6">
        <div className="rounded-xl border bg-gradient-to-r from-primary/5 to-transparent p-6 text-center">
          <p className="text-sm text-muted-foreground">
            Still need help? We typically respond within 24 hours.
          </p>
        </div>
      </div>
    </div>
  )
}
