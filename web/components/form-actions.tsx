"use client"

import { Loader2 } from "lucide-react"
import { Button } from "@/components/ui/button"

interface FormActionsProps {
  isSubmitting?: boolean
  submitLabel?: string
  editLabel?: string
  isEditing?: boolean
  onCancel?: () => void
}

export function FormActions({ isSubmitting, submitLabel = "Create", editLabel = "Update", isEditing, onCancel }: FormActionsProps) {
  return (
    <div className="flex justify-end gap-4 pt-4 border-t">
      {onCancel && (
        <Button type="button" variant="outline" onClick={onCancel} disabled={isSubmitting}>
          Cancel
        </Button>
      )}
      <Button type="submit" disabled={isSubmitting}>
        {isSubmitting ? <><Loader2 className="mr-2 h-4 w-4 animate-spin" /> Saving...</> : isEditing ? editLabel : submitLabel}
      </Button>
    </div>
  )
}
