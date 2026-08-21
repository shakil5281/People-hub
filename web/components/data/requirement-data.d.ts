import { z } from "zod";
import type { Department } from "./organization-data";
export interface Requirement {
    id: string;
    position: string;
    department_id: string;
    vacancies: number;
    applicants: number;
    status: "Open" | "Closed";
    priority: "High" | "Medium" | "Low";
    description: string;
    created_at: string;
    updated_at: string;
    department?: Department;
}
export declare const requirementSchema: z.ZodObject<{
    position: z.ZodString;
    department_id: z.ZodString;
    vacancies: z.ZodNumber;
    applicants: z.ZodNumber;
    status: z.ZodEnum<{
        Open: "Open";
        Closed: "Closed";
    }>;
    priority: z.ZodEnum<{
        High: "High";
        Medium: "Medium";
        Low: "Low";
    }>;
    description: z.ZodOptional<z.ZodString>;
}, z.core.$strip>;
export type RequirementFormData = z.infer<typeof requirementSchema>;
export declare const statusOptions: ({
    value: "Open";
    label: string;
} | {
    value: "Closed";
    label: string;
})[];
export declare const priorityOptions: ({
    value: "High";
    label: string;
} | {
    value: "Medium";
    label: string;
} | {
    value: "Low";
    label: string;
})[];
export declare const positionOptions: {
    value: string;
    label: string;
}[];
