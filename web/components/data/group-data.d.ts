import { z } from "zod";
export interface Group {
    id: string;
    name: string;
}
export declare const groupSchema: z.ZodObject<{
    name: z.ZodString;
}, z.core.$strip>;
export type GroupFormData = z.infer<typeof groupSchema>;
