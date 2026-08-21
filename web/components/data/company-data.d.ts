import { z } from "zod";
export interface Company {
    id: string;
    company_name_en: string;
    company_name_bn: string;
    slug: string;
    address: string;
    phone: string;
    status: "active" | "inactive";
    created_at: string;
    updated_at: string;
}
export declare const companySchema: z.ZodObject<{
    company_name_en: z.ZodString;
    company_name_bn: z.ZodDefault<z.ZodString>;
    address: z.ZodString;
    phone: z.ZodString;
    status: z.ZodDefault<z.ZodEnum<{
        active: "active";
        inactive: "inactive";
    }>>;
}, z.core.$strip>;
export type CompanyFormData = z.infer<typeof companySchema>;
export declare const statusOptions: {
    value: "active" | "inactive";
    label: string;
}[];
