import { z } from "zod";
import type { Department, Designation } from "./organization-data";
export interface IdCard {
    id: string;
    employee: string;
    employee_id: string;
    designation_id: string;
    department_id: string;
    card_no: string;
    issued: string;
    expiry: string;
    status: "Active" | "Expired" | "Lost" | "Damaged";
    created_at: string;
    updated_at: string;
    department?: Department;
    designation?: Designation;
}
export declare const idCardSchema: z.ZodObject<{
    employee: z.ZodString;
    employee_id: z.ZodString;
    designation_id: z.ZodString;
    department_id: z.ZodString;
    card_no: z.ZodString;
    issued: z.ZodString;
    expiry: z.ZodString;
    status: z.ZodEnum<{
        Active: "Active";
        Expired: "Expired";
        Lost: "Lost";
        Damaged: "Damaged";
    }>;
}, z.core.$strip>;
export type IdCardFormData = z.infer<typeof idCardSchema>;
export declare const idCardStatusOptions: ({
    value: "Active";
    label: string;
} | {
    value: "Expired";
    label: string;
} | {
    value: "Lost";
    label: string;
} | {
    value: "Damaged";
    label: string;
})[];
