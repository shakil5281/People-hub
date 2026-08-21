import { z } from "zod";
export interface Shift {
    id: string;
    company_id: string;
    name: string;
    shift_type: string;
    start_time: string;
    end_time: string;
    late_grace_minutes: number;
    weekend_days: string;
    status: "active" | "inactive";
    created_at: string;
    updated_at: string;
}
export declare const shiftSchema: z.ZodObject<{
    company_id: z.ZodString;
    name: z.ZodString;
    shift_type: z.ZodDefault<z.ZodEnum<{
        day: "day";
        night: "night";
        general: "general";
    }>>;
    start_time: z.ZodString;
    end_time: z.ZodString;
    late_grace_minutes: z.ZodDefault<z.ZodNumber>;
    weekend_days: z.ZodDefault<z.ZodString>;
    status: z.ZodDefault<z.ZodEnum<{
        active: "active";
        inactive: "inactive";
    }>>;
}, z.core.$strip>;
export type ShiftFormData = z.infer<typeof shiftSchema>;
export declare const statusOptions: ({
    value: "active";
    label: string;
} | {
    value: "inactive";
    label: string;
})[];
export declare const dayOptions: {
    value: string;
    label: string;
}[];
