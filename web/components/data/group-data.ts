"use client"

import { z } from "zod"

export interface Group {
  id: string
  name: string
}

export const groupSchema = z.object({
  name: z.string().min(1, "Name is required"),
})

export type GroupFormData = z.infer<typeof groupSchema>
