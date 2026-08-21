import { z } from "zod";
import type { Department } from "./organization-data";
export interface Separation {
    id: string;
    employee: string;
    employee_id: string;
    department_id: string;
    type: "Resign" | "Lefty" | "Close";
    date: string;
    status: "Approved" | "Pending" | "Rejected" | "Processed" | "Cancelled";
    reason: string;
    created_at: string;
    updated_at: string;
    department?: Department;
}
export declare const separationSchema: z.ZodObject<{
    employee: z.ZodString;
    employee_id: z.ZodString;
    department_id: z.ZodString;
    type: z.ZodEnum<{
        Resign: "Resign";
        Lefty: "Lefty";
        Close: "Close";
    }>;
    date: z.ZodString;
    status: z.ZodEnum<{
        Approved: "Approved";
        Pending: "Pending";
        Rejected: "Rejected";
        Processed: "Processed";
        Cancelled: "Cancelled";
    }>;
    reason: z.ZodOptional<z.ZodString>;
}, z.core.$strip>;
export type SeparationFormData = z.infer<typeof separationSchema>;
export declare const separationTypeOptions: ({
    value: "Resign";
    label: string;
} | {
    value: "Lefty";
    label: string;
} | {
    value: "Close";
    label: string;
})[];
export declare const separationStatusOptions: ({
    value: "Approved";
    label: string;
} | {
    value: "Pending";
    label: string;
} | {
    value: "Rejected";
    label: string;
} | {
    value: "Processed";
    label: string;
} | {
    value: "Cancelled";
    label: string;
})[];
