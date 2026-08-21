"use client"

import { z } from "zod"

export interface Floor {
  id: string
  name: string
}

export const floorSchema = z.object({
  name: z.string().min(1, "Name is required"),
})

export type FloorFormData = z.infer<typeof floorSchema>
